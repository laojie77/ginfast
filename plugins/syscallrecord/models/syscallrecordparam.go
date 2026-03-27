package models

import (
	"time"
	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCallRecordListRequest sys_call_record列表请求参数
type SysCallRecordListRequest struct {
	models.BasePaging
	models.Validator
	UserId *int `form:"userId"` // 业务员id
	Mobile *string `form:"mobile"` // 电话
	Status *int `form:"status"` // 状态:0等待处理,1已完成,2未接听/挂断/拒绝,3待拨号
	Type *int `form:"type"` // 通话类型:1呼出/拨号,2呼入/接听
	CreatedAt *time.Time `form:"createdAt"` // CreatedAt
}

// Validate 验证请求参数
func (r *SysCallRecordListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle 获取查询条件
func (r *SysCallRecordListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
        if r.UserId != nil {
            db = db.Where("user_id = ?", *r.UserId)
        }
        if r.Mobile != nil {
            db = db.Where("mobile = ?", *r.Mobile)
        }
        if r.Status != nil {
            db = db.Where("status = ?", *r.Status)
        }
        if r.Type != nil {
            db = db.Where("type = ?", *r.Type)
        }
        if r.CreatedAt != nil {
            // 默认等于查询
            db = db.Where("created_at = ?", *r.CreatedAt)
        }
		return db
	}
}

// SysCallRecordCreateRequest 创建sys_call_record请求参数
type SysCallRecordCreateRequest struct {
	models.Validator
}

// Validate 验证请求参数
func (r *SysCallRecordCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCallRecordUpdateRequest 更新sys_call_record请求参数
type SysCallRecordUpdateRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCallRecordUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCallRecordDeleteRequest 删除sys_call_record请求参数
type SysCallRecordDeleteRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCallRecordDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCallRecordGetByIDRequest 根据ID获取sys_call_record请求参数
type SysCallRecordGetByIDRequest struct {
	models.Validator
	Id int `uri:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysCallRecordGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}