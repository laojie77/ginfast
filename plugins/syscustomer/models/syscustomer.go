package models

import (
	"context"
	"gin-fast/app/global/app"
	"gin-fast/exampleutils/snowflakehelper"
	"gin-fast/plugins/syscustomertraces/models"
	"time"

	"gorm.io/gorm"
)

// SysCustomer sys_customer 模型结构体
type SysCustomer struct {
	Id                 int            `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`
	Num                string         `gorm:"column:num;size:64;index" json:"num"`
	Name               string         `gorm:"column:name" json:"name"`
	Mobile             string         `gorm:"column:mobile;index" json:"mobile"`
	MdMobile           string         `gorm:"column:md5_mobile;index" json:"mdMobile"`
	MoneyDemand        int            `gorm:"column:money_demand" json:"moneyDemand"`
	ChannelID          int            `gorm:"column:channel_id;index" json:"channelId"`
	UserID             int            `gorm:"column:user_id;index" json:"userId"`
	CustomerStar       int            `gorm:"column:customer_star" json:"customerStar"`
	Status             int            `gorm:"column:status;not null;default:1;index" json:"status"`
	Intention          int            `gorm:"column:intention;not null;default:0" json:"intention"`
	From               int            `gorm:"column:from;not null;default:1" json:"from"`
	LastTime           *time.Time     `gorm:"column:last_time" json:"lastTime"`
	Extra              string         `gorm:"column:extra;type:json" json:"extra"`
	SinglePieceType    int            `gorm:"column:single_piece_type" json:"singlePieceType"`
	Sex                int            `gorm:"column:sex;not null;default:0" json:"sex"`
	AllotTime          *time.Time     `gorm:"column:allot_time" json:"allotTime"`
	DeptID             int            `gorm:"column:dept_id;index" json:"deptId"`
	TenantID           int            `gorm:"column:tenant_id" json:"tenantId"`
	Remarks            string         `gorm:"column:remarks" json:"remarks"`
	Age                int            `gorm:"column:age" json:"age"`
	City               string         `gorm:"column:city" json:"city"`
	CustomerComment    string         `gorm:"column:customer_comment" json:"customerComment"`
	Source             string         `gorm:"column:source" json:"source"`
	NewData            int            `gorm:"column:new_data" json:"newData"`
	RedistributionTime *time.Time     `gorm:"column:redistribution_time" json:"redistributionTime"`
	IsReassign         int            `gorm:"column:is_reassign;default:0" json:"isReassign"`
	BatchId            int            `gorm:"column:batch_id" json:"batchId"`
	IsRead             int            `gorm:"column:is_read" json:"isRead"`
	IsPublic           int            `gorm:"column:is_public" json:"isPublic"`
	IsQuit             int            `gorm:"column:is_quit" json:"isQuit"`
	IsRepeat           int            `gorm:"column:is_repeat" json:"isRepeat"`
	IsRubbish          int            `gorm:"column:is_rubbish" json:"isRubbish"`
	RemarkTime         *time.Time     `gorm:"column:remark_time" json:"remarkTime"`
	DispatchTime       *time.Time     `gorm:"column:dispatch_time" json:"dispatchTime"`
	IsRemind           int            `gorm:"column:is_remind" json:"isRemind"`
	IsSms              int            `gorm:"column:is_sms;default:0" json:"isSms"`
	StarStatus         int            `gorm:"column:star_status;not null;default:0" json:"starStatus"`
	IsExchange         int            `gorm:"column:is_exchange" json:"isExchange"`
	PublicTypeTime     *time.Time     `gorm:"column:public_type_time" json:"publicTypeTime"`
	PublicType         int            `gorm:"column:public_type" json:"publicType"`
	IsLock             int            `gorm:"column:is_lock;not null;default:0" json:"isLock"`
	CreatedAt          *time.Time     `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt          *time.Time     `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt          gorm.DeletedAt `gorm:"column:deleted_at" json:"deletedAt"`
	// 关联关系
	CustomerTracesList []models.SysCustomerTraces `gorm:"foreignKey:CustomerID;references:Id" json:"customerTracesList"` // 客户跟进列表
}

// SysCustomerList sys_customer 列表
type SysCustomerList []SysCustomer

// NewSysCustomer 创建 sys_customer 实例
func NewSysCustomer() *SysCustomer {
	return &SysCustomer{}
}

// NewSysCustomerList 创建 sys_customer 列表实例
func NewSysCustomerList() *SysCustomerList {
	return &SysCustomerList{}
}

// TableName 指定表名
func (SysCustomer) TableName() string {
	return "sys_customer"
}

// GetByID 根据 ID 获取 sys_customer
func (m *SysCustomer) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

// Create 创建 sys_customer 记录
func (m *SysCustomer) Create(c context.Context) error {
	if m.Num == "" {
		id, err := snowflakehelper.GenerateID()
		if err != nil {
			return err
		}
		m.Num = id
	}
	return app.DB().WithContext(c).Create(m).Error
}

// Update 更新 sys_customer 记录
func (m *SysCustomer) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

// Delete 软删除 sys_customer 记录
func (m *SysCustomer) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

// IsEmpty 检查模型是否为空
func (m *SysCustomer) IsEmpty() bool {
	return m == nil || m.Id == 0
}

// Find 查询 sys_customer 列表
func (l *SysCustomerList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCustomer{}).Scopes(funcs...).Find(l).Error
}

// GetTotal 获取 sys_customer 总数
func (l *SysCustomerList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCustomer{}).Scopes(query...).Count(&count).Error
	return count, err
}
