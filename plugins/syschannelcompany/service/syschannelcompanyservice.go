package service

import (
	"gin-fast/app/utils/tenanthelper"
	"gin-fast/plugins/syschannelcompany/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysChannelCompanyService sys_channel_company服务
type SysChannelCompanyService struct{}

// NewSysChannelCompanyService 创建sys_channel_company服务
func NewSysChannelCompanyService() *SysChannelCompanyService {
	return &SysChannelCompanyService{}
}

// Create 创建sys_channel_company
func (s *SysChannelCompanyService) Create(c *gin.Context, req models.SysChannelCompanyCreateRequest) (*models.SysChannelCompany, error) {
	// 创建sys_channel_company记录
	sysChannelCompany := models.NewSysChannelCompany()
	sysChannelCompany.ChannelId = req.ChannelId
	sysChannelCompany.TenantId = req.TenantId
	sysChannelCompany.City = req.City
	sysChannelCompany.HiddenName = req.HiddenName
	sysChannelCompany.IsStar = req.IsStar
	sysChannelCompany.FieldMappings = req.FieldMappings
	// 保存到数据库
	if err := sysChannelCompany.Create(c); err != nil {
		return nil, err
	}

	return sysChannelCompany, nil
}

// Update 更新sys_channel_company
func (s *SysChannelCompanyService) Update(c *gin.Context, req models.SysChannelCompanyUpdateRequest) error {
	// 查找sys_channel_company记录
	sysChannelCompany := models.NewSysChannelCompany()
	if err := sysChannelCompany.GetByID(c, req.Id); err != nil {
		return err
	}
	// 更新sys_channel_company信息
	sysChannelCompany.ChannelId = req.ChannelId
	sysChannelCompany.TenantId = req.TenantId
	sysChannelCompany.City = req.City
	sysChannelCompany.HiddenName = req.HiddenName
	sysChannelCompany.IsStar = req.IsStar
	sysChannelCompany.FieldMappings = req.FieldMappings
	// 保存到数据库
	if err := sysChannelCompany.Update(c); err != nil {
		return err
	}
	return nil
}

// Delete 删除sys_channel_company
func (s *SysChannelCompanyService) Delete(c *gin.Context, id int) error {
	// 查找sys_channel_company记录
	sysChannelCompany := models.NewSysChannelCompany()
	if err := sysChannelCompany.GetByID(c, id); err != nil {
		return err
	}

	// 删除数据库记录
	if err := sysChannelCompany.Delete(c); err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取sys_channel_company
func (s *SysChannelCompanyService) GetByID(c *gin.Context, id int) (*models.SysChannelCompany, error) {
	// 查找sys_channel_company记录
	sysChannelCompany := models.NewSysChannelCompany()
	if err := sysChannelCompany.GetByID(c, id); err != nil {
		return nil, err
	}

	return sysChannelCompany, nil
}

// List sys_channel_company列表（分页查询）
func (s *SysChannelCompanyService) List(c *gin.Context, req models.SysChannelCompanyListRequest) (*models.SysChannelCompanyList, int64, error) {
	// 获取总数
	sysChannelCompanyList := models.NewSysChannelCompanyList()
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle()}
	scopes = append(scopes, tenanthelper.TenantScope(c))
	total, err := sysChannelCompanyList.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	scopes = append(scopes, req.Paginate())
	// 获取分页数据
	err = sysChannelCompanyList.Find(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	return sysChannelCompanyList, total, nil
}
