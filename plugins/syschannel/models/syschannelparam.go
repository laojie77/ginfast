package models

import (
	"gin-fast/app/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysChannelListRequest sys_channel列表请求参数
type SysChannelListRequest struct {
	models.BasePaging
	models.Validator
	ChannelName *string `form:"channelName"` // 渠道名称
	ChannelKey  *string `form:"channelKey"`  // 渠道码
	HiddenCode  *string `form:"hiddenCode"`  // 渠道隐藏名称
	Status      *int    `form:"status"`      // 状态
}

// Validate 验证请求参数
func (r *SysChannelListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle 获取查询条件
func (r *SysChannelListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.ChannelName != nil {
			db = db.Where("channel_name LIKE ?", "%"+*r.ChannelName+"%")
		}
		if r.ChannelKey != nil {
			db = db.Where("channel_key = ?", *r.ChannelKey)
		}
		if r.HiddenCode != nil {
			db = db.Where("hidden_code = ?", *r.HiddenCode)
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		return db
	}
}

// SysChannelCreateRequest 创建sys_channel请求参数
type SysChannelCreateRequest struct {
	models.Validator
	ChannelName      string `form:"channelName" validate:"required" message:"渠道名称不能为空"` // 渠道名称
	ChannelKey       string `form:"channelKey" validate:"required" message:"渠道码不能为空"`   // 渠道码
	HiddenCode       string `form:"hiddenCode"`                                         // 渠道隐藏名称
	StarUrl          string `form:"starUrl"`                                            // 星级回传地址
	Remark           string `form:"remark"`                                             // 备注
	Status           int    `form:"status"`                                             // 状态
	StartTime        string `form:"startTime" validate:"required" message:"开始时间不能为空"`   // 开始时间
	EndTime          string `form:"endTime" validate:"required" message:"结束时间不能为空"`     // 结束时间
	Type             int    `form:"type"`                                               // 数据类型
	MessageType      int    `form:"messageType"`                                        // 短信
	IsRepeat         int    `form:"isRepeat"`                                           // 是否查重
	SuccessCode      string `form:"successCode"`                                        // 成功返回码
	SuccessCodeField string `form:"successCodeField"`                                   // 成功码字段名
}

// Validate 验证请求参数
func (r *SysChannelCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysChannelUpdateRequest 更新sys_channel请求参数
type SysChannelUpdateRequest struct {
	models.Validator
	Id               int    `form:"id" validate:"required" message:"Id不能为空"`            // Id
	ChannelName      string `form:"channelName" validate:"required" message:"渠道名称不能为空"` // 渠道名称
	ChannelKey       string `form:"channelKey" validate:"required" message:"渠道码不能为空"`   // 渠道码
	HiddenCode       string `form:"hiddenCode"`                                         // 渠道隐藏名称
	Institution      string `form:"institution"`                                        // 机构ID
	StarUrl          string `form:"starUrl"`                                            // 星级回传地址
	Remark           string `form:"remark"`                                             // 备注
	Status           int    `form:"status"`                                             // 状态
	StartTime        string `form:"startTime" validate:"required" message:"开始时间不能为空"`   // 开始时间
	EndTime          string `form:"endTime" validate:"required" message:"结束时间不能为空"`     // 结束时间
	Type             int    `form:"type"`                                               // 数据类型
	MessageType      int    `form:"messageType"`                                        // 短信
	IsOur            uint   `form:"isOur"`                                              // 是否自营
	IsRepeat         int    `form:"isRepeat"`                                           // 是否查重
	SuccessCode      string `form:"successCode"`                                        // 成功返回码
	SuccessCodeField string `form:"successCodeField"`                                   // 成功码字段名
}

// Validate 验证请求参数
func (r *SysChannelUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysChannelDeleteRequest 删除sys_channel请求参数
type SysChannelDeleteRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysChannelDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysChannelGetByIDRequest 根据ID获取sys_channel请求参数
type SysChannelGetByIDRequest struct {
	models.Validator
	Id int `uri:"id" validate:"required" message:"Id不能为空"` // Id
}

// Validate 验证请求参数
func (r *SysChannelGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
