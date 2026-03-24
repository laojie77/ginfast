package service

import (
	"encoding/json"
	"strings"

	"gin-fast/app/global/app"
	"gin-fast/app/utils/common"
	"gin-fast/app/utils/datascope"
	"gin-fast/app/utils/tenanthelper"
	"gin-fast/plugins/syscustomer/models"
	traceModels "gin-fast/plugins/syscustomertraces/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type SysCustomerService struct{}

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

func (s *SysCustomerService) Create(c *gin.Context, req models.SysCustomerCreateRequest) (*models.SysCustomer, error) {
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
	if err := sysCustomer.Create(c); err != nil {
		return nil, err
	}

	return sysCustomer, nil
}

func (s *SysCustomerService) Update(c *gin.Context, req models.SysCustomerUpdateRequest) error {
	sysCustomer, err := s.getScopedCustomerByID(c, req.Id)
	if err != nil {
		return err
	}

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
	scopes := []func(*gorm.DB) *gorm.DB{req.Handle()}
	scopes = append(scopes, datascope.GetDataScopeUser(c), tenanthelper.TenantScope(c))
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
