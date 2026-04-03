package models

import (
	"context"
	"gin-fast/app/global/app"
	"time"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	CustomerExportTaskBizTypeSysCustomer = "syscustomer" // 业务类型：系统客户

	CustomerExportTaskStatusQueued   = "queued"   // 状态：排队中（待处理）
	CustomerExportTaskStatusRunning  = "running"  // 状态：执行中
	CustomerExportTaskStatusSuccess  = "success"  // 状态：成功（已完成）
	CustomerExportTaskStatusFailed   = "failed"   // 状态：失败
	CustomerExportTaskStatusCanceled = "canceled" // 状态：已取消
)

// SysCustomerExportTask 记录一次客户导出任务的完整信息，
// 包括是谁发起的、当前做到哪一步、导出的文件在哪里，以及失败原因。
type SysCustomerExportTask struct {
	ID            uint           `gorm:"primaryKey;comment:导出任务ID" json:"id"`
	TenantID      uint           `gorm:"column:tenant_id;not null;comment:所属租户ID;index:idx_customer_export_task_tenant_status,priority:1;index:idx_customer_export_task_tenant_user_status,priority:1" json:"tenantId"`
	UserID        uint           `gorm:"column:user_id;not null;comment:发起导出的用户ID;index:idx_customer_export_task_tenant_user_status,priority:2" json:"userId"`
	UserNickName  string         `gorm:"-" json:"userNickName"`
	BizType       string         `gorm:"column:biz_type;type:varchar(50);not null;default:'syscustomer';comment:任务所属业务类型;index:idx_customer_export_task_biz_status,priority:1" json:"bizType"`
	Status        string         `gorm:"column:status;type:varchar(20);not null;default:'queued';comment:当前任务状态;index:idx_customer_export_task_biz_status,priority:2;index:idx_customer_export_task_tenant_status,priority:2;index:idx_customer_export_task_tenant_user_status,priority:3" json:"status"`
	RequestJSON   string         `gorm:"column:request_json;type:longtext;not null;comment:导出请求参数快照（JSON）" json:"requestJson"`
	SnapshotMaxID int            `gorm:"column:snapshot_max_id;not null;default:0;comment:本次导出快照范围内的最大客户ID" json:"snapshotMaxId"`
	Total         int64          `gorm:"column:total;not null;default:0;comment:预计导出的数据总数" json:"total"`
	Processed     int64          `gorm:"column:processed;not null;default:0;comment:已完成处理的数据条数" json:"processed"`
	Progress      int            `gorm:"column:progress;not null;default:0;comment:当前导出进度（百分比）" json:"progress"`
	AffixID       uint           `gorm:"column:affix_id;not null;default:0;comment:导出完成后生成的附件ID" json:"affixId"`
	FileName      string         `gorm:"column:file_name;type:varchar(255);default:'';comment:导出文件名" json:"fileName"`
	ErrorMessage  string         `gorm:"column:error_message;type:varchar(1000);default:'';comment:导出失败原因" json:"errorMessage"`
	StartedAt     *time.Time     `gorm:"column:started_at;comment:任务开始执行时间" json:"startedAt"`
	FinishedAt    *time.Time     `gorm:"column:finished_at;comment:任务完成时间" json:"finishedAt"`
	CreatedAt     *time.Time     `gorm:"column:created_at;not null;comment:任务创建时间" json:"createdAt"`
	UpdatedAt     *time.Time     `gorm:"column:updated_at;not null;comment:任务最近更新时间" json:"updatedAt"`
	DeletedAt     gorm.DeletedAt `gorm:"column:deleted_at;index;comment:软删除时间" json:"deletedAt"`
}

// SysCustomerExportTaskList 用于承接批量查询出来的导出任务列表。
type SysCustomerExportTaskList []SysCustomerExportTask

func NewSysCustomerExportTask() *SysCustomerExportTask {
	return &SysCustomerExportTask{}
}

func NewSysCustomerExportTaskList() *SysCustomerExportTaskList {
	return &SysCustomerExportTaskList{}
}

func (SysCustomerExportTask) TableName() string {
	return "sys_customer_export_tasks"
}

func customerExportTaskDB(c context.Context) *gorm.DB {
	return app.DB().WithContext(c).Clauses(dbresolver.Write)
}

func (m *SysCustomerExportTask) GetByID(c context.Context, id uint) error {
	return customerExportTaskDB(c).First(m, id).Error
}

func (m *SysCustomerExportTask) Create(c context.Context) error {
	return customerExportTaskDB(c).Create(m).Error
}

func (m *SysCustomerExportTask) Update(c context.Context) error {
	return customerExportTaskDB(c).Save(m).Error
}

func (m *SysCustomerExportTask) Delete(c context.Context) error {
	return customerExportTaskDB(c).Delete(m).Error
}

func (m *SysCustomerExportTask) IsEmpty() bool {
	return m == nil || m.ID == 0
}

func (l *SysCustomerExportTaskList) Find(c context.Context, funcs ...func(*gorm.DB) *gorm.DB) error {
	return customerExportTaskDB(c).Model(&SysCustomerExportTask{}).Scopes(funcs...).Find(l).Error
}

func (l *SysCustomerExportTaskList) GetTotal(c context.Context, query ...func(*gorm.DB) *gorm.DB) (int64, error) {
	var count int64
	err := customerExportTaskDB(c).Model(&SysCustomerExportTask{}).Scopes(query...).Count(&count).Error
	return count, err
}
