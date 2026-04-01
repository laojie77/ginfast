package service

import (
	"context"
	"errors"
	"fmt"
	models2 "gin-fast/plugins/sysnotice/models"
	"time"

	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"gin-fast/app/utils/common"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type NoticeDispatcher interface {
	Dispatch(ctx context.Context, notice *models2.SysNotice, recipients models2.SysNoticeRecipientList) error
}

type noopNoticeDispatcher struct{}

func (noopNoticeDispatcher) Dispatch(ctx context.Context, notice *models2.SysNotice, recipients models2.SysNoticeRecipientList) error {
	return nil
}

type SysNoticeService struct {
	dispatcher NoticeDispatcher
}

func NewSysNoticeService() *SysNoticeService {
	return &SysNoticeService{
		dispatcher: noopNoticeDispatcher{},
	}
}

func (s *SysNoticeService) Create(c *gin.Context, req *models2.SysNoticeAddRequest) (*models2.SysNotice, error) {
	currentUserID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	if err := s.validateTargets(c, tenantID, req.Targets); err != nil {
		return nil, err
	}

	notice := &models2.SysNotice{
		Title:         req.Title,
		Content:       req.Content,
		Type:          req.Type,
		Level:         req.Level,
		PublishStatus: models2.SysNoticePublishStatusDraft,
		TenantID:      tenantID,
		CreatedBy:     currentUserID,
	}

	shouldDispatch := false
	err := app.DB().WithContext(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(notice).Error; err != nil {
			return err
		}
		if err := s.replaceTargetsTx(tx, notice.ID, tenantID, req.Targets); err != nil {
			return err
		}
		if req.PublishNow {
			if err := s.publishNoticeTx(c, tx, notice, currentUserID, tenantID); err != nil {
				return err
			}
			shouldDispatch = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if shouldDispatch {
		s.dispatchPublishedNotice(c, notice.ID)
	}

	return s.GetByID(c, notice.ID)
}

func (s *SysNoticeService) Update(c *gin.Context, req *models2.SysNoticeUpdateRequest) (*models2.SysNotice, error) {
	currentUserID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	if err := s.validateTargets(c, tenantID, req.Targets); err != nil {
		return nil, err
	}

	notice, err := s.getNoticeForManage(c, req.ID, tenantID)
	if err != nil {
		return nil, err
	}
	if notice.PublishStatus != models2.SysNoticePublishStatusDraft {
		return nil, errors.New("仅未发布的通知允许编辑")
	}

	shouldDispatch := false
	err = app.DB().WithContext(c).Transaction(func(tx *gorm.DB) error {
		updates := map[string]interface{}{
			"title":   req.Title,
			"content": req.Content,
			"type":    req.Type,
			"level":   req.Level,
		}
		if err := tx.Model(&models2.SysNotice{}).Where("id = ? AND tenant_id = ?", req.ID, tenantID).Updates(updates).Error; err != nil {
			return err
		}
		if err := s.replaceTargetsTx(tx, req.ID, tenantID, req.Targets); err != nil {
			return err
		}
		if req.PublishNow {
			if err := s.publishNoticeTx(c, tx, notice, currentUserID, tenantID); err != nil {
				return err
			}
			shouldDispatch = true
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	if shouldDispatch {
		s.dispatchPublishedNotice(c, req.ID)
	}

	return s.GetByID(c, req.ID)
}

func (s *SysNoticeService) Publish(c *gin.Context, noticeID uint) (*models2.SysNotice, error) {
	currentUserID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	notice, err := s.getNoticeForManage(c, noticeID, tenantID)
	if err != nil {
		return nil, err
	}
	if notice.PublishStatus == models2.SysNoticePublishStatusPublished {
		return nil, errors.New("仅未发布的通知允许发布")
	}

	err = app.DB().WithContext(c).Transaction(func(tx *gorm.DB) error {
		return s.publishNoticeTx(c, tx, notice, currentUserID, tenantID)
	})
	if err != nil {
		return nil, err
	}

	s.dispatchPublishedNotice(c, noticeID)

	return s.GetByID(c, noticeID)
}

func (s *SysNoticeService) Revoke(c *gin.Context, noticeID uint) (*models2.SysNotice, error) {
	tenantID := common.GetCurrentTenantID(c)
	notice, err := s.getNoticeForManage(c, noticeID, tenantID)
	if err != nil {
		return nil, err
	}
	if notice.PublishStatus != models2.SysNoticePublishStatusPublished {
		return nil, errors.New("仅已发布的通知允许撤回")
	}

	now := time.Now()
	err = app.DB().WithContext(c).Model(&models2.SysNotice{}).
		Where("id = ? AND tenant_id = ?", noticeID, tenantID).
		Updates(map[string]interface{}{
			"publish_status": models2.SysNoticePublishStatusRevoked,
			"revoke_time":    now,
		}).Error
	if err != nil {
		return nil, err
	}

	return s.GetByID(c, noticeID)
}

func (s *SysNoticeService) Delete(c *gin.Context, noticeID uint) error {
	tenantID := common.GetCurrentTenantID(c)
	notice, err := s.getNoticeForManage(c, noticeID, tenantID)
	if err != nil {
		return err
	}
	if notice.PublishStatus == models2.SysNoticePublishStatusPublished {
		return errors.New("已发布的通知请先撤回后再删除")
	}

	return app.DB().WithContext(c).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("notice_id = ?", noticeID).Delete(&models2.SysNoticeTarget{}).Error; err != nil {
			return err
		}
		if err := tx.Where("notice_id = ?", noticeID).Delete(&models2.SysNoticeRecipient{}).Error; err != nil {
			return err
		}
		if err := tx.Where("id = ? AND tenant_id = ?", noticeID, tenantID).Delete(&models2.SysNotice{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *SysNoticeService) GetByID(c *gin.Context, noticeID uint) (*models2.SysNotice, error) {
	tenantID := common.GetCurrentTenantID(c)
	notice, err := s.getNoticeWithTargets(c, noticeID, tenantID)
	if err != nil {
		return nil, err
	}
	if err := s.fillTargetNames(c, notice.Targets, tenantID); err != nil {
		return nil, err
	}
	if err := s.fillNoticeStats(c, models2.SysNoticeList{notice}); err != nil {
		return nil, err
	}
	return notice, nil
}

func (s *SysNoticeService) List(c *gin.Context, req *models2.SysNoticeListRequest) (models2.SysNoticeList, int64, error) {
	tenantID := common.GetCurrentTenantID(c)
	list := models2.NewSysNoticeList()
	baseScope := func(db *gorm.DB) *gorm.DB {
		return db.Where("tenant_id = ?", tenantID)
	}

	total, err := list.GetTotal(c, baseScope, req.Handle())
	if err != nil {
		return nil, 0, err
	}

	err = list.Find(c, baseScope, req.Handle(), req.Paginate(), func(db *gorm.DB) *gorm.DB {
		return db.Preload("Publisher").Preload("Targets").Order("updated_at DESC, id DESC")
	})
	if err != nil {
		return nil, 0, err
	}

	for _, notice := range list {
		if err := s.fillTargetNames(c, notice.Targets, tenantID); err != nil {
			return nil, 0, err
		}
	}
	if err := s.fillNoticeStats(c, list); err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (s *SysNoticeService) InboxList(c *gin.Context, req *models2.SysNoticeInboxListRequest) (models2.SysNoticeRecipientList, int64, int64, error) {
	userID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	list := models2.NewSysNoticeRecipientList()

	baseScope := func(db *gorm.DB) *gorm.DB {
		return db.Joins("JOIN sys_notice ON sys_notice.id = sys_notice_recipient.notice_id").
			Where("sys_notice_recipient.user_id = ? AND sys_notice_recipient.tenant_id = ? AND sys_notice.publish_status = ?",
				userID, tenantID, models2.SysNoticePublishStatusPublished)
	}

	total, err := list.GetTotal(c, baseScope, req.Handle())
	if err != nil {
		return nil, 0, 0, err
	}

	var unreadCount int64
	err = app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
		Scopes(baseScope).
		Where("sys_notice_recipient.read_status = ?", models2.SysNoticeReadStatusUnread).
		Count(&unreadCount).Error
	if err != nil {
		return nil, 0, 0, err
	}

	err = list.Find(c, baseScope, req.Handle(), req.Paginate(), func(db *gorm.DB) *gorm.DB {
		return db.Preload("Notice", func(noticeDB *gorm.DB) *gorm.DB {
			return noticeDB.Preload("Publisher")
		}).Order("sys_notice_recipient.publish_time DESC, sys_notice_recipient.id DESC")
	})
	if err != nil {
		return nil, 0, 0, err
	}

	return list, total, unreadCount, nil
}

func (s *SysNoticeService) GetInboxDetail(c *gin.Context, noticeID uint) (*models2.SysNoticeRecipient, error) {
	userID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)

	recipient := models2.NewSysNoticeRecipient()
	err := app.DB().WithContext(c).
		Joins("JOIN sys_notice ON sys_notice.id = sys_notice_recipient.notice_id").
		Where("sys_notice_recipient.notice_id = ? AND sys_notice_recipient.user_id = ? AND sys_notice_recipient.tenant_id = ? AND sys_notice.publish_status = ?",
			noticeID, userID, tenantID, models2.SysNoticePublishStatusPublished).
		Preload("Notice", func(noticeDB *gorm.DB) *gorm.DB {
			return noticeDB.Preload("Publisher").Preload("Targets")
		}).
		First(recipient).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("通知不存在或您无权查看")
		}
		return nil, err
	}

	if err := s.fillTargetNames(c, recipient.Notice.Targets, tenantID); err != nil {
		return nil, err
	}
	if err := s.fillNoticeStats(c, models2.SysNoticeList{&recipient.Notice}); err != nil {
		return nil, err
	}

	return recipient, nil
}

func (s *SysNoticeService) MarkRead(c *gin.Context, noticeID uint) error {
	userID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	now := time.Now()

	result := app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
		Where("notice_id = ? AND user_id = ? AND tenant_id = ? AND read_status = ?", noticeID, userID, tenantID, models2.SysNoticeReadStatusUnread).
		Updates(map[string]interface{}{
			"read_status": models2.SysNoticeReadStatusRead,
			"read_time":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		err := app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
			Where("notice_id = ? AND user_id = ? AND tenant_id = ?", noticeID, userID, tenantID).
			Count(&count).Error
		if err != nil {
			return err
		}
		if count == 0 {
			return errors.New("通知不存在")
		}
	}
	return nil
}

func (s *SysNoticeService) MarkAllRead(c *gin.Context) (int64, error) {
	userID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	now := time.Now()

	publishedSubQuery := app.DB().WithContext(c).Model(&models2.SysNotice{}).
		Select("id").
		Where("tenant_id = ? AND publish_status = ?", tenantID, models2.SysNoticePublishStatusPublished)

	result := app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
		Where("user_id = ? AND tenant_id = ? AND read_status = ? AND notice_id IN (?)",
			userID, tenantID, models2.SysNoticeReadStatusUnread, publishedSubQuery).
		Updates(map[string]interface{}{
			"read_status": models2.SysNoticeReadStatusRead,
			"read_time":   now,
		})
	if result.Error != nil {
		return 0, result.Error
	}
	return result.RowsAffected, nil
}

func (s *SysNoticeService) UnreadCount(c *gin.Context) (int64, error) {
	userID := common.GetCurrentUserID(c)
	tenantID := common.GetCurrentTenantID(c)
	var unreadCount int64
	err := app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
		Joins("JOIN sys_notice ON sys_notice.id = sys_notice_recipient.notice_id").
		Where("sys_notice_recipient.user_id = ? AND sys_notice_recipient.tenant_id = ? AND sys_notice_recipient.read_status = ? AND sys_notice.publish_status = ?",
			userID, tenantID, models2.SysNoticeReadStatusUnread, models2.SysNoticePublishStatusPublished).
		Count(&unreadCount).Error
	if err != nil {
		return 0, err
	}
	return unreadCount, nil
}

func (s *SysNoticeService) validateTargets(c *gin.Context, tenantID uint, targets []models2.SysNoticeTargetItemRequest) error {
	if len(targets) == 0 {
		return errors.New("请至少选择一个通知目标")
	}

	targetKeys := make(map[string]struct{})
	var userIDs []uint
	var roleIDs []uint
	var deptIDs []uint

	for _, target := range targets {
		if target.TargetType != models2.SysNoticeTargetTypeAll && target.TargetID == 0 {
			return errors.New("非全体通知必须指定目标ID")
		}
		key := fmt.Sprintf("%d-%d-%t", target.TargetType, target.TargetID, target.IncludeChildren)
		if _, exists := targetKeys[key]; exists {
			return errors.New("通知目标存在重复项")
		}
		targetKeys[key] = struct{}{}

		switch target.TargetType {
		case models2.SysNoticeTargetTypeUser:
			userIDs = append(userIDs, target.TargetID)
		case models2.SysNoticeTargetTypeRole:
			roleIDs = append(roleIDs, target.TargetID)
		case models2.SysNoticeTargetTypeDept:
			deptIDs = append(deptIDs, target.TargetID)
		}
	}

	if err := s.ensureUsersExist(c, tenantID, userIDs); err != nil {
		return err
	}
	if err := s.ensureRolesExist(c, tenantID, roleIDs); err != nil {
		return err
	}
	if err := s.ensureDepartmentsExist(c, tenantID, deptIDs); err != nil {
		return err
	}

	return nil
}

func (s *SysNoticeService) ensureUsersExist(c *gin.Context, tenantID uint, userIDs []uint) error {
	if len(userIDs) == 0 {
		return nil
	}
	var count int64
	err := app.DB().WithContext(c).Model(&models.User{}).
		Where("tenant_id = ? AND id IN ?", tenantID, userIDs).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count != int64(len(uniqueUintSlice(userIDs))) {
		return errors.New("通知目标中存在无效用户")
	}
	return nil
}

func (s *SysNoticeService) ensureRolesExist(c *gin.Context, tenantID uint, roleIDs []uint) error {
	if len(roleIDs) == 0 {
		return nil
	}
	var count int64
	err := app.DB().WithContext(c).Model(&models.SysRole{}).
		Where("tenant_id = ? AND id IN ?", tenantID, roleIDs).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count != int64(len(uniqueUintSlice(roleIDs))) {
		return errors.New("通知目标中存在无效角色")
	}
	return nil
}

func (s *SysNoticeService) ensureDepartmentsExist(c *gin.Context, tenantID uint, deptIDs []uint) error {
	if len(deptIDs) == 0 {
		return nil
	}
	var count int64
	err := app.DB().WithContext(c).Model(&models.SysDepartment{}).
		Where("tenant_id = ? AND id IN ?", tenantID, deptIDs).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count != int64(len(uniqueUintSlice(deptIDs))) {
		return errors.New("通知目标中存在无效部门")
	}
	return nil
}

func (s *SysNoticeService) replaceTargetsTx(tx *gorm.DB, noticeID, tenantID uint, targets []models2.SysNoticeTargetItemRequest) error {
	if err := tx.Where("notice_id = ?", noticeID).Delete(&models2.SysNoticeTarget{}).Error; err != nil {
		return err
	}

	rows := make([]models2.SysNoticeTarget, 0, len(targets))
	for _, target := range targets {
		row := models2.SysNoticeTarget{
			NoticeID:        noticeID,
			TargetType:      target.TargetType,
			TargetID:        target.TargetID,
			IncludeChildren: target.IncludeChildren && target.TargetType == models2.SysNoticeTargetTypeDept,
			TenantID:        tenantID,
		}
		if row.TargetType == models2.SysNoticeTargetTypeAll {
			row.TargetID = 0
			row.IncludeChildren = false
		}
		rows = append(rows, row)
	}

	if len(rows) == 0 {
		return errors.New("请至少选择一个通知目标")
	}

	return tx.Create(&rows).Error
}

func (s *SysNoticeService) publishNoticeTx(c *gin.Context, tx *gorm.DB, notice *models2.SysNotice, publisherID, tenantID uint) error {
	var targets models2.SysNoticeTargetList
	if err := tx.Where("notice_id = ?", notice.ID).Find(&targets).Error; err != nil {
		return err
	}
	if len(targets) == 0 {
		return errors.New("通知缺少投放目标，无法发布")
	}

	userIDs, err := s.resolveRecipientUserIDsTx(c, tx, tenantID, targets)
	if err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return errors.New("没有匹配到有效的接收人")
	}

	now := time.Now()
	recipients := make([]models2.SysNoticeRecipient, 0, len(userIDs))
	for _, userID := range userIDs {
		recipients = append(recipients, models2.SysNoticeRecipient{
			NoticeID:    notice.ID,
			UserID:      userID,
			ReadStatus:  models2.SysNoticeReadStatusUnread,
			TenantID:    tenantID,
			PublishTime: &now,
		})
	}

	if err := tx.Where("notice_id = ?", notice.ID).Delete(&models2.SysNoticeRecipient{}).Error; err != nil {
		return err
	}
	if err := tx.Create(&recipients).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"publisher_id":   publisherID,
		"publish_status": models2.SysNoticePublishStatusPublished,
		"publish_time":   now,
		"revoke_time":    nil,
	}
	if err := tx.Model(&models2.SysNotice{}).
		Where("id = ? AND tenant_id = ?", notice.ID, tenantID).
		Updates(updates).Error; err != nil {
		return err
	}

	notice.PublishStatus = models2.SysNoticePublishStatusPublished
	notice.PublishTime = &now
	notice.RevokeTime = nil
	notice.PublisherID = publisherID

	return nil
}

func (s *SysNoticeService) resolveRecipientUserIDsTx(c *gin.Context, tx *gorm.DB, tenantID uint, targets models2.SysNoticeTargetList) ([]uint, error) {
	userIDSet := make(map[uint]struct{})
	var deptTargets []models2.SysNoticeTarget
	var roleIDs []uint
	var directUserIDs []uint
	hasAll := false

	for _, target := range targets {
		switch target.TargetType {
		case models2.SysNoticeTargetTypeAll:
			hasAll = true
		case models2.SysNoticeTargetTypeUser:
			directUserIDs = append(directUserIDs, target.TargetID)
		case models2.SysNoticeTargetTypeRole:
			roleIDs = append(roleIDs, target.TargetID)
		case models2.SysNoticeTargetTypeDept:
			deptTargets = append(deptTargets, *target)
		}
	}

	if hasAll {
		var allUserIDs []uint
		if err := tx.Model(&models.User{}).
			Where("tenant_id = ? AND status = ?", tenantID, 1).
			Pluck("id", &allUserIDs).Error; err != nil {
			return nil, err
		}
		for _, userID := range allUserIDs {
			userIDSet[userID] = struct{}{}
		}
	}

	if len(directUserIDs) > 0 {
		var validUserIDs []uint
		if err := tx.Model(&models.User{}).
			Where("tenant_id = ? AND status = ? AND id IN ?", tenantID, 1, uniqueUintSlice(directUserIDs)).
			Pluck("id", &validUserIDs).Error; err != nil {
			return nil, err
		}
		for _, userID := range validUserIDs {
			userIDSet[userID] = struct{}{}
		}
	}

	if len(roleIDs) > 0 {
		var roleUserIDs []uint
		if err := tx.Table("sys_user_role").
			Joins("JOIN sys_users ON sys_users.id = sys_user_role.user_id").
			Joins("JOIN sys_role ON sys_role.id = sys_user_role.role_id").
			Where("sys_user_role.role_id IN ? AND sys_role.tenant_id = ? AND sys_users.tenant_id = ? AND sys_users.status = ?",
				uniqueUintSlice(roleIDs), tenantID, tenantID, 1).
			Distinct().
			Pluck("sys_user_role.user_id", &roleUserIDs).Error; err != nil {
			return nil, err
		}
		for _, userID := range roleUserIDs {
			userIDSet[userID] = struct{}{}
		}
	}

	if len(deptTargets) > 0 {
		deptIDs, err := s.expandDepartmentTargetsTx(tx, tenantID, deptTargets)
		if err != nil {
			return nil, err
		}
		if len(deptIDs) > 0 {
			var deptUserIDs []uint
			if err := tx.Model(&models.User{}).
				Where("tenant_id = ? AND status = ? AND dept_id IN ?", tenantID, 1, deptIDs).
				Pluck("id", &deptUserIDs).Error; err != nil {
				return nil, err
			}
			for _, userID := range deptUserIDs {
				userIDSet[userID] = struct{}{}
			}
		}
	}

	return mapKeysToSlice(userIDSet), nil
}

func (s *SysNoticeService) expandDepartmentTargetsTx(tx *gorm.DB, tenantID uint, targets []models2.SysNoticeTarget) ([]uint, error) {
	if len(targets) == 0 {
		return nil, nil
	}

	deptIDSet := make(map[uint]struct{})
	needTree := false
	for _, target := range targets {
		if target.TargetID == 0 {
			continue
		}
		deptIDSet[target.TargetID] = struct{}{}
		if target.IncludeChildren {
			needTree = true
		}
	}

	if !needTree {
		return mapKeysToSlice(deptIDSet), nil
	}

	var allDepartments models.SysDepartmentList
	if err := tx.Where("tenant_id = ?", tenantID).Find(&allDepartments).Error; err != nil {
		return nil, err
	}
	departmentTree := allDepartments.BuildTree()

	for _, target := range targets {
		if !target.IncludeChildren || target.TargetID == 0 {
			continue
		}
		childIDs, err := getDepartmentAndChildrenIDs(departmentTree, target.TargetID)
		if err != nil {
			return nil, err
		}
		for _, deptID := range childIDs {
			deptIDSet[deptID] = struct{}{}
		}
	}

	return mapKeysToSlice(deptIDSet), nil
}

func getDepartmentAndChildrenIDs(departmentTree models.SysDepartmentList, deptID uint) ([]uint, error) {
	if departmentTree.IsEmpty() {
		return []uint{deptID}, nil
	}

	var findDepartment func(depts models.SysDepartmentList, targetID uint) *models.SysDepartment
	findDepartment = func(depts models.SysDepartmentList, targetID uint) *models.SysDepartment {
		for _, dept := range depts {
			if dept.ID == targetID {
				return dept
			}
			if len(dept.Children) > 0 {
				if found := findDepartment(dept.Children, targetID); found != nil {
					return found
				}
			}
		}
		return nil
	}

	targetDept := findDepartment(departmentTree, deptID)
	if targetDept == nil {
		return []uint{deptID}, nil
	}

	var result []uint
	var collect func(dept *models.SysDepartment)
	collect = func(dept *models.SysDepartment) {
		result = append(result, dept.ID)
		for _, child := range dept.Children {
			collect(child)
		}
	}
	collect(targetDept)

	return result, nil
}

func (s *SysNoticeService) getNoticeForManage(c *gin.Context, noticeID, tenantID uint) (*models2.SysNotice, error) {
	notice := models2.NewSysNotice()
	err := app.DB().WithContext(c).
		Where("id = ? AND tenant_id = ?", noticeID, tenantID).
		First(notice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("通知不存在")
		}
		return nil, err
	}
	return notice, nil
}

func (s *SysNoticeService) getNoticeWithTargets(c *gin.Context, noticeID, tenantID uint) (*models2.SysNotice, error) {
	notice := models2.NewSysNotice()
	err := app.DB().WithContext(c).
		Where("id = ? AND tenant_id = ?", noticeID, tenantID).
		Preload("Publisher").
		Preload("Targets").
		First(notice).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("通知不存在")
		}
		return nil, err
	}
	return notice, nil
}

func (s *SysNoticeService) fillTargetNames(c *gin.Context, targets models2.SysNoticeTargetList, tenantID uint) error {
	if len(targets) == 0 {
		return nil
	}

	var userIDs []uint
	var roleIDs []uint
	var deptIDs []uint

	for _, target := range targets {
		switch target.TargetType {
		case models2.SysNoticeTargetTypeAll:
			target.TargetName = "全体用户"
		case models2.SysNoticeTargetTypeUser:
			userIDs = append(userIDs, target.TargetID)
		case models2.SysNoticeTargetTypeRole:
			roleIDs = append(roleIDs, target.TargetID)
		case models2.SysNoticeTargetTypeDept:
			deptIDs = append(deptIDs, target.TargetID)
		}
	}

	userNameMap, err := s.loadUserNameMap(c, tenantID, userIDs)
	if err != nil {
		return err
	}
	roleNameMap, err := s.loadRoleNameMap(c, tenantID, roleIDs)
	if err != nil {
		return err
	}
	deptNameMap, err := s.loadDepartmentNameMap(c, tenantID, deptIDs)
	if err != nil {
		return err
	}

	for _, target := range targets {
		switch target.TargetType {
		case models2.SysNoticeTargetTypeUser:
			target.TargetName = userNameMap[target.TargetID]
		case models2.SysNoticeTargetTypeRole:
			target.TargetName = roleNameMap[target.TargetID]
		case models2.SysNoticeTargetTypeDept:
			target.TargetName = deptNameMap[target.TargetID]
			if target.IncludeChildren && target.TargetName != "" {
				target.TargetName += "（含下级）"
			}
		}
	}

	return nil
}

func (s *SysNoticeService) loadUserNameMap(c *gin.Context, tenantID uint, userIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(userIDs) == 0 {
		return result, nil
	}
	var users []models.User
	err := app.DB().WithContext(c).
		Select("id", "username", "nick_name").
		Where("tenant_id = ? AND id IN ?", tenantID, uniqueUintSlice(userIDs)).
		Find(&users).Error
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.NickName != "" {
			result[user.ID] = user.NickName
		} else {
			result[user.ID] = user.Username
		}
	}
	return result, nil
}

func (s *SysNoticeService) loadRoleNameMap(c *gin.Context, tenantID uint, roleIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(roleIDs) == 0 {
		return result, nil
	}
	var roles models.SysRoleList
	err := app.DB().WithContext(c).
		Select("id", "name").
		Where("tenant_id = ? AND id IN ?", tenantID, uniqueUintSlice(roleIDs)).
		Find(&roles).Error
	if err != nil {
		return nil, err
	}
	for _, role := range roles {
		result[role.ID] = role.Name
	}
	return result, nil
}

func (s *SysNoticeService) loadDepartmentNameMap(c *gin.Context, tenantID uint, deptIDs []uint) (map[uint]string, error) {
	result := make(map[uint]string)
	if len(deptIDs) == 0 {
		return result, nil
	}
	var departments models.SysDepartmentList
	err := app.DB().WithContext(c).
		Select("id", "name").
		Where("tenant_id = ? AND id IN ?", tenantID, uniqueUintSlice(deptIDs)).
		Find(&departments).Error
	if err != nil {
		return nil, err
	}
	for _, dept := range departments {
		result[dept.ID] = dept.Name
	}
	return result, nil
}

func (s *SysNoticeService) fillNoticeStats(c *gin.Context, notices models2.SysNoticeList) error {
	if len(notices) == 0 {
		return nil
	}

	var noticeIDs []uint
	for _, notice := range notices {
		if notice != nil && notice.ID > 0 {
			noticeIDs = append(noticeIDs, notice.ID)
		}
	}
	if len(noticeIDs) == 0 {
		return nil
	}

	type noticeCountRow struct {
		NoticeID uint
		Count    int64
	}

	var recipientRows []noticeCountRow
	err := app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
		Select("notice_id, COUNT(*) AS count").
		Where("notice_id IN ?", uniqueUintSlice(noticeIDs)).
		Group("notice_id").
		Scan(&recipientRows).Error
	if err != nil {
		return err
	}

	var unreadRows []noticeCountRow
	err = app.DB().WithContext(c).Model(&models2.SysNoticeRecipient{}).
		Select("notice_id, COUNT(*) AS count").
		Where("notice_id IN ? AND read_status = ?", uniqueUintSlice(noticeIDs), models2.SysNoticeReadStatusUnread).
		Group("notice_id").
		Scan(&unreadRows).Error
	if err != nil {
		return err
	}

	recipientCountMap := make(map[uint]int64)
	for _, row := range recipientRows {
		recipientCountMap[row.NoticeID] = row.Count
	}
	unreadCountMap := make(map[uint]int64)
	for _, row := range unreadRows {
		unreadCountMap[row.NoticeID] = row.Count
	}

	for _, notice := range notices {
		if notice == nil {
			continue
		}
		notice.RecipientCount = recipientCountMap[notice.ID]
		notice.UnreadCount = unreadCountMap[notice.ID]
	}

	return nil
}

func (s *SysNoticeService) dispatchPublishedNotice(c *gin.Context, noticeID uint) {
	if s.dispatcher == nil {
		return
	}

	notice, err := s.GetByID(c, noticeID)
	if err != nil {
		app.ZapLog.Warn("加载已发布通知详情失败", zap.Error(err), zap.Uint("notice_id", noticeID))
		return
	}

	recipients := models2.NewSysNoticeRecipientList()
	if err := recipients.Find(c, func(db *gorm.DB) *gorm.DB {
		return db.Where("notice_id = ?", noticeID)
	}); err != nil {
		app.ZapLog.Warn("加载通知接收人失败", zap.Error(err), zap.Uint("notice_id", noticeID))
		return
	}

	if err := s.dispatcher.Dispatch(c, notice, recipients); err != nil {
		app.ZapLog.Warn("通知实时分发失败", zap.Error(err), zap.Uint("notice_id", noticeID))
	}
}

func uniqueUintSlice(values []uint) []uint {
	if len(values) == 0 {
		return values
	}
	seen := make(map[uint]struct{}, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 {
			continue
		}
		if _, exists := seen[value]; exists {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}

func mapKeysToSlice(values map[uint]struct{}) []uint {
	result := make([]uint, 0, len(values))
	for value := range values {
		result = append(result, value)
	}
	return result
}
