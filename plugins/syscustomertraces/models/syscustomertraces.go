package models

import (
	"context"
	"time"

	"gin-fast/app/global/app"

	"gorm.io/gorm"
)

// SysCustomerTraces sys_customer_traces 模型结构体
type SysCustomerTraces struct {
	Id          uint64         `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`     // Id
	CustomerID  int64          `gorm:"column:customer_id;not null;index" json:"customerId"`       // 客户
	UserID      int            `gorm:"column:user_id;not null;index" json:"userId"`               // 操作用户
	TenantID    uint           `gorm:"column:tenant_id;not null;default:0;index" json:"tenantId"` // 所属租户
	Data        string         `gorm:"column:data;not null" json:"data"`                          // 跟进内容
	CreatedAt   *time.Time     `gorm:"column:created_at;index" json:"createdAt"`                  // CreatedAt
	UpdatedAt   *time.Time     `gorm:"column:updated_at" json:"updatedAt"`                        // UpdatedAt
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at" json:"deletedAt"`                        // DeletedAt
	UserName    string         `gorm:"column:user_name;->" json:"userName"`                       // 用户名称
	Avatar      string         `gorm:"column:avatar;->" json:"avatar"`                            // 用户头像
	CustomerNum string         `gorm:"column:customer_num;->" json:"customerNum"`                 // 客户编号
}

// SysCustomerTracesWithUser 包含用户信息的跟进记录
type SysCustomerTracesWithUser struct {
	SysCustomerTraces
}

// SysCustomerTracesList sys_customer_traces 列表
type SysCustomerTracesList []SysCustomerTraces

// SysCustomerTracesWithUserList 包含用户信息的跟进记录列表
type SysCustomerTracesWithUserList []SysCustomerTracesWithUser

// NewSysCustomerTraces 创建 sys_customer_traces 实例
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
		Select("sys_customer_traces.*, sys_users.nick_name as user_name, sys_users.avatar as avatar, sys_customer.num as customer_num").
		Joins("LEFT JOIN sys_users ON sys_customer_traces.user_id = sys_users.id").
		Joins("LEFT JOIN sys_customer ON sys_customer_traces.customer_id = sys_customer.id").
		Scopes(funcs...).
		Find(l).Error
}

// GetTotalWithUser 获取包含用户信息的跟进记录总数
func (l *SysCustomerTracesWithUserList) GetTotalWithUser(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).
		Table("sys_customer_traces").
		Joins("LEFT JOIN sys_users ON sys_customer_traces.user_id = sys_users.id").
		Joins("LEFT JOIN sys_customer ON sys_customer_traces.customer_id = sys_customer.id").
		Scopes(query...).
		Count(&count).Error
	return count, err
}

// TableName 指定表名
func (SysCustomerTraces) TableName() string {
	return "sys_customer_traces"
}

// GetByID 根据 ID 获取 sys_customer_traces
func (m *SysCustomerTraces) GetByID(c context.Context, id uint64) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建 sys_customer_traces 记录
func (m *SysCustomerTraces) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新 sys_customer_traces 记录
func (m *SysCustomerTraces) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除 sys_customer_traces 记录
func (m *SysCustomerTraces) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysCustomerTraces) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询 sys_customer_traces 列表
func (l *SysCustomerTracesList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCustomerTraces{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取 sys_customer_traces 总数
func (l *SysCustomerTracesList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCustomerTraces{}).Scopes(query...).Count(&count).Error
	return count, err
}
