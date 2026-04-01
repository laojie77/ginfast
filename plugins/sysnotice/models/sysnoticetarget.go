package models

import (
	"context"
	"time"

	"gin-fast/app/global/app"

	"gorm.io/gorm"
)

// SysNoticeTarget 通知目标规则
type SysNoticeTarget struct {
	ID              uint      `gorm:"primarykey" json:"id"`
	NoticeID        uint      `gorm:"column:notice_id;not null;comment:通知ID" json:"noticeId"`
	TargetType      int8      `gorm:"column:target_type;not null;comment:目标类型;commit:目标类型（1: 全体, 2: 用户，3：本部门及下级，4： 角色）" json:"targetType"`
	TargetID        uint      `gorm:"column:target_id;default:0;comment:目标ID" json:"targetId"`
	IncludeChildren bool      `gorm:"column:include_children;default:false;comment:是否包含下级部门" json:"includeChildren"`
	TenantID        uint      `gorm:"column:tenant_id;default:0;comment:租户ID" json:"tenantId"`
	CreatedAt       time.Time `gorm:"column:created_at" json:"createdAt"`
	TargetName      string    `gorm:"-" json:"targetName"`
}

func (SysNoticeTarget) TableName() string {
	return "sys_notice_target"
}

func NewSysNoticeTarget() *SysNoticeTarget {
	return &SysNoticeTarget{}
}

type SysNoticeTargetList []*SysNoticeTarget

func NewSysNoticeTargetList() SysNoticeTargetList {
	return SysNoticeTargetList{}
}

func (list SysNoticeTargetList) IsEmpty() bool {
	return len(list) == 0
}

func (list *SysNoticeTargetList) Find(ctx context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(ctx).Scopes(funcs...).Find(list).Error
}
