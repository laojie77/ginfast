package controllers

import (
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

// Create creates a sys_customer record.
func (c *SysCustomerController) Create(ctx *gin.Context) {
	var req models.SysCustomerCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	sysCustomer, err := c.SysCustomerService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建sys_customer失败", err)
		return
	}

	c.Success(ctx, gin.H{
		"id": sysCustomer.Id,
	})
}

// Update updates a sys_customer record.
func (c *SysCustomerController) Update(ctx *gin.Context) {
	var req models.SysCustomerUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	if err := c.SysCustomerService.Update(ctx, req); err != nil {
		c.FailAndAbort(ctx, "更新sys_customer失败", err)
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
		c.FailAndAbort(ctx, "删除sys_customer失败", err)
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
		c.FailAndAbort(ctx, "sys_customer不存在", err)
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
			c.FailAndAbort(ctx, "提交sys_customer导出任务失败", err)
		} else {
			c.Success(ctx, result)
		}
		return
	}
	if exportFlag == "1" || exportFlag == "true" {
		if err := c.SysCustomerService.ExportCSV(ctx, req); err != nil {
			c.FailAndAbort(ctx, "导出sys_customer失败", err)
		}
		return
	}

	sysCustomerList, total, err := c.SysCustomerService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取sys_customer列表失败", err)
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
