package models

import (
	"context"
	"time"

	"gin-fast/app/global/app"

	"gorm.io/gorm"
	"gorm.io/plugin/dbresolver"
)

const (
	CustomerImportBatchStatusPending   = "pending"
	CustomerImportBatchStatusRunning   = "running"
	CustomerImportBatchStatusCanceling = "canceling"
	CustomerImportBatchStatusCanceled  = "canceled"
	CustomerImportBatchStatusSuccess   = "success"
	CustomerImportBatchStatusPartial   = "partial"
	CustomerImportBatchStatusFailed    = "failed"
)

type SysCustomerImportBatch struct {
	ID               int            `gorm:"column:id;primaryKey;autoIncrement;comment:导入批次ID" json:"id"`
	TenantID         int            `gorm:"column:tenant_id;not null;default:0;index:idx_customer_import_batch_tenant_created,priority:1;comment:租户ID" json:"tenantId"`
	UserID           int            `gorm:"column:user_id;not null;default:0;index:idx_customer_import_batch_tenant_user,priority:2;comment:导入用户ID" json:"userId"`
	DeptID           int            `gorm:"column:dept_id;not null;default:0;comment:导入部门ID" json:"deptId"`
	Scene            string         `gorm:"column:scene;type:varchar(20);not null;default:'my';comment:导入场景" json:"scene"`
	ChannelCompanyID int            `gorm:"column:channel_company_id;not null;default:0;comment:渠道公司ID" json:"channelCompanyId"`
	ChannelID        int            `gorm:"column:channel_id;not null;default:0;comment:渠道主表ID" json:"channelId"`
	ChannelKey       string         `gorm:"column:channel_key;type:varchar(100);not null;default:'';comment:渠道编码" json:"channelKey"`
	ChannelName      string         `gorm:"column:channel_name;type:varchar(150);not null;default:'';comment:渠道展示名称" json:"channelName"`
	StartRow         int            `gorm:"column:start_row;not null;default:2;comment:导入起始行号" json:"startRow"`
	Remark           string         `gorm:"column:remark;type:varchar(255);comment:批次备注" json:"remark"`
	FileAffixID      uint           `gorm:"column:file_affix_id;not null;default:0;comment:导入文件附件ID" json:"fileAffixId"`
	FileName         string         `gorm:"column:file_name;type:varchar(255);not null;default:'';comment:导入文件名" json:"fileName"`
	FileURL          string         `gorm:"column:file_url;type:varchar(255);not null;default:'';comment:导入文件URL" json:"fileUrl"`
	FilePath         string         `gorm:"column:file_path;type:varchar(255);not null;default:'';comment:导入文件存储路径" json:"filePath"`
	FileDeletedAt    *time.Time     `gorm:"column:file_deleted_at;comment:导入文件清理时间" json:"fileDeletedAt"`
	TotalCount       int            `gorm:"column:total_count;not null;default:0;comment:总行数" json:"totalCount"`
	ProcessedCount   int            `gorm:"column:processed_count;not null;default:0;comment:已处理行数" json:"processedCount"`
	SuccessCount     int            `gorm:"column:success_count;not null;default:0;comment:成功行数" json:"successCount"`
	FailedCount      int            `gorm:"column:failed_count;not null;default:0;comment:失败行数" json:"failedCount"`
	DuplicateCount   int            `gorm:"column:duplicate_count;not null;default:0;comment:重复数据" json:"duplicateCount"`
	Progress         int            `gorm:"column:progress;not null;default:0;comment:当前进度百分比" json:"progress"`
	ResumeRow        int            `gorm:"column:resume_row;not null;default:0;comment:建议继续导入的行号" json:"resumeRow"`
	Status           string         `gorm:"column:status;type:varchar(20);not null;default:'pending';index:idx_customer_import_batch_tenant_user,priority:3;comment:导入状态" json:"status"`
	ErrorMessage     string         `gorm:"column:error_message;type:varchar(1000);not null;default:'';comment:失败原因" json:"errorMessage"`
	FailurePreview   string         `gorm:"column:failure_preview;type:longtext;not null;comment:失败预览JSON" json:"-"`
	StartedAt        *time.Time     `gorm:"column:started_at;comment:导入开始时间" json:"startedAt"`
	FinishedAt       *time.Time     `gorm:"column:finished_at;comment:导入结束时间" json:"finishedAt"`
	CreatedAt        *time.Time     `gorm:"column:created_at;precision:3" json:"createdAt"`
	UpdatedAt        *time.Time     `gorm:"column:updated_at;precision:3" json:"updatedAt"`
	DeletedAt        gorm.DeletedAt `gorm:"column:deleted_at;precision:3;index" json:"deletedAt"`
}

func (SysCustomerImportBatch) TableName() string {
	return "sys_customer_import_batches"
}

func customerImportBatchDB(c context.Context) *gorm.DB {
	return app.DB().WithContext(c).Clauses(dbresolver.Write)
}

func (m *SysCustomerImportBatch) Create(c context.Context) error {
	return customerImportBatchDB(c).Create(m).Error
}

func (m *SysCustomerImportBatch) Update(c context.Context) error {
	return customerImportBatchDB(c).Save(m).Error
}
