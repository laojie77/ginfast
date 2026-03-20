package models

import (
	"gin-fast/app/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Extra字段属性定义常量
const (
	ExtraPropertyOccupation      = "occupation"
	ExtraPropertyHouse           = "house"
	ExtraPropertyCar             = "car"
	ExtraPropertyCreditCard      = "creditcard"
	ExtraPropertyInsurance       = "insurance"
	ExtraPropertyHouseFund       = "housefund"
	ExtraPropertySocialInsurance = "socialinsurance"
	ExtraPropertyZhimaScore      = "zhimascore"
	ExtraPropertySalaryMoney     = "salarymoney"
	ExtraPropertyEducation       = "education"
	ProgressRemark               = "progress_remark"
	IntentionValidId             = "intention_valid_id"
)

// 所有属性列表
var AllExtraProperties = []string{
	ExtraPropertyOccupation,
	ExtraPropertyHouse,
	ExtraPropertyCar,
	ExtraPropertyCreditCard,
	ExtraPropertyInsurance,
	ExtraPropertyHouseFund,
	ExtraPropertySocialInsurance,
	ExtraPropertyZhimaScore,
	ExtraPropertySalaryMoney,
	ExtraPropertyEducation,
}

// 属性显示名称映射
var ExtraPropertyLabels = map[string]string{
	ExtraPropertyOccupation:      "职业",
	ExtraPropertyHouse:           "房产",
	ExtraPropertyCar:             "车产",
	ExtraPropertyCreditCard:      "信用卡",
	ExtraPropertyInsurance:       "商业保险",
	ExtraPropertyHouseFund:       "公积金",
	ExtraPropertySocialInsurance: "社保",
	ExtraPropertyZhimaScore:      "芝麻分",
	ExtraPropertySalaryMoney:     "月薪",
	ExtraPropertyEducation:       "学历",
}

// SysCustomerListRequest sys_customer列表请求参数
type SysCustomerListRequest struct {
	models.BasePaging
	models.Validator
	Num             *string    `form:"num"`             // 客户编号
	Name            *string    `form:"name"`            // 客户姓名
	Mobile          *string    `form:"mobile"`          // 手机号
	MoneyDemand     *int       `form:"moneyDemand"`     // 需求金额
	ChannelId       *int       `form:"channelId"`       // 渠道来源
	UserId          *int       `form:"userId"`          // 跟进人
	CustomerStar    *int       `form:"customerStar"`    // 星级
	Status          *int       `form:"status"`          // 业务阶段
	Intention       *int       `form:"intention"`       // 客户有效
	SinglePieceType *int       `form:"singlePieceType"` // 贷款类型
	AllotTime       *time.Time `form:"allotTime"`       // 分配时间
	DepartmentId    *int       `form:"departmentId"`    // 所属部门
	City            *string    `form:"city"`            // 所在城市
	IsReassign      *int       `form:"isReassign"`      // 再分配
	IsQuit          *int       `form:"isQuit"`          // 离职数据
	IsRepeat        *int       `form:"isRepeat"`        // 重复标记
	StarStatus      *int       `form:"starStatus"`      // 星级回传
	IsLock          *int       `form:"isLock"`          // 是否锁定
}

// Validate 验证请求参数
func (r *SysCustomerListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// Handle 获取查询条件
func (r *SysCustomerListRequest) Handle() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if r.Num != nil {
			db = db.Where("num LIKE ?", "%"+*r.Num+"%")
		}
		if r.Name != nil {
			db = db.Where("name LIKE ?", "%"+*r.Name+"%")
		}
		if r.Mobile != nil {
			db = db.Where("mobile LIKE ?", "%"+*r.Mobile+"%")
		}
		if r.MoneyDemand != nil {
			db = db.Where("money_demand = ?", *r.MoneyDemand)
		}
		if r.ChannelId != nil {
			db = db.Where("channel_id = ?", *r.ChannelId)
		}
		if r.UserId != nil {
			db = db.Where("user_id = ?", *r.UserId)
		}
		if r.CustomerStar != nil {
			db = db.Where("customer_star = ?", *r.CustomerStar)
		}
		if r.Status != nil {
			db = db.Where("status = ?", *r.Status)
		}
		if r.Intention != nil {
			db = db.Where("intention = ?", *r.Intention)
		}
		if r.SinglePieceType != nil {
			db = db.Where("single_piece_type = ?", *r.SinglePieceType)
		}
		if r.AllotTime != nil {
			db = db.Where("allot_time = ?", *r.AllotTime)
		}
		if r.DepartmentId != nil {
			db = db.Where("department_id = ?", *r.DepartmentId)
		}
		if r.City != nil {
			db = db.Where("city = ?", *r.City)
		}
		if r.IsReassign != nil {
			db = db.Where("is_reassign = ?", *r.IsReassign)
		}
		if r.IsQuit != nil {
			db = db.Where("is_quit = ?", *r.IsQuit)
		}
		if r.IsRepeat != nil {
			db = db.Where("is_repeat = ?", *r.IsRepeat)
		}
		if r.StarStatus != nil {
			db = db.Where("star_status = ?", *r.StarStatus)
		}
		if r.IsLock != nil {
			db = db.Where("is_lock = ?", *r.IsLock)
		}
		return db
	}
}

// SysCustomerCreateRequest 创建sys_customer请求参数
type SysCustomerCreateRequest struct {
	models.Validator
	Num          string `form:"num"`                                          // 客户编号
	Name         string `form:"name"`                                         // 客户姓名
	Mobile       string `form:"mobile" validate:"required" message:"手机号不能为空"` // 手机号
	MoneyDemand  int    `form:"moneyDemand"`                                  // 需求金额
	ChannelId    int    `form:"channelId"`                                    // 渠道来源
	CustomerStar int    `form:"customerStar"`                                 // 星级
	Status       int    `form:"status"`                                       // 业务阶段
	Intention    int    `form:"intention"`                                    // 客户有效
	Extra        string `form:"extra"`                                        // 扩展属性
	Sex          int    `form:"sex"`                                          // 性别
	DepartmentId int    `form:"departmentId"`                                 // 所属部门
	Remarks      string `form:"remarks"`                                      // 客户备注
	Age          int    `form:"age"`                                          // 年龄
	From         int    `form:"from"`                                         // 客户来源
}

// Validate 验证请求参数
func (r *SysCustomerCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerUpdateRequest 更新sys_customer请求参数
type SysCustomerUpdateRequest struct {
	models.Validator
	Id              int    `form:"id" validate:"required" message:"主键ID不能为空"`    // 主键ID
	Num             string `form:"num"`                                          // 客户编号
	Name            string `form:"name"`                                         // 客户姓名
	Mobile          string `form:"mobile" validate:"required" message:"手机号不能为空"` // 手机号
	MoneyDemand     int    `form:"moneyDemand"`                                  // 需求金额
	ChannelId       int    `form:"channelId"`                                    // 渠道来源
	CustomerStar    int    `form:"customerStar"`                                 // 星级
	Status          int    `form:"status"`                                       // 业务阶段
	Intention       int    `form:"intention"`                                    // 客户有效
	Extra           string `form:"extra"`                                        // 扩展属性
	Sex             int    `form:"sex"`                                          // 性别
	DepartmentId    int    `form:"departmentId"`                                 // 所属部门
	Remarks         string `form:"remarks"`                                      // 客户备注
	Age             int    `form:"age"`                                          // 年龄
	IsLock          int    `form:"isLock"`                                       // 年龄
	SinglePieceType int    `form:"singlePieceType"`                              // 借贷类型
	From            int    `form:"from"`                                         // 客户来源
}

// Validate 验证请求参数
func (r *SysCustomerUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerDeleteRequest 删除sys_customer请求参数
type SysCustomerDeleteRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"主键ID不能为空"` // 主键ID
}

// Validate 验证请求参数
func (r *SysCustomerDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerGetByIDRequest 根据ID获取sys_customer请求参数
type SysCustomerGetByIDRequest struct {
	models.Validator
	Id int `uri:"id" validate:"required" message:"主键ID不能为空"` // 主键ID
}

// Validate 验证请求参数
func (r *SysCustomerGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
