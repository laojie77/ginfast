package models

import (
	"context"
	models "gin-fast/app/models"
	"time"

	"gin-fast/app/global/app"

	"gorm.io/gorm"
)

const (
	SysNoticePublishStatusDraft     int8 = 0
	SysNoticePublishStatusPublished int8 = 1
	SysNoticePublishStatusRevoked   int8 = -1

	SysNoticeTargetTypeAll  int8 = 1
	SysNoticeTargetTypeUser int8 = 2
	SysNoticeTargetTypeDept int8 = 3
	SysNoticeTargetTypeRole int8 = 4

	SysNoticeReadStatusUnread int8 = 0
	SysNoticeReadStatusRead   int8 = 1
)

// SysNotice 系统通知公告
type SysNotice struct {
	models.BaseModel
	Title          string              `gorm:"column:title;size:100;not null;comment:通知标题" json:"title"`
	Content        string              `gorm:"column:content;type:longtext;not null;comment:通知内容" json:"content"`
	Type           int8                `gorm:"column:type;not null;comment:通知类型(字典code：noticeType)" json:"type"`
	Level          string              `gorm:"column:level;size:10;not null;comment:通知等级(字典code：noticeLevel)" json:"level"`
	PublisherID    uint                `gorm:"column:publisher_id;default:0;comment:发布人ID" json:"publisherId"`
	Publisher      *models.User        `gorm:"foreignKey:publisher_id;references:id" json:"publisher,omitempty"`
	PublishStatus  int8                `gorm:"column:publish_status;default:0;comment:发布状态" json:"publishStatus"`
	PublishTime    *time.Time          `gorm:"column:publish_time;comment:发布时间" json:"publishTime"`
	RevokeTime     *time.Time          `gorm:"column:revoke_time;comment:撤回时间" json:"revokeTime"`
	TenantID       uint                `gorm:"column:tenant_id;default:0;comment:租户ID" json:"tenantId"`
	CreatedBy      uint                `gorm:"column:created_by;default:0;comment:创建人ID" json:"createdBy"`
	Targets        SysNoticeTargetList `gorm:"foreignKey:notice_id;references:id" json:"targets,omitempty"`
	RecipientCount int64               `gorm:"-" json:"recipientCount"`
	UnreadCount    int64               `gorm:"-" json:"unreadCount"`
}

func (SysNotice) TableName() string {
	return "sys_notice"
}

func NewSysNotice() *SysNotice {
	return &SysNotice{}
}

func (n *SysNotice) IsEmpty() bool {
	return n == nil || n.ID == 0
}

func (n *SysNotice) Find(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(ctx).Scopes(funcs...).Find(n).Error
}

func (n *SysNotice) Create(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(ctx).Scopes(funcs...).Create(n).Error
}

func (n *SysNotice) Update(ctx context.Context) error {
	return app.DB().WithContext(ctx).Save(n).Error
}

func (n *SysNotice) Delete(ctx context.Context) error {
	return app.DB().WithContext(ctx).Delete(n).Error
}

type SysNoticeList []*SysNotice

func NewSysNoticeList() SysNoticeList {
	return SysNoticeList{}
}

func (list SysNoticeList) IsEmpty() bool {
	return len(list) == 0
}

func (list *SysNoticeList) Find(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(ctx).Scopes(funcs...).Find(list).Error
}

func (list *SysNoticeList) GetTotal(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var total int64
	err := app.DB().WithContext(ctx).Model(&SysNotice{}).Scopes(funcs...).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}
