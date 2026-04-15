package controllers

import (
	"mime/multipart"
	"strconv"
	"strings"

	"gin-fast/app/controllers"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/syscustomer/models"
	"gin-fast/plugins/syscustomer/service"

	"github.com/gin-gonic/gin"
)

// SysCustomerController sys_customer controller.
type SysCustomerController struct {
	controllers.Common
	SysCustomerService *service.SysCustomerService
}

// NewSysCustomerController creates a sys_customer controller.
func NewSysCustomerController() *SysCustomerController {
	return &SysCustomerController{
		SysCustomerService: service.NewSysCustomerService(),
	}
}

func getCustomerImportFile(ctx *gin.Context) (*multipart.FileHeader, error) {
	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		return nil, err
	}
	return fileHeader, nil
}

func resolveCustomerImportSubmitErrorMessage(err error) string {
	if err == nil {
		return "导入客户失败"
	}
	if service.IsCustomerImportPublicError(err) {
		return strings.TrimSpace(err.Error())
	}
	return "导入客户失败"
}

// Create creates a sys_customer record.
func (c *SysCustomerController) Create(ctx *gin.Context) {
	fileHeader, fileErr := getCustomerImportFile(ctx)
	if fileErr == nil && fileHeader != nil {
		var importReq models.SysCustomerImportRequest
		if err := importReq.Validate(ctx); err != nil {
			c.FailAndAbort(ctx, err.Error(), err)
			return
		}

		result, err := c.SysCustomerService.ImportCustomers(ctx, importReq, fileHeader)
		if err != nil {
			c.FailAndAbort(ctx, resolveCustomerImportSubmitErrorMessage(err), err)
			return
		}

		c.SuccessWithMessage(ctx, "导入任务已提交", result)
		return
	}

	var req models.SysCustomerCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	sysCustomer, err := c.SysCustomerService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建客户失败", err)
		return
	}

	c.Success(ctx, gin.H{
		"id": sysCustomer.Id,
	})
}

func (c *SysCustomerController) Import(ctx *gin.Context) {
	contentType := strings.ToLower(strings.TrimSpace(ctx.GetHeader("Content-Type")))
	if strings.Contains(contentType, "application/json") {
		var cancelReq models.SysCustomerImportCancelRequest
		if err := cancelReq.Validate(ctx); err != nil {
			c.FailAndAbort(ctx, err.Error(), err)
			return
		}

		claims := common.GetClaims(ctx)
		if claims == nil {
			c.FailAndAbort(ctx, "登录失效", nil)
			return
		}

		result, err := c.SysCustomerService.CancelImportBatch(ctx, claims.TenantID, claims.UserID, cancelReq.BatchID)
		if err != nil {
			c.FailAndAbort(ctx, "终止导入任务失败", err)
			return
		}
		if result == nil {
			c.FailAndAbort(ctx, "导入任务不存在", nil)
			return
		}

		c.SuccessWithMessage(ctx, "终止导入指令已提交", result)
		return
	}

	var req models.SysCustomerImportRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	fileHeader, err := getCustomerImportFile(ctx)
	if err != nil {
		c.FailAndAbort(ctx, "请上传导入文件", err)
		return
	}

	result, err := c.SysCustomerService.ImportCustomers(ctx, req, fileHeader)
	if err != nil {
		c.FailAndAbort(ctx, resolveCustomerImportSubmitErrorMessage(err), err)
		return
	}

	c.SuccessWithMessage(ctx, "导入任务已提交", result)
}

func (c *SysCustomerController) GetImportBatch(ctx *gin.Context) {
	batchID, err := strconv.Atoi(strings.TrimSpace(ctx.Param("batchId")))
	if err != nil || batchID <= 0 {
		c.FailAndAbort(ctx, "导入批次ID无效", err)
		return
	}

	claims := common.GetClaims(ctx)
	if claims == nil {
		c.FailAndAbort(ctx, "当前登录状态已失效", nil)
		return
	}

	result, err := c.SysCustomerService.GetImportBatch(ctx, claims.TenantID, claims.UserID, batchID)
	if err != nil {
		c.FailAndAbort(ctx, "获取导入任务状态失败", err)
		return
	}
	if result == nil {
		c.FailAndAbort(ctx, "导入任务不存在", nil)
		return
	}

	c.Success(ctx, result)
}

func (c *SysCustomerController) GetLatestImportBatch(ctx *gin.Context) {
	claims := common.GetClaims(ctx)
	if claims == nil {
		c.FailAndAbort(ctx, "当前登录状态已失效", nil)
		return
	}

	result, err := c.SysCustomerService.GetLatestImportBatch(ctx, claims.TenantID, claims.UserID)
	if err != nil {
		c.FailAndAbort(ctx, "获取最新导入任务失败", err)
		return
	}
	if result == nil {
		c.Success(ctx, nil)
		return
	}

	c.Success(ctx, result)
}

// Update updates a sys_customer record.
func (c *SysCustomerController) Update(ctx *gin.Context) {
	var req models.SysCustomerUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	if err := c.SysCustomerService.Update(ctx, req); err != nil {
		c.FailAndAbort(ctx, "更新客户失败", err)
		return
	}

	c.SuccessWithMessage(ctx, "更新成功")
}

// Delete deletes a sys_customer record.
func (c *SysCustomerController) Delete(ctx *gin.Context) {
	var req models.SysCustomerDeleteRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	if err := c.SysCustomerService.Delete(ctx, req.Id); err != nil {
		c.FailAndAbort(ctx, "删除客户失败", err)
		return
	}

	c.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetByID returns a sys_customer record by ID.
func (c *SysCustomerController) GetByID(ctx *gin.Context) {
	var req models.SysCustomerGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	sysCustomer, err := c.SysCustomerService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "客户不存在", err)
		return
	}

	c.Success(ctx, sysCustomer)
}

func (c *SysCustomerController) GetExportTask(ctx *gin.Context) {
	taskID, err := strconv.ParseUint(strings.TrimSpace(ctx.Param("taskId")), 10, 64)
	if err != nil || taskID == 0 {
		c.FailAndAbort(ctx, "导出任务ID无效", err)
		return
	}

	claims := common.GetClaims(ctx)
	if claims == nil {
		c.FailAndAbort(ctx, "当前登录状态已失效", nil)
		return
	}

	task, err := c.SysCustomerService.GetExportTask(ctx, claims.TenantID, claims.UserID, uint(taskID))
	if err != nil {
		c.FailAndAbort(ctx, "获取导出任务状态失败", err)
		return
	}
	if task == nil {
		c.FailAndAbort(ctx, "导出任务不存在", nil)
		return
	}

	c.Success(ctx, models.SysCustomerExportTaskResult{
		ID:           task.ID,
		Status:       task.Status,
		Total:        task.Total,
		Processed:    task.Processed,
		Progress:     task.Progress,
		AffixID:      task.AffixID,
		FileName:     task.FileName,
		ErrorMessage: task.ErrorMessage,
		StartedAt:    task.StartedAt,
		FinishedAt:   task.FinishedAt,
		UpdatedAt:    task.UpdatedAt,
	})
}

// List returns the sys_customer list or streams the export file when requested.
func (c *SysCustomerController) List(ctx *gin.Context) {
	var req models.SysCustomerListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	exportFlag := strings.TrimSpace(strings.ToLower(ctx.Query("export")))
	if exportFlag == "submit" {
		result, err := c.SysCustomerService.SubmitExport(ctx, req)
		if err != nil {
			c.FailAndAbort(ctx, "提交客户导出任务失败", err)
		} else {
			c.Success(ctx, result)
		}
		return
	}
	if exportFlag == "1" || exportFlag == "true" {
		if err := c.SysCustomerService.ExportCSV(ctx, req); err != nil {
			c.FailAndAbort(ctx, "导出客户失败", err)
		}
		return
	}

	sysCustomerList, total, err := c.SysCustomerService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取客户列表失败", err)
		return
	}

	c.Success(ctx, gin.H{
		"list":  sysCustomerList,
		"total": total,
	})
}

// UpdateCustomerStatusTrace updates quick status fields and trace data.
func (c *SysCustomerController) UpdateCustomerStatusTrace(ctx *gin.Context) {
	var req models.CustomerQuickStatusUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	req.UserID = int(common.GetCurrentUserID(ctx))
	if err := c.SysCustomerService.CustomerQuickStatusUpdate(ctx, req); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	c.SuccessWithMessage(ctx, "更新成功")
}
