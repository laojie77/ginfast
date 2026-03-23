package service

import (
	"gin-fast/app/global/app"
	"gin-fast/app/utils/common"
	"gin-fast/app/utils/datascope"
	"gin-fast/app/utils/tenanthelper"
	"gin-fast/plugins/syscustomer/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCustomerService sys_customer服务
type SysCustomerService struct{}

// NewSysCustomerService 创建sys_customer服务
func NewSysCustomerService() *SysCustomerService {
	return &SysCustomerService{}
}

func (s *SysCustomerService) customerTracesPreload(c *gin.Context) func(db *gorm.DB) *gorm.DB {
	claims := common.GetClaims(c)

	return func(db *gorm.DB) *gorm.DB {
		db = db.
			Select("sys_customer_traces.*, sys_users.nick_name AS user_name, sys_users.avatar AS avatar").
			Joins("LEFT JOIN sys_users ON sys_customer_traces.user_id = sys_users.id").
			Order("sys_customer_traces.created_at ASC, sys_customer_traces.id ASC")

		if claims != nil {
			return db.Where("sys_customer_traces.tenant_id = ?", claims.TenantID)
		}

		return db.Where("1 = 0")
	}
}

func (s *SysCustomerService) getScopedCustomerByID(c *gin.Context, id int) (*models.SysCustomer, error) {
	sysCustomer := models.NewSysCustomer()
	err := app.DB().WithContext(c).
		Model(&models.SysCustomer{}).
		Preload("CustomerTracesList", s.customerTracesPreload(c)).
		Scopes(datascope.GetDataScopeUser(c), tenanthelper.TenantScope(c)).
		First(sysCustomer, id).Error
	if err != nil {
		return nil, err
	}
	return sysCustomer, nil
}

// Create 创建sys_customer
func (s *SysCustomerService) Create(c *gin.Context, req models.SysCustomerCreateRequest) (*models.SysCustomer, error) {
	// 创建sys_customer记录
	sysCustomer := models.NewSysCustomer()
	sysCustomer.Name = req.Name
	sysCustomer.Mobile = req.Mobile
	sysCustomer.MoneyDemand = req.MoneyDemand
	sysCustomer.ChannelID = req.ChannelID
	sysCustomer.CustomerStar = req.CustomerStar
	sysCustomer.Status = req.Status
	sysCustomer.Intention = req.Intention
	sysCustomer.Extra = req.Extra
	sysCustomer.Sex = req.Sex
	sysCustomer.Remarks = req.Remarks
	sysCustomer.Age = req.Age
	if currentUserID := common.GetCurrentUserID(c); currentUserID > 0 {
		sysCustomer.UserID = int(currentUserID)
	}
	// 保存到数据库
	if err := sysCustomer.Create(c); err != nil {
		return nil, err
	}

	return sysCustomer, nil
}

// Update 更新sys_customer
func (s *SysCustomerService) Update(c *gin.Context, req models.SysCustomerUpdateRequest) error {
	// 查找sys_customer记录
	sysCustomer, err := s.getScopedCustomerByID(c, req.Id)
	if err != nil {
		return err
	}
	// 更新sys_customer信息
	sysCustomer.Name = req.Name
	sysCustomer.Mobile = req.Mobile
	sysCustomer.MoneyDemand = req.MoneyDemand
	sysCustomer.ChannelID = req.ChannelID
	sysCustomer.CustomerStar = req.CustomerStar
	sysCustomer.Status = req.Status
	sysCustomer.Intention = req.Intention
	sysCustomer.Extra = req.Extra
	sysCustomer.Sex = req.Sex
	sysCustomer.Remarks = req.Remarks
	sysCustomer.Age = req.Age
	sysCustomer.IsLock = req.IsLock
	sysCustomer.SinglePieceType = req.SinglePieceType
	// 保存到数据库
	if err := sysCustomer.Update(c); err != nil {
		return err
	}
	return nil
}

// Delete 删除sys_customer
func (s *SysCustomerService) Delete(c *gin.Context, id int) error {
	// 查找sys_customer记录
	sysCustomer, err := s.getScopedCustomerByID(c, id)
	if err != nil {
		return err
	}

	// 删除数据库记录
	if err := sysCustomer.Delete(c); err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取sys_customer
func (s *SysCustomerService) GetByID(c *gin.Context, id int) (*models.SysCustomer, error) {
	return s.getScopedCustomerByID(c, id)
}

// List sys_customer列表（分页查询）
func (s *SysCustomerService) List(c *gin.Context, req models.SysCustomerListRequest) (*models.SysCustomerList, int64, error) {
	// 获取总数
	sysCustomerList := models.NewSysCustomerList()
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle()}
	scopes = append(scopes, datascope.GetDataScopeUser(c), tenanthelper.TenantScope(c))
	total, err := sysCustomerList.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	scopes = append(scopes, req.Paginate())
	// 获取分页数据
	err = app.DB().WithContext(c).
		Model(&models.SysCustomer{}).
		Preload("CustomerTracesList", s.customerTracesPreload(c)).
		Scopes(scopes...).
		Find(sysCustomerList).Error
	if err != nil {
		return nil, 0, err
	}

	return sysCustomerList, total, nil
}
