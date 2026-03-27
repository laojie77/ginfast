package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/syscallrecord/models"
	"gin-fast/plugins/syscallrecord/service"

	"github.com/gin-gonic/gin"
)

// SysCallRecordController sys_call_record控制器
type SysCallRecordController struct {
	controllers.Common
	SysCallRecordService *service.SysCallRecordService
}

// NewSysCallRecordController 创建sys_call_record控制器
func NewSysCallRecordController() *SysCallRecordController {
	return &SysCallRecordController{
		SysCallRecordService: service.NewSysCallRecordService(),
	}
}

// Create 创建sys_call_record
func (c *SysCallRecordController) Create(ctx *gin.Context) {
	var req models.SysCallRecordCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}
	
	sysCallRecord, err := c.SysCallRecordService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建sys_call_record失败", err)
	}

	c.Success(ctx, gin.H{
		"id": sysCallRecord.Id,
	})
}

// Update 更新sys_call_record
func (c *SysCallRecordController) Update(ctx *gin.Context) {
	var req models.SysCallRecordUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysCallRecordService.Update(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "更新sys_call_record失败", err)
	}

	c.SuccessWithMessage(ctx, "更新成功")
}

// Delete 删除sys_call_record
func (c *SysCallRecordController) Delete(ctx *gin.Context) {
	var req models.SysCallRecordDeleteRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysCallRecordService.Delete(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "删除sys_call_record失败", err)
	}

	c.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetByID 根据ID获取sys_call_record信息
func (c *SysCallRecordController) GetByID(ctx *gin.Context) {
	var req models.SysCallRecordGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCallRecord, err := c.SysCallRecordService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "sys_call_record不存在", err)
	}

	c.Success(ctx, sysCallRecord)
}

// List sys_call_record列表（分页查询）
func (c *SysCallRecordController) List(ctx *gin.Context) {
	var req models.SysCallRecordListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCallRecordList, total, err := c.SysCallRecordService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取sys_call_record列表失败", err)
	}

	c.Success(ctx, gin.H{
		"list":  sysCallRecordList,
		"total": total,
	})
}