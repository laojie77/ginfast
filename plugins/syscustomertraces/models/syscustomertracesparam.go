package models

import (
	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCustomerTracesListRequest sys_customer_traces列表请求参数
type SysCustomerTracesListRequest struct {
	models.BasePaging
	models.Validator
	CustomerId *int64  `form:"customerId"` // 客户
	UserID     *int    `form:"userId"`     // 操作用户
	Data       *string `form:"data"`       // 跟进内容
}

// Validate 验证请求参数
func (r *SysCustomerTracesListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle 获取查询条件
func (r *SysCustomerTracesListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.CustomerId != nil {
			db = db.Where("sys_customer_traces.customer_id = ?", *r.CustomerId)
		}
		if r.UserID != nil {
			db = db.Where("sys_customer_traces.user_id = ?", *r.UserID)
		}
		if r.Data != nil {
			// 默认等于查询
			db = db.Where("sys_customer_traces.data = ?", *r.Data)
		}
		// 按创建时间倒序排列，明确指定表名
		db = db.Order("sys_customer_traces.created_at DESC")
		return db
	}
}

// SysCustomerTracesCreateRequest 创建sys_customer_traces请求参数
type SysCustomerTracesCreateRequest struct {
	models.Validator
	CustomerId int64  `json:"customerId" validate:"required" message:"客户ID不能为空"` // 客户
	UserID     int    `json:"userId"`                                            // 操作用户
	Data       string `json:"data" validate:"required" message:"跟进内容不能为空"`       // 跟进内容
}

// Validate 验证请求参数
func (r *SysCustomerTracesCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerTracesUpdateRequest 更新sys_customer_traces请求参数
type SysCustomerTracesUpdateRequest struct {
	models.Validator
	Id         uint64  `json:"id" validate:"required" message:"Id不能为空"` // Id
	CustomerId *int64  `json:"customerId"`                              // 客户
	UserID     *int    `json:"userId"`                                  // 操作用户
	Data       *string `json:"data"`                                    // 跟进内容
}

// Validate 验证请求参数
func (r *SysCustomerTracesUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerTracesDeleteRequest 删除sys_customer_traces请求参数
type SysCustomerTracesDeleteRequest struct {
	models.Validator
	Id uint64 `json:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCustomerTracesDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerTracesGetByIDRequest 根据ID获取sys_customer_traces请求参数
type SysCustomerTracesGetByIDRequest struct {
	models.Validator
	Id uint64 `uri:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCustomerTracesGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
