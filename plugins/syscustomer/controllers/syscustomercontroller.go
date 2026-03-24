package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/syscustomer/models"
	"gin-fast/plugins/syscustomer/service"

	"github.com/gin-gonic/gin"
)

// SysCustomerController sys_customer控制器
type SysCustomerController struct {
	controllers.Common
	SysCustomerService *service.SysCustomerService
}

// NewSysCustomerController 创建sys_customer控制器
func NewSysCustomerController() *SysCustomerController {
	return &SysCustomerController{
		SysCustomerService: service.NewSysCustomerService(),
	}
}

// Create 创建sys_customer
func (c *SysCustomerController) Create(ctx *gin.Context) {
	var req models.SysCustomerCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCustomer, err := c.SysCustomerService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建sys_customer失败", err)
	}

	c.Success(ctx, gin.H{
		"id": sysCustomer.Id,
	})
}

// Update 更新sys_customer
func (c *SysCustomerController) Update(ctx *gin.Context) {
	var req models.SysCustomerUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysCustomerService.Update(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "更新sys_customer失败", err)
	}

	c.SuccessWithMessage(ctx, "更新成功")
}

// Delete 删除sys_customer
func (c *SysCustomerController) Delete(ctx *gin.Context) {
	var req models.SysCustomerDeleteRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysCustomerService.Delete(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "删除sys_customer失败", err)
	}

	c.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetByID 根据ID获取sys_customer信息
func (c *SysCustomerController) GetByID(ctx *gin.Context) {
	var req models.SysCustomerGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCustomer, err := c.SysCustomerService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "sys_customer不存在", err)
	}

	c.Success(ctx, sysCustomer)
}

// List sys_customer列表（分页查询）
func (c *SysCustomerController) List(ctx *gin.Context) {
	var req models.SysCustomerListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysCustomerList, total, err := c.SysCustomerService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取sys_customer列表失败", err)
	}

	c.Success(ctx, gin.H{
		"list":  sysCustomerList,
		"total": total,
	})
}

// Update 更新sys_customer
func (c *SysCustomerController) UpdateCustomerStatusTrace(ctx *gin.Context) {
	var req models.CustomerStatusTracesUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}
	req.UserID = int(common.GetCurrentUserID(ctx))
	err := c.SysCustomerService.CustomerStatusTracesUpdate(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "更新客户状态失败", err)
	}

	c.SuccessWithMessage(ctx, "更新成功")
}
