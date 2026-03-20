package service

import (
	"gin-fast/app/utils/common"
	"gin-fast/plugins/syscustomertraces/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCustomerTracesService sys_customer_traces服务
type SysCustomerTracesService struct{}

func sysCustomerTracesTenantScope(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		claims := common.GetClaims(c)
		if claims != nil {
			return db.Where("sys_customer_traces.tenant_id = ?", claims.TenantID)
		}
		return db.Where("1 = 0")
	}
}

// NewSysCustomerTracesService 创建sys_customer_traces服务
func NewSysCustomerTracesService() *SysCustomerTracesService {
	return &SysCustomerTracesService{}
}

// Create 创建sys_customer_traces
func (s *SysCustomerTracesService) Create(c *gin.Context, req models.SysCustomerTracesCreateRequest) (*models.SysCustomerTraces, error) {
	// 创建sys_customer_traces记录
	sysCustomerTraces := models.NewSysCustomerTraces()
	sysCustomerTraces.CustomerId = req.CustomerId
	sysCustomerTraces.UserId = req.UserId
	sysCustomerTraces.Data = req.Data

	// 保存到数据库
	if err := sysCustomerTraces.Create(c); err != nil {
		return nil, err
	}

	return sysCustomerTraces, nil
}

// Update 更新sys_customer_traces
func (s *SysCustomerTracesService) Update(c *gin.Context, req models.SysCustomerTracesUpdateRequest) error {
	// 查找sys_customer_traces记录
	sysCustomerTraces := models.NewSysCustomerTraces()
	if err := sysCustomerTraces.GetByID(c, req.Id); err != nil {
		return err
	}

	// 更新sys_customer_traces信息
	if req.CustomerId != nil {
		sysCustomerTraces.CustomerId = *req.CustomerId
	}
	if req.UserId != nil {
		sysCustomerTraces.UserId = *req.UserId
	}
	if req.Data != nil {
		sysCustomerTraces.Data = *req.Data
	}

	// 保存到数据库
	if err := sysCustomerTraces.Update(c); err != nil {
		return err
	}
	return nil
}

// Delete 删除sys_customer_traces
func (s *SysCustomerTracesService) Delete(c *gin.Context, id uint64) error {
	// 查找sys_customer_traces记录
	sysCustomerTraces := models.NewSysCustomerTraces()
	if err := sysCustomerTraces.GetByID(c, id); err != nil {
		return err
	}

	// 删除数据库记录
	if err := sysCustomerTraces.Delete(c); err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取sys_customer_traces
func (s *SysCustomerTracesService) GetByID(c *gin.Context, id uint64) (*models.SysCustomerTraces, error) {
	// 查找sys_customer_traces记录
	sysCustomerTraces := models.NewSysCustomerTraces()
	if err := sysCustomerTraces.GetByID(c, id); err != nil {
		return nil, err
	}

	return sysCustomerTraces, nil
}

// List sys_customer_traces列表（分页查询）
func (s *SysCustomerTracesService) List(c *gin.Context, req models.SysCustomerTracesListRequest) (*models.SysCustomerTracesWithUserList, int64, error) {
	// 获取总数
	sysCustomerTracesWithUserList := models.NewSysCustomerTracesWithUserList()
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle()}
	scopes = append(scopes, sysCustomerTracesTenantScope(c))
	total, err := sysCustomerTracesWithUserList.GetTotalWithUser(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	scopes = append(scopes, req.Paginate())
	// 获取分页数据
	err = sysCustomerTracesWithUserList.FindWithUser(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	return sysCustomerTracesWithUserList, total, nil
}
