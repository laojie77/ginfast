package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	"gin-fast/plugins/syscustomer/models"
	noticeService "gin-fast/plugins/sysnotice/service"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	// customer_export_worker_count 未配置时的默认并发数。
	// 这里的“worker”可以理解为后台消费导出任务的工人线程/协程数量。
	defaultCustomerExportWorkerCount = 2
	// 调度器兜底轮询数据库的间隔。
	// 正常情况下新任务入队后会主动唤醒调度器；这个定时器只是防止漏掉唤醒信号。
	customerExportDispatchInterval = 1 * time.Minute
	// 每次调度时最多扫描多少条排队任务。
	// 目的是避免单次扫描过大，影响数据库压力和调度响应速度。
	customerExportQueueScanLimit = 20
	// 控制失败原因写入数据库时的最大长度，
	// 避免异常信息过长导致字段超限或把无关堆栈直接塞进表里。
	customerExportFailureMessageMaxLen = 900
	// 判断一个“排队中/执行中”的任务是否已经卡死。
	// 超过这个时间仍没有继续推进，就会被系统回收并标记失败。
	customerExportTaskActiveTimeout = 30 * time.Minute
)

// customerExportTaskPayload 保存任务入队时的查询条件。
// 后台真正执行导出时，会直接按这份快照来跑，避免前端条件变化影响结果。
type customerExportTaskPayload struct {
	Request models.SysCustomerListRequest `json:"request"`
}

// customerExportDispatcher 负责从队列里挑出可执行任务，
// 再把任务分发给后台 worker 去实际导出。
type customerExportDispatcher struct {
	// taskCh 是真正交给 worker 执行的任务 ID 通道。
	taskCh chan uint
	// wakeCh 是调度器的唤醒信号通道。
	// 新任务入队、worker 完成任务后，都会尝试往这里发一个信号，促使调度器立即再跑一轮。
	wakeCh chan struct{}
}

var (
	customerExportDispatcherOnce sync.Once
	customerExportDispatcherInst *customerExportDispatcher
)

func customerExportTaskDB(ctx context.Context) *gorm.DB {
	return app.DB().WithContext(ctx).Clauses(dbresolver.Write)
}

// customerExportTaskLastActiveAt 取任务最后一次“还在动”的时间点。
// 优先级是 updated_at > started_at > created_at，用来判断任务是不是已经卡住。
func customerExportTaskLastActiveAt(task *models.SysCustomerExportTask) *time.Time {
	if task == nil {
		return nil
	}
	if task.UpdatedAt != nil && !task.UpdatedAt.IsZero() {
		return task.UpdatedAt
	}
	if task.StartedAt != nil && !task.StartedAt.IsZero() {
		return task.StartedAt
	}
	if task.CreatedAt != nil && !task.CreatedAt.IsZero() {
		return task.CreatedAt
	}
	return nil
}

func isCustomerExportTaskStale(task *models.SysCustomerExportTask) bool {
	lastActiveAt := customerExportTaskLastActiveAt(task)
	if lastActiveAt == nil {
		return false
	}
	return time.Since(*lastActiveAt) >= customerExportTaskActiveTimeout
}

// recycleStaleActiveCustomerExportTask 回收超时未完成的任务。
// 这里会把任务直接改成 failed，避免前端一直看到一个永远不结束的“运行中”状态。
func (s *SysCustomerService) recycleStaleActiveCustomerExportTask(ctx context.Context, task *models.SysCustomerExportTask) (bool, error) {
	if task == nil || task.ID == 0 {
		return false, nil
	}

	now := time.Now()
	tx := customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Where("id = ?", task.ID).
		Where("status = ?", task.Status).
		Updates(map[string]interface{}{
			"status":        models.CustomerExportTaskStatusFailed,
			"finished_at":   now,
			"updated_at":    now,
			"error_message": "导出任务长时间未完成，系统已自动回收，请重新发起导出。",
		})
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

// StartCustomerExportDispatcher 启动客户导出任务调度器。
// 服务启动时会顺手把上次意外中断、但还能重试的任务重新放回队列。
func StartCustomerExportDispatcher() {
	customerExportDispatcherOnce.Do(func() {
		workerCount := getCustomerExportWorkerCount()
		dispatcher := &customerExportDispatcher{
			taskCh: make(chan uint, workerCount),
			wakeCh: make(chan struct{}, 1),
		}
		customerExportDispatcherInst = dispatcher

		service := NewSysCustomerService()
		if err := service.requeueRecoverableCustomerExportTasks(context.Background()); err != nil {
			app.ZapLog.Warn("恢复客户导出任务失败", zap.Error(err))
		}

		for i := 0; i < workerCount; i++ {
			go dispatcher.workerLoop(i + 1)
		}
		go dispatcher.dispatchLoop()
		dispatcher.wake()
	})
}

func wakeCustomerExportDispatcher() {
	if customerExportDispatcherInst != nil {
		customerExportDispatcherInst.wake()
	}
}

// wake 尝试向调度器发送一次“立刻检查队列”的信号。
// 通道容量只有 1，所以如果已经有未消费的唤醒信号，就直接跳过，避免重复堆积。
func (d *customerExportDispatcher) wake() {
	select {
	case d.wakeCh <- struct{}{}:
	default:
	}
}

func (d *customerExportDispatcher) dispatchLoop() {
	// 正常情况下由 wake 事件触发即时调度；
	// 定时器只是兜底，防止某些场景漏掉唤醒信号。
	ticker := time.NewTicker(customerExportDispatchInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			d.dispatchOnce()
		case <-d.wakeCh:
			d.dispatchOnce()
		}
	}
}

// dispatchOnce 执行一轮调度：
// 1. 先查当前已经在跑的任务，计算还剩多少空闲 worker；
// 2. 再从排队队列里按顺序挑任务；
// 3. 同一时刻限制“同租户只跑一个任务、同用户只跑一个任务”；
// 4. 抢占成功后把任务 ID 投递给 worker 处理。
func (d *customerExportDispatcher) dispatchOnce() {
	service := NewSysCustomerService()
	ctx := context.Background()

	runningTasks, err := service.listRunningCustomerExportTasks(ctx)
	if err != nil {
		app.ZapLog.Warn("查询执行中的客户导出任务失败", zap.Error(err))
		return
	}

	workerCount := getCustomerExportWorkerCount()
	slots := workerCount - len(runningTasks)
	if slots <= 0 {
		return
	}

	// 这两个集合用于做运行中任务去重，避免同一个租户或同一个用户并发导出。
	runningTenant := make(map[uint]struct{}, len(runningTasks))
	runningUser := make(map[string]struct{}, len(runningTasks))
	for _, task := range runningTasks {
		runningTenant[task.TenantID] = struct{}{}
		runningUser[buildCustomerExportTaskUserKey(task.TenantID, task.UserID)] = struct{}{}
	}

	queuedTasks, err := service.listQueuedCustomerExportTasks(ctx, customerExportQueueScanLimit)
	if err != nil {
		app.ZapLog.Warn("查询待执行的客户导出任务失败", zap.Error(err))
		return
	}

	for _, task := range queuedTasks {
		if slots <= 0 {
			return
		}

		if _, exists := runningTenant[task.TenantID]; exists {
			continue
		}
		userKey := buildCustomerExportTaskUserKey(task.TenantID, task.UserID)
		if _, exists := runningUser[userKey]; exists {
			continue
		}

		claimed, claimErr := service.claimCustomerExportTask(ctx, task.ID)
		if claimErr != nil {
			app.ZapLog.Warn("抢占客户导出任务失败",
				zap.Uint("task_id", task.ID),
				zap.Error(claimErr),
			)
			continue
		}
		if !claimed {
			continue
		}

		runningTenant[task.TenantID] = struct{}{}
		runningUser[userKey] = struct{}{}
		slots--
		d.taskCh <- task.ID
	}
}

func (d *customerExportDispatcher) workerLoop(workerID int) {
	for taskID := range d.taskCh {
		service := NewSysCustomerService()
		if err := service.executeQueuedCustomerExportTask(context.Background(), taskID); err != nil {
			app.ZapLog.Error("执行客户导出任务失败",
				zap.Int("worker_id", workerID),
				zap.Uint("task_id", taskID),
				zap.Error(err),
			)
		}
		d.wake()
	}
}

// getCustomerExportWorkerCount 读取配置项 platform.customer_export_worker_count。
// 它表示“客户导出任务最多同时开几个后台 worker 并发执行”。
// 例如配置成 3，就代表最多同时处理 3 个客户导出任务。
// 这里还做了兜底限制：
// 1. 小于等于 0 时，回退到默认值 2；
// 2. 大于 8 时，强制按 8 处理，避免并发过高把数据库或磁盘导出压垮。
func getCustomerExportWorkerCount() int {
	workerCount := app.ConfigYml.GetInt("platform.customer_export_worker_count")
	if workerCount <= 0 {
		return defaultCustomerExportWorkerCount
	}
	if workerCount > 8 {
		return 8
	}
	return workerCount
}

// buildCustomerExportTaskUserKey 把租户 ID 和用户 ID 拼成一个唯一键，
// 方便在内存里快速判断“这个用户是不是已经有任务在跑”。
func buildCustomerExportTaskUserKey(tenantID, userID uint) string {
	return fmt.Sprintf("%d:%d", tenantID, userID)
}

// truncateCustomerExportErrorMessage 截断导出失败信息，避免写入数据库时过长。
func truncateCustomerExportErrorMessage(err error) string {
	if err == nil {
		return ""
	}

	message := strings.TrimSpace(err.Error())
	if len(message) <= customerExportFailureMessageMaxLen {
		return message
	}
	return message[:customerExportFailureMessageMaxLen]
}

// enqueueCustomerExportTask 创建一条新的导出任务，
// 并把当时的查询条件一并保存下来，交给后台异步处理。
func (s *SysCustomerService) enqueueCustomerExportTask(
	ctx context.Context,
	req models.SysCustomerListRequest,
	claims app.Claims,
	total int64,
	snapshotMaxID int,
) (*models.SysCustomerExportTask, error) {
	payload, err := json.Marshal(customerExportTaskPayload{
		Request: cloneSysCustomerListRequest(req),
	})
	if err != nil {
		return nil, err
	}

	task := models.NewSysCustomerExportTask()
	task.TenantID = claims.TenantID
	task.UserID = claims.UserID
	task.BizType = models.CustomerExportTaskBizTypeSysCustomer
	task.Status = models.CustomerExportTaskStatusQueued
	task.RequestJSON = string(payload)
	task.Total = total
	task.SnapshotMaxID = snapshotMaxID

	if err = task.Create(ctx); err != nil {
		return nil, err
	}

	wakeCustomerExportDispatcher()
	return task, nil
}

// findActiveCustomerExportTask 查询当前用户是否已经有未完成的导出任务。
// 如果查到的是超时卡住的任务，会先回收，再继续往下查一次。
func (s *SysCustomerService) findActiveCustomerExportTask(ctx context.Context, tenantID, userID uint) (*models.SysCustomerExportTask, error) {
	for attempt := 0; attempt < 2; attempt++ {
		var task models.SysCustomerExportTask
		tx := customerExportTaskDB(ctx).
			Where("tenant_id = ?", tenantID).
			Where("user_id = ?", userID).
			Where("biz_type = ?", models.CustomerExportTaskBizTypeSysCustomer).
			Where("status IN ?", []string{models.CustomerExportTaskStatusQueued, models.CustomerExportTaskStatusRunning}).
			Order("id DESC").
			First(&task)
		if tx.Error != nil {
			return nil, tx.Error
		}
		if tx.RowsAffected == 0 || task.ID == 0 {
			return nil, nil
		}
		if !isCustomerExportTaskStale(&task) {
			return &task, nil
		}

		recycled, recycleErr := s.recycleStaleActiveCustomerExportTask(ctx, &task)
		if recycleErr != nil {
			return nil, recycleErr
		}
		if !recycled {
			continue
		}
	}

	return nil, nil
}

// GetExportTask 按“任务 ID + 租户 + 用户 + 业务类型”查询单个导出任务。
// 这样可以保证用户只能看到自己租户下、自己发起的那条客户导出任务。
func (s *SysCustomerService) GetExportTask(ctx context.Context, tenantID, userID, taskID uint) (*models.SysCustomerExportTask, error) {
	if taskID == 0 || tenantID == 0 || userID == 0 {
		return nil, nil
	}

	var task models.SysCustomerExportTask
	tx := customerExportTaskDB(ctx).
		Where("id = ?", taskID).
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Where("biz_type = ?", models.CustomerExportTaskBizTypeSysCustomer).
		First(&task)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 || task.ID == 0 {
		return nil, nil
	}
	return &task, nil
}

// requeueRecoverableCustomerExportTasks 在服务重启后回收“运行中”但实际已经中断的任务，
// 把它们改回排队状态，等调度器重新处理。
func (s *SysCustomerService) requeueRecoverableCustomerExportTasks(ctx context.Context) error {
	now := time.Now()
	return customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Where("biz_type = ?", models.CustomerExportTaskBizTypeSysCustomer).
		Where("status = ?", models.CustomerExportTaskStatusRunning).
		Updates(map[string]interface{}{
			"status":        models.CustomerExportTaskStatusQueued,
			"started_at":    nil,
			"finished_at":   nil,
			"processed":     0,
			"progress":      0,
			"error_message": "",
			"updated_at":    now,
		}).Error
}

// listRunningCustomerExportTasks 查询当前正在执行的导出任务。
// 调度器会用这份结果控制并发数，并避免同租户或同用户任务同时运行。
func (s *SysCustomerService) listRunningCustomerExportTasks(ctx context.Context) ([]models.SysCustomerExportTask, error) {
	var tasks []models.SysCustomerExportTask
	err := customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Select("id", "tenant_id", "user_id").
		Where("biz_type = ?", models.CustomerExportTaskBizTypeSysCustomer).
		Where("status = ?", models.CustomerExportTaskStatusRunning).
		Order("id ASC").
		Find(&tasks).Error
	return tasks, err
}

// listQueuedCustomerExportTasks 按先进先出的顺序取出待执行任务，
// 供调度器挑选并分配给 worker。
func (s *SysCustomerService) listQueuedCustomerExportTasks(ctx context.Context, limit int) ([]models.SysCustomerExportTask, error) {
	var tasks []models.SysCustomerExportTask
	err := customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Select("id", "tenant_id", "user_id", "status").
		Where("biz_type = ?", models.CustomerExportTaskBizTypeSysCustomer).
		Where("status = ?", models.CustomerExportTaskStatusQueued).
		Order("id ASC").
		Limit(limit).
		Find(&tasks).Error
	return tasks, err
}

// claimCustomerExportTask 尝试把排队中的任务改成执行中。
// 返回 true 说明当前实例抢占成功，可以继续处理这个任务。
func (s *SysCustomerService) claimCustomerExportTask(ctx context.Context, taskID uint) (bool, error) {
	now := time.Now()
	tx := customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Where("id = ?", taskID).
		Where("biz_type = ?", models.CustomerExportTaskBizTypeSysCustomer).
		Where("status = ?", models.CustomerExportTaskStatusQueued).
		Updates(map[string]interface{}{
			"status":        models.CustomerExportTaskStatusRunning,
			"started_at":    now,
			"finished_at":   nil,
			"processed":     0,
			"progress":      0,
			"error_message": "",
			"updated_at":    now,
		})
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

// executeQueuedCustomerExportTask 真正执行导出任务：
// 读取快照条件、生成文件、更新进度，并在结束后发送通知。
func (s *SysCustomerService) executeQueuedCustomerExportTask(ctx context.Context, taskID uint) error {
	task := models.NewSysCustomerExportTask()
	if err := task.GetByID(ctx, taskID); err != nil {
		return err
	}
	if task.Status != models.CustomerExportTaskStatusRunning {
		return nil
	}

	var payload customerExportTaskPayload
	if err := json.Unmarshal([]byte(task.RequestJSON), &payload); err != nil {
		updateErr := s.markCustomerExportTaskFailed(ctx, task.ID, err)
		s.notifyAsyncExportFailed(ctx, app.Claims{
			ClaimsUser: app.ClaimsUser{
				UserID:   task.UserID,
				TenantID: task.TenantID,
			},
		}, task.ID, err)
		if updateErr != nil {
			return updateErr
		}
		return err
	}

	claims := app.Claims{
		ClaimsUser: app.ClaimsUser{
			UserID:   task.UserID,
			TenantID: task.TenantID,
		},
	}
	authCtx := buildExportAuthContext(claims)

	// 导出前先清一遍当前租户的过期历史文件，避免本地存储无限增长。
	if err := s.cleanupExpiredAsyncExports(ctx, task.TenantID); err != nil {
		app.ZapLog.Warn("清理过期客户导出文件失败",
			zap.Uint("task_id", task.ID),
			zap.Uint("tenant_id", task.TenantID),
			zap.Error(err),
		)
	}

	progressHook := func(processed int64) {
		if progressErr := s.updateCustomerExportTaskProgress(context.Background(), task.ID, processed, task.Total); progressErr != nil {
			app.ZapLog.Warn("更新客户导出任务进度失败",
				zap.Uint("task_id", task.ID),
				zap.Error(progressErr),
			)
		}
	}

	// generateAsyncExportFile 会按任务快照重新查询数据并生成导出附件，
	// 中途通过 progressHook 持续回写进度。
	affix, err := s.generateAsyncExportFile(
		ctx,
		authCtx,
		payload.Request,
		claims,
		task.Total,
		task.SnapshotMaxID,
		progressHook,
	)
	if err != nil {
		if updateErr := s.markCustomerExportTaskFailed(ctx, task.ID, err); updateErr != nil {
			app.ZapLog.Warn("更新客户导出任务失败状态失败",
				zap.Uint("task_id", task.ID),
				zap.Error(updateErr),
			)
		}
		s.notifyAsyncExportFailed(ctx, claims, task.ID, err)
		return err
	}

	if err = s.markCustomerExportTaskSucceeded(ctx, task.ID, affix, task.Total); err != nil {
		return err
	}

	if err = noticeService.NewNoticeBusinessService().NotifyExportReady(
		ctx,
		claims.TenantID,
		claims.UserID,
		"客户导出已完成",
		fmt.Sprintf("已为您生成 %d 条客户数据导出文件，点击即可下载。", task.Total),
		fmt.Sprintf("%d", affix.ID),
		map[string]interface{}{
			"affixId":  affix.ID,
			"fileName": affix.Name,
			"total":    task.Total,
			"taskId":   task.ID,
		},
	); err != nil {
		app.ZapLog.Warn("发送客户导出完成通知失败",
			zap.Uint("task_id", task.ID),
			zap.Uint("tenant_id", claims.TenantID),
			zap.Uint("user_id", claims.UserID),
			zap.Uint("affix_id", affix.ID),
			zap.Error(err),
		)
	}

	return nil
}

// updateCustomerExportTaskProgress 按当前已处理条数更新进度。
// 这里会把进度上限卡在 99%，最终 100% 只在成功收尾时写入。
func (s *SysCustomerService) updateCustomerExportTaskProgress(ctx context.Context, taskID uint, processed int64, total int64) error {
	progress := 0
	if total > 0 {
		progress = int((processed * 100) / total)
		if progress > 99 {
			progress = 99
		}
	}

	return customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Where("id = ?", taskID).
		Where("status = ?", models.CustomerExportTaskStatusRunning).
		Updates(map[string]interface{}{
			"processed": processed,
			"progress":  progress,
		}).Error
}

// markCustomerExportTaskSucceeded 在导出成功后统一收尾：
// 写入 success 状态、100% 进度、完成时间，以及生成出来的附件信息。
func (s *SysCustomerService) markCustomerExportTaskSucceeded(
	ctx context.Context,
	taskID uint,
	affix *appModels.SysAffix,
	total int64,
) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":      models.CustomerExportTaskStatusSuccess,
		"processed":   total,
		"progress":    100,
		"finished_at": now,
		"updated_at":  now,
	}
	if affix != nil {
		updates["affix_id"] = affix.ID
		updates["file_name"] = affix.Name
	}

	return customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Where("id = ?", taskID).
		Updates(updates).Error
}

// markCustomerExportTaskFailed 在导出失败后统一收尾：
// 写入 failed 状态、完成时间，并记录截断后的失败原因。
func (s *SysCustomerService) markCustomerExportTaskFailed(ctx context.Context, taskID uint, exportErr error) error {
	now := time.Now()
	return customerExportTaskDB(ctx).
		Model(&models.SysCustomerExportTask{}).
		Where("id = ?", taskID).
		Updates(map[string]interface{}{
			"status":        models.CustomerExportTaskStatusFailed,
			"finished_at":   now,
			"updated_at":    now,
			"error_message": truncateCustomerExportErrorMessage(exportErr),
		}).Error
}
