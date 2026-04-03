package service

import (
	"gin-fast/app/global/app"
	appModels "gin-fast/app/models"
	"gin-fast/app/utils/tenanthelper"
	customerModels "gin-fast/plugins/syscustomer/models"
	pluginModels "gin-fast/plugins/syscustomerexporttasks/models"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SysCustomerExportTasksService provides read-only access to export tasks.
type SysCustomerExportTasksService struct{}

// NewSysCustomerExportTasksService creates a task query service.
func NewSysCustomerExportTasksService() *SysCustomerExportTasksService {
	return &SysCustomerExportTasksService{}
}

// GetByID returns a single task within the current tenant.
func (s *SysCustomerExportTasksService) GetByID(c *gin.Context, id uint64) (*customerModels.SysCustomerExportTask, error) {
	sysCustomerExportTaskList := customerModels.NewSysCustomerExportTaskList()
	scopes := append(buildTaskScopes(c, pluginModels.SysCustomerExportTasksListRequest{}), func(db *gorm.DB) *gorm.DB {
		return db.Where("id = ?", uint(id)).Limit(1)
	})
	if err := sysCustomerExportTaskList.Find(c, scopes...); err != nil {
		return nil, err
	}
	if len(*sysCustomerExportTaskList) == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	if err := s.fillTaskDisplayFields(c, sysCustomerExportTaskList); err != nil {
		return nil, err
	}

	task := (*sysCustomerExportTaskList)[0]
	return &task, nil
}

// List returns paged customer export tasks within the current tenant.
func (s *SysCustomerExportTasksService) List(c *gin.Context, req pluginModels.SysCustomerExportTasksListRequest) (*customerModels.SysCustomerExportTaskList, int64, error) {
	sysCustomerExportTaskList := customerModels.NewSysCustomerExportTaskList()
	scopes := buildTaskScopes(c, req)

	total, err := sysCustomerExportTaskList.GetTotal(c, scopes...)
	if err != nil {
		return nil, 0, err
	}

	scopes = append(scopes, req.Paginate())
	if err = sysCustomerExportTaskList.Find(c, scopes...); err != nil {
		return nil, 0, err
	}

	if err = s.fillTaskDisplayFields(c, sysCustomerExportTaskList); err != nil {
		return nil, 0, err
	}

	return sysCustomerExportTaskList, total, nil
}

func (s *SysCustomerExportTasksService) fillTaskDisplayFields(c *gin.Context, list *customerModels.SysCustomerExportTaskList) error {
	if list == nil || len(*list) == 0 {
		return nil
	}

	userNameMap, err := s.loadUserNickNameMap(c, list)
	if err != nil {
		return err
	}

	for index := range *list {
		task := &(*list)[index]
		task.UserNickName = userNameMap[task.UserID]
	}

	return nil
}

func (s *SysCustomerExportTasksService) loadUserNickNameMap(c *gin.Context, list *customerModels.SysCustomerExportTaskList) (map[uint]string, error) {
	userIDSet := make(map[uint]struct{}, len(*list))
	for _, task := range *list {
		if task.UserID > 0 {
			userIDSet[task.UserID] = struct{}{}
		}
	}

	userIDs := make([]uint, 0, len(userIDSet))
	for userID := range userIDSet {
		userIDs = append(userIDs, userID)
	}
	if len(userIDs) == 0 {
		return map[uint]string{}, nil
	}

	var users []appModels.User
	if err := app.DB().WithContext(c).
		Model(&appModels.User{}).
		Scopes(tenanthelper.TenantScope(c)).
		Select("id", "username", "nick_name").
		Where("id IN ?", userIDs).
		Find(&users).Error; err != nil {
		return nil, err
	}

	result := make(map[uint]string, len(users))
	for _, user := range users {
		if user.NickName != "" {
			result[user.ID] = user.NickName
			continue
		}
		result[user.ID] = user.Username
	}

	return result, nil
}

func buildTaskScopes(c *gin.Context, req pluginModels.SysCustomerExportTasksListRequest) []func(*gorm.DB) *gorm.DB {
	return []func(*gorm.DB) *gorm.DB{
		tenanthelper.TenantScope(c),
		req.Handle(),
		func(db *gorm.DB) *gorm.DB {
			if req.UserName != nil {
				keyword := strings.TrimSpace(*req.UserName)
				if keyword != "" {
					subQuery := app.DB().WithContext(c).
						Model(&appModels.User{}).
						Scopes(tenanthelper.TenantScope(c)).
						Select("id").
						Where("username LIKE ? OR nick_name LIKE ?", "%"+keyword+"%", "%"+keyword+"%")
					db = db.Where("user_id IN (?)", subQuery)
				}
			}

			return db.Where("biz_type = ?", customerModels.CustomerExportTaskBizTypeSysCustomer).
				Order("created_at DESC").
				Order("id DESC")
		},
	}
}
