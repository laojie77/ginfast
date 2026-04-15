package models

import (
	"errors"
	"strings"
	"time"

	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
)

type SysCustomerImportRequest struct {
	models.Validator
	Scene            string `form:"scene" validate:"required" message:"导入场景不能为空"`
	ChannelCompanyID int    `form:"channelCompanyId" validate:"required" message:"请选择导入渠道"`
	StartRow         int    `form:"startRow"`
	Remark           string `form:"remark"  message:"批次备注不能超过255个字符"`
}

func (r *SysCustomerImportRequest) Validate(c *gin.Context) error {
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}

	r.Scene = strings.ToLower(strings.TrimSpace(r.Scene))
	r.Remark = strings.TrimSpace(r.Remark)
	if r.StartRow < 2 {
		r.StartRow = 2
	}

	return nil
}

const CustomerImportActionCancel = "cancel"

type SysCustomerImportCancelRequest struct {
	models.Validator
	Action  string `json:"action" form:"action" validate:"required" message:"导入操作不能为空"`
	BatchID int    `json:"batchId" form:"batchId" validate:"required" message:"导入批次不能为空"`
}

func (r *SysCustomerImportCancelRequest) Validate(c *gin.Context) error {
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}

	r.Action = strings.ToLower(strings.TrimSpace(r.Action))
	if r.Action != CustomerImportActionCancel {
		return errors.New("不支持的导入操作")
	}
	if r.BatchID <= 0 {
		return errors.New("导入批次不能为空")
	}

	return nil
}

type SysCustomerImportFailure struct {
	Row     int    `json:"row"`
	Name    string `json:"name,omitempty"`
	Mobile  string `json:"mobile,omitempty"`
	Message string `json:"message"`
}

type SysCustomerImportSubmitResult struct {
	BatchID  int    `json:"batchId"`
	Status   string `json:"status"`
	Existing bool   `json:"existing,omitempty"`
	Message  string `json:"message,omitempty"`
}

type SysCustomerImportBatchResult struct {
	ID             int                        `json:"id"`
	Status         string                     `json:"status"`
	StartRow       int                        `json:"startRow"`
	ResumeRow      int                        `json:"resumeRow,omitempty"`
	Interrupted    bool                       `json:"interrupted"`
	TotalCount     int                        `json:"totalCount"`
	ProcessedCount int                        `json:"processedCount"`
	SuccessCount   int                        `json:"successCount"`
	FailedCount    int                        `json:"failedCount"`
	DuplicateCount int                        `json:"duplicateCount"`
	Progress       int                        `json:"progress"`
	Remark         string                     `json:"remark,omitempty"`
	ErrorMessage   string                     `json:"errorMessage,omitempty"`
	Failures       []SysCustomerImportFailure `json:"failures,omitempty"`
	StartedAt      *time.Time                 `json:"startedAt,omitempty"`
	FinishedAt     *time.Time                 `json:"finishedAt,omitempty"`
	UpdatedAt      *time.Time                 `json:"updatedAt,omitempty"`
	FileName       string                     `json:"fileName,omitempty"`
}
