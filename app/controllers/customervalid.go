package controllers

import (
	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"gin-fast/app/utils/tenanthelper"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"strconv"
)

type CustomerValidController struct {
	Common
}

// List 获取客户有效性标签列表
func (c *CustomerValidController) List(context *gin.Context) {
	var req models.CustomerValidListParams
	if err := context.ShouldBindQuery(&req); err != nil {
		c.FailAndAbort(context, err.Error(), err)
		return
	}

	// 设置默认分页参数
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// 创建列表和查询条件
	customerValidList := make([]*models.CustomerValid, 0)
	
	// 构建查询条件
	query := func(db *gorm.DB) *gorm.DB {
		if req.Type != nil {
			db = db.Where("type = ?", *req.Type)
		}
		if req.Name != "" {
			db = db.Where("name LIKE ?", "%"+req.Name+"%")
		}
		if req.Status != nil {
			db = db.Where("status = ?", *req.Status)
		}
		return db
	}

	// 获取总数
	var total int64
	err := app.DB().Model(&models.CustomerValid{}).Scopes(tenanthelper.TenantScope(context), query).Count(&total).Error
	if err != nil {
		c.FailAndAbort(context, "获取客户有效性标签总数失败", err)
		return
	}

	// 获取分页数据
	err = app.DB().Scopes(tenanthelper.TenantScope(context), req.Paginate(), query).Find(&customerValidList).Error
	if err != nil {
		c.FailAndAbort(context, "获取客户有效性标签列表失败", err)
		return
	}

	c.Success(context, gin.H{
		"list":     customerValidList,
		"total":    total,
		"pageNum":  req.PageNum,
		"pageSize": req.PageSize,
	})
}

// Create 创建客户有效性标签
func (c *CustomerValidController) Create(context *gin.Context) {
	var req models.CustomerValidCreateRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		c.FailAndAbort(context, err.Error(), err)
		return
	}

	// 创建数据
	customerValid := models.CustomerValid{
		Type:   req.Type,
		Name:   req.Name,
		Status: req.Status,
	}

	// TenantID 会通过 GORM hook 自动设置
	if err := app.DB().WithContext(context).Create(&customerValid).Error; err != nil {
		c.FailAndAbort(context, "创建客户有效性标签失败", err)
		return
	}

	c.SuccessWithMessage(context, "创建成功", customerValid)
}

// Update 更新客户有效性标签
func (c *CustomerValidController) Update(context *gin.Context) {
	var req models.CustomerValidUpdateRequest
	if err := context.ShouldBindJSON(&req); err != nil {
		c.FailAndAbort(context, err.Error(), err)
		return
	}

	// 查找记录
	var customerValid models.CustomerValid
	db := app.DB().Scopes(tenanthelper.TenantScope(context)).Where("id = ?", req.ID)
	
	if err := db.First(&customerValid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.FailAndAbort(context, "记录不存在", err)
		} else {
			c.FailAndAbort(context, "查询记录失败", err)
		}
		return
	}

	// 更新数据
	customerValid.Type = req.Type
	customerValid.Name = req.Name
	customerValid.Status = req.Status

	if err := app.DB().Save(&customerValid).Error; err != nil {
		c.FailAndAbort(context, "更新客户有效性标签失败", err)
		return
	}

	c.SuccessWithMessage(context, "更新成功", customerValid)
}

// Delete 删除客户有效性标签
func (c *CustomerValidController) Delete(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.FailAndAbort(context, "无效的ID", err)
		return
	}

	// 查找记录
	var customerValid models.CustomerValid
	db := app.DB().Scopes(tenanthelper.TenantScope(context)).Where("id = ?", id)
	
	if err := db.First(&customerValid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.FailAndAbort(context, "记录不存在", err)
		} else {
			c.FailAndAbort(context, "查询记录失败", err)
		}
		return
	}

	// 删除记录
	if err := app.DB().Delete(&customerValid).Error; err != nil {
		c.FailAndAbort(context, "删除客户有效性标签失败", err)
		return
	}

	c.SuccessWithMessage(context, "删除成功", nil)
}

// Detail 获取客户有效性标签详情
func (c *CustomerValidController) Detail(context *gin.Context) {
	idStr := context.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.FailAndAbort(context, "无效的ID", err)
		return
	}

	// 查找记录
	var customerValid models.CustomerValid
	db := app.DB().Scopes(tenanthelper.TenantScope(context)).Where("id = ?", id)
	
	if err := db.First(&customerValid).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.FailAndAbort(context, "记录不存在", err)
		} else {
			c.FailAndAbort(context, "查询记录失败", err)
		}
		return
	}

	c.Success(context, customerValid)
}