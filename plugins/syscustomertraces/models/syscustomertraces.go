package models

import (
	"context"
	"gin-fast/app/global/app"
	"time"

	"gorm.io/gorm"
)

// SysCustomerTraces sys_customer_traces 模型结构体
type SysCustomerTraces struct {
	Id         uint64         `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`     // Id
	CustomerId int64          `gorm:"column:customer_id;not null;index" json:"customerId"`       // 客户
	UserId     int            `gorm:"column:user_id;not null;index" json:"userId"`               // 操作用户
	TenantID   uint           `gorm:"column:tenant_id;not null;default:0;index" json:"tenantId"` // 所属平台公司
	Data       string         `gorm:"column:data;not null" json:"data"`                          // 跟进内容
	CreatedAt  *time.Time     `gorm:"column:created_at;index" json:"createdAt"`                  // CreatedAt
	UpdatedAt  *time.Time     `gorm:"column:updated_at" json:"updatedAt"`                        // UpdatedAt
	DeletedAt  gorm.DeletedAt `gorm:"column:deleted_at" json:"deletedAt"`                        // DeletedAt
}

// SysCustomerTracesWithUser 包含用户信息的跟进记录
type SysCustomerTracesWithUser struct {
	SysCustomerTraces
	UserName string `json:"userName"` // 用户昵称
	Avatar   string `json:"avatar"`   // 用户头像
}

// SysCustomerTracesList sys_customer_traces 列表
type SysCustomerTracesList []SysCustomerTraces

// SysCustomerTracesWithUserList 包含用户信息的跟进记录列表
type SysCustomerTracesWithUserList []SysCustomerTracesWithUser

// NewSysCustomerTraces 创建sys_customer_traces实例
func NewSysCustomerTraces() *SysCustomerTraces {
	return &SysCustomerTraces{}
}

// NewSysCustomerTracesWithUserList 创建包含用户信息的跟进记录列表实例
func NewSysCustomerTracesWithUserList() *SysCustomerTracesWithUserList {
	return &SysCustomerTracesWithUserList{}
}

// FindWithUser 查询包含用户信息的跟进记录列表
func (l *SysCustomerTracesWithUserList) FindWithUser(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).
		Table("sys_customer_traces").
		Select("sys_customer_traces.*, sys_users.nick_name as user_name, sys_users.avatar as avatar").
		Joins("LEFT JOIN sys_users ON sys_customer_traces.user_id = sys_users.id").
		Scopes(funcs...).
		Find(l).Error
}

// GetTotalWithUser 获取包含用户信息的跟进记录总数
func (l *SysCustomerTracesWithUserList) GetTotalWithUser(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).
		Table("sys_customer_traces").
		Joins("LEFT JOIN sys_users ON sys_customer_traces.user_id = sys_users.id").
		Scopes(query...).
		Count(&count).Error
	return count, err
}

// TableName 指定表名
func (SysCustomerTraces) TableName() string {
	return "sys_customer_traces"
}

// GetByID 根据ID获取sys_customer_traces
func (m *SysCustomerTraces) GetByID(c context.Context, id uint64) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建sys_customer_traces记录
func (m *SysCustomerTraces) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新sys_customer_traces记录
func (m *SysCustomerTraces) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除sys_customer_traces记录
func (m *SysCustomerTraces) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysCustomerTraces) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询sys_customer_traces列表
func (l *SysCustomerTracesList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCustomerTraces{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取sys_customer_traces总数
func (l *SysCustomerTracesList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCustomerTraces{}).Scopes(query...).Count(&count).Error
	return count, err
}
