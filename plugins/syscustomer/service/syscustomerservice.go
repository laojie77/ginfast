package service

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"strconv"
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
const customerListTotalCacheTTL = 15 * time.Second
const customerListTotalCachePrefix = "syscustomer:list:total:"

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
		Scopes(datascope.GetDataScopeByColumn(c, ""), tenanthelper.TenantScope(c)).
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

func (s *SysCustomerService) buildCustomerListBaseQuery(c *gin.Context, req models.SysCustomerListRequest, currentUserID int) *gorm.DB {
	return app.DB().WithContext(c).
		Model(&models.SysCustomer{}).
		Scopes(
			req.Handle(),
			func(db *gorm.DB) *gorm.DB {
				return req.ApplyListScene(db, currentUserID)
			},
			datascope.GetDataScopeByColumn(c, ""),
			tenanthelper.TenantScope(c),
		)
}

func customerListTotalCacheKey(c *gin.Context, req models.SysCustomerListRequest, currentUserID int) string {
	payload := struct {
		TenantID        uint       `json:"tenantId"`
		CurrentUserID   int        `json:"currentUserId"`
		Scene           string     `json:"scene"`
		Num             *string    `json:"num,omitempty"`
		Name            *string    `json:"name,omitempty"`
		Mobile          *string    `json:"mobile,omitempty"`
		MoneyDemand     *int       `json:"moneyDemand,omitempty"`
		ChannelID       *int       `json:"channelId,omitempty"`
		UserID          *int       `json:"userId,omitempty"`
		CustomerStar    *int       `json:"customerStar,omitempty"`
		Status          *int       `json:"status,omitempty"`
		Intention       *int       `json:"intention,omitempty"`
		SinglePieceType *int       `json:"singlePieceType,omitempty"`
		AllotTime       *time.Time `json:"allotTime,omitempty"`
		DeptID          *int       `json:"deptId,omitempty"`
		City            *string    `json:"city,omitempty"`
		IsReassign      *int       `json:"isReassign,omitempty"`
		IsPublic        *int       `json:"isPublic,omitempty"`
		IsQuit          *int       `json:"isQuit,omitempty"`
		IsRepeat        *int       `json:"isRepeat,omitempty"`
		StarStatus      *int       `json:"starStatus,omitempty"`
		IsExchange      *int       `json:"isExchange,omitempty"`
		IsLock          *int       `json:"isLock,omitempty"`
		IsRead          *int       `json:"isRead,omitempty"`
	}{
		TenantID:        common.GetCurrentTenantID(c),
		CurrentUserID:   currentUserID,
		Scene:           strings.ToLower(strings.TrimSpace(derefString(req.Scene))),
		Num:             req.Num,
		Name:            req.Name,
		Mobile:          req.Mobile,
		MoneyDemand:     req.MoneyDemand,
		ChannelID:       req.ChannelId,
		UserID:          req.UserID,
		CustomerStar:    req.CustomerStar,
		Status:          req.Status,
		Intention:       req.Intention,
		SinglePieceType: req.SinglePieceType,
		AllotTime:       req.AllotTime,
		DeptID:          req.DeptID,
		City:            req.City,
		IsReassign:      req.IsReassign,
		IsPublic:        req.IsPublic,
		IsQuit:          req.IsQuit,
		IsRepeat:        req.IsRepeat,
		StarStatus:      req.StarStatus,
		IsExchange:      req.IsExchange,
		IsLock:          req.IsLock,
		IsRead:          req.IsRead,
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return ""
	}

	return customerListTotalCachePrefix + fmt.Sprintf("%x", md5.Sum(raw))
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func getCachedCustomerListTotal(ctx context.Context, cacheKey string) (int64, bool) {
	if cacheKey == "" {
		return 0, false
	}

	raw, err := app.Cache.Get(ctx, cacheKey)
	if err != nil || strings.TrimSpace(raw) == "" {
		return 0, false
	}

	total, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || total < 0 {
		return 0, false
	}

	return total, true
}

func cacheCustomerListTotal(ctx context.Context, cacheKey string, total int64) {
	if cacheKey == "" || total < 0 {
		return
	}

	_ = app.Cache.Set(ctx, cacheKey, strconv.FormatInt(total, 10), customerListTotalCacheTTL)
}

func resolveExactCustomerListTotal(req models.SysCustomerListRequest, pageIDs []int) (int64, bool) {
	if req.PageNum <= 0 || req.PageSize <= 0 {
		return 0, false
	}

	pageSize := req.PageSize
	pageLen := len(pageIDs)
	offset := (req.PageNum - 1) * pageSize

	if req.PageNum == 1 && pageLen == 0 {
		return 0, true
	}

	if pageLen > 0 && pageLen < pageSize {
		return int64(offset + pageLen), true
	}

	return 0, false
}

type customerListCountResult struct {
	total int64
	err   error
}

func customerListSelectColumns() []string {
	return []string{
		"id",
		"num",
		"name",
		"mobile",
		"money_demand",
		"channel_id",
		"user_id",
		"customer_star",
		"status",
		"intention",
		"`from`",
		"allot_time",
		"dept_id",
		"remarks",
		"city",
		"customer_comment",
		"is_reassign",
		"is_read",
		"is_quit",
		"is_repeat",
		"is_sms",
		"star_status",
		"is_lock",
	}
}

func reorderCustomerListByPageIDs(list *models.SysCustomerList, pageIDs []int) {
	if list == nil || len(*list) <= 1 || len(pageIDs) <= 1 {
		return
	}

	indexByID := make(map[int]int, len(pageIDs))
	for idx, id := range pageIDs {
		indexByID[id] = idx
	}

	ordered := make(models.SysCustomerList, len(pageIDs))
	for _, customer := range *list {
		if idx, ok := indexByID[customer.Id]; ok {
			ordered[idx] = customer
		}
	}

	filtered := ordered[:0]
	for _, customer := range ordered {
		if customer.Id > 0 {
			filtered = append(filtered, customer)
		}
	}
	*list = filtered
}

func (s *SysCustomerService) loadLatestCustomerTraces(c *gin.Context, customerIDs []int) (map[int][]traceModels.SysCustomerTraces, error) {
	result := make(map[int][]traceModels.SysCustomerTraces, len(customerIDs))
	if len(customerIDs) == 0 {
		return result, nil
	}

	claims := common.GetClaims(c)
	if claims == nil {
		return result, nil
	}

	latestTraceIDQuery := app.DB().WithContext(c).
		Model(&traceModels.SysCustomerTraces{}).
		Select("MAX(id)").
		Where("tenant_id = ?", claims.TenantID).
		Where("customer_id IN ?", customerIDs).
		Group("customer_id")

	var traces []traceModels.SysCustomerTraces
	if err := app.DB().WithContext(c).
		Model(&traceModels.SysCustomerTraces{}).
		Select("sys_customer_traces.id, sys_customer_traces.customer_id, sys_customer_traces.data, sys_customer_traces.created_at, sys_users.nick_name AS user_name").
		Joins("LEFT JOIN sys_users ON sys_customer_traces.user_id = sys_users.id").
		Where("sys_customer_traces.id IN (?)", latestTraceIDQuery).
		Find(&traces).Error; err != nil {
		return nil, err
	}

	for _, trace := range traces {
		result[int(trace.CustomerID)] = []traceModels.SysCustomerTraces{trace}
	}

	return result, nil
}

func attachLatestCustomerTraces(list *models.SysCustomerList, traceMap map[int][]traceModels.SysCustomerTraces) {
	if list == nil {
		return
	}

	for i := range *list {
		customer := &(*list)[i]
		customer.CustomerTracesList = traceMap[customer.Id]
	}
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
		// Create uses Select(selectedFields), so TenantID must be included explicitly here.
		sysCustomer.TenantID = int(tenantID)
		selectedFields = appendSelectedField(selectedFields, "TenantID")

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
	cacheCtx := context.Background()
	totalCacheKey := customerListTotalCacheKey(c, req, currentUserID)
	total, totalFromCache := getCachedCustomerListTotal(cacheCtx, totalCacheKey)

	var pageIDs []int
	pageQuery := s.buildCustomerListBaseQuery(c, req, currentUserID)
	if err := pageQuery.
		Scopes(req.ApplyDefaultOrder, req.Paginate()).
		Pluck("id", &pageIDs).Error; err != nil {
		return nil, 0, err
	}

	if !totalFromCache {
		if exactTotal, ok := resolveExactCustomerListTotal(req, pageIDs); ok {
			total = exactTotal
			totalFromCache = true
			cacheCustomerListTotal(cacheCtx, totalCacheKey, total)
		}
	}

	var countResultCh chan customerListCountResult
	if !totalFromCache {
		countResultCh = make(chan customerListCountResult, 1)
		countContext := c.Copy()
		go func() {
			var counted int64
			err := s.buildCustomerListBaseQuery(countContext, req, currentUserID).Count(&counted).Error
			countResultCh <- customerListCountResult{total: counted, err: err}
		}()
	}

	if len(pageIDs) == 0 {
		if countResultCh != nil {
			result := <-countResultCh
			if result.err != nil {
				return nil, 0, result.err
			}
			total = result.total
			cacheCustomerListTotal(cacheCtx, totalCacheKey, total)
		}
		return sysCustomerList, total, nil
	}

	dataQuery := app.DB().WithContext(c).
		Model(&models.SysCustomer{}).
		Select(customerListSelectColumns()).
		Where("id IN ?", pageIDs)

	if err := dataQuery.Find(sysCustomerList).Error; err != nil {
		return nil, 0, err
	}
	reorderCustomerListByPageIDs(sysCustomerList, pageIDs)

	traceMap, err := s.loadLatestCustomerTraces(c, pageIDs)
	if err != nil {
		return nil, 0, err
	}
	attachLatestCustomerTraces(sysCustomerList, traceMap)

	if countResultCh != nil {
		result := <-countResultCh
		if result.err != nil {
			return nil, 0, result.err
		}
		total = result.total
		cacheCustomerListTotal(cacheCtx, totalCacheKey, total)
	}

	return sysCustomerList, total, nil
}

func (s *SysCustomerService) listLegacy(c *gin.Context, req models.SysCustomerListRequest) (*models.SysCustomerList, int64, error) {
	sysCustomerList := models.NewSysCustomerList()
	currentUserID := int(common.GetCurrentUserID(c))
	scopes := []func(*gorm.DB) *gorm.DB{
		req.Handle(),
		func(db *gorm.DB) *gorm.DB {
			// 场景默认条件依赖当前登录用户，因此在 service 层把 userID 注入给 request。
			return req.ApplyListScene(db, currentUserID)
		},
	}
	scopes = append(scopes, req.ApplyDefaultOrder, datascope.GetDataScopeByColumn(c, ""), tenanthelper.TenantScope(c))
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

	//if currentUserID := common.GetCurrentUserID(c); currentUserID > 0 && sysCustomer.UserID != int(currentUserID) {
	//	return fmt.Errorf("无权操作该客户")
	//}

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
	if req.IsRead != nil {
		updates["is_read"] = *req.IsRead
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
