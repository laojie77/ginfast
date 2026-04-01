package models

import (
	"context"
	models2 "gin-fast/app/models"
	"time"

	"gin-fast/app/global/app"

	"gorm.io/gorm"
)

// SysNoticeRecipient 通知接收人快照
type SysNoticeRecipient struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	NoticeID    uint       `gorm:"column:notice_id;not null;comment:通知ID" json:"noticeId"`
	UserID      uint       `gorm:"column:user_id;not null;comment:接收人ID" json:"userId"`
	ReadStatus  int8       `gorm:"column:read_status;default:0;comment:读取状态" json:"readStatus"`
	ReadTime    *time.Time `gorm:"column:read_time;comment:阅读时间" json:"readTime"`
	TenantID    uint       `gorm:"column:tenant_id;default:0;comment:租户ID" json:"tenantId"`
	PublishTime *time.Time `gorm:"column:publish_time;comment:发布时间快照" json:"publishTime"`
	CreatedAt time.Time    `gorm:"column:created_at" json:"createdAt"`
	Notice    SysNotice    `gorm:"foreignKey:notice_id;references:id" json:"notice"`
	User      models2.User `gorm:"foreignKey:user_id;references:id" json:"user"`
}

func (SysNoticeRecipient) TableName() string {
	return "sys_notice_recipient"
}

func NewSysNoticeRecipient() *SysNoticeRecipient {
	return &SysNoticeRecipient{}
}

type SysNoticeRecipientList []*SysNoticeRecipient

func NewSysNoticeRecipientList() SysNoticeRecipientList {
	return SysNoticeRecipientList{}
}

func (list SysNoticeRecipientList) IsEmpty() bool {
	return len(list) == 0
}

func (list *SysNoticeRecipientList) Find(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(ctx).Scopes(funcs...).Find(list).Error
}

func (list *SysNoticeRecipientList) GetTotal(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var total int64
	err := app.DB().WithContext(ctx).Model(&SysNoticeRecipient{}).Scopes(funcs...).Count(&total).Error
	if err != nil {
		return 0, err
	}
	return total, nil
}
