package models

import (
	"time"

	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysCallRecordListRequest struct {
	models.BasePaging
	models.Validator
	UserId     *int       `form:"userId"`
	CustomerID *string    `form:"customer_id"`
	Mobile     *string    `form:"mobile"`
	Status     *int       `form:"status"`
	Type       *int       `form:"type"`
	CreatedAt  *time.Time `form:"createdAt"`
}

func (r *SysCallRecordListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

func (r *SysCallRecordListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.UserId != nil {
			db = db.Where("sys_call_record.user_id = ?", *r.UserId)
		}
		if r.CustomerID != nil && *r.CustomerID != "" {
			db = db.Where("sys_customer.num LIKE ?", "%"+*r.CustomerID+"%")
		}
		if r.Mobile != nil {
			db = db.Where("sys_call_record.mobile = ?", *r.Mobile)
		}
		if r.Status != nil {
			db = db.Where("sys_call_record.status = ?", *r.Status)
		}
		if r.Type != nil {
			db = db.Where("sys_call_record.type = ?", *r.Type)
		}
		if r.CreatedAt != nil {
			db = db.Where("sys_call_record.created_at = ?", *r.CreatedAt)
		}
		return db
	}
}

type SysCallRecordCreateRequest struct {
	models.Validator
}

func (r *SysCallRecordCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type SysCallRecordUpdateRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"Id不能为空"`
}

func (r *SysCallRecordUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type SysCallRecordDeleteRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"Id不能为空"`
}

func (r *SysCallRecordDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type SysCallRecordGetByIDRequest struct {
	models.Validator
	Id int `uri:"id" validate:"required" message:"Id不能为空"`
}

func (r *SysCallRecordGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
