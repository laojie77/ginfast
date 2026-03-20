package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/syschannelcompany/models"
	"gin-fast/plugins/syschannelcompany/service"

	"github.com/gin-gonic/gin"
)

// SysChannelCompanyController sys_channel_company控制器
type SysChannelCompanyController struct {
	controllers.Common
	SysChannelCompanyService *service.SysChannelCompanyService
}

// NewSysChannelCompanyController 创建sys_channel_company控制器
func NewSysChannelCompanyController() *SysChannelCompanyController {
	return &SysChannelCompanyController{
		SysChannelCompanyService: service.NewSysChannelCompanyService(),
	}
}

// Create 创建sys_channel_company
func (c *SysChannelCompanyController) Create(ctx *gin.Context) {
	var req models.SysChannelCompanyCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}
	
	sysChannelCompany, err := c.SysChannelCompanyService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建sys_channel_company失败", err)
	}

	c.Success(ctx, gin.H{
		"id": sysChannelCompany.Id,
	})
}

// Update 更新sys_channel_company
func (c *SysChannelCompanyController) Update(ctx *gin.Context) {
	var req models.SysChannelCompanyUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysChannelCompanyService.Update(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "更新sys_channel_company失败", err)
	}

	c.SuccessWithMessage(ctx, "更新成功")
}

// Delete 删除sys_channel_company
func (c *SysChannelCompanyController) Delete(ctx *gin.Context) {
	var req models.SysChannelCompanyDeleteRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysChannelCompanyService.Delete(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "删除sys_channel_company失败", err)
	}

	c.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetByID 根据ID获取sys_channel_company信息
func (c *SysChannelCompanyController) GetByID(ctx *gin.Context) {
	var req models.SysChannelCompanyGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysChannelCompany, err := c.SysChannelCompanyService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "sys_channel_company不存在", err)
	}

	c.Success(ctx, sysChannelCompany)
}

// List sys_channel_company列表（分页查询）
func (c *SysChannelCompanyController) List(ctx *gin.Context) {
	var req models.SysChannelCompanyListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysChannelCompanyList, total, err := c.SysChannelCompanyService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取sys_channel_company列表失败", err)
	}

	c.Success(ctx, gin.H{
		"list":  sysChannelCompanyList,
		"total": total,
	})
}