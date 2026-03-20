package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/plugins/syschannelcompany/models"
	"time"

	"gorm.io/gorm"
)

// SysChannel sys_channel 模型结构体
type SysChannel struct {
	Id               int            `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`            // Id
	ChannelName      string         `gorm:"column:channel_name;not null" json:"channelName"`                  // 渠道名称
	ChannelKey       string         `gorm:"column:channel_key;not null" json:"channelKey"`                    // 渠道码
	HiddenCode       string         `gorm:"column:hidden_code" json:"hiddenCode"`                             // 渠道隐藏名称
	SecretKey        string         `gorm:"column:secret_key;not null" json:"secretKey"`                      // 渠道密钥
	StarUrl          string         `gorm:"column:star_url" json:"starUrl"`                                   // 星级回传地址
	Total            int            `gorm:"column:total;default:0" json:"total"`                              // 总进件数量
	Rate             string         `gorm:"column:rate;default:0" json:"rate"`                                // 优质率
	Remark           string         `gorm:"column:remark" json:"remark"`                                      // 备注
	Status           int            `gorm:"column:status;default:0" json:"status"`                            // 状态
	StartTime        string         `gorm:"column:start_time" json:"startTime"`                               // 开始时间
	EndTime          string         `gorm:"column:end_time" json:"endTime"`                                   // 结束时间
	Type             int            `gorm:"column:type" json:"type"`                                          // 数据类型
	MessageType      int            `gorm:"column:message_type;default:0" json:"messageType"`                 // 短信
	IsRepeat         int            `gorm:"column:is_repeat;not null;default:1" json:"isRepeat"`              // 是否查重
	SuccessCode      string         `gorm:"column:success_code;default:'0'" json:"successCode"`               // 成功返回码
	SuccessCodeField string         `gorm:"column:success_code_field;default:'code'" json:"successCodeField"` // 成功码字段名
	CreatedAt        time.Time      `json:"createdAt"`
	UpdatedAt        time.Time      `json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"index" json:"deletedAt"`

	// 关联关系
	ChannelCompanyList []models.SysChannelCompany `gorm:"foreignKey:ChannelId" json:"channelCompanyList"` // 渠道配置列表

}

// SysChannelList sys_channel 列表
type SysChannelList []SysChannel

// NewSysChannel 创建sys_channel实例
func NewSysChannel() *SysChannel {
	return &SysChannel{}
}

// NewSysChannelList 创建sys_channel列表实例
func NewSysChannelList() *SysChannelList {
	return &SysChannelList{}
}

// TableName 指定表名
func (SysChannel) TableName() string {
	return "sys_channel"
}

// GetByID 根据ID获取sys_channel
func (m *SysChannel) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建sys_channel记录
func (m *SysChannel) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新sys_channel记录
func (m *SysChannel) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除sys_channel记录
func (m *SysChannel) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysChannel) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询sys_channel列表
func (l *SysChannelList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysChannel{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取sys_channel总数
func (l *SysChannelList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysChannel{}).Scopes(query...).Count(&count).Error
	return count, err
}
