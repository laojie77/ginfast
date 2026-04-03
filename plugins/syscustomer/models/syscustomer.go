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
	Id                 int                             `gorm:"column:id;primaryKey;autoIncrement;comment:主键ID" json:"id"`                                                           // 主键ID
	Num                string                          `gorm:"column:num;size:18;comment:客户编号" json:"num"`                                                                          // 客户编号
	Name               string                          `gorm:"column:name;size:80;comment:客户姓名" json:"name"`                                                                        // 客户姓名
	Mobile             string                          `gorm:"column:mobile;size:20;index:idx_mobile;comment:手机号" json:"mobile"`                                                    // 手机号
	MdMobile           string                          `gorm:"column:md5_mobile;size:32;index:idx_md5_mobile;comment:手机号MD5值，用于隐私查询" json:"mdMobile"`                               // 手机号MD5值，用于隐私查询
	MoneyDemand        int                             `gorm:"column:money_demand;default:0;comment:需求金额（单位：万元）" json:"moneyDemand"`                                                // 需求金额（单位：万元）
	ChannelID          int                             `gorm:"column:channel_id;index:idx_channel_id;comment:渠道来源ID" json:"channelId"`                                              // 渠道来源ID
	UserID             int                             `gorm:"column:user_id;index:idx_user_id;comment:所属销售员ID" json:"userId"`                                                      // 所属销售员ID
	CustomerStar       *int                            `gorm:"column:customer_star;comment:星级：0-5（0表示无星级）" json:"customerStar"`                                                     // 星级：0-5（0表示无星级）
	Status             int                             `gorm:"column:status;not null;default:0;index:idx_status;comment:状态：0未受理，1待跟进，2预约上门，3待签约，4已签约，5已进件，6已放款，7已结算" json:"status"` // 状态
	Intention          int                             `gorm:"column:intention;not null;default:0;comment:客户有效:0未确认1待确认2有效客户3无效客户4黑名单" json:"intention"`                            // 客户有效
	From               int                             `gorm:"column:from;not null;default:1;comment:客户来源:1系统分配，2再分配，3手动创建" json:"from"`                                            // 客户来源
	LastTime           *time.Time                      `gorm:"column:last_time;comment:最后操作时间" json:"lastTime"`                                                                     // 最后操作时间
	SinglePieceType    int                             `gorm:"column:single_piece_type;comment:贷款类型：1房抵贷，2车抵贷，3信用贷，4保单贷" json:"singlePieceType"`                                    // 贷款类型
	Sex                int                             `gorm:"column:sex;not null;default:2;comment:性别：1男，0女，2未知" json:"sex"`                                                       // 性别
	AllotTime          *time.Time                      `gorm:"column:allot_time;comment:分配时间（初次分配）" json:"allotTime"`                                                               // 分配时间（初次分配）
	DeptID             int                             `gorm:"column:dept_id;index:idx_department_id;comment:当前所属部门ID" json:"deptId"`                                               // 当前所属部门ID
	TenantID           int                             `gorm:"column:tenant_id;comment:所属公司ID" json:"tenantId"`                                                                     // 所属公司ID
	Remarks            string                          `gorm:"column:remarks;type:text;comment:客户备注" json:"remarks"`                                                                // 客户备注
	Age                int                             `gorm:"column:age;comment:年龄" json:"age"`                                                                                    // 年龄
	City               string                          `gorm:"column:city;size:45;comment:所在城市" json:"city"`                                                                        // 所在城市
	CustomerComment    string                          `gorm:"column:customer_comment;type:longtext;comment:上级评价" json:"customerComment"`                                           // 上级评价
	Source             string                          `gorm:"column:source;size:15;comment:渠道编码" json:"source"`                                                                    // 渠道编码
	NewData            int                             `gorm:"column:new_data;comment:再分配时是否为新数据：1是，0否" json:"newData"`                                                             // 再分配时是否为新数据
	RedistributionTime *time.Time                      `gorm:"column:redistribution_time;comment:再分配时间" json:"redistributionTime"`                                                  // 再分配时间
	IsReassign         int                             `gorm:"column:is_reassign;default:0;comment:是否再分配：1是" json:"isReassign"`                                                     // 是否再分配
	BatchId            int                             `gorm:"column:batch_id;comment:导入批次ID" json:"batchId"`                                                                       // 导入批次ID
	IsRead             int                             `gorm:"column:is_read;default:0;comment:是否已读：1是" json:"isRead"`                                                              // 是否已读
	IsPublic           int                             `gorm:"column:is_public;default:0;comment:是否在公共池：1是" json:"isPublic"`                                                        // 是否在公共池
	IsQuit             int                             `gorm:"column:is_quit;default:0;comment:是否离职数据：1是" json:"isQuit"`                                                            // 是否离职数据
	IsRepeat           int                             `gorm:"column:is_repeat;comment:重复标记：1重复，2黑名单" json:"isRepeat"`                                                              // 重复标记
	IsRubbish          int                             `gorm:"column:is_rubbish;default:0;comment:是否垃圾库：1是" json:"isRubbish"`                                                       // 是否垃圾库
	DispatchTime       *time.Time                      `gorm:"column:dispatch_time;comment:待调度时间" json:"dispatchTime"`                                                              // 待调度时间
	IsRemind           int                             `gorm:"column:is_remind;default:0;comment:是否提醒：1是，0否" json:"isRemind"`                                                       // 是否提醒
	IsSms              int                             `gorm:"column:is_sms;default:0;comment:短信发送状态：0未发送，1已发送" json:"isSms"`                                                       // 短信发送状态
	StarStatus         int                             `gorm:"column:star_status;not null;default:0;comment:星级是否已回传：1是，0否，2未回传" json:"starStatus"`                                  // 星级是否已回传
	IsExchange         int                             `gorm:"column:is_exchange;default:0;comment:待流转标记：1待流转" json:"isExchange"`                                                   // 待流转标记
	PublicTypeTime     *time.Time                      `gorm:"column:public_type_time;comment:抓取公共池时间" json:"publicTypeTime"`                                                       // 抓取公共池时间
	PublicType         int                             `gorm:"column:public_type;comment:抓取类型：0手动，1自动" json:"publicType"`                                                           // 抓取类型
	IsLock             int                             `gorm:"column:is_lock;not null;default:0;comment:是否锁定：1已锁定" json:"isLock"`                                                   // 是否锁定
	Extra              string                          `gorm:"column:extra;type:json;comment:扩展字段" json:"extra"`                                                                    // 扩展字段
	CreatedAt          *time.Time                      `gorm:"column:created_at;precision:3" json:"createdAt"`                                                                      // 创建时间
	UpdatedAt          *time.Time                      `gorm:"column:updated_at;precision:3" json:"updatedAt"`                                                                      // 更新时间
	DeletedAt          gorm.DeletedAt                  `gorm:"column:deleted_at;precision:3" json:"deletedAt"`                                                                      // 删除时间
	CustomerTracesList []traceModels.SysCustomerTraces `gorm:"foreignKey:CustomerID;references:Id" json:"customerTracesList"`                                                       // 客户跟进记录
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
