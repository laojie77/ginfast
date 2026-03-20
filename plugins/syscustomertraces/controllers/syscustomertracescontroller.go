package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/syscustomertraces/models"
	"gin-fast/plugins/syscustomertraces/service"

	"github.com/gin-gonic/gin"
)

// SysCustomerTracesController sys_customer_traces控制器
type SysCustomerTracesController struct {
	controllers.Common
	SysCustomerTracesService *service.SysCustomerTracesService
}

// NewSysCustomerTracesController 创建sys_customer_traces控制器
func NewSysCustomerTracesController() *SysCustomerTracesController {
	return &SysCustomerTracesController{
		SysCustomerTracesService: service.NewSysCustomerTracesService(),
	}
}

// Create 创建sys_customer_traces
func (c *SysCustomerTracesController) Create(ctx *gin.Context) {
	var req models.SysCustomerTracesCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	// 如果UserId为0，则使用当前登录用户ID
	if req.UserId == 0 {
		req.UserId = int(common.GetCurrentUserID(ctx))
	}

	sysCustomerTraces, err := c.SysCustomerTracesService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建sys_customer_traces失败", err)
	}

	c.Success(ctx, gin.H{
		"id": sysCustomerTraces.Id,
	})
}

// Update 更新sys_customer_traces
func (c *SysCustomerTracesController) Update(ctx *gin.Context) {
	var req models.SysCustomerTracesUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysCustomerTracesService.Update(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "更新sys_customer_traces失败", err)
	}

	c.SuccessWithMessage(ctx, "更新成功")
}

// Delete 删除sys_customer_traces
func (c *SysCustomerTracesController) Delete(ctx *gin.Context) {
	var req models.SysCustomerTracesDeleteRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysCustomerTracesService.Delete(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "删除sys_customer_traces失败", err)
	}

	c.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetByID 根据ID获取sys_customer_traces信息
func (c *SysCustomerTracesController) GetByID(ctx *gin.Context) {
	var req models.SysCustomerTracesGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCustomerTraces, err := c.SysCustomerTracesService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "sys_customer_traces不存在", err)
	}

	c.Success(ctx, sysCustomerTraces)
}

// List sys_customer_traces列表（分页查询）
func (c *SysCustomerTracesController) List(ctx *gin.Context) {
	var req models.SysCustomerTracesListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCustomerTracesList, total, err := c.SysCustomerTracesService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取sys_customer_traces列表失败", err)
	}

	c.Success(ctx, gin.H{
		"list":  sysCustomerTracesList,
		"total": total,
	})
}
