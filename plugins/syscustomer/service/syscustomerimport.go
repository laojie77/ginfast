package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	"gin-fast/app/utils/common"
	"gin-fast/exampleutils/snowflakehelper"
	channelModels "gin-fast/plugins/syschannel/models"
	channelCompanyModels "gin-fast/plugins/syschannelcompany/models"
	"gin-fast/plugins/syscustomer/models"
	noticeService "gin-fast/plugins/sysnotice/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	customerImportFailurePreviewLimit = 20
	customerImportBatchFileMaxErrLen  = 900
	customerImportInsertBatchSize     = 500
	customerImportProgressFlushStep   = 100
	customerImportTempRelativeDir     = "import/syscustomer"
	customerImportMaxFileSizeMB       = 100
)

type customerImportChannelContext struct {
	ChannelCompanyID int
	ChannelID        int
	ChannelKey       string
	ChannelName      string
}

type customerImportRow struct {
	RowIndex     int
	Name         string
	Mobile       string
	Age          int
	CustomerStar int
	MoneyDemand  int
	Remarks      string
	Sex          int
}

type customerImportFailure struct {
	Row     int
	Name    string
	Mobile  string
	Message string
}

type customerImportFailurePreviewPayload struct {
	Failures       []models.SysCustomerImportFailure `json:"failures"`
	DuplicateCount int                               `json:"duplicateCount,omitempty"`
}

type customerImportPendingCustomer struct {
	RowIndex int
	Customer models.SysCustomer
}

type customerImportPublicError struct {
	message string
}

var errCustomerImportCanceled = errors.New("导入任务已手动终止")

func (e *customerImportPublicError) Error() string {
	return e.message
}

func newCustomerImportPublicError(format string, args ...interface{}) error {
	return &customerImportPublicError{
		message: fmt.Sprintf(format, args...),
	}
}

func IsCustomerImportPublicError(err error) bool {
	var target *customerImportPublicError
	return errors.As(err, &target)
}

func isCustomerImportCanceledError(err error) bool {
	return errors.Is(err, errCustomerImportCanceled)
}

func (s *SysCustomerService) ImportCustomers(
	c *gin.Context,
	req models.SysCustomerImportRequest,
	fileHeader *multipart.FileHeader,
) (*models.SysCustomerImportSubmitResult, error) {
	// 提交导入时只做参数校验、保存文件和创建批次记录，
	// 真正的 Excel 解析交给后台队列，避免请求长时间阻塞。
	StartCustomerImportDispatcher()

	if err := s.validateCustomerImportFile(fileHeader); err != nil {
		return nil, err
	}

	tenantID := int(common.GetCurrentTenantID(c))
	userID := int(common.GetCurrentUserID(c))
	if tenantID <= 0 || userID <= 0 {
		return nil, newCustomerImportPublicError("当前登录状态已失效")
	}

	existingBatch, err := s.findActiveCustomerImportBatch(c, tenantID, userID)
	if err != nil {
		return nil, err
	}
	if existingBatch != nil {
		return &models.SysCustomerImportSubmitResult{
			BatchID:  existingBatch.ID,
			Status:   existingBatch.Status,
			Existing: true,
			Message:  "当前已有导入任务正在处理中，请稍后查看进度。",
		}, nil
	}

	channelCtx, err := s.getCustomerImportChannelContext(c, tenantID, req.ChannelCompanyID)
	if err != nil {
		return nil, err
	}

	deptID, err := s.getCurrentUserDeptID(c, uint(userID))
	if err != nil {
		return nil, err
	}

	filePath, err := s.saveCustomerImportProcessingFile(fileHeader)
	if err != nil {
		return nil, err
	}

	batch := &models.SysCustomerImportBatch{
		TenantID:         tenantID,
		UserID:           userID,
		DeptID:           deptID,
		Scene:            req.Scene,
		ChannelCompanyID: channelCtx.ChannelCompanyID,
		ChannelID:        channelCtx.ChannelID,
		ChannelKey:       channelCtx.ChannelKey,
		ChannelName:      channelCtx.ChannelName,
		StartRow:         normalizeCustomerImportStartRow(req.StartRow),
		Remark:           req.Remark,
		Status:           models.CustomerImportBatchStatusPending,
		FileName:         strings.TrimSpace(fileHeader.Filename),
		FilePath:         filePath,
		ResumeRow:        normalizeCustomerImportStartRow(req.StartRow),
	}
	if err = batch.Create(c); err != nil {
		_ = os.Remove(filePath)
		return nil, err
	}

	wakeCustomerImportDispatcher()
	s.tryCleanupExpiredCustomerImportFiles(context.Background(), uint(tenantID), uint(userID), batch.ID)

	return &models.SysCustomerImportSubmitResult{
		BatchID: batch.ID,
		Status:  batch.Status,
		Message: "导入任务已提交，系统正在后台处理中。",
	}, nil
}

func (s *SysCustomerService) GetImportBatch(ctx context.Context, tenantID, userID uint, batchID int) (*models.SysCustomerImportBatchResult, error) {
	if batchID <= 0 || tenantID == 0 || userID == 0 {
		return nil, nil
	}

	var batch models.SysCustomerImportBatch
	tx := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batchID).
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		First(&batch)
	if tx.Error != nil {
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 || batch.ID == 0 {
		return nil, nil
	}

	return buildCustomerImportBatchResult(&batch), nil
}

func (s *SysCustomerService) GetLatestImportBatch(ctx context.Context, tenantID, userID uint) (*models.SysCustomerImportBatchResult, error) {
	if tenantID == 0 || userID == 0 {
		return nil, nil
	}

	var batch models.SysCustomerImportBatch
	tx := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		Where("status IN ?", []string{
			models.CustomerImportBatchStatusPending,
			models.CustomerImportBatchStatusRunning,
			models.CustomerImportBatchStatusCanceling,
			models.CustomerImportBatchStatusPartial,
			models.CustomerImportBatchStatusCanceled,
			models.CustomerImportBatchStatusFailed,
		}).
		Order("id DESC").
		First(&batch)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 || batch.ID == 0 {
		return nil, nil
	}

	return buildCustomerImportBatchResult(&batch), nil
}

func (s *SysCustomerService) CancelImportBatch(
	ctx context.Context,
	tenantID, userID uint,
	batchID int,
) (*models.SysCustomerImportBatchResult, error) {
	if batchID <= 0 || tenantID == 0 || userID == 0 {
		return nil, nil
	}

	var batch models.SysCustomerImportBatch
	tx := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batchID).
		Where("tenant_id = ?", tenantID).
		Where("user_id = ?", userID).
		First(&batch)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, tx.Error
	}
	if tx.RowsAffected == 0 || batch.ID == 0 {
		return nil, nil
	}

	now := time.Now()
	switch batch.Status {
	case models.CustomerImportBatchStatusPending:
		resumeRow := 0
		message := "导入任务已手动终止，当前批次无已导入客户数据。"
		if err := app.DB().WithContext(ctx).
			Model(&models.SysCustomerImportBatch{}).
			Where("id = ?", batch.ID).
			Updates(map[string]interface{}{
				"status":        models.CustomerImportBatchStatusCanceled,
				"resume_row":    resumeRow,
				"error_message": message,
				"finished_at":   now,
				"updated_at":    now,
			}).Error; err != nil {
			return nil, err
		}

		batch.Status = models.CustomerImportBatchStatusCanceled
		batch.ResumeRow = resumeRow
		batch.ErrorMessage = message
		batch.FinishedAt = &now
		batch.UpdatedAt = &now
		s.pushCustomerImportResultNotice(
			ctx,
			&batch,
			batch.Status,
			batch.TotalCount,
			batch.ProcessedCount,
			batch.SuccessCount,
			batch.FailedCount,
			batch.DuplicateCount,
			batch.Progress,
			batch.ResumeRow,
			message,
		)
		return buildCustomerImportBatchResult(&batch), nil
	case models.CustomerImportBatchStatusRunning:
		if err := app.DB().WithContext(ctx).
			Model(&models.SysCustomerImportBatch{}).
			Where("id = ?", batch.ID).
			Where("status = ?", models.CustomerImportBatchStatusRunning).
			Updates(map[string]interface{}{
				"status":     models.CustomerImportBatchStatusCanceling,
				"updated_at": now,
			}).Error; err != nil {
			return nil, err
		}

		batch.Status = models.CustomerImportBatchStatusCanceling
		batch.UpdatedAt = &now
		return buildCustomerImportBatchResult(&batch), nil
	case models.CustomerImportBatchStatusCanceling,
		models.CustomerImportBatchStatusCanceled,
		models.CustomerImportBatchStatusSuccess,
		models.CustomerImportBatchStatusPartial,
		models.CustomerImportBatchStatusFailed:
		return buildCustomerImportBatchResult(&batch), nil
	default:
		return buildCustomerImportBatchResult(&batch), nil
	}
}

func buildCustomerImportRealtimeExtra(
	batch *models.SysCustomerImportBatch,
	totalCount, processedCount, successCount, failedCount, duplicateCount, progress, resumeRow int,
	errorMessage string,
) map[string]interface{} {
	if batch == nil {
		return nil
	}

	extra := map[string]interface{}{
		"batchId":        batch.ID,
		"status":         batch.Status,
		"startRow":       batch.StartRow,
		"resumeRow":      resumeRow,
		"interrupted":    resumeRow > 0 && batch.Status != models.CustomerImportBatchStatusSuccess && processedCount < totalCount,
		"totalCount":     totalCount,
		"processedCount": processedCount,
		"successCount":   successCount,
		"failedCount":    failedCount,
		"duplicateCount": duplicateCount,
		"progress":       progress,
		"fileName":       strings.TrimSpace(batch.FileName),
		"errorMessage":   strings.TrimSpace(errorMessage),
	}
	if remark := strings.TrimSpace(batch.Remark); remark != "" {
		extra["remark"] = remark
	}
	return extra
}

func (s *SysCustomerService) pushCustomerImportProgressNotice(
	ctx context.Context,
	batch *models.SysCustomerImportBatch,
	totalCount, processedCount, successCount, failedCount, duplicateCount, progress, resumeRow int,
) {
	if batch == nil {
		return
	}

	content := buildCustomerImportProgressContent(totalCount, processedCount, successCount, failedCount, duplicateCount)
	_ = noticeService.NewNoticeBusinessService().NotifyCustomerImportProgress(
		ctx,
		uint(batch.TenantID),
		uint(batch.UserID),
		"客户导入进行中",
		content,
		buildCustomerImportRealtimeExtra(
			batch,
			totalCount,
			processedCount,
			successCount,
			failedCount,
			duplicateCount,
			progress,
			resumeRow,
			"",
		),
	)
}

func (s *SysCustomerService) pushCustomerImportResultNotice(
	ctx context.Context,
	batch *models.SysCustomerImportBatch,
	status string,
	totalCount, processedCount, successCount, failedCount, duplicateCount, progress, resumeRow int,
	errorMessage string,
) {
	if batch == nil {
		return
	}

	title := "客户导入完成"
	level := "success"
	switch status {
	case models.CustomerImportBatchStatusPartial:
		title = "客户导入部分完成"
		level = "warning"
	case models.CustomerImportBatchStatusCanceled:
		title = "客户导入已终止"
		level = "warning"
	case models.CustomerImportBatchStatusFailed:
		title = "客户导入失败"
		level = "error"
	}

	content := buildCustomerImportCountSummary(successCount, failedCount, duplicateCount)
	if strings.TrimSpace(errorMessage) != "" {
		content = strings.TrimSpace(errorMessage)
	}

	tempBatch := *batch
	tempBatch.Status = status
	_ = noticeService.NewNoticeBusinessService().NotifyCustomerImportResult(
		ctx,
		uint(batch.TenantID),
		uint(batch.UserID),
		title,
		content,
		level,
		buildCustomerImportRealtimeExtra(
			&tempBatch,
			totalCount,
			processedCount,
			successCount,
			failedCount,
			duplicateCount,
			progress,
			resumeRow,
			errorMessage,
		),
	)
}

func buildCustomerImportBatchResult(batch *models.SysCustomerImportBatch) *models.SysCustomerImportBatchResult {
	if batch == nil || batch.ID == 0 {
		return nil
	}

	failures, duplicateCount := decodeCustomerImportFailurePreview(batch.FailurePreview)
	if batch.DuplicateCount > 0 || duplicateCount == 0 {
		duplicateCount = batch.DuplicateCount
	}
	interrupted := batch.ResumeRow > 0 &&
		batch.Status != models.CustomerImportBatchStatusSuccess &&
		batch.ProcessedCount < batch.TotalCount

	return &models.SysCustomerImportBatchResult{
		ID:             batch.ID,
		Status:         batch.Status,
		StartRow:       batch.StartRow,
		ResumeRow:      batch.ResumeRow,
		Interrupted:    interrupted,
		TotalCount:     batch.TotalCount,
		ProcessedCount: batch.ProcessedCount,
		SuccessCount:   batch.SuccessCount,
		FailedCount:    batch.FailedCount,
		DuplicateCount: duplicateCount,
		Progress:       batch.Progress,
		Remark:         strings.TrimSpace(batch.Remark),
		ErrorMessage:   strings.TrimSpace(batch.ErrorMessage),
		Failures:       failures,
		StartedAt:      batch.StartedAt,
		FinishedAt:     batch.FinishedAt,
		UpdatedAt:      batch.UpdatedAt,
		FileName:       strings.TrimSpace(batch.FileName),
	}
}

func (s *SysCustomerService) getCustomerImportBatchStatus(ctx context.Context, batchID int) (string, error) {
	if batchID <= 0 {
		return "", nil
	}

	var row struct {
		Status string `gorm:"column:status"`
	}
	if err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Select("status").
		Where("id = ?", batchID).
		Take(&row).Error; err != nil {
		return "", err
	}

	return strings.TrimSpace(row.Status), nil
}

func (s *SysCustomerService) isCustomerImportBatchCanceling(ctx context.Context, batchID int) (bool, error) {
	status, err := s.getCustomerImportBatchStatus(ctx, batchID)
	if err != nil {
		return false, err
	}
	return status == models.CustomerImportBatchStatusCanceling || status == models.CustomerImportBatchStatusCanceled, nil
}

func (s *SysCustomerService) deleteImportedCustomersByBatchID(ctx context.Context, batchID int) (int64, error) {
	if batchID <= 0 {
		return 0, nil
	}

	// 手动终止导入时，需要把当前批次已经写入的客户做物理删除，
	// 不能只打 deleted_at 软删除标记，否则列表虽然看不见，数据实际还留在库里。
	tx := app.DB().WithContext(ctx).
		Unscoped().
		Where("batch_id = ?", batchID).
		Delete(&models.SysCustomer{})
	if tx.Error != nil {
		return 0, tx.Error
	}
	return tx.RowsAffected, nil
}

func normalizeCustomerImportStartRow(startRow int) int {
	if startRow < 2 {
		return 2
	}
	return startRow
}

func calculateCustomerImportProgress(totalCount, processedCount int) int {
	if totalCount <= 0 {
		return 0
	}
	if processedCount <= 0 {
		return 0
	}
	if processedCount >= totalCount {
		return 100
	}
	return int(float64(processedCount) * 100 / float64(totalCount))
}

func buildCustomerImportCountSummary(successCount, failedCount, duplicateCount int) string {
	summary := fmt.Sprintf("导入成功 %d 条，导入失败 %d 条", successCount, failedCount)
	if duplicateCount > 0 {
		summary = fmt.Sprintf("%s，重复数据：%d", summary, duplicateCount)
	}
	return summary + "。"
}

func buildCustomerImportProgressContent(totalCount, processedCount, successCount, failedCount, duplicateCount int) string {
	content := fmt.Sprintf(
		"已处理 %d/%d 行，导入成功 %d 条，导入失败 %d 条",
		processedCount,
		totalCount,
		successCount,
		failedCount,
	)
	if duplicateCount > 0 {
		content = fmt.Sprintf("%s，重复数据：%d", content, duplicateCount)
	}
	return content + "。"
}

func buildCustomerImportSummaryMessage(
	successCount int,
	failedCount int,
	duplicateCount int,
	failures []models.SysCustomerImportFailure,
	interrupted bool,
	resumeRow int,
	importErr error,
) string {
	parts := make([]string, 0, 2)
	if interrupted || failedCount > 0 || duplicateCount > 0 {
		parts = append(parts, buildCustomerImportCountSummary(successCount, failedCount, duplicateCount))
	}

	if interrupted {
		reason := ""
		if importErr != nil {
			reason = strings.TrimSpace(importErr.Error())
		}
		if resumeRow > 0 {
			if reason != "" {
				parts = append(parts, fmt.Sprintf("导入在第 %d 行附近中断，建议从第 %d 行继续导入。原因：%s", resumeRow, resumeRow, reason))
				return strings.Join(parts, "")
			}
			parts = append(parts, fmt.Sprintf("导入在第 %d 行附近中断，建议从第 %d 行继续导入。", resumeRow, resumeRow))
			return strings.Join(parts, "")
		}
		if reason != "" {
			parts = append(parts, fmt.Sprintf("导入任务中断，原因：%s", reason))
			return strings.Join(parts, "")
		}
		parts = append(parts, "导入任务中断，请重新发起导入。")
		return strings.Join(parts, "")
	}

	if failedCount <= 0 {
		return strings.Join(parts, "")
	}

	//if len(failures) > 0 {
	//	first := failures[0]
	//	parts = append(parts, fmt.Sprintf("共有 %d 行因字段不合规被跳过，首条失败：第 %d 行 %s", failedCount, first.Row, strings.TrimSpace(first.Message)))
	//	return strings.Join(parts, "")
	//}

	parts = append(parts, fmt.Sprintf("共有 %d 行因字段不合规被跳过。", failedCount))
	return strings.Join(parts, "")
}

func buildCustomerImportCanceledMessage(successCount, failedCount, duplicateCount int, deletedCount int64) string {
	if successCount == 0 && failedCount == 0 && duplicateCount == 0 && deletedCount == 0 {
		return "导入任务已手动终止，当前批次无已导入客户数据。"
	}

	parts := []string{buildCustomerImportCountSummary(successCount, failedCount, duplicateCount)}
	parts = append(parts, fmt.Sprintf("导入任务已手动终止，已按批次删除 %d 条已导入客户数据。", deletedCount))
	return strings.Join(parts, "")
}

func encodeCustomerImportFailurePreview(failures []models.SysCustomerImportFailure, duplicateCount int) string {
	data, err := json.Marshal(customerImportFailurePreviewPayload{
		Failures:       failures,
		DuplicateCount: duplicateCount,
	})
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeCustomerImportFailurePreview(raw string) ([]models.SysCustomerImportFailure, int) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, 0
	}

	if strings.HasPrefix(raw, "[") {
		var failures []models.SysCustomerImportFailure
		if err := json.Unmarshal([]byte(raw), &failures); err != nil {
			return nil, 0
		}
		return failures, 0
	}

	var payload customerImportFailurePreviewPayload
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, 0
	}
	return payload.Failures, payload.DuplicateCount
}

func appendCustomerImportFailurePreview(
	failures []models.SysCustomerImportFailure,
	failure customerImportFailure,
) []models.SysCustomerImportFailure {
	if len(failures) >= customerImportFailurePreviewLimit {
		return failures
	}
	return append(failures, models.SysCustomerImportFailure{
		Row:     failure.Row,
		Name:    failure.Name,
		Mobile:  failure.Mobile,
		Message: failure.Message,
	})
}

func truncateCustomerImportError(err error) string {
	if err == nil {
		return ""
	}
	message := strings.TrimSpace(err.Error())
	if len(message) <= customerImportBatchFileMaxErrLen {
		return message
	}
	return message[:customerImportBatchFileMaxErrLen]
}

func (s *SysCustomerService) validateCustomerImportFile(fileHeader *multipart.FileHeader) error {
	if fileHeader == nil {
		return newCustomerImportPublicError("请先选择导入文件")
	}
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if ext != ".xlsx" {
		return newCustomerImportPublicError("当前仅支持导入 .xlsx 文件")
	}
	maxSize := int64(customerImportMaxFileSizeMB) * 1024 * 1024
	if fileHeader.Size > maxSize {
		return newCustomerImportPublicError("导入文件大小超过限制，最大允许 %d MB", customerImportMaxFileSizeMB)
	}
	return nil
}

func (s *SysCustomerService) getCustomerImportChannelContext(ctx context.Context, tenantID int, channelCompanyID int) (*customerImportChannelContext, error) {
	if channelCompanyID <= 0 {
		return nil, newCustomerImportPublicError("请选择导入渠道公司")
	}

	type channelJoinRow struct {
		ChannelCompanyID int    `gorm:"column:channel_company_id"`
		ChannelID        int    `gorm:"column:channel_id"`
		ChannelKey       string `gorm:"column:channel_key"`
		ChannelName      string `gorm:"column:channel_name"`
		HiddenName       string `gorm:"column:hidden_name"`
	}

	var row channelJoinRow
	// 这里必须使用 Take/Scan，不能使用 First。
	// First 会根据目标结构体推断主键并附加 ORDER BY cc.channel_company_id，
	// 但 sys_channel_company 的真实主键列是 cc.id，这正是导入时报错的根因。
	err := app.DB().WithContext(ctx).
		Table(channelCompanyModels.SysChannelCompany{}.TableName()+" AS cc").
		Select("cc.id AS channel_company_id", "cc.channel_id", "c.channel_key", "cc.hidden_name", "c.channel_name").
		Joins("JOIN "+channelModels.SysChannel{}.TableName()+" AS c ON c.id = cc.channel_id").
		Where("cc.id = ?", channelCompanyID).
		Where("cc.tenant_id = ?", tenantID).
		Take(&row).Error
	if err == gorm.ErrRecordNotFound {
		return nil, newCustomerImportPublicError("所选渠道公司不存在或无权访问")
	}
	if err != nil {
		return nil, err
	}

	channelName := strings.TrimSpace(row.HiddenName)
	if channelName == "" {
		channelName = strings.TrimSpace(row.ChannelName)
	}

	return &customerImportChannelContext{
		ChannelCompanyID: row.ChannelCompanyID,
		ChannelID:        row.ChannelID,
		ChannelKey:       strings.TrimSpace(row.ChannelKey),
		ChannelName:      channelName,
	}, nil
}

func (s *SysCustomerService) getTenantCityByIDForImport(ctx context.Context, tenantID uint) (string, error) {
	var tenant struct {
		City string `gorm:"column:city"`
	}
	err := app.DB().WithContext(ctx).Table(appModels.Tenant{}.TableName()).Select("city").Where("id = ?", tenantID).First(&tenant).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(tenant.City), nil
}

func resolveCustomerImportStorageRoot() string {
	root := strings.TrimSpace(app.UploadService.GetUploadConfig().LocalPath)
	if root == "" {
		return os.TempDir()
	}
	if filepath.IsAbs(root) {
		return root
	}
	return filepath.Join(app.BasePath, root)
}

func (s *SysCustomerService) saveCustomerImportProcessingFile(fileHeader *multipart.FileHeader) (string, error) {
	root := resolveCustomerImportStorageRoot()
	dateDir := time.Now().Format("2006-01-02")
	targetDir := filepath.Join(root, filepath.FromSlash(customerImportTempRelativeDir), dateDir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	fileName := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102150405"), uuid.NewString(), ext)
	filePath := filepath.Join(targetDir, fileName)

	src, err := fileHeader.Open()
	if err != nil {
		return "", err
	}
	defer func() { _ = src.Close() }()

	dst, err := os.Create(filePath)
	if err != nil {
		return "", err
	}
	defer func() { _ = dst.Close() }()

	if _, err = io.Copy(dst, src); err != nil {
		return "", err
	}
	if err = dst.Sync(); err != nil {
		return "", err
	}

	return filePath, nil
}

func validateCustomerImportProcessingFile(filePath string) (string, error) {
	trimmedPath := strings.TrimSpace(filePath)
	if trimmedPath == "" {
		return "", fmt.Errorf("导入文件不存在或已被清理，请重新上传文件")
	}

	fileInfo, err := os.Stat(trimmedPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("导入文件不存在或已被清理，请重新上传文件")
		}
		return "", fmt.Errorf("读取导入文件失败: %w", err)
	}
	if fileInfo.IsDir() {
		return "", fmt.Errorf("导入文件路径无效，请重新上传文件")
	}

	return trimmedPath, nil
}

func openCustomerImportWorkbook(filePath string) (*excelize.File, string, error) {
	validatedPath, err := validateCustomerImportProcessingFile(filePath)
	if err != nil {
		return nil, "", err
	}

	book, err := excelize.OpenFile(validatedPath)
	if err != nil {
		return nil, "", fmt.Errorf("读取导入文件失败: %w", err)
	}

	sheetName := strings.TrimSpace(book.GetSheetName(0))
	if sheetName == "" {
		_ = book.Close()
		return nil, "", fmt.Errorf("导入文件缺少工作表")
	}

	return book, sheetName, nil
}

func (s *SysCustomerService) countCustomerImportRows(filePath string, startRow int) (int, error) {
	book, sheetName, err := openCustomerImportWorkbook(filePath)
	if err != nil {
		return 0, err
	}
	defer func() { _ = book.Close() }()

	rows, err := book.Rows(sheetName)
	if err != nil {
		return 0, fmt.Errorf("读取工作表数据失败: %w", err)
	}
	defer func() { _ = rows.Close() }()

	rowIndex := 0
	headerMap := map[string]int{}
	total := 0
	for rows.Next() {
		rowIndex++
		columns, columnErr := rows.Columns()
		if columnErr != nil {
			return 0, fmt.Errorf("读取第 %d 行失败: %w", rowIndex, columnErr)
		}
		if rowIndex == 1 {
			headerMap, err = buildCustomerImportHeaderMap(columns)
			if err != nil {
				return 0, err
			}
			continue
		}
		if rowIndex < startRow {
			continue
		}
		if isCustomerImportRowEmpty(columns, headerMap) {
			continue
		}
		total++
	}

	return total, nil
}

func isCustomerImportRowEmpty(row []string, headerMap map[string]int) bool {
	relevantHeaders := []string{"客户姓名", "手机号", "年龄", "星级", "需求金额", "客户备注", "性别"}
	for _, header := range relevantHeaders {
		index, ok := headerMap[header]
		if !ok || index >= len(row) {
			continue
		}
		if strings.TrimSpace(row[index]) != "" {
			return false
		}
	}
	return true
}

func buildCustomerImportHeaderMap(headerRow []string) (map[string]int, error) {
	requiredHeaders := []string{"客户姓名", "手机号", "年龄", "星级", "需求金额", "客户备注", "性别"}
	headerMap := make(map[string]int, len(headerRow))
	for index, raw := range headerRow {
		name := strings.TrimSpace(raw)
		if name == "" {
			continue
		}
		headerMap[name] = index
	}

	for _, header := range requiredHeaders {
		if _, ok := headerMap[header]; !ok {
			return nil, fmt.Errorf("导入模板缺少字段：%s", header)
		}
	}
	return headerMap, nil
}

func parseCustomerImportRow(rowIndex int, row []string, headerMap map[string]int) (customerImportRow, bool, error) {
	getCell := func(header string) string {
		index, ok := headerMap[header]
		if !ok || index >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[index])
	}

	name := strings.TrimSpace(getCell("客户姓名"))
	mobile := strings.TrimSpace(getCell("手机号"))
	ageText := strings.TrimSpace(getCell("年龄"))
	starText := strings.TrimSpace(getCell("星级"))
	moneyText := strings.TrimSpace(getCell("需求金额"))
	remarks := strings.TrimSpace(getCell("客户备注"))
	sexText := strings.TrimSpace(getCell("性别"))

	if name == "" && mobile == "" && ageText == "" && starText == "" && moneyText == "" && remarks == "" && sexText == "" {
		return customerImportRow{}, true, nil
	}

	if name == "" {
		name = "未命名客户"
	}
	if mobile == "" {
		return customerImportRow{}, false, fmt.Errorf("第 %d 行手机号不能为空", rowIndex)
	}

	age, err := parseCustomerImportInteger(ageText, 0, "年龄", rowIndex)
	if err != nil {
		return customerImportRow{}, false, err
	}
	star, err := parseCustomerImportInteger(starText, 0, "星级", rowIndex)
	if err != nil {
		return customerImportRow{}, false, err
	}
	moneyDemand, err := parseCustomerImportInteger(moneyText, 0, "需求金额", rowIndex)
	if err != nil {
		return customerImportRow{}, false, err
	}
	sex, err := parseCustomerImportSex(sexText, rowIndex)
	if err != nil {
		return customerImportRow{}, false, err
	}

	return customerImportRow{
		RowIndex:     rowIndex,
		Name:         name,
		Mobile:       mobile,
		Age:          age,
		CustomerStar: star,
		MoneyDemand:  moneyDemand,
		Remarks:      remarks,
		Sex:          sex,
	}, false, nil
}

func parseCustomerImportInteger(raw string, defaultValue int, field string, rowIndex int) (int, error) {
	value := strings.TrimSpace(raw)
	if value == "" {
		return defaultValue, nil
	}

	if strings.Contains(value, ".") {
		floatValue, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return 0, fmt.Errorf("第 %d 行%s格式不正确", rowIndex, field)
		}
		return int(floatValue), nil
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("第 %d 行%s格式不正确", rowIndex, field)
	}
	return intValue, nil
}

func parseCustomerImportSex(raw string, rowIndex int) (int, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "", "未知", "unknown", "-":
		return 2, nil
	case "男", "1", "male", "m":
		return 1, nil
	case "女", "0", "female", "f":
		return 0, nil
	default:
		return 0, fmt.Errorf("第 %d 行性别格式不正确，仅支持男/女/空值", rowIndex)
	}
}

func (s *SysCustomerService) buildImportedCustomerExtra(row customerImportRow) (string, error) {
	extraPayload := map[string]interface{}{}
	_ = row

	// The current import template has no qualification columns yet.
	// Keep a valid empty JSON payload and extend this map when extra fields are introduced.
	extraBytes, err := json.Marshal(extraPayload)
	if err != nil {
		return "", err
	}

	return string(extraBytes), nil
}

func (s *SysCustomerService) buildImportedCustomer(
	row customerImportRow,
	batchID int,
	scene string,
	channelCtx *customerImportChannelContext,
	deptID int,
	tenantID int,
	userID int,
	city string,
	now time.Time,
) (*models.SysCustomer, error) {
	mobile := s.normalizeMobile(row.Mobile)
	if !models.IsValidImportCustomerMobile(mobile) {
		return nil, fmt.Errorf("手机号格式不正确")
	}

	num, err := snowflakehelper.GenerateID()
	if err != nil {
		return nil, err
	}

	extra, err := s.buildImportedCustomerExtra(row)
	if err != nil {
		return nil, err
	}

	customerStar := row.CustomerStar
	customer := &models.SysCustomer{
		Num:         num,
		Name:        strings.TrimSpace(row.Name),
		Mobile:      mobile,
		MdMobile:    fmt.Sprintf("%x", md5.Sum([]byte(mobile))),
		MoneyDemand: row.MoneyDemand,
		// sys_customer.channel_id 历史上保存的是“渠道公司 ID”，
		// 这里继续沿用该口径，保证列表筛选和历史数据兼容。
		ChannelID:    channelCtx.ChannelCompanyID,
		UserID:       userID,
		CustomerStar: &customerStar,
		Status:       1,
		Intention:    0,
		From:         4,
		Sex:          row.Sex,
		AllotTime:    &now,
		DeptID:       deptID,
		TenantID:     tenantID,
		Remarks:      strings.TrimSpace(row.Remarks),
		Age:          row.Age,
		City:         strings.TrimSpace(city),
		Source:       channelCtx.ChannelKey,
		BatchId:      batchID,
		Extra:        extra,
	}

	switch strings.ToLower(strings.TrimSpace(scene)) {
	case models.CustomerListScenePublic:
		customer.IsPublic = 1
	case models.CustomerListSceneReassign:
		customer.NewData = 1
		customer.IsReassign = 1
		customer.RedistributionTime = &now
	}

	return customer, nil
}

func (s *SysCustomerService) flushImportedCustomers(ctx context.Context, pending []customerImportPendingCustomer) error {
	if len(pending) == 0 {
		return nil
	}

	customers := make([]models.SysCustomer, 0, len(pending))
	for _, item := range pending {
		customers = append(customers, item.Customer)
	}

	return app.DB().WithContext(ctx).CreateInBatches(&customers, customerImportInsertBatchSize).Error
}

func (s *SysCustomerService) updateCustomerImportBatchProgress(
	ctx context.Context,
	batch *models.SysCustomerImportBatch,
	batchID int,
	totalCount int,
	processedCount int,
	successCount int,
	failedCount int,
	duplicateCount int,
	resumeRow int,
	failures []models.SysCustomerImportFailure,
) error {
	updates := map[string]interface{}{
		"total_count":     totalCount,
		"processed_count": processedCount,
		"success_count":   successCount,
		"failed_count":    failedCount,
		"duplicate_count": duplicateCount,
		"progress":        calculateCustomerImportProgress(totalCount, processedCount),
		"resume_row":      resumeRow,
		"failure_preview": encodeCustomerImportFailurePreview(failures, duplicateCount),
		"updated_at":      time.Now(),
	}
	if err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batchID).
		Updates(updates).Error; err != nil {
		return err
	}

	s.pushCustomerImportProgressNotice(
		ctx,
		batch,
		totalCount,
		processedCount,
		successCount,
		failedCount,
		duplicateCount,
		updates["progress"].(int),
		resumeRow,
	)
	return nil
}

func (s *SysCustomerService) finalizeCustomerImportBatch(
	ctx context.Context,
	batch *models.SysCustomerImportBatch,
	batchID int,
	status string,
	totalCount int,
	processedCount int,
	successCount int,
	failedCount int,
	duplicateCount int,
	resumeRow int,
	errorMessage string,
	failures []models.SysCustomerImportFailure,
) error {
	progress := calculateCustomerImportProgress(totalCount, processedCount)
	if status == models.CustomerImportBatchStatusSuccess {
		progress = 100
		resumeRow = 0
	}
	if processedCount >= totalCount && status != models.CustomerImportBatchStatusRunning {
		resumeRow = 0
	}

	updates := map[string]interface{}{
		"status":          status,
		"total_count":     totalCount,
		"processed_count": processedCount,
		"success_count":   successCount,
		"failed_count":    failedCount,
		"duplicate_count": duplicateCount,
		"progress":        progress,
		"resume_row":      resumeRow,
		"error_message":   errorMessage,
		"failure_preview": encodeCustomerImportFailurePreview(failures, duplicateCount),
		"finished_at":     time.Now(),
		"updated_at":      time.Now(),
	}

	if err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batchID).
		Updates(updates).Error; err != nil {
		return err
	}

	s.pushCustomerImportResultNotice(
		ctx,
		batch,
		status,
		totalCount,
		processedCount,
		successCount,
		failedCount,
		duplicateCount,
		progress,
		resumeRow,
		errorMessage,
	)
	return nil
}

func (s *SysCustomerService) executeQueuedCustomerImportBatch(ctx context.Context, batchID int) error {
	var batch models.SysCustomerImportBatch
	if err := app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batchID).
		First(&batch).Error; err != nil {
		return err
	}

	if batch.ID == 0 {
		return nil
	}

	totalCount, err := s.countCustomerImportRows(batch.FilePath, normalizeCustomerImportStartRow(batch.StartRow))
	if err != nil {
		finalErr := truncateCustomerImportError(err)
		if updateErr := s.finalizeCustomerImportBatch(
			ctx,
			&batch,
			batch.ID,
			models.CustomerImportBatchStatusFailed,
			0,
			0,
			0,
			0,
			0,
			normalizeCustomerImportStartRow(batch.StartRow),
			finalErr,
			nil,
		); updateErr != nil {
			return updateErr
		}
		return err
	}

	if totalCount == 0 {
		err = fmt.Errorf("从第 %d 行开始没有可导入的数据", normalizeCustomerImportStartRow(batch.StartRow))
		finalErr := truncateCustomerImportError(err)
		if updateErr := s.finalizeCustomerImportBatch(
			ctx,
			&batch,
			batch.ID,
			models.CustomerImportBatchStatusFailed,
			0,
			0,
			0,
			0,
			0,
			0,
			finalErr,
			nil,
		); updateErr != nil {
			return updateErr
		}
		return err
	}

	if err = app.DB().WithContext(ctx).
		Model(&models.SysCustomerImportBatch{}).
		Where("id = ?", batch.ID).
		Updates(map[string]interface{}{
			"total_count": totalCount,
			"updated_at":  time.Now(),
		}).Error; err != nil {
		return err
	}

	city, err := s.getTenantCityByIDForImport(ctx, uint(batch.TenantID))
	if err != nil {
		finalErr := truncateCustomerImportError(err)
		if updateErr := s.finalizeCustomerImportBatch(
			ctx,
			&batch,
			batch.ID,
			models.CustomerImportBatchStatusFailed,
			totalCount,
			0,
			0,
			0,
			0,
			batch.StartRow,
			finalErr,
			nil,
		); updateErr != nil {
			return updateErr
		}
		return err
	}

	book, sheetName, err := openCustomerImportWorkbook(batch.FilePath)
	if err != nil {
		finalErr := truncateCustomerImportError(err)
		if updateErr := s.finalizeCustomerImportBatch(
			ctx,
			&batch,
			batch.ID,
			models.CustomerImportBatchStatusFailed,
			totalCount,
			0,
			0,
			0,
			0,
			batch.StartRow,
			finalErr,
			nil,
		); updateErr != nil {
			return updateErr
		}
		return err
	}
	defer func() { _ = book.Close() }()

	rows, err := book.Rows(sheetName)
	if err != nil {
		finalErr := truncateCustomerImportError(err)
		if updateErr := s.finalizeCustomerImportBatch(
			ctx,
			&batch,
			batch.ID,
			models.CustomerImportBatchStatusFailed,
			totalCount,
			0,
			0,
			0,
			0,
			batch.StartRow,
			finalErr,
			nil,
		); updateErr != nil {
			return updateErr
		}
		return err
	}
	defer func() { _ = rows.Close() }()

	headerMap := map[string]int{}
	rowIndex := 0
	processedCount := 0
	successCount := 0
	failedCount := 0
	duplicateCount := 0
	resumeRow := batch.StartRow
	failures := make([]models.SysCustomerImportFailure, 0, customerImportFailurePreviewLimit)
	pending := make([]customerImportPendingCustomer, 0, customerImportInsertBatchSize)
	seenMobiles := make(map[string]struct{}, customerImportInsertBatchSize)
	now := time.Now()
	lastProgressFlushAt := time.Now()
	lastCancelCheckAt := time.Time{}
	var interruptErr error

	channelCtx := &customerImportChannelContext{
		ChannelCompanyID: batch.ChannelCompanyID,
		ChannelID:        batch.ChannelID,
		ChannelKey:       batch.ChannelKey,
		ChannelName:      batch.ChannelName,
	}

	flushProgress := func(force bool) error {
		if !force &&
			time.Since(lastProgressFlushAt) < time.Second &&
			processedCount%customerImportProgressFlushStep != 0 {
			return nil
		}
		lastProgressFlushAt = time.Now()
		return s.updateCustomerImportBatchProgress(
			ctx,
			&batch,
			batch.ID,
			totalCount,
			processedCount,
			successCount,
			failedCount,
			duplicateCount,
			resumeRow,
			failures,
		)
	}

	checkCancelRequested := func(force bool) (bool, error) {
		if !force && !lastCancelCheckAt.IsZero() && time.Since(lastCancelCheckAt) < 500*time.Millisecond {
			return false, nil
		}
		lastCancelCheckAt = time.Now()
		return s.isCustomerImportBatchCanceling(ctx, batch.ID)
	}

	for rows.Next() {
		rowIndex++
		columns, columnErr := rows.Columns()
		if columnErr != nil {
			interruptErr = fmt.Errorf("读取第 %d 行失败: %w", rowIndex, columnErr)
			resumeRow = maxCustomerImportRow(resumeRow, rowIndex)
			break
		}

		if rowIndex == 1 {
			headerMap, err = buildCustomerImportHeaderMap(columns)
			if err != nil {
				interruptErr = err
				resumeRow = batch.StartRow
				break
			}
			continue
		}

		if rowIndex < batch.StartRow {
			continue
		}

		cancelRequested, cancelErr := checkCancelRequested(false)
		if cancelErr != nil {
			return cancelErr
		}
		if cancelRequested {
			interruptErr = errCustomerImportCanceled
			if len(pending) > 0 {
				resumeRow = pending[0].RowIndex
			} else {
				resumeRow = maxCustomerImportRow(batch.StartRow, rowIndex)
			}
			break
		}

		if isCustomerImportRowEmpty(columns, headerMap) {
			continue
		}

		rowData, skip, rowErr := parseCustomerImportRow(rowIndex, columns, headerMap)
		if rowErr != nil {
			failedCount++
			processedCount++
			resumeRow = rowIndex + 1
			failures = appendCustomerImportFailurePreview(failures, customerImportFailure{
				Row:     rowIndex,
				Name:    "",
				Mobile:  "",
				Message: rowErr.Error(),
			})
			if err = flushProgress(false); err != nil {
				return err
			}
			continue
		}
		if skip {
			continue
		}

		customer, buildErr := s.buildImportedCustomer(
			rowData,
			batch.ID,
			batch.Scene,
			channelCtx,
			batch.DeptID,
			batch.TenantID,
			batch.UserID,
			city,
			now,
		)
		if buildErr != nil {
			failedCount++
			processedCount++
			resumeRow = rowIndex + 1
			failures = appendCustomerImportFailurePreview(failures, customerImportFailure{
				Row:     rowIndex,
				Name:    rowData.Name,
				Mobile:  rowData.Mobile,
				Message: buildErr.Error(),
			})
			if err = flushProgress(false); err != nil {
				return err
			}
			continue
		}

		if _, exists := seenMobiles[customer.Mobile]; exists {
			duplicateCount++
			processedCount++
			resumeRow = rowIndex + 1
			if err = flushProgress(false); err != nil {
				return err
			}
			continue
		}

		seenMobiles[customer.Mobile] = struct{}{}

		pending = append(pending, customerImportPendingCustomer{
			RowIndex: rowIndex,
			Customer: *customer,
		})

		if len(pending) < customerImportInsertBatchSize {
			continue
		}

		cancelRequested, cancelErr = checkCancelRequested(true)
		if cancelErr != nil {
			return cancelErr
		}
		if cancelRequested {
			interruptErr = errCustomerImportCanceled
			resumeRow = pending[0].RowIndex
			break
		}

		if err = s.flushImportedCustomers(ctx, pending); err != nil {
			interruptErr = fmt.Errorf("从第 %d 行开始的批次写入失败: %w", pending[0].RowIndex, err)
			resumeRow = pending[0].RowIndex
			break
		}

		successCount += len(pending)
		processedCount += len(pending)
		resumeRow = pending[len(pending)-1].RowIndex + 1
		pending = pending[:0]

		if err = flushProgress(true); err != nil {
			return err
		}
	}

	if interruptErr == nil && len(pending) > 0 {
		cancelRequested, cancelErr := checkCancelRequested(true)
		if cancelErr != nil {
			return cancelErr
		}
		if cancelRequested {
			interruptErr = errCustomerImportCanceled
			resumeRow = pending[0].RowIndex
		} else if err = s.flushImportedCustomers(ctx, pending); err != nil {
			interruptErr = fmt.Errorf("从第 %d 行开始的批次写入失败: %w", pending[0].RowIndex, err)
			resumeRow = pending[0].RowIndex
		} else {
			successCount += len(pending)
			processedCount += len(pending)
			resumeRow = pending[len(pending)-1].RowIndex + 1
		}
	}

	if interruptErr == nil {
		cancelRequested, cancelErr := checkCancelRequested(true)
		if cancelErr != nil {
			return cancelErr
		}
		if cancelRequested && processedCount < totalCount {
			interruptErr = errCustomerImportCanceled
			if len(pending) > 0 {
				resumeRow = pending[0].RowIndex
			} else if resumeRow <= 0 {
				resumeRow = batch.StartRow
			}
		}
	}

	deletedCount := int64(0)
	var cancelCleanupErr error
	if isCustomerImportCanceledError(interruptErr) {
		deletedCount, cancelCleanupErr = s.deleteImportedCustomersByBatchID(ctx, batch.ID)
		resumeRow = 0
	}

	status := models.CustomerImportBatchStatusSuccess
	if interruptErr != nil {
		if isCustomerImportCanceledError(interruptErr) && cancelCleanupErr == nil {
			status = models.CustomerImportBatchStatusCanceled
		} else if processedCount > 0 {
			status = models.CustomerImportBatchStatusPartial
		} else {
			status = models.CustomerImportBatchStatusFailed
		}
	} else if successCount == 0 && failedCount > 0 {
		status = models.CustomerImportBatchStatusFailed
	} else if failedCount > 0 {
		status = models.CustomerImportBatchStatusPartial
	}

	errorMessage := ""
	if cancelCleanupErr != nil {
		errorMessage = fmt.Sprintf("导入任务已终止，但按 batch_id 彻底删除已导入客户失败：%s", strings.TrimSpace(cancelCleanupErr.Error()))
	} else if isCustomerImportCanceledError(interruptErr) {
		errorMessage = buildCustomerImportCanceledMessage(successCount, failedCount, duplicateCount, deletedCount)
	} else {
		errorMessage = buildCustomerImportSummaryMessage(
			successCount,
			failedCount,
			duplicateCount,
			failures,
			interruptErr != nil,
			resumeRow,
			interruptErr,
		)
	}

	if err = s.finalizeCustomerImportBatch(
		ctx,
		&batch,
		batch.ID,
		status,
		totalCount,
		processedCount,
		successCount,
		failedCount,
		duplicateCount,
		resumeRow,
		truncateCustomerImportError(fmt.Errorf("%s", strings.TrimSpace(errorMessage))),
		failures,
	); err != nil {
		return err
	}

	s.tryCleanupExpiredCustomerImportFiles(ctx, uint(batch.TenantID), uint(batch.UserID), batch.ID)
	return interruptErr
}

func maxCustomerImportRow(current, next int) int {
	if next > current {
		return next
	}
	return current
}

func getCustomerImportCleanupDuration() time.Duration {
	days := app.ConfigYml.GetInt("platform.export_clean_days")
	if days <= 1 {
		days = 1
	}
	return time.Duration(days) * 24 * time.Hour
}

func (s *SysCustomerService) cleanupExpiredCustomerImportFiles(ctx context.Context, tenantID uint, excludeBatchID int) error {
	if tenantID == 0 {
		return nil
	}

	cutoff := time.Now().Add(-getCustomerImportCleanupDuration())
	var batches []models.SysCustomerImportBatch
	tx := app.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("file_deleted_at IS NULL").
		Where("(file_path <> '' OR file_url <> '')").
		Where("status IN ?", []string{
			models.CustomerImportBatchStatusSuccess,
			models.CustomerImportBatchStatusPartial,
			models.CustomerImportBatchStatusFailed,
		}).
		Where("finished_at IS NOT NULL").
		Where("finished_at < ?", cutoff)
	if excludeBatchID > 0 {
		tx = tx.Where("id <> ?", excludeBatchID)
	}
	if err := tx.Find(&batches).Error; err != nil {
		return err
	}

	for _, batch := range batches {
		deleteTarget := strings.TrimSpace(batch.FilePath)
		if deleteTarget == "" {
			deleteTarget = strings.TrimSpace(batch.FileURL)
		}
		if deleteTarget != "" {
			deleteErr := error(nil)
			if _, statErr := os.Stat(deleteTarget); statErr == nil {
				deleteErr = os.Remove(deleteTarget)
			} else {
				deleteErr = app.UploadService.DeleteFile(deleteTarget)
			}
			if deleteErr != nil {
				app.ZapLog.Warn("delete expired customer import file failed",
					zap.Int("batch_id", batch.ID),
					zap.Uint("tenant_id", tenantID),
					zap.Error(deleteErr),
				)
			}
		}

		if batch.FileAffixID > 0 {
			if err := s.deleteCustomerImportAffix(ctx, batch.FileAffixID); err != nil {
				app.ZapLog.Warn("delete expired customer import affix failed",
					zap.Int("batch_id", batch.ID),
					zap.Uint("tenant_id", tenantID),
					zap.Uint("affix_id", batch.FileAffixID),
					zap.Error(err),
				)
			}
		}

		now := time.Now()
		if err := app.DB().WithContext(ctx).
			Model(&models.SysCustomerImportBatch{}).
			Where("id = ?", batch.ID).
			Updates(map[string]interface{}{
				"file_affix_id":   0,
				"file_url":        "",
				"file_path":       "",
				"file_deleted_at": now,
			}).Error; err != nil {
			app.ZapLog.Warn("clear expired customer import batch file fields failed",
				zap.Int("batch_id", batch.ID),
				zap.Uint("tenant_id", tenantID),
				zap.Error(err),
			)
		}
	}

	return nil
}

func (s *SysCustomerService) tryCleanupExpiredCustomerImportFiles(ctx context.Context, tenantID uint, userID uint, excludeBatchID int) {
	if err := s.cleanupExpiredCustomerImportFiles(ctx, tenantID, excludeBatchID); err != nil {
		app.ZapLog.Warn("cleanup expired customer import files failed",
			zap.Uint("tenant_id", tenantID),
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
	}
}
