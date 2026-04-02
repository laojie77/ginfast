package models

import (
	models2 "gin-fast/app/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysNoticeTargetItemRequest struct {
	TargetType      int8 `json:"targetType" form:"targetType" validate:"required|in:1,2,3,4" message:"目标类型不正确"`
	TargetID        uint `json:"targetId" form:"targetId"`
	IncludeChildren bool `json:"includeChildren" form:"includeChildren"`
}

type SysNoticeAddRequest struct {
	models2.Validator
	Title      string                       `json:"title" form:"title" validate:"required" message:"通知标题不能为空"`
	Content    string                       `json:"content" form:"content" validate:"required" message:"通知内容不能为空"`
	Type       int8                         `json:"type" form:"type" validate:"required" message:"通知类型不能为空"`
	Level      string                       `json:"level" form:"level" validate:"required" message:"通知等级不能为空"`
	Targets    []SysNoticeTargetItemRequest `json:"targets" form:"targets" validate:"required" message:"通知目标不能为空"`
	PublishNow bool                         `json:"publishNow" form:"publishNow"`
}

func (r *SysNoticeAddRequest) Validate(c *gin.Context) error {
	return r.Check(c, r)
}

type SysNoticeUpdateRequest struct {
	models2.Validator
	ID         uint                         `json:"id" form:"id" validate:"required" message:"通知ID不能为空"`
	Title      string                       `json:"title" form:"title" validate:"required" message:"通知标题不能为空"`
	Content    string                       `json:"content" form:"content" validate:"required" message:"通知内容不能为空"`
	Type       int8                         `json:"type" form:"type" validate:"required" message:"通知类型不能为空"`
	Level      string                       `json:"level" form:"level" validate:"required" message:"通知等级不能为空"`
	Targets    []SysNoticeTargetItemRequest `json:"targets" form:"targets" validate:"required" message:"通知目标不能为空"`
	PublishNow bool                         `json:"publishNow" form:"publishNow"`
}

func (r *SysNoticeUpdateRequest) Validate(c *gin.Context) error {
	return r.Check(c, r)
}

type SysNoticeListRequest struct {
	models2.BasePaging
	models2.Validator
	Title         string   `json:"title" form:"title"`
	Type          *int8    `json:"type" form:"type"`
	Level         string   `json:"level" form:"level"`
	PublishStatus *int8    `json:"publishStatus" form:"publishStatus"`
	PublishTime   []string `json:"publishTime" form:"publishTime"`
}

func (r *SysNoticeListRequest) Validate(c *gin.Context) error {
	return r.Check(c, r)
}

func (r *SysNoticeListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.Title != "" {
			db = db.Where("title LIKE ?", "%"+r.Title+"%")
		}
		if r.Type != nil {
			db = db.Where("type = ?", *r.Type)
		}
		if r.Level != "" {
			db = db.Where("level = ?", r.Level)
		}
		if r.PublishStatus != nil {
			db = db.Where("publish_status = ?", *r.PublishStatus)
		}
		if len(r.PublishTime) > 1 {
			db = db.Where("publish_time BETWEEN ? AND ?", r.PublishTime[0], r.PublishTime[1])
		}
		return db
	}
}

type SysNoticeActionRequest struct {
	models2.Validator
	ID uint `json:"id" form:"id" validate:"required" message:"通知ID不能为空"`
}

func (r *SysNoticeActionRequest) Validate(c *gin.Context) error {
	return r.Check(c, r)
}

type SysNoticeInboxListRequest struct {
	models2.BasePaging
	models2.Validator
	Title      string `json:"title" form:"title"`
	Type       *int8  `json:"type" form:"type"`
	Level      string `json:"level" form:"level"`
	ReadStatus *int8  `json:"readStatus" form:"readStatus"`
}

func (r *SysNoticeInboxListRequest) Validate(c *gin.Context) error {
	return r.Check(c, r)
}

func (r *SysNoticeInboxListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ReadStatus != nil {
			db = db.Where("sys_notice_recipient.read_status = ?", *r.ReadStatus)
		}
		if r.Title != "" {
			db = db.Where("sys_notice.title LIKE ?", "%"+r.Title+"%")
		}
		if r.Type != nil {
			db = db.Where("sys_notice.type = ?", *r.Type)
		}
		if r.Level != "" {
			db = db.Where("sys_notice.level = ?", r.Level)
		}
		return db
	}
}

type SysNoticeRealtimeMockRequest struct {
	models2.Validator
	Scene       string `json:"scene" form:"scene" validate:"required" message:"实时通知场景不能为空"`
	Title       string `json:"title" form:"title" validate:"required" message:"实时通知标题不能为空"`
	Content     string `json:"content" form:"content" validate:"required" message:"实时通知内容不能为空"`
	Level       string `json:"level" form:"level"`
	ActionKind  string `json:"actionKind" form:"actionKind"`
	ActionLabel string `json:"actionLabel" form:"actionLabel"`
	ActionValue string `json:"actionValue" form:"actionValue"`
	OpenMode    string `json:"openMode" form:"openMode"`
}

func (r *SysNoticeRealtimeMockRequest) Validate(c *gin.Context) error {
	return r.Check(c, r)
}
