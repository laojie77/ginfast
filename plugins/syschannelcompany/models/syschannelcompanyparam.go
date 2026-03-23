package models

import (
	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysChannelCompanyListRequest sys_channel_company列表请求参数
type SysChannelCompanyListRequest struct {
	models.BasePaging
	models.Validator
	ChannelID  *int    `form:"channelId"`  // 渠道ID
	TenantID   *int    `form:"tenantId"`   // 公司平台
	City       *string `form:"city"`       // 城市
	HiddenName *string `form:"hiddenName"` // 渠道名称
}

// Validate 验证请求参数
func (r *SysChannelCompanyListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle 获取查询条件
func (r *SysChannelCompanyListRequest) Handle() func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ChannelID != nil {
			db = db.Where("channel_id = ?", *r.ChannelID)
		}
		if r.TenantID != nil {
			db = db.Where("tenant_id = ?", *r.TenantID)
		}
		if r.City != nil {
			db = db.Where("city LIKE ?", "%"+*r.City+"%")
		}
		if r.HiddenName != nil {
			db = db.Where("hidden_name LIKE ?", "%"+*r.HiddenName+"%")
		}
		return db
	}
}

// SysChannelCompanyCreateRequest 创建sys_channel_company请求参数
type SysChannelCompanyCreateRequest struct {
	models.Validator
	ChannelID     int    `form:"channelId" validate:"required" message:"渠道ID不能为空"`       // 渠道ID
	TenantID      int    `form:"tenantId" validate:"required" message:"公司平台不能为空"`        // 公司平台
	City          string `form:"city" validate:"required" message:"城市不能为空"`              // 城市
	HiddenName    string `form:"hiddenName" validate:"required" message:"渠道别名不能为空"`      // 渠道别名
	IsStar        int    `form:"isStar"`                                                 // 是否回传
	FieldMappings string `form:"fieldMappings" validate:"required" message:"字段映射配置不能为空"` // 字段映射配置
}

// Validate 验证请求参数
func (r *SysChannelCompanyCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysChannelCompanyUpdateRequest 更新sys_channel_company请求参数
type SysChannelCompanyUpdateRequest struct {
	models.Validator
	Id            int    `form:"id" validate:"required" message:"Id不能为空"`                // Id
	ChannelID     int    `form:"channelId" validate:"required" message:"渠道ID不能为空"`       // 渠道ID
	TenantID      int    `form:"tenantId" validate:"required" message:"公司平台不能为空"`        // 公司平台
	City          string `form:"city" validate:"required" message:"城市不能为空"`              // 城市
	HiddenName    string `form:"hiddenName" validate:"required" message:"渠道别名不能为空"`      // 渠道别名
	IsStar        int    `form:"isStar"`                                                 // 是否回传
	FieldMappings string `form:"fieldMappings" validate:"required" message:"字段映射配置不能为空"` // 字段映射配置
}

// Validate 验证请求参数
func (r *SysChannelCompanyUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysChannelCompanyDeleteRequest 删除sys_channel_company请求参数
type SysChannelCompanyDeleteRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysChannelCompanyDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysChannelCompanyGetByIDRequest 根据ID获取sys_channel_company请求参数
type SysChannelCompanyGetByIDRequest struct {
	models.Validator
	Id int `uri:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysChannelCompanyGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
