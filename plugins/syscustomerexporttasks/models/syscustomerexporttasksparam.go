package models

import (
	"time"

	appModels "gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCustomerExportTasksListRequest defines supported task list filters.
type SysCustomerExportTasksListRequest struct {
	appModels.BasePaging
	appModels.Validator
	UserName  *string     `form:"userName"`
	Status    *string     `form:"status"`
	CreatedAt []time.Time `form:"createdAt"`
}

// Validate validates request params.
func (r *SysCustomerExportTasksListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle builds list query conditions.
func (r *SysCustomerExportTasksListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		if len(r.CreatedAt) >= 2 {
			db = db.Where("created_at BETWEEN ? AND ?", r.CreatedAt[0], r.CreatedAt[1])
		}
		return db
	}
}

// SysCustomerExportTasksGetByIDRequest defines task detail query params.
type SysCustomerExportTasksGetByIDRequest struct {
	appModels.Validator
	Id uint64 `uri:"id" validate:"required" message:"导出任务主键ID不能为空"`
}

// Validate validates request params.
func (r *SysCustomerExportTasksGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
