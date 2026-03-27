package models

import (
	"strings"

	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCustomerTracesListRequest sys_customer_traces 列表请求参数
type SysCustomerTracesListRequest struct {
	models.BasePaging
	models.Validator
	CustomerID  *int64  `form:"customerId"`  // 客户ID
	CustomerNum *string `form:"customerNum"` // 客户编号
	UserID      *int    `form:"userId"`      // 操作用户
	Data        *string `form:"data"`        // 跟进内容
}

// Validate 验证请求参数
func (r *SysCustomerTracesListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle 获取查询条件
func (r *SysCustomerTracesListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.CustomerID != nil {
			db = db.Where("sys_customer_traces.customer_id = ?", *r.CustomerID)
		}
		if r.CustomerNum != nil && strings.TrimSpace(*r.CustomerNum) != "" {
			db = db.Where("sys_customer.num LIKE ?", "%"+strings.TrimSpace(*r.CustomerNum)+"%")
		}
		if r.UserID != nil {
			db = db.Where("sys_customer_traces.user_id = ?", *r.UserID)
		}
		if r.Data != nil {
			db = db.Where("sys_customer_traces.data = ?", *r.Data)
		}
		db = db.Order("sys_customer_traces.created_at DESC")
		return db
	}
}

// SysCustomerTracesCreateRequest 创建 sys_customer_traces 请求参数
type SysCustomerTracesCreateRequest struct {
	models.Validator
	CustomerID int64  `json:"customerId" validate:"required" message:"客户ID不能为空"` // 客户
	UserID     int    `json:"userId"`                                            // 操作用户
	Data       string `json:"data" validate:"required" message:"跟进内容不能为空"`       // 跟进内容
}

// Validate 验证请求参数
func (r *SysCustomerTracesCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerTracesUpdateRequest 更新 sys_customer_traces 请求参数
type SysCustomerTracesUpdateRequest struct {
	models.Validator
	Id         uint64  `json:"id" validate:"required" message:"Id不能为空"` // Id
	CustomerID *int64  `json:"customerId"`                              // 客户
	UserID     *int    `json:"userId"`                                  // 操作用户
	Data       *string `json:"data"`                                    // 跟进内容
}

// Validate 验证请求参数
func (r *SysCustomerTracesUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerTracesDeleteRequest 删除 sys_customer_traces 请求参数
type SysCustomerTracesDeleteRequest struct {
	models.Validator
	Id uint64 `json:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCustomerTracesDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerTracesGetByIDRequest 根据ID获取 sys_customer_traces 请求参数
type SysCustomerTracesGetByIDRequest struct {
	models.Validator
	Id uint64 `uri:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCustomerTracesGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
