package service

import (
	"gin-fast/app/utils/common"
	"gin-fast/plugins/syscallrecord/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysCallRecordService struct{}

func NewSysCallRecordService() *SysCallRecordService {
	return &SysCallRecordService{}
}

func syscallRecordTenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		claims := common.GetClaims(c)
		if claims != nil {
			return db.Where("sys_call_record.tenant_id = ?", claims.TenantID)
		}
		return db.Where("1 = 0")
	}
}

func (s *SysCallRecordService) Create(c *gin.Context, req models.SysCallRecordCreateRequest) (*models.SysCallRecord, error) {
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.Create(c); err != nil {
		return nil, err
	}

	return sysCallRecord, nil
}

func (s *SysCallRecordService) Update(c *gin.Context, req models.SysCallRecordUpdateRequest) error {
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.GetByID(c, req.Id); err != nil {
		return err
	}
	if err := sysCallRecord.Update(c); err != nil {
		return err
	}
	return nil
}

func (s *SysCallRecordService) Delete(c *gin.Context, id int) error {
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.GetByID(c, id); err != nil {
		return err
	}

	if err := sysCallRecord.Delete(c); err != nil {
		return err
	}

	return nil
}

func (s *SysCallRecordService) GetByID(c *gin.Context, id int) (*models.SysCallRecord, error) {
	sysCallRecord := models.NewSysCallRecord()
	if err := sysCallRecord.GetByIDWithCustomer(c, id, syscallRecordTenantScope(c)); err != nil {
		return nil, err
	}

	return sysCallRecord, nil
}

func (s *SysCallRecordService) List(c *gin.Context, req models.SysCallRecordListRequest) (*models.SysCallRecordList, int64, error) {
	sysCallRecordList := models.NewSysCallRecordList()
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle(), syscallRecordTenantScope(c)}
	total, err := sysCallRecordList.GetTotalWithCustomer(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	scopes = append(scopes, req.Paginate())
	if err = sysCallRecordList.FindWithCustomer(c, scopes...); err != nil {
		return nil, 0, err
	}

	return sysCallRecordList, total, nil
}
