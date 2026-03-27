package models

import (
	"context"
	"time"

	"gin-fast/app/global/app"
	"gorm.io/gorm"
)

type SysCallRecord struct {
	Id          int        `gorm:"column:id;primaryKey;not null;autoIncrement" json:"id"`
	UserId      int        `gorm:"column:user_id" json:"userId"`
	UserName    string     `gorm:"column:user_name" json:"userName"`
	CustomerID  int        `gorm:"column:customer_id" json:"customerId"`
	CustomerNum string     `gorm:"column:customer_num;->" json:"customer_id"`
	Mobile      string     `gorm:"column:mobile" json:"mobile"`
	Status      int        `gorm:"column:status;default:0;index" json:"status"`
	Type        int        `gorm:"column:type" json:"type"`
	StartTime   *time.Time `gorm:"column:start_time" json:"startTime"`
	EndTime     *time.Time `gorm:"column:end_time" json:"endTime"`
	Vdieo       string     `gorm:"column:vdieo" json:"vdieo"`
	Duration    int        `gorm:"column:duration" json:"duration"`
	Version     string     `gorm:"column:version" json:"version"`
	CreatedAt   *time.Time `gorm:"column:created_at" json:"createdAt"`
	TenantId    uint       `gorm:"column:tenant_id;default:0" json:"tenantId"`
}

type SysCallRecordList []SysCallRecord

func NewSysCallRecord() *SysCallRecord {
	return &SysCallRecord{}
}

func NewSysCallRecordList() *SysCallRecordList {
	return &SysCallRecordList{}
}

func (SysCallRecord) TableName() string {
	return "sys_call_record"
}

func (m *SysCallRecord) GetByID(c context.Context, id int) error {
	return app.DB().WithContext(c).First(m, id).Error
}

func (m *SysCallRecord) GetByIDWithCustomer(c context.Context, id int, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).
		Table("sys_call_record").
		Select("sys_call_record.*, sys_customer.num AS customer_num").
		Joins("LEFT JOIN sys_customer ON sys_call_record.customer_id = sys_customer.id").
		Scopes(funcs...).
		Where("sys_call_record.id = ?", id).
		First(m).Error
}

func (m *SysCallRecord) Create(c context.Context) error {
	return app.DB().WithContext(c).Create(m).Error
}

func (m *SysCallRecord) Update(c context.Context) error {
	return app.DB().WithContext(c).Save(m).Error
}

func (m *SysCallRecord) Delete(c context.Context) error {
	return app.DB().WithContext(c).Delete(m).Error
}

func (m *SysCallRecord) IsEmpty() bool {
	return m == nil || m.Id == 0
}

func (l *SysCallRecordList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).Model(&SysCallRecord{}).Scopes(funcs...).Find(l).Error
}

func (l *SysCallRecordList) FindWithCustomer(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return app.DB().WithContext(c).
		Table("sys_call_record").
		Select("sys_call_record.*, sys_customer.num AS customer_num").
		Joins("LEFT JOIN sys_customer ON sys_call_record.customer_id = sys_customer.id").
		Scopes(funcs...).
		Find(l).Error
}

func (l *SysCallRecordList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).Model(&SysCallRecord{}).Scopes(query...).Count(&count).Error
	return count, err
}

func (l *SysCallRecordList) GetTotalWithCustomer(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := app.DB().WithContext(c).
		Table("sys_call_record").
		Joins("LEFT JOIN sys_customer ON sys_call_record.customer_id = sys_customer.id").
		Scopes(query...).
		Count(&count).Error
	return count, err
}
