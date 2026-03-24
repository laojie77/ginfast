package models

import (
	"gin-fast/app/models"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

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

// SysCustomerListRequest sys_customer 列表请求参数
type SysCustomerListRequest struct {
	models.BasePaging
	models.Validator
	Num             *string    `form:"num"`
	Name            *string    `form:"name"`
	Mobile          *string    `form:"mobile"`
	MoneyDemand     *int       `form:"moneyDemand"`
	ChannelId       *int       `form:"channelId"`
	UserID          *int       `form:"userId"`
	CustomerStar    *int       `form:"customerStar"`
	Status          *int       `form:"status"`
	Intention       *int       `form:"intention"`
	SinglePieceType *int       `form:"singlePieceType"`
	AllotTime       *time.Time `form:"allotTime"`
	DeptID          *int       `form:"deptId"`
	City            *string    `form:"city"`
	IsReassign      *int       `form:"isReassign"`
	IsQuit          *int       `form:"isQuit"`
	IsRepeat        *int       `form:"isRepeat"`
	StarStatus      *int       `form:"starStatus"`
	IsLock          *int       `form:"isLock"`
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
		if r.UserID != nil {
			db = db.Where("user_id = ?", *r.UserID)
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
		if r.DeptID != nil {
			db = db.Where("dept_id = ?", *r.DeptID)
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

// SysCustomerCreateRequest 创建 sys_customer 请求参数
type SysCustomerCreateRequest struct {
	models.Validator
	Num          string `form:"num"`
	Name         string `form:"name"`
	Mobile       string `form:"mobile" validate:"required" message:"手机号不能为空"`
	MoneyDemand  int    `form:"moneyDemand"`
	ChannelID    int    `form:"channelId"`
	CustomerStar *int   `form:"customerStar"`
	Status       *int   `form:"status"`
	Intention    *int   `form:"intention"`
	Extra        string `form:"extra"`
	Sex          *int   `form:"sex"`
	DeptID       *int   `form:"deptId"`
	Remarks      string `form:"remarks"`
	Age          *int   `form:"age"`
	From         *int   `form:"from"`
}

// Validate 验证请求参数
func (r *SysCustomerCreateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerUpdateRequest 更新 sys_customer 请求参数
type SysCustomerUpdateRequest struct {
	models.Validator
	Id              int    `form:"id" validate:"required" message:"主键ID不能为空"`
	Num             string `form:"num"`
	Name            string `form:"name"`
	Mobile          string `form:"mobile" validate:"required" message:"手机号不能为空"`
	MoneyDemand     int    `form:"moneyDemand"`
	ChannelID       int    `form:"channelId"`
	CustomerStar    *int   `form:"customerStar"`
	Status          int    `form:"status"`
	Intention       int    `form:"intention"`
	Extra           string `form:"extra"`
	Sex             int    `form:"sex"`
	DeptID          int    `form:"deptId"`
	Remarks         string `form:"remarks"`
	Age             int    `form:"age"`
	IsLock          int    `form:"isLock"`
	SinglePieceType int    `form:"singlePieceType"`
	From            int    `form:"from"`
}

// Validate 验证请求参数
func (r *SysCustomerUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerDeleteRequest 删除 sys_customer 请求参数
type SysCustomerDeleteRequest struct {
	models.Validator
	Id int `form:"id" validate:"required" message:"主键ID不能为空"`
}

// Validate 验证请求参数
func (r *SysCustomerDeleteRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// SysCustomerGetByIDRequest 根据 ID 获取 sys_customer 请求参数
type SysCustomerGetByIDRequest struct {
	models.Validator
	Id int `uri:"id" validate:"required" message:"主键ID不能为空"`
}

// Validate 验证请求参数
func (r *SysCustomerGetByIDRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type CustomerStatusTracesUpdateRequest struct {
	models.Validator
	CustomerID   int64  `form:"id" validate:"required" message:"客户ID不能为空"` // 客户
	UserID       int    `form:"userId"`                                    // 操作用户
	Data         string `form:"data" `                                     // 跟进内容
	CustomerStar string `form:"customerStar"`                              // 星级
}

// Validate 验证请求参数
func (r *CustomerStatusTracesUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

type CustomerQuickStatusUpdateRequest struct {
	models.Validator
	CustomerID       int64   `json:"customerId" form:"customerId" validate:"required" message:"瀹㈡埛ID涓嶈兘涓虹┖"`
	UserID           int     `json:"userId" form:"userId"`
	Data             string  `json:"data" form:"data"`
	Status           *int    `json:"status" form:"status"`
	Intention        *int    `json:"intention" form:"intention"`
	CustomerStar     *int    `json:"customerStar" form:"customerStar"`
	ProgressRemark   *string `json:"progressRemark" form:"progressRemark"`
	IntentionValidID *int    `json:"intentionValidId" form:"intentionValidId"`
}

func (r *CustomerQuickStatusUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
