package service

import (
	"gin-fast/plugins/syscallrecord/models"
	"github.com/gin-gonic/gin"
    "gorm.io/gorm"
	"gin-fast/app/utils/datascope"
	"gin-fast/app/utils/tenanthelper"
)

// SysCallRecordService sys_call_record服务
type SysCallRecordService struct{}

// NewSysCallRecordService 创建sys_call_record服务
func NewSysCallRecordService() *SysCallRecordService {
	return &SysCallRecordService{}
}

// Create 创建sys_call_record
func (s *SysCallRecordService) Create(c *gin.Context, req models.SysCallRecordCreateRequest) (*models.SysCallRecord, error) {
	// 创建sys_call_record记录
	sysCallRecord := models.NewSysCallRecord()
	// 保存到数据库
	if err := sysCallRecord.Create(c); err != nil {
		return nil, err
	}

	return sysCallRecord, nil
}

// Update 更新sys_call_record
func (s *SysCallRecordService) Update(c *gin.Context, req models.SysCallRecordUpdateRequest) error {
	// 查找sys_call_record记录
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.GetByID(c, req.Id); err != nil {
		return err
	}
	// 更新sys_call_record信息
	// 保存到数据库
	if err := sysCallRecord.Update(c); err != nil {
		return err
	}
	return nil
}

// Delete 删除sys_call_record
func (s *SysCallRecordService) Delete(c *gin.Context, id int) error {
	// 查找sys_call_record记录
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.GetByID(c, id); err != nil {
		return err
	}

	// 删除数据库记录
	if err := sysCallRecord.Delete(c); err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取sys_call_record
func (s *SysCallRecordService) GetByID(c *gin.Context, id int) (*models.SysCallRecord, error) {
	// 查找sys_call_record记录
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.GetByID(c, id); err != nil {
		return nil, err
	}

	return sysCallRecord, nil
}

// List sys_call_record列表（分页查询）
func (s *SysCallRecordService) List(c *gin.Context, req models.SysCallRecordListRequest) (*models.SysCallRecordList, int64, error) {
	// 获取总数
	sysCallRecordList := models.NewSysCallRecordList()
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle()}
	scopes = append(scopes, tenanthelper.TenantScope(c))
	total, err := sysCallRecordList.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
    scopes = append(scopes, req.Paginate())
	// 获取分页数据
	err = sysCallRecordList.Find(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	return sysCallRecordList, total, nil
}