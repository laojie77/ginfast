package models

import (
	"context"
	"time"
	"gin-fast/app/global/app"
	"gorm.io/gorm"
)

// SysCallRecord sys_call_record 模型结构体
type SysCallRecord struct {
	Id int `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"` // Id
	UserId int `gorm:"column:user_id" json:"userId"` // 业务员id
	UserName string `gorm:"column:user_name" json:"userName"` // 业务员姓名
	Name string `gorm:"column:name" json:"name"` // 客户姓名
	Mobile string `gorm:"column:mobile" json:"mobile"` // 电话
	Status int `gorm:"column:status;default:0;index" json:"status"` // 状态:0等待处理,1已完成,2未接听/挂断/拒绝,3待拨号
	Type int `gorm:"column:type" json:"type"` // 通话类型:1呼出/拨号,2呼入/接听
	StartTime *time.Time `gorm:"column:start_time" json:"startTime"` // 开始时间
	EndTime *time.Time `gorm:"column:end_time" json:"endTime"` // 结束时间
	Vdieo string `gorm:"column:vdieo" json:"vdieo"` // 录音地址
	Duration int `gorm:"column:duration" json:"duration"` // 通话时长
	Version string `gorm:"column:version" json:"version"` // 版本
	CreatedAt *time.Time `gorm:"column:created_at" json:"createdAt"` // CreatedAt
	TenantId uint `gorm:"column:tenant_id;default:0" json:"tenantId"` // 租户ID字段
}

// SysCallRecordList sys_call_record 列表
type SysCallRecordList []SysCallRecord

// NewSysCallRecord 创建sys_call_record实例
func NewSysCallRecord() *SysCallRecord {
	return &SysCallRecord{}
}

// NewSysCallRecordList 创建sys_call_record列表实例
func NewSysCallRecordList() *SysCallRecordList {
	return &SysCallRecordList{}
}

// TableName 指定表名
func (SysCallRecord) TableName() string {
	return "sys_call_record"
}

// GetByID 根据ID获取sys_call_record
func (m *SysCallRecord) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建sys_call_record记录
func (m *SysCallRecord) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新sys_call_record记录
func (m *SysCallRecord) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除sys_call_record记录
func (m *SysCallRecord) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysCallRecord) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询sys_call_record列表
func (l *SysCallRecordList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCallRecord{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取sys_call_record总数
func (l *SysCallRecordList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCallRecord{}).Scopes(query...).Count(&count).Error
	return count, err
}