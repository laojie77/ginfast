package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/syscustomerexporttasks/models"
	"gin-fast/plugins/syscustomerexporttasks/service"

	"github.com/gin-gonic/gin"
)

// SysCustomerExportTasksController handles customer export task queries.
type SysCustomerExportTasksController struct {
	controllers.Common
	SysCustomerExportTasksService *service.SysCustomerExportTasksService
}

// NewSysCustomerExportTasksController creates a task query controller.
func NewSysCustomerExportTasksController() *SysCustomerExportTasksController {
	return &SysCustomerExportTasksController{
		SysCustomerExportTasksService: service.NewSysCustomerExportTasksService(),
	}
}

// GetByID returns a task detail record.
func (c *SysCustomerExportTasksController) GetByID(ctx *gin.Context) {
	var req models.SysCustomerExportTasksGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	sysCustomerExportTask, err := c.SysCustomerExportTasksService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "获取导出任务详情失败", err)
		return
	}

	c.Success(ctx, sysCustomerExportTask)
}

// List returns the customer export task list.
func (c *SysCustomerExportTasksController) List(ctx *gin.Context) {
	var req models.SysCustomerExportTasksListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
		return
	}

	sysCustomerExportTaskList, total, err := c.SysCustomerExportTasksService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取导出任务列表失败", err)
		return
	}

	c.Success(ctx, gin.H{
		"list":  sysCustomerExportTaskList,
		"total": total,
	})
}
