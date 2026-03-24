package models

import (
	"context"
	"time"

	"gin-fast/app/global/app"
	"gin-fast/exampleutils/snowflakehelper"
	traceModels "gin-fast/plugins/syscustomertraces/models"

	"gorm.io/gorm"
)

// SysCustomer sys_customer model
type SysCustomer struct {
	Id                 int                             `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`
	Num                string                          `gorm:"column:num;size:64;index" json:"num"`
	Name               string                          `gorm:"column:name" json:"name"`
	Mobile             string                          `gorm:"column:mobile;index" json:"mobile"`
	MdMobile           string                          `gorm:"column:md5_mobile;index" json:"mdMobile"`
	MoneyDemand        int                             `gorm:"column:money_demand;default:0" json:"moneyDemand"`
	ChannelID          int                             `gorm:"column:channel_id;index" json:"channelId"`
	UserID             int                             `gorm:"column:user_id;index" json:"userId"`
	CustomerStar       *int                            `gorm:"column:customer_star;default:null;comment:客户星级 1-5" json:"customerStar"`
	Status             int                             `gorm:"column:status;not null;default:0;index" json:"status"`
	Intention          int                             `gorm:"column:intention;not null;default:0" json:"intention"`
	From               int                             `gorm:"column:from;not null;default:1" json:"from"`
	LastTime           *time.Time                      `gorm:"column:last_time" json:"lastTime"`
	Extra              string                          `gorm:"column:extra;type:json" json:"extra"`
	SinglePieceType    int                             `gorm:"column:single_piece_type" json:"singlePieceType"`
	Sex                int                             `gorm:"column:sex;not null;default:2" json:"sex"`
	AllotTime          time.Time                       `gorm:"column:allot_time" json:"allotTime"`
	DeptID             int                             `gorm:"column:dept_id;index" json:"deptId"`
	TenantID           int                             `gorm:"column:tenant_id" json:"tenantId"`
	Remarks            string                          `gorm:"column:remarks" json:"remarks"`
	Age                int                             `gorm:"column:age" json:"age"`
	City               string                          `gorm:"column:city" json:"city"`
	CustomerComment    string                          `gorm:"column:customer_comment" json:"customerComment"`
	Source             string                          `gorm:"column:source" json:"source"`
	NewData            int                             `gorm:"column:new_data" json:"newData"`
	RedistributionTime *time.Time                      `gorm:"column:redistribution_time" json:"redistributionTime"`
	IsReassign         int                             `gorm:"column:is_reassign;default:0" json:"isReassign"`
	BatchId            int                             `gorm:"column:batch_id" json:"batchId"`
	IsRead             int                             `gorm:"column:is_read" json:"isRead"`
	IsPublic           int                             `gorm:"column:is_public" json:"isPublic"`
	IsQuit             int                             `gorm:"column:is_quit" json:"isQuit"`
	IsRepeat           int                             `gorm:"column:is_repeat" json:"isRepeat"`
	IsRubbish          int                             `gorm:"column:is_rubbish" json:"isRubbish"`
	RemarkTime         *time.Time                      `gorm:"column:remark_time" json:"remarkTime"`
	DispatchTime       *time.Time                      `gorm:"column:dispatch_time" json:"dispatchTime"`
	IsRemind           int                             `gorm:"column:is_remind" json:"isRemind"`
	IsSms              int                             `gorm:"column:is_sms;default:0" json:"isSms"`
	StarStatus         int                             `gorm:"column:star_status;not null;default:0" json:"starStatus"`
	IsExchange         int                             `gorm:"column:is_exchange" json:"isExchange"`
	PublicTypeTime     *time.Time                      `gorm:"column:public_type_time" json:"publicTypeTime"`
	PublicType         int                             `gorm:"column:public_type" json:"publicType"`
	IsLock             int                             `gorm:"column:is_lock;not null;default:0" json:"isLock"`
	CreatedAt          *time.Time                      `gorm:"column:created_at" json:"createdAt"`
	UpdatedAt          *time.Time                      `gorm:"column:updated_at" json:"updatedAt"`
	DeletedAt          gorm.DeletedAt                  `gorm:"column:deleted_at" json:"deletedAt"`
	CustomerTracesList []traceModels.SysCustomerTraces `gorm:"foreignKey:CustomerID;references:Id" json:"customerTracesList"`
}

type SysCustomerList []SysCustomer
type CustomerStatusTraces struct{}

func NewSysCustomer() *SysCustomer {
	return &SysCustomer{}
}

func NewSysCustomerList() *SysCustomerList {
	return &SysCustomerList{}
}

func (SysCustomer) TableName() string {
	return "sys_customer"
}

func (m *SysCustomer) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

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

func (m *SysCustomer) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *SysCustomer) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (m *SysCustomer) IsEmpty() bool {
	return m == nil || m.Id == 0
}

func (l *SysCustomerList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCustomer{}).Scopes(funcs...).Find(l).Error
}

func (l *SysCustomerList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCustomer{}).Scopes(query...).Count(&count).Error
	return count, err
}

func (l *SysCustomer) UpdateCustomerStatusTrace(c context.Context) error {
	return app.DB().WithContext(c).Updates(l).Error
}
