package service

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	"gin-fast/app/utils/common"
	"gin-fast/app/utils/datascope"
	"gin-fast/app/utils/tenanthelper"
	"gin-fast/exampleutils/snowflakehelper"
	"gin-fast/plugins/syscustomer/models"
	traceModels "gin-fast/plugins/syscustomertraces/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysCustomerService struct{}

const customerDefaultFrom = 3

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

func (s *SysCustomerService) getCurrentUserDeptID(c *gin.Context, userID uint) (int, error) {
	var user appModels.User
	if err := app.DB().WithContext(c).Select("dept_id").Where("id = ?", userID).First(&user).Error; err != nil {
		return 0, err
	}
	return int(user.DeptID), nil
}

func (s *SysCustomerService) getTenantCityByID(c *gin.Context, tenantID uint) (string, error) {
	var tenant appModels.Tenant
	err := app.DB().WithContext(c).Select("city").Where("id = ?", tenantID).First(&tenant).Error
	if err == gorm.ErrRecordNotFound {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(tenant.City), nil
}

func (s *SysCustomerService) normalizeCreateExtra(extra string) (string, error) {
	extra = strings.TrimSpace(extra)
	if extra == "" {
		return "{}", nil
	}

	var extraObj map[string]interface{}
	if err := json.Unmarshal([]byte(extra), &extraObj); err != nil {
		return "", fmt.Errorf("extra 字段不是合法的 JSON: %w", err)
	}

	extraBytes, err := json.Marshal(extraObj)
	if err != nil {
		return "", err
	}

	return string(extraBytes), nil
}

func (s *SysCustomerService) normalizeMobile(mobile string) string {
	mobile = strings.TrimSpace(mobile)
	mobile = strings.ReplaceAll(mobile, " ", "")
	return mobile
}

func (s *SysCustomerService) buildMobileHash(mobile string) string {
	if mobile == "" {
		return ""
	}
	return fmt.Sprintf("%x", md5.Sum([]byte(mobile)))
}

func appendSelectedField(fields []string, field string) []string {
	for _, current := range fields {
		if current == field {
			return fields
		}
	}
	return append(fields, field)
}

func (s *SysCustomerService) Create(c *gin.Context, req models.SysCustomerCreateRequest) (*models.SysCustomer, error) {
	mobile := s.normalizeMobile(req.Mobile)
	extra, err := s.normalizeCreateExtra(req.Extra)
	if err != nil {
		return nil, err
	}

	num := strings.TrimSpace(req.Num)
	if num == "" {
		num, err = snowflakehelper.GenerateID()
		if err != nil {
			return nil, err
		}
	}
	now := time.Now()
	sysCustomer := &models.SysCustomer{
		Num:         num,
		Name:        strings.TrimSpace(req.Name),
		Mobile:      mobile,
		MdMobile:    s.buildMobileHash(mobile),
		MoneyDemand: req.MoneyDemand,
		ChannelID:   req.ChannelID,
		Extra:       extra,
		AllotTime:   &now,
		Remarks:     strings.TrimSpace(req.Remarks),
	}

	selectedFields := []string{"Num", "Name", "Mobile", "MdMobile", "MoneyDemand", "ChannelID", "Extra", "AllotTime", "Remarks"}
	currentUserID := common.GetCurrentUserID(c)
	if currentUserID > 0 {
		sysCustomer.UserID = int(currentUserID)
		selectedFields = appendSelectedField(selectedFields, "UserID")
	}

	if tenantID := common.GetCurrentTenantID(c); tenantID > 0 {
		//sysCustomer.TenantID = int(tenantID)
		//selectedFields = appendSelectedField(selectedFields, "TenantID")

		city, cityErr := s.getTenantCityByID(c, tenantID)
		if cityErr != nil {
			return nil, cityErr
		}
		if city != "" {
			sysCustomer.City = city
			selectedFields = appendSelectedField(selectedFields, "City")
		}
	}

	if req.CustomerStar != nil {
		sysCustomer.CustomerStar = req.CustomerStar
		selectedFields = appendSelectedField(selectedFields, "CustomerStar")
	}

	if req.Status != nil {
		sysCustomer.Status = *req.Status
		selectedFields = appendSelectedField(selectedFields, "Status")
	}

	if req.Intention != nil {
		sysCustomer.Intention = *req.Intention
		selectedFields = appendSelectedField(selectedFields, "Intention")
	}

	if req.Sex != nil {
		sysCustomer.Sex = *req.Sex
		selectedFields = appendSelectedField(selectedFields, "Sex")
	}

	if req.Age != nil {
		sysCustomer.Age = *req.Age
		selectedFields = appendSelectedField(selectedFields, "Age")
	}

	if req.From != nil {
		sysCustomer.From = *req.From
	} else {
		sysCustomer.From = customerDefaultFrom
	}
	selectedFields = appendSelectedField(selectedFields, "From")

	if req.DeptID != nil && *req.DeptID > 0 {
		sysCustomer.DeptID = *req.DeptID
		selectedFields = appendSelectedField(selectedFields, "DeptID")
	} else if currentUserID > 0 {
		deptID, deptErr := s.getCurrentUserDeptID(c, currentUserID)
		if deptErr != nil {
			return nil, deptErr
		}
		if deptID > 0 {
			sysCustomer.DeptID = deptID
			selectedFields = appendSelectedField(selectedFields, "DeptID")
		}
	}

	if err := app.DB().WithContext(c).Select(selectedFields).Create(sysCustomer).Error; err != nil {
		return nil, err
	}

	return sysCustomer, nil
}

func (s *SysCustomerService) Update(c *gin.Context, req models.SysCustomerUpdateRequest) error {
	sysCustomer, err := s.getScopedCustomerByID(c, req.Id)
	if err != nil {
		return err
	}

	mobile := s.normalizeMobile(req.Mobile)
	extra, err := s.normalizeCreateExtra(req.Extra)
	if err != nil {
		return err
	}

	sysCustomer.Name = strings.TrimSpace(req.Name)
	sysCustomer.Mobile = mobile
	sysCustomer.MdMobile = s.buildMobileHash(mobile)
	sysCustomer.MoneyDemand = req.MoneyDemand
	sysCustomer.ChannelID = req.ChannelID
	if req.CustomerStar != nil {
		sysCustomer.CustomerStar = req.CustomerStar
	}
	sysCustomer.Status = req.Status
	sysCustomer.Intention = req.Intention
	sysCustomer.Extra = extra
	sysCustomer.Sex = req.Sex
	sysCustomer.Remarks = strings.TrimSpace(req.Remarks)
	sysCustomer.Age = req.Age
	sysCustomer.IsLock = req.IsLock
	sysCustomer.SinglePieceType = req.SinglePieceType

	if currentUserID := common.GetCurrentUserID(c); currentUserID > 0 && sysCustomer.UserID == int(currentUserID) {
		now := time.Now()
		sysCustomer.LastTime = &now
	}
	if err := sysCustomer.Update(c); err != nil {
		return err
	}
	return nil
}

func (s *SysCustomerService) Delete(c *gin.Context, id int) error {
	sysCustomer, err := s.getScopedCustomerByID(c, id)
	if err != nil {
		return err
	}

	if err := sysCustomer.Delete(c); err != nil {
		return err
	}

	return nil
}

func (s *SysCustomerService) GetByID(c *gin.Context, id int) (*models.SysCustomer, error) {
	return s.getScopedCustomerByID(c, id)
}

func (s *SysCustomerService) List(c *gin.Context, req models.SysCustomerListRequest) (*models.SysCustomerList, int64, error) {
	sysCustomerList := models.NewSysCustomerList()
	currentUserID := int(common.GetCurrentUserID(c))
	scopes := []func(*gorm.DB) *gorm.DB{
		req.Handle(),
		func(db *gorm.DB) *gorm.DB {
			// 场景默认条件依赖当前登录用户，因此在 service 层把 userID 注入给 request。
			return req.ApplyListScene(db, currentUserID)
		},
	}
	scopes = append(scopes, req.ApplyDefaultOrder, datascope.GetDataScopeUser(c), tenanthelper.TenantScope(c))
	total, err := sysCustomerList.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}
	scopes = append(scopes, req.Paginate())

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

func (s *SysCustomerService) CustomerQuickStatusUpdate(c *gin.Context, req models.CustomerQuickStatusUpdateRequest) error {
	sysCustomer, err := s.getScopedCustomerByID(c, int(req.CustomerID))
	if err != nil {
		return err
	}

	updates := map[string]interface{}{}
	extraObj := map[string]interface{}{}
	extraChanged := false

	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Intention != nil {
		updates["intention"] = *req.Intention
	}
	if req.CustomerStar != nil {
		updates["customer_star"] = *req.CustomerStar
	}

	if req.ProgressRemark != nil || req.IntentionValidID != nil {
		if strings.TrimSpace(sysCustomer.Extra) != "" {
			if err := json.Unmarshal([]byte(sysCustomer.Extra), &extraObj); err != nil {
				extraObj = map[string]interface{}{}
			}
		}
	}

	if req.ProgressRemark != nil {
		progressRemark := strings.TrimSpace(*req.ProgressRemark)
		if progressRemark == "" {
			delete(extraObj, models.ProgressRemark)
		} else {
			extraObj[models.ProgressRemark] = progressRemark
		}
		extraChanged = true
	}

	if req.IntentionValidID != nil {
		if *req.IntentionValidID > 0 {
			extraObj[models.IntentionValidId] = *req.IntentionValidID
		} else {
			delete(extraObj, models.IntentionValidId)
		}
		extraChanged = true
	}

	if extraChanged {
		extraBytes, err := json.Marshal(extraObj)
		if err != nil {
			return err
		}
		updates["extra"] = string(extraBytes)
	}

	return app.DB().WithContext(c).Transaction(func(tx *gorm.DB) error {
		if len(updates) > 0 {
			if err := tx.Model(&models.SysCustomer{}).Where("id = ?", sysCustomer.Id).Updates(updates).Error; err != nil {
				return err
			}
		}

		traceData := strings.TrimSpace(req.Data)
		if traceData == "" {
			return nil
		}

		trace := traceModels.SysCustomerTraces{
			CustomerID: int64(sysCustomer.Id),
			UserID:     req.UserID,
			Data:       traceData,
		}

		return tx.Create(&trace).Error
	})
}
