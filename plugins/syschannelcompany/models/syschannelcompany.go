package models

import (
	"context"
	"gin-fast/app/global/app"
	"time"

	"gorm.io/gorm"
)

// SysChannelCompany sys_channel_company 模型结构体
type SysChannelCompany struct {
	Id            int            `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`             // Id
	ChannelID     int            `gorm:"column:channel_id" json:"channelId"`                                // 渠道ID
	TenantID      int            `gorm:"column:tenant_id" json:"tenantId"`                                  // 公司平台
	City          string         `gorm:"column:city" json:"city"`                                           // 城市
	HiddenName    string         `gorm:"column:hidden_name" json:"hiddenName"`                              // 渠道别名
	Rate          float64        `gorm:"column:rate;default:0.00" json:"rate"`                              // 优质率
	Quantity      int            `gorm:"column:quantity;default:0" json:"quantity"`                         // 总进件数
	IsStar        int            `gorm:"column:is_star;not null;default:0" json:"isStar"`                   // 是否回传
	FieldMappings string         `gorm:"column:field_mappings" json:"fieldMappings"`                        // 字段映射配置
	CreatedAt     *time.Time     `gorm:"column:created_at;default:'CURRENT_TIMESTAMP(3)'" json:"createdAt"` // 创建时间
	UpdatedAt     *time.Time     `gorm:"column:updated_at" json:"updatedAt"`                                // 更新时间
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at" json:"deletedAt"`                                // 删除时间
}

// SysChannelCompanyList sys_channel_company 列表
type SysChannelCompanyList []SysChannelCompany

// NewSysChannelCompany 创建sys_channel_company实例
func NewSysChannelCompany() *SysChannelCompany {
	return &SysChannelCompany{}
}

// NewSysChannelCompanyList 创建sys_channel_company列表实例
func NewSysChannelCompanyList() *SysChannelCompanyList {
	return &SysChannelCompanyList{}
}

// TableName 指定表名
func (SysChannelCompany) TableName() string {
	return "sys_channel_company"
}

// GetByID 根据ID获取sys_channel_company
func (m *SysChannelCompany) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建sys_channel_company记录
func (m *SysChannelCompany) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新sys_channel_company记录
func (m *SysChannelCompany) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除sys_channel_company记录
func (m *SysChannelCompany) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysChannelCompany) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询sys_channel_company列表
func (l *SysChannelCompanyList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysChannelCompany{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取sys_channel_company总数
func (l *SysChannelCompanyList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysChannelCompany{}).Scopes(query...).Count(&count).Error
	return count, err
}
