package service

import (
	"fmt"
	"gin-fast/app/utils/common"
	"gin-fast/plugins/syschannel/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysChannelService sys_channel服务
type SysChannelService struct{}

// NewSysChannelService 创建sys_channel服务
func NewSysChannelService() *SysChannelService {
	return &SysChannelService{}
}

// Create 创建sys_channel
func (s *SysChannelService) Create(c *gin.Context, req models.SysChannelCreateRequest) (*models.SysChannel, error) {
	// 创建sys_channel记录
	sysChannel := models.NewSysChannel()
	sysChannel.ChannelName = req.ChannelName
	sysChannel.ChannelKey = req.ChannelKey
	secretKey, err := common.GenerateRandomKey(16)
	if err != nil {
		return nil, fmt.Errorf("生成密钥失败: %v", err)
	}
	sysChannel.SecretKey = secretKey
	sysChannel.StarUrl = req.StarUrl
	sysChannel.Remark = req.Remark
	sysChannel.Status = req.Status
	sysChannel.StartTime = req.StartTime
	sysChannel.EndTime = req.EndTime
	sysChannel.Type = req.Type
	sysChannel.MessageType = req.MessageType
	sysChannel.IsRepeat = req.IsRepeat
	sysChannel.SuccessCode = req.SuccessCode
	sysChannel.SuccessCodeField = req.SuccessCodeField
	// 保存到数据库
	if err := sysChannel.Create(c); err != nil {
		return nil, err
	}

	return sysChannel, nil
}

// Update 更新sys_channel
func (s *SysChannelService) Update(c *gin.Context, req models.SysChannelUpdateRequest) error {
	// 查找sys_channel记录
	sysChannel := models.NewSysChannel()
	if err := sysChannel.GetByID(c, req.Id); err != nil {
		return err
	}

	// 更新sys_channel信息
	sysChannel.ChannelName = req.ChannelName
	sysChannel.ChannelKey = req.ChannelKey
	sysChannel.HiddenCode = req.HiddenCode
	sysChannel.StarUrl = req.StarUrl
	sysChannel.Remark = req.Remark
	sysChannel.Status = req.Status
	sysChannel.StartTime = req.StartTime
	sysChannel.EndTime = req.EndTime
	sysChannel.Type = req.Type
	sysChannel.MessageType = req.MessageType
	sysChannel.IsRepeat = req.IsRepeat
	sysChannel.SuccessCode = req.SuccessCode
	sysChannel.SuccessCodeField = req.SuccessCodeField

	// 保存到数据库
	if err := sysChannel.Update(c); err != nil {
		return fmt.Errorf("更新数据库失败: %w", err)
	}
	return nil
}

// Delete 删除sys_channel
func (s *SysChannelService) Delete(c *gin.Context, id int) error {
	// 查找sys_channel记录
	sysChannel := models.NewSysChannel()
	if err := sysChannel.GetByID(c, id); err != nil {
		return err
	}

	// 删除数据库记录
	if err := sysChannel.Delete(c); err != nil {
		return err
	}

	return nil
}

// GetByID 根据ID获取sys_channel
func (s *SysChannelService) GetByID(c *gin.Context, id int) (*models.SysChannel, error) {
	// 查找sys_channel记录
	sysChannel := models.NewSysChannel()
	if err := sysChannel.GetByID(c, id); err != nil {
		return nil, err
	}

	return sysChannel, nil
}

// List sys_channel列表（分页查询）
func (s *SysChannelService) List(c *gin.Context, req models.SysChannelListRequest) (*models.SysChannelList, int64, error) {
	// 获取总数
	sysChannelList := models.NewSysChannelList()
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle()}
	total, err := sysChannelList.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	scopes = append(scopes, req.Paginate())
	// 预加载城市配置列表
	scopes = append(scopes, func(db *gorm.DB) *gorm.DB {
		return db.Preload("ChannelCompanyList")
	})
	// 获取分页数据
	err = sysChannelList.Find(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	return sysChannelList, total, nil
}
