package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/exampleutils/snowflakehelper"
	"time"

	"gorm.io/gorm"
)

// SysCustomer sys_customer 模型结构体
type SysCustomer struct {
	Id                 int            `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`   // 主键ID
	Num                string         `gorm:"column:num;size:64;index" json:"num"`                     // 客户编号
	Name               string         `gorm:"column:name" json:"name"`                                 // 客户姓名
	Mobile             string         `gorm:"column:mobile;index" json:"mobile"`                       // 手机号
	MdMobile           string         `gorm:"column:md5_mobile;index" json:"mdMobile"`                 // 手机号MD5值，用于隐私查询
	MoneyDemand        int            `gorm:"column:money_demand" json:"moneyDemand"`                  // 需求金额
	ChannelId          int            `gorm:"column:channel_id;index" json:"channelId"`                // 渠道来源
	UserId             int            `gorm:"column:user_id;index" json:"userId"`                      // 跟进人
	CustomerStar       int            `gorm:"column:customer_star" json:"customerStar"`                // 星级
	Status             int            `gorm:"column:status;not null;default:1;index" json:"status"`    // 业务阶段
	Intention          int            `gorm:"column:intention;not null;default:0" json:"intention"`    // 客户有效
	LastTime           *time.Time     `gorm:"column:last_time" json:"lastTime"`                        // 最后操作时间
	Extra              string         `gorm:"column:extra;type:json" json:"extra"`                     // 扩展属性
	SinglePieceType    int            `gorm:"column:single_piece_type" json:"singlePieceType"`         // 贷款类型
	Sex                int            `gorm:"column:sex;not null;default:0" json:"sex"`                // 性别
	AllotTime          *time.Time     `gorm:"column:allot_time" json:"allotTime"`                      // 分配时间
	DepartmentId       int            `gorm:"column:department_id;index" json:"departmentId"`          // 所属部门
	TenantId           int            `gorm:"column:tenant_id" json:"tenantId"`                        // 所属公司ID
	Remarks            string         `gorm:"column:remarks" json:"remarks"`                           // 客户备注
	Age                int            `gorm:"column:age" json:"age"`                                   // 年龄
	City               string         `gorm:"column:city" json:"city"`                                 // 所在城市
	CustomerComment    string         `gorm:"column:customer_comment" json:"customerComment"`          // 上级评价
	Source             string         `gorm:"column:source" json:"source"`                             // 渠道编码
	NewData            int            `gorm:"column:new_data" json:"newData"`                          // 新数据
	RedistributionTime *time.Time     `gorm:"column:redistribution_time" json:"redistributionTime"`    // 再分配时间
	IsReassign         int            `gorm:"column:is_reassign;default:0" json:"isReassign"`          // 再分配
	BatchId            int            `gorm:"column:batch_id" json:"batchId"`                          // 导入批次ID
	IsRead             int            `gorm:"column:is_read" json:"isRead"`                            // 是否已读
	IsPublic           int            `gorm:"column:is_public" json:"isPublic"`                        // 公共池
	IsQuit             int            `gorm:"column:is_quit" json:"isQuit"`                            // 离职数据
	IsRepeat           int            `gorm:"column:is_repeat" json:"isRepeat"`                        // 重复标记
	IsRubbish          int            `gorm:"column:is_rubbish" json:"isRubbish"`                      // 是否垃圾库：1是
	RemarkTime         *time.Time     `gorm:"column:remark_time" json:"remarkTime"`                    // 最后备注时间
	DispatchTime       *time.Time     `gorm:"column:dispatch_time" json:"dispatchTime"`                // 待调度时间
	IsRemind           int            `gorm:"column:is_remind" json:"isRemind"`                        // 是否提醒
	IsSms              int            `gorm:"column:is_sms;default:0" json:"isSms"`                    // 短信发送状态
	StarStatus         int            `gorm:"column:star_status;not null;default:0" json:"starStatus"` // 星级回传
	IsExchange         int            `gorm:"column:is_exchange" json:"isExchange"`                    // 待流转标记：1待流转
	PublicTypeTime     *time.Time     `gorm:"column:public_type_time" json:"publicTypeTime"`           // 抓取公共池时间
	PublicType         int            `gorm:"column:public_type" json:"publicType"`                    // 抓取类型
	IsLock             int            `gorm:"column:is_lock;not null;default:0" json:"isLock"`         // 是否锁定
	CreatedAt          *time.Time     `gorm:"column:created_at" json:"createdAt"`                      // CreatedAt
	UpdatedAt          *time.Time     `gorm:"column:updated_at" json:"updatedAt"`                      // UpdatedAt
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at" json:"deletedAt"`                      // DeletedAt
}

// SysCustomerList sys_customer 列表
type SysCustomerList []SysCustomer

// NewSysCustomer 创建sys_customer实例
func NewSysCustomer() *SysCustomer {
	return &SysCustomer{}
}

// NewSysCustomerList 创建sys_customer列表实例
func NewSysCustomerList() *SysCustomerList {
	return &SysCustomerList{}
}

// TableName 指定表名
func (SysCustomer) TableName() string {
	return "sys_customer"
}

// GetByID 根据ID获取sys_customer
func (m *SysCustomer) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建sys_customer记录
func (m *SysCustomer) Create(c context.Context) error {
	// 如果客户编号为空，自动生成雪花ID
	if m.Num == "" {
		id, err := snowflakehelper.GenerateID()
		if err != nil {
			return err
		}
		m.Num = id
	}
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新sys_customer记录
func (m *SysCustomer) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除sys_customer记录
func (m *SysCustomer) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysCustomer) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询sys_customer列表
func (l *SysCustomerList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCustomer{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取sys_customer总数
func (l *SysCustomerList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCustomer{}).Scopes(query...).Count(&count).Error
	return count, err
}
