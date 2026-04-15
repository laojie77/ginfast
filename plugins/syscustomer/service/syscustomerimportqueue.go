package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	"gin-fast/plugins/syscustomer/models"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	defaultCustomerImportWorkerCount = 1
	customerImportQueueScanLimit     = 10
)

type customerImportDispatcher struct {
	taskCh chan int
	wakeCh chan struct{}
}

var (
	customerImportDispatcherOnce sync.Once
	customerImportDispatcherInst *customerImportDispatcher
)

func StartCustomerImportDispatcher() {
	customerImportDispatcherOnce.Do(func() {
		workerCount := getCustomerImportWorkerCount()
		dispatcher := &customerImportDispatcher{
			taskCh: make(chan int, workerCount),
			wakeCh: make(chan struct{}, 1),
		}
		customerImportDispatcherInst = dispatcher

		service := NewSysCustomerService()
		if err := service.recoverInterruptedCustomerImportBatches(context.Background()); err != nil {
			app.ZapLog.Warn("recover customer import batches failed", zap.Error(err))
		}

		for i := 0; i < workerCount; i++ {
			go dispatcher.workerLoop(i + 1)
		}
		go dispatcher.dispatchLoop()
		dispatcher.wake()
	})
}

func getCustomerImportWorkerCount() int {
	workerCount := app.ConfigYml.GetInt("platform.customer_import_worker_count")
	if workerCount <= 0 {
		return defaultCustomerImportWorkerCount
	}
	if workerCount > 4 {
		return 4
	}
	return workerCount
}

func wakeCustomerImportDispatcher() {
	if customerImportDispatcherInst != nil {
		customerImportDispatcherInst.wake()
	}
}

func (d *customerImportDispatcher) wake() {
	select {
	case d.wakeCh <- struct{}{}:
	default:
	}
}

func (d *customerImportDispatcher) dispatchLoop() {
	ticker := time.NewTicker(5 * time.Second)
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

func (d *customerImportDispatcher) dispatchOnce() {
	service := NewSysCustomerService()
	ctx := context.Background()

	runningTasks, err := service.listRunningCustomerImportBatches(ctx)
	if err != nil {
		app.ZapLog.Warn("list running customer import batches failed", zap.Error(err))
		return
	}

	workerCount := getCustomerImportWorkerCount()
	slots := workerCount - len(runningTasks)
	if slots <= 0 {
		return
	}

	runningUsers := make(map[string]struct{}, len(runningTasks))
	for _, task := range runningTasks {
		runningUsers[buildCustomerImportTaskUserKey(uint(task.TenantID), uint(task.UserID))] = struct{}{}
	}

	pendingTasks, err := service.listPendingCustomerImportBatches(ctx, customerImportQueueScanLimit)
	if err != nil {
		app.ZapLog.Warn("list pending customer import batches failed", zap.Error(err))
		return
	}

	for _, task := range pendingTasks {
		if slots <= 0 {
			return
		}

		userKey := buildCustomerImportTaskUserKey(uint(task.TenantID), uint(task.UserID))
		if _, exists := runningUsers[userKey]; exists {
			continue
		}

		claimed, claimErr := service.claimCustomerImportBatch(ctx, task.ID)
		if claimErr != nil {
			app.ZapLog.Warn("claim customer import batch failed",
				zap.Int("batch_id", task.ID),
				zap.Error(claimErr),
			)
			continue
		}
		if !claimed {
			continue
		}

		runningUsers[userKey] = struct{}{}
		slots--
		d.taskCh <- task.ID
	}
}

func (d *customerImportDispatcher) workerLoop(workerID int) {
	for batchID := range d.taskCh {
		service := NewSysCustomerService()
		if err := service.executeQueuedCustomerImportBatch(context.Background(), batchID); err != nil {
			app.ZapLog.Warn("execute customer import batch finished with error",
				zap.Int("worker_id", workerID),
				zap.Int("batch_id", batchID),
				zap.Error(err),
			)
		}
		d.wake()
	}
}

func buildCustomerImportTaskUserKey(tenantID, userID uint) string {
	return fmt.Sprintf("%d:%d", tenantID, userID)
}

func (s *SysCustomerService) findActiveCustomerImportBatch(ctx context.Context, tenantID, userID int) (*models.SysCustomerImportBatch, error) {
	var batch models.SysCustomerImportBatch
	tx := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Where("status IN ?", []string{
			models.CustomerImportBatchStatusPending,
			models.CustomerImportBatchStatusRunning,
			models.CustomerImportBatchStatusCanceling,
		}).
		Order("id DESC").
		First(&batch)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tx.Error
	}
	if batch.ID == 0 {
		return nil, nil
	}
	return &batch, nil
}

func (s *SysCustomerService) listRunningCustomerImportBatches(ctx context.Context) ([]models.SysCustomerImportBatch, error) {
	var batches []models.SysCustomerImportBatch
	err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Select("id", "tenant_id", "user_id").
		Where("status IN ?", []string{
			models.CustomerImportBatchStatusRunning,
			models.CustomerImportBatchStatusCanceling,
		}).
		Find(&batches).Error
	return batches, err
}

func (s *SysCustomerService) listPendingCustomerImportBatches(ctx context.Context, limit int) ([]models.SysCustomerImportBatch, error) {
	if limit <= 0 {
		limit = customerImportQueueScanLimit
	}

	var batches []models.SysCustomerImportBatch
	err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("status = ?", models.CustomerImportBatchStatusPending).
		Order("id ASC").
		Limit(limit).
		Find(&batches).Error
	return batches, err
}

func (s *SysCustomerService) claimCustomerImportBatch(ctx context.Context, batchID int) (bool, error) {
	now := time.Now()
	tx := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batchID).
		Where("status = ?", models.CustomerImportBatchStatusPending).
		Updates(map[string]interface{}{
			"status":     models.CustomerImportBatchStatusRunning,
			"started_at": now,
			"updated_at": now,
		})
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

func (s *SysCustomerService) recoverInterruptedCustomerImportBatches(ctx context.Context) error {
	now := time.Now()
	var runningBatches []models.SysCustomerImportBatch
	if err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("status IN ?", []string{
			models.CustomerImportBatchStatusRunning,
			models.CustomerImportBatchStatusCanceling,
		}).
		Find(&runningBatches).Error; err != nil {
		return err
	}

	for _, batch := range runningBatches {
		resumeRow := batch.ResumeRow
		if resumeRow <= 0 {
			resumeRow = batch.StartRow
		}

		status := models.CustomerImportBatchStatusFailed
		if batch.Status == models.CustomerImportBatchStatusCanceling {
			status = models.CustomerImportBatchStatusCanceled
		} else if batch.ProcessedCount > 0 {
			status = models.CustomerImportBatchStatusPartial
		}

		message := fmt.Sprintf("导入任务因服务重启中断，建议从第 %d 行继续导入。", resumeRow)
		deletedCount := int64(0)
		if status == models.CustomerImportBatchStatusCanceled {
			var err error
			deletedCount, err = s.deleteImportedCustomersByBatchID(ctx, batch.ID)
			if err != nil {
				return err
			}
			resumeRow = 0
			message = buildCustomerImportCanceledMessage(batch.SuccessCount, batch.FailedCount, batch.DuplicateCount, deletedCount)
		}
		if strings.TrimSpace(batch.ErrorMessage) != "" {
			message = strings.TrimSpace(batch.ErrorMessage)
		}

		if err := app.DB().WithContext(ctx).
			Model(&models.SysCustomerImportBatch{}).
			Where("id = ?", batch.ID).
			Updates(map[string]interface{}{
				"status":        status,
				"resume_row":    resumeRow,
				"error_message": message,
				"finished_at":   now,
				"updated_at":    now,
			}).Error; err != nil {
			return err
		}
	}

	return nil
}

func (s *SysCustomerService) deleteCustomerImportAffix(ctx context.Context, affixID uint) error {
	if affixID == 0 {
		return nil
	}
	return app.DB().WithContext(ctx).Delete(&appModels.SysAffix{}, affixID).Error
}
