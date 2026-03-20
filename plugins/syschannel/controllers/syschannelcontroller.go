package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/syschannel/models"
	"gin-fast/plugins/syschannel/service"

	"github.com/gin-gonic/gin"
)

// SysChannelController sys_channel控制器
type SysChannelController struct {
	controllers.Common
	SysChannelService *service.SysChannelService
}

// NewSysChannelController 创建sys_channel控制器
func NewSysChannelController() *SysChannelController {
	return &SysChannelController{
		SysChannelService: service.NewSysChannelService(),
	}
}

// Create 创建sys_channel
func (c *SysChannelController) Create(ctx *gin.Context) {
	var req models.SysChannelCreateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysChannel, err := c.SysChannelService.Create(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "创建sys_channel失败", err)
	}

	c.Success(ctx, gin.H{
		"id": sysChannel.Id,
	})
}

// Update 更新sys_channel
func (c *SysChannelController) Update(ctx *gin.Context) {
	var req models.SysChannelUpdateRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}
	err := c.SysChannelService.Update(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "更新失败", err)
	}

	c.SuccessWithMessage(ctx, "更新成功")
}

// Delete 删除sys_channel
func (c *SysChannelController) Delete(ctx *gin.Context) {
	var req models.SysChannelDeleteRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	err := c.SysChannelService.Delete(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "删除sys_channel失败", err)
	}

	c.SuccessWithMessage(ctx, "删除成功", nil)
}

// GetByID 根据ID获取sys_channel信息
func (c *SysChannelController) GetByID(ctx *gin.Context) {
	var req models.SysChannelGetByIDRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysChannel, err := c.SysChannelService.GetByID(ctx, req.Id)
	if err != nil {
		c.FailAndAbort(ctx, "sys_channel不存在", err)
	}

	c.Success(ctx, sysChannel)
}

// List sys_channel列表（分页查询）
func (c *SysChannelController) List(ctx *gin.Context) {
	var req models.SysChannelListRequest
	if err := req.Validate(ctx); err != nil {
		c.FailAndAbort(ctx, err.Error(), err)
	}

	sysChannelList, total, err := c.SysChannelService.List(ctx, req)
	if err != nil {
		c.FailAndAbort(ctx, "获取sys_channel列表失败", err)
	}

	c.Success(ctx, gin.H{
		"list":  sysChannelList,
		"total": total,
	})
}
