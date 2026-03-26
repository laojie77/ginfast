package models

import (
	"errors"
	"gin-fast/app/models"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var customerMobilePattern = regexp.MustCompile(`^1[3-9]\d{9}$`)

func normalizeCustomerMobile(mobile string) string {
	return strings.Join(strings.Fields(strings.TrimSpace(mobile)), "")
}

func isValidCustomerMobile(mobile string) bool {
	return customerMobilePattern.MatchString(normalizeCustomerMobile(mobile))
}

const (
	//职业
	ExtraPropertyOccupation = "occupation"
	//房产
	ExtraPropertyHouse = "house"
	//车产
	ExtraPropertyCar = "car"
	//信用卡
	ExtraPropertyCreditCard = "creditcard"
	//商业保险
	ExtraPropertyInsurance = "insurance"
	//公积金
	ExtraPropertyHouseFund = "housefund"
	//社保
	ExtraPropertySocialInsurance = "socialinsurance"
	//芝麻分
	ExtraPropertyZhimaScore = "zhimascore"
	//月薪
	ExtraPropertySalaryMoney = "salarymoney"
	//学历
	ExtraPropertyEducation = "education"
	//上级评价
	ProgressRemark = "progress_remark"
	//客户有效二级状态
	IntentionValidId = "intention_valid_id"
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

const (
	// 客户列表场景：全部客户（默认排除公共池、待流转、锁定）。
	CustomerListSceneAll = "all"
	// 客户列表场景：当前登录用户自己的客户。
	CustomerListSceneMy = "my"
	// 客户列表场景：公共池客户。
	CustomerListScenePublic = "public"
	// 客户列表场景：待流转客户。
	CustomerListSceneExchange = "exchange"
	// 客户列表场景：当前登录用户名下的再分配客户。
	CustomerListSceneReassign = "reassign"
	// 客户列表场景：当前登录用户名下的锁定客户。
	CustomerListSceneLocked = "locked"
	// 无效客户
	CustomerListSceneIntention2 = "intention2"
	// 黑名单客户
	CustomerListSceneIntention3 = "intention3"
	//待处理客户
	CustomerListSceneStatus0 = "status0"
)

// SysCustomerListRequest sys_customer 列表请求参数
// SysCustomerListRequest 客户列表查询参数。
type SysCustomerListRequest struct {
	models.BasePaging
	models.Validator
	// Scene 列表场景，不同场景会带出不同的默认过滤条件。
	Scene *string `form:"scene"`
	// 以下字段为显式查询条件；一旦前端传值，优先级高于场景默认条件。
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
	IsPublic        *int       `form:"isPublic"`
	IsQuit          *int       `form:"isQuit"`
	IsRepeat        *int       `form:"isRepeat"`
	StarStatus      *int       `form:"starStatus"`
	IsExchange      *int       `form:"isExchange"`
	IsLock          *int       `form:"isLock"`
	IsRead          *int       `form:"isRead"`
}

// Validate 验证请求参数
// Validate 校验客户列表查询参数。
func (r *SysCustomerListRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}

// getScene 统一读取并标准化场景值，避免大小写和空格影响判断。
func (r *SysCustomerListRequest) getScene() string {
	if r.Scene == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*r.Scene))
}

// applyCustomerListDefaultIntFilter 仅在前端未显式传值时，补充场景默认条件。
var customerListAllowedOperators = map[string]struct{}{
	"=":  {},
	"!=": {},
	">":  {},
	">=": {},
	"<":  {},
	"<=": {},
}

// applyCustomerListDefaultIntFilter only applies the scene default when the caller
// does not explicitly provide a value. The operator defaults to "=".
func applyCustomerListDefaultIntFilter(db *gorm.DB, field string, value int, current *int, operators ...string) *gorm.DB {
	if current != nil {
		return db
	}

	operator := "="
	if len(operators) > 0 {
		candidate := strings.TrimSpace(operators[0])
		if _, ok := customerListAllowedOperators[candidate]; ok {
			operator = candidate
		}
	}

	switch operator {
	case "!=":
		return db.Where(clause.Neq{Column: field, Value: value})
	case ">":
		return db.Where(clause.Gt{Column: field, Value: value})
	case ">=":
		return db.Where(clause.Gte{Column: field, Value: value})
	case "<":
		return db.Where(clause.Lt{Column: field, Value: value})
	case "<=":
		return db.Where(clause.Lte{Column: field, Value: value})
	default:
		return db.Where(clause.Eq{Column: field, Value: value})
	}
}

// ApplyListScene 根据列表场景补充默认过滤条件。
// userID 是当前登录用户 ID，供“我的客户 / 再分配 / 锁定”等场景追加默认 user_id 条件。
// 如果前端显式传了 userId、isLock、isReassign 等参数，则以显式参数为准，不再套用默认场景值。
func (r *SysCustomerListRequest) ApplyListScene(db *gorm.DB, userID int) *gorm.DB {
	switch r.getScene() {
	case CustomerListSceneAll:
		// 全部客户默认排除公共池、待流转、锁定客户。
		db = applyCustomerListDefaultIntFilter(db, "is_public", 0, r.IsPublic)
		db = applyCustomerListDefaultIntFilter(db, "is_exchange", 0, r.IsExchange)
		db = applyCustomerListDefaultIntFilter(db, "is_lock", 0, r.IsLock)
		db = applyCustomerListDefaultIntFilter(db, "intention", 3, r.Intention, "!=")
	case CustomerListSceneMy:
		// 我的客户场景本身不额外限制状态，只在下方补充当前登录人的 user_id。
		db = applyCustomerListDefaultIntFilter(db, "is_public", 0, r.IsPublic)
		db = applyCustomerListDefaultIntFilter(db, "is_exchange", 0, r.IsExchange)
		db = applyCustomerListDefaultIntFilter(db, "is_reassign", 0, r.IsReassign)
		db = applyCustomerListDefaultIntFilter(db, "intention", 3, r.Intention, "!=")
		db = applyCustomerListDefaultIntFilter(db, "status", 0, r.Status, "!=")
	case CustomerListScenePublic:
		//公共池客户
		db = applyCustomerListDefaultIntFilter(db, "is_public", 1, r.IsPublic)
	case CustomerListSceneExchange:
		//带流转客户
		db = applyCustomerListDefaultIntFilter(db, "is_exchange", 1, r.IsExchange)
	case CustomerListSceneReassign:
		//再分配客户
		db = applyCustomerListDefaultIntFilter(db, "is_reassign", 1, r.IsReassign)
	case CustomerListSceneLocked:
		//锁定客户
		db = applyCustomerListDefaultIntFilter(db, "is_lock", 1, r.IsLock)
	case CustomerListSceneIntention2:
		//无效客户
		db = applyCustomerListDefaultIntFilter(db, "intention", 2, r.Intention)
	case CustomerListSceneIntention3:
		//黑名单
		db = applyCustomerListDefaultIntFilter(db, "intention", 3, r.Intention)
	case CustomerListSceneStatus0:
		//待处理客户
		db = applyCustomerListDefaultIntFilter(db, "status", 0, r.Status)
	}

	if r.UserID != nil || userID <= 0 {
		return db
	}

	switch r.getScene() {
	case CustomerListSceneMy, CustomerListSceneReassign, CustomerListSceneLocked, CustomerListSceneStatus0:
		// 这些场景默认只看当前登录用户名下的数据。
		db = db.Where("user_id = ?", userID)
	}

	return db
}

// Handle 获取查询条件
// ApplyDefaultOrder applies the customer list default sort when the caller
// does not provide an explicit order. The latest assignment-related time wins.
func (r *SysCustomerListRequest) ApplyDefaultOrder(db *gorm.DB) *gorm.DB {
	if strings.TrimSpace(r.Order) != "" {
		return db
	}

	return db.
		Order("GREATEST(COALESCE(redistribution_time, '1000-01-01 00:00:00'), COALESCE(allot_time, '1000-01-01 00:00:00')) DESC").
		Order("id DESC")
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
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}

	r.Mobile = normalizeCustomerMobile(r.Mobile)
	if !isValidCustomerMobile(r.Mobile) {
		return errors.New("手机号格式不正确")
	}

	return nil
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
	if err := r.Validator.Check(c, r); err != nil {
		return err
	}

	r.Mobile = normalizeCustomerMobile(r.Mobile)
	if !isValidCustomerMobile(r.Mobile) {
		return errors.New("手机号格式不正确")
	}

	return nil
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
	IsRead           *int    `json:"isRead" form:"isRead"`
	ProgressRemark   *string `json:"progressRemark" form:"progressRemark"`
	IntentionValidID *int    `json:"intentionValidId" form:"intentionValidId"`
}

func (r *CustomerQuickStatusUpdateRequest) Validate(c *gin.Context) error {
	return r.Validator.Check(c, r)
}
