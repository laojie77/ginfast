package controllers

import (
	"fmt"
	"time"

	"gin-fast/app/controllers"
	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	"gin-fast/app/utils/common"
	customerModels "gin-fast/plugins/syscustomer/models"
	"gin-fast/plugins/syscustomertraces/models"
	"gin-fast/plugins/syscustomertraces/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		return
	}

	currentUserID := common.GetCurrentUserID(ctx)
	requestUserID := req.UserID
	isSameUser := currentUserID > 0 && requestUserID == int(currentUserID)
	now := time.Now()
	sysCustomerTraces := models.NewSysCustomerTraces()

	err := app.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if currentUserID > 0 {
			req.UserID = int(currentUserID)
		}
		sysCustomerTraces.CustomerID = req.CustomerID
		sysCustomerTraces.UserID = req.UserID
		sysCustomerTraces.Data = req.Data

		if err := tx.Create(sysCustomerTraces).Error; err != nil {
			return err
		}

		updates := map[string]interface{}{}
		if isSameUser {
			updates["remark_time"] = now
		} else if currentUserID > 0 {
			userName := "System"
			currentUser := appModels.NewUser()
			if err := currentUser.GetUserByID(ctx, currentUserID); err == nil {
				userName = currentUser.NickName
			}
			updates["customer_comment"] = fmt.Sprintf("%s - %s - %s", now.Format("2006-01-02 15:04:05"), userName, req.Data)
		}

		if len(updates) == 0 {
			return nil
		}

		return tx.Model(&customerModels.SysCustomer{}).
			Where("id = ?", req.CustomerID).
			Updates(updates).Error
	})
	if err != nil {
		c.FailAndAbort(ctx, "创建客户跟进失败", err)
		return
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
