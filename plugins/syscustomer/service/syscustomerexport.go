package service

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gin-fast/app/global/app"
	appConsts "gin-fast/app/global/consts"
	appModels "gin-fast/app/models"
	"gin-fast/app/utils/common"
	"gin-fast/app/utils/datascope"
	"gin-fast/app/utils/filehelper"
	"gin-fast/app/utils/tenanthelper"
	"gin-fast/exampleutils/snowflakehelper"
	channelCompanyModels "gin-fast/plugins/syschannelcompany/models"
	"gin-fast/plugins/syscustomer/models"
	noticeService "gin-fast/plugins/sysnotice/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

const (
	customerExportBatchSize        = 2000
	customerAsyncExportRelativeDir = "export/syscustomer"
	customerExportModeSync         = "sync"
	customerExportModeAsync        = "async"
	customerExportFileSuffix       = ".csv"
)

type sysCustomerExportLookup struct {
	dictMaps          map[string]map[string]string
	channelNames      map[int]string
	userNames         map[int]string
	departmentNames   map[int]string
	customerValidName map[int]string
}

type sysDictExportItem struct {
	Code  string
	Value string
	Name  string
}

type customerExtraExportInfo struct {
	ProgressRemark   string
	IntentionValidID int
}

// SubmitExport 先判断本次导出走同步还是异步，避免大批量导出阻塞请求。
func (s *SysCustomerService) SubmitExport(c *gin.Context, req models.SysCustomerListRequest) (*models.SysCustomerExportSubmitResult, error) {
	total, err := s.CountExport(c, req)
	if err != nil {
		return nil, err
	}

	result := &models.SysCustomerExportSubmitResult{
		Mode:  customerExportModeSync,
		Total: total,
	}
	if total <= getCustomerAsyncExportThreshold() {
		return result, nil
	}

	claims := common.GetClaims(c)
	if claims == nil {
		return nil, fmt.Errorf("当前登录状态已失效")
	}

	result.Mode = customerExportModeAsync
	result.Message = fmt.Sprintf("本次导出共 %d 条数据，已转为异步导出，完成后会通过右上角实时通知提醒下载。", total)

	reqCopy := cloneSysCustomerListRequest(req)
	claimsCopy := cloneExportClaims(claims)
	go s.runAsyncExport(reqCopy, claimsCopy, total)

	return result, nil
}

func (s *SysCustomerService) CountExport(c *gin.Context, req models.SysCustomerListRequest) (int64, error) {
	var total int64
	err := buildCustomerExportQuery(c, c, req).Count(&total).Error
	return total, err
}

// ExportCSV 执行同步导出，并在导出前按配置顺手清理当前租户的历史导出文件。
func (s *SysCustomerService) ExportCSV(c *gin.Context, req models.SysCustomerListRequest) error {
	if claims := common.GetClaims(c); claims != nil {
		s.tryCleanupExpiredAsyncExports(c, claims.TenantID, claims.UserID)
	}

	total, err := s.CountExport(c, req)
	if err != nil {
		return err
	}

	filename, err := buildCustomerExportFilename(req, total)
	if err != nil {
		return err
	}
	c.Header("Content-Type", "text/csv; charset=utf-8")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Transfer-Encoding", "binary")
	c.Header("Cache-Control", "no-cache")

	if err := writeCSVUTF8BOM(c.Writer); err != nil {
		return err
	}

	var flusher http.Flusher
	if value, ok := c.Writer.(http.Flusher); ok {
		flusher = value
	}

	return s.exportCSVToWriter(c, c, c.Writer, req, flusher)
}

// exportCSVToWriter 统一输出 CSV 内容。
// 同步导出写入响应流，异步导出写入本地文件，导出字段保持一致。
func (s *SysCustomerService) exportCSVToWriter(
	dbCtx context.Context,
	authCtx *gin.Context,
	output io.Writer,
	req models.SysCustomerListRequest,
	flusher http.Flusher,
) error {
	lookup, err := s.loadExportLookup(dbCtx, authCtx)
	if err != nil {
		return err
	}

	writer := csv.NewWriter(output)
	defer writer.Flush()

	headers := []string{
		"客户编号",
		"客户姓名",
		"手机号",
		"业务阶段",
		"客户有效",
		"星级",
		"渠道来源",
		"跟进人",
		"需求金额",
		"客户备注",
		"分配时间",
		"所属部门",
		"所在城市",
		"客户来源",
		"再分配",
		"离职数据",
		"重复标记",
		"短信状态",
		"星级回传",
		"锁定状态",
		"创建时间",
		"更新时间",
	}
	if err = writer.Write(headers); err != nil {
		return err
	}
	writer.Flush()
	if err = writer.Error(); err != nil {
		return err
	}
	if flusher != nil {
		flusher.Flush()
	}

	baseQuery := buildCustomerExportQuery(dbCtx, authCtx, req)
	lastID := 0

	for {
		var batch []models.SysCustomer
		query := baseQuery.Session(&gorm.Session{})
		if err = query.Where("id > ?", lastID).Order("id ASC").Limit(customerExportBatchSize).Find(&batch).Error; err != nil {
			return err
		}
		if len(batch) == 0 {
			break
		}

		for _, item := range batch {
			record := s.buildCustomerExportRecord(item, lookup)
			if err = writer.Write(record); err != nil {
				return err
			}
			lastID = item.Id
		}

		writer.Flush()
		if err = writer.Error(); err != nil {
			return err
		}
		if flusher != nil {
			flusher.Flush()
		}
	}

	return nil
}

func buildCustomerExportQuery(dbCtx context.Context, authCtx *gin.Context, req models.SysCustomerListRequest) *gorm.DB {
	currentUserID := int(common.GetCurrentUserID(authCtx))
	return app.DB().WithContext(dbCtx).
		Model(&models.SysCustomer{}).
		Scopes(
			req.Handle(),
			func(db *gorm.DB) *gorm.DB {
				return req.ApplyListScene(db, currentUserID)
			},
			datascope.GetDataScopeByColumn(authCtx, ""),
			tenanthelper.TenantScope(authCtx),
		)
}

// runAsyncExport 执行异步导出任务。
// 这里会先按配置清理当前租户的历史导出文件，再生成新的导出文件并推送实时通知。
func (s *SysCustomerService) runAsyncExport(req models.SysCustomerListRequest, claims app.Claims, total int64) {
	authCtx := buildExportAuthContext(claims)
	dbCtx := context.Background()

	if err := s.cleanupExpiredAsyncExports(dbCtx, claims.TenantID); err != nil {
		app.ZapLog.Warn("清理过期客户导出文件失败",
			zap.Error(err),
			zap.Uint("user_id", claims.UserID),
			zap.Uint("tenant_id", claims.TenantID),
		)
	}

	affix, err := s.generateAsyncExportFile(dbCtx, authCtx, req, claims, total)
	if err != nil {
		app.ZapLog.Error("异步导出客户失败",
			zap.Error(err),
			zap.Uint("user_id", claims.UserID),
			zap.Uint("tenant_id", claims.TenantID),
		)
		s.notifyAsyncExportFailed(dbCtx, claims, err)
		return
	}

	if err := noticeService.NewNoticeBusinessService().NotifyExportReady(
		dbCtx,
		claims.TenantID,
		claims.UserID,
		"客户导出已完成",
		fmt.Sprintf("已为您生成 %d 条客户数据导出文件，点击即可下载。", total),
		strconv.FormatUint(uint64(affix.ID), 10),
		map[string]interface{}{
			"affixId":  affix.ID,
			"fileName": affix.Name,
			"total":    total,
		},
	); err != nil {
		app.ZapLog.Warn("发送客户导出完成通知失败",
			zap.Error(err),
			zap.Uint("user_id", claims.UserID),
			zap.Uint("tenant_id", claims.TenantID),
			zap.Uint("affix_id", affix.ID),
		)
	}
}

func (s *SysCustomerService) notifyAsyncExportFailed(ctx context.Context, claims app.Claims, exportErr error) {
	if err := noticeService.NewNoticeBusinessService().NotifyUsers(ctx, noticeService.NoticeBusinessNotifyRequest{
		TenantID: claims.TenantID,
		UserIDs:  []uint{claims.UserID},
		Payload: noticeService.NoticeBusinessPayload{
			Scene:   noticeService.NoticeBusinessSceneSystemInternal,
			Title:   "客户导出失败",
			Content: fmt.Sprintf("客户导出任务执行失败，请稍后重试。原因：%s", strings.TrimSpace(exportErr.Error())),
			Level:   "error",
		},
	}); err != nil {
		app.ZapLog.Warn("发送客户导出失败通知失败",
			zap.Error(err),
			zap.Uint("user_id", claims.UserID),
			zap.Uint("tenant_id", claims.TenantID),
		)
	}
}

// generateAsyncExportFile 负责把异步导出结果写入本地上传目录，
// 并同步生成 sys_affix 记录，供通知弹窗直接下载。
func (s *SysCustomerService) generateAsyncExportFile(
	dbCtx context.Context,
	authCtx *gin.Context,
	req models.SysCustomerListRequest,
	claims app.Claims,
	total int64,
) (*appModels.SysAffix, error) {
	uploadConfig := app.UploadService.GetUploadConfig()
	localPath := strings.TrimSpace(uploadConfig.LocalPath)
	if localPath == "" {
		return nil, fmt.Errorf("未配置导出文件本地存储目录")
	}

	dateDir := time.Now().Format("2006-01-02")
	relativeDir := path.Join(customerAsyncExportRelativeDir, dateDir)
	targetDir := filepath.Join(localPath, filepath.FromSlash(relativeDir))
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return nil, err
	}

	filename, err := buildCustomerExportFilename(req, total)
	if err != nil {
		return nil, err
	}
	filePath := filepath.Join(targetDir, filename)
	file, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}

	success := false
	defer func() {
		_ = file.Close()
		if !success {
			_ = os.Remove(filePath)
		}
	}()

	if err = writeCSVUTF8BOM(file); err != nil {
		return nil, err
	}
	if err = s.exportCSVToWriter(dbCtx, authCtx, file, req, nil); err != nil {
		return nil, err
	}
	if err = file.Sync(); err != nil {
		return nil, err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	affix := appModels.NewSysAffix()
	affix.Name = filename
	affix.Path = filePath
	affix.Url = app.UploadService.GetFileUrl(path.Join(relativeDir, filename))
	affix.Size = int(fileInfo.Size())
	affix.Suffix = customerExportFileSuffix
	affix.Ftype = filehelper.GetFileTypeBySuffix(customerExportFileSuffix)
	affix.CreatedBy = claims.UserID
	affix.TenantID = claims.TenantID

	if err = affix.Create(dbCtx); err != nil {
		return nil, err
	}

	success = true
	return affix, nil
}

// cleanupExpiredAsyncExports 仅清理客户导出生成的历史文件与附件记录。
// 范围固定在 export/syscustomer，避免影响其他上传文件。
func (s *SysCustomerService) cleanupExpiredAsyncExports(ctx context.Context, tenantID uint) error {
	if tenantID == 0 {
		return nil
	}

	cutoff := time.Now().Add(-getCustomerExportCleanupDuration())
	urlPrefix := strings.TrimSpace(app.UploadService.GetFileUrl(customerAsyncExportRelativeDir))
	if urlPrefix == "" {
		return fmt.Errorf("客户导出文件 URL 前缀不能为空")
	}

	var affixes []appModels.SysAffix
	if err := app.DB().WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("url LIKE ?", urlPrefix+"%").
		Where("created_at < ?", cutoff).
		Find(&affixes).Error; err != nil {
		return err
	}

	for _, affix := range affixes {
		if strings.TrimSpace(affix.Path) != "" {
			if err := app.UploadService.DeleteFile(affix.Path); err != nil && !os.IsNotExist(err) {
				app.ZapLog.Warn("删除过期客户导出文件失败",
					zap.Error(err),
					zap.Uint("affix_id", affix.ID),
					zap.Uint("tenant_id", tenantID),
				)
			}
		}

		if err := app.DB().WithContext(ctx).Delete(&appModels.SysAffix{}, affix.ID).Error; err != nil {
			app.ZapLog.Warn("删除过期客户导出附件记录失败",
				zap.Error(err),
				zap.Uint("affix_id", affix.ID),
				zap.Uint("tenant_id", tenantID),
			)
		}
	}

	return nil
}

// tryCleanupExpiredAsyncExports 是导出链路里的兜底清理。
// 清理失败只记录日志，不影响当前导出继续执行。
func (s *SysCustomerService) tryCleanupExpiredAsyncExports(ctx context.Context, tenantID uint, userID uint) {
	if err := s.cleanupExpiredAsyncExports(ctx, tenantID); err != nil {
		app.ZapLog.Warn("清理过期客户导出文件失败",
			zap.Error(err),
			zap.Uint("user_id", userID),
			zap.Uint("tenant_id", tenantID),
		)
	}
}

// getCustomerExportCleanupDuration 从配置读取导出文件保留天数。
// export_clean = 0 表示当前导出前立马清理历史客户导出文件；
// export_clean < 0 或未配置时默认按 3 天处理。
func getCustomerExportCleanupDuration() time.Duration {
	days := app.ConfigYml.GetInt("platform.export_clean")
	if days == 0 {
		return 0
	}
	if days < 0 {
		days = 3
	}
	return time.Duration(days) * 24 * time.Hour
}

// getCustomerAsyncExportThreshold 从配置读取异步导出阈值。
// export_async_threshold 小于 0 时按 0 处理，方便测试时强制全部走异步。
func getCustomerAsyncExportThreshold() int64 {
	threshold := app.ConfigYml.GetInt("platform.export_async_threshold")
	if threshold < 0 {
		return 0
	}
	return int64(threshold)
}

func cloneSysCustomerListRequest(req models.SysCustomerListRequest) models.SysCustomerListRequest {
	payload, err := json.Marshal(req)
	if err != nil {
		return req
	}

	var cloned models.SysCustomerListRequest
	if err = json.Unmarshal(payload, &cloned); err != nil {
		return req
	}

	return cloned
}

func cloneExportClaims(claims *app.Claims) app.Claims {
	if claims == nil {
		return app.Claims{}
	}

	cloned := *claims
	return cloned
}

func buildExportAuthContext(claims app.Claims) *gin.Context {
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	ctx.Set(appConsts.BindContextKeyName, &claims)
	return ctx
}

func buildCustomerExportFilename(req models.SysCustomerListRequest, total int64) (string, error) {
	now := time.Now()
	uniqueID, err := snowflakehelper.GenerateIDUint64()
	if err != nil {
		return "", err
	}

	sceneName := resolveCustomerExportSceneName(req.Scene)
	shortCode := strings.ToUpper(strconv.FormatUint(uniqueID, 36))
	return fmt.Sprintf("%s_%s_%s_%s_%d%s", sceneName, now.Format("20060102"), now.Format("150405"), shortCode, total, customerExportFileSuffix), nil
}

func resolveCustomerExportSceneName(scene *string) string {
	if scene == nil {
		return "全部客户"
	}

	switch strings.ToLower(strings.TrimSpace(*scene)) {
	case models.CustomerListSceneAll:
		return "全部客户"
	case models.CustomerListSceneMy:
		return "我的客户"
	case models.CustomerListScenePublic:
		return "公共池客户"
	case models.CustomerListSceneExchange:
		return "待流转客户"
	case models.CustomerListSceneReassign:
		return "再分配客户"
	case models.CustomerListSceneLocked:
		return "锁定客户"
	case models.CustomerListSceneIntention2:
		return "无效客户"
	case models.CustomerListSceneIntention3:
		return "黑名单客户"
	case models.CustomerListSceneStatus0:
		return "待处理客户"
	default:
		return "全部客户"
	}
}

func writeCSVUTF8BOM(writer io.Writer) error {
	_, err := writer.Write([]byte{0xEF, 0xBB, 0xBF})
	return err
}

func (s *SysCustomerService) loadExportLookup(dbCtx context.Context, authCtx *gin.Context) (*sysCustomerExportLookup, error) {
	dictMaps, err := s.loadCustomerDictMaps(dbCtx)
	if err != nil {
		return nil, err
	}

	channelNames, err := s.loadChannelNames(dbCtx, authCtx)
	if err != nil {
		return nil, err
	}

	userNames, err := s.loadUserNames(dbCtx, authCtx)
	if err != nil {
		return nil, err
	}

	departmentNames, err := s.loadDepartmentNames(dbCtx, authCtx)
	if err != nil {
		return nil, err
	}

	customerValidNames, err := s.loadCustomerValidNames(dbCtx, authCtx)
	if err != nil {
		return nil, err
	}

	return &sysCustomerExportLookup{
		dictMaps:          dictMaps,
		channelNames:      channelNames,
		userNames:         userNames,
		departmentNames:   departmentNames,
		customerValidName: customerValidNames,
	}, nil
}

func (s *SysCustomerService) loadCustomerDictMaps(dbCtx context.Context) (map[string]map[string]string, error) {
	dictCodes := []string{
		"customerStar",
		"progressStatusArr",
		"intentionStatusArr",
		"isStatus",
		"starStatus",
		"from",
		"isSms",
		"singlePieceTypeArr",
	}

	var rows []sysDictExportItem
	err := app.DB().WithContext(dbCtx).
		Table("sys_dict_item").
		Select("sys_dict.code AS code, sys_dict_item.value AS value, sys_dict_item.name AS name").
		Joins("JOIN sys_dict ON sys_dict_item.dict_id = sys_dict.id").
		Where("sys_dict.code IN ?", dictCodes).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[string]map[string]string, len(dictCodes))
	for _, row := range rows {
		if _, ok := result[row.Code]; !ok {
			result[row.Code] = map[string]string{}
		}
		result[row.Code][row.Value] = row.Name
	}
	return result, nil
}

func (s *SysCustomerService) loadChannelNames(dbCtx context.Context, authCtx *gin.Context) (map[int]string, error) {
	var rows []channelCompanyModels.SysChannelCompany
	err := app.DB().WithContext(dbCtx).
		Model(&channelCompanyModels.SysChannelCompany{}).
		Scopes(tenanthelper.TenantScope(authCtx)).
		Select("id", "hidden_name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(rows))
	for _, row := range rows {
		result[row.Id] = row.HiddenName
	}
	return result, nil
}

func (s *SysCustomerService) loadUserNames(dbCtx context.Context, authCtx *gin.Context) (map[int]string, error) {
	var rows []appModels.User
	err := app.DB().WithContext(dbCtx).
		Model(&appModels.User{}).
		Scopes(tenanthelper.TenantScope(authCtx)).
		Select("id", "nick_name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(rows))
	for _, row := range rows {
		result[int(row.ID)] = row.NickName
	}
	return result, nil
}

func (s *SysCustomerService) loadDepartmentNames(dbCtx context.Context, authCtx *gin.Context) (map[int]string, error) {
	var rows []appModels.SysDepartment
	err := app.DB().WithContext(dbCtx).
		Model(&appModels.SysDepartment{}).
		Scopes(tenanthelper.TenantScope(authCtx)).
		Select("id", "name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(rows))
	for _, row := range rows {
		result[int(row.ID)] = row.Name
	}
	return result, nil
}

func (s *SysCustomerService) loadCustomerValidNames(dbCtx context.Context, authCtx *gin.Context) (map[int]string, error) {
	var rows []appModels.CustomerValid
	err := app.DB().WithContext(dbCtx).
		Model(&appModels.CustomerValid{}).
		Scopes(tenanthelper.TenantScope(authCtx)).
		Select("id", "name").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int]string, len(rows))
	for _, row := range rows {
		result[int(row.ID)] = row.Name
	}
	return result, nil
}

func (s *SysCustomerService) buildCustomerExportRecord(item models.SysCustomer, lookup *sysCustomerExportLookup) []string {
	extraInfo := parseCustomerExtraExportInfo(item.Extra)
	statusName := dictLookupName(lookup.dictMaps, "progressStatusArr", item.Status, strconv.Itoa(item.Status))
	intentionName := dictLookupName(lookup.dictMaps, "intentionStatusArr", item.Intention, strconv.Itoa(item.Intention))
	customerStarName := "未定级"
	if item.CustomerStar != nil {
		customerStarName = dictLookupName(lookup.dictMaps, "customerStar", *item.CustomerStar, strconv.Itoa(*item.CustomerStar))
	}

	statusDisplay := statusName
	if extraInfo.ProgressRemark != "" {
		statusDisplay = strings.TrimSpace(statusDisplay + " - " + extraInfo.ProgressRemark)
	}

	intentionDisplay := intentionName
	if item.Intention != 0 && extraInfo.IntentionValidID > 0 {
		if validName := lookup.customerValidName[extraInfo.IntentionValidID]; validName != "" {
			intentionDisplay = strings.TrimSpace(intentionDisplay + " - " + validName)
		}
	}

	record := []string{
		sanitizeCSVCell(item.Num),
		sanitizeCSVCell(defaultString(item.Name, "未命名客户")),
		sanitizeCSVCell(item.Mobile),
		sanitizeCSVCell(statusDisplay),
		sanitizeCSVCell(intentionDisplay),
		sanitizeCSVCell(customerStarName),
		sanitizeCSVCell(lookup.channelNames[item.ChannelID]),
		sanitizeCSVCell(lookup.userNames[item.UserID]),
		sanitizeCSVCell(formatIntValue(item.MoneyDemand)),
		sanitizeCSVCell(item.Remarks),
		sanitizeCSVCell(formatTimePointer(item.AllotTime)),
		sanitizeCSVCell(lookup.departmentNames[item.DeptID]),
		sanitizeCSVCell(item.City),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "from", item.From, strconv.Itoa(item.From))),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "isStatus", item.IsReassign, strconv.Itoa(item.IsReassign))),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "isStatus", item.IsQuit, strconv.Itoa(item.IsQuit))),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "isStatus", item.IsRepeat, strconv.Itoa(item.IsRepeat))),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "isSms", item.IsSms, strconv.Itoa(item.IsSms))),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "starStatus", item.StarStatus, strconv.Itoa(item.StarStatus))),
		sanitizeCSVCell(dictLookupName(lookup.dictMaps, "isStatus", item.IsLock, strconv.Itoa(item.IsLock))),
		sanitizeCSVCell(formatTimePointer(item.CreatedAt)),
		sanitizeCSVCell(formatTimePointer(item.UpdatedAt)),
	}

	return record
}

func parseCustomerExtraExportInfo(extra string) customerExtraExportInfo {
	info := customerExtraExportInfo{}
	if strings.TrimSpace(extra) == "" {
		return info
	}

	var extraMap map[string]interface{}
	if err := json.Unmarshal([]byte(extra), &extraMap); err != nil {
		return info
	}

	if progressRemark, ok := extraMap[models.ProgressRemark].(string); ok {
		info.ProgressRemark = strings.TrimSpace(progressRemark)
	}

	switch value := extraMap[models.IntentionValidId].(type) {
	case float64:
		info.IntentionValidID = int(value)
	case float32:
		info.IntentionValidID = int(value)
	case int:
		info.IntentionValidID = value
	case int64:
		info.IntentionValidID = int(value)
	case json.Number:
		if parsed, err := value.Int64(); err == nil {
			info.IntentionValidID = int(parsed)
		}
	case string:
		if parsed, err := strconv.Atoi(strings.TrimSpace(value)); err == nil {
			info.IntentionValidID = parsed
		}
	}

	return info
}

func dictLookupName(dictMaps map[string]map[string]string, dictCode string, value int, fallback string) string {
	options, ok := dictMaps[dictCode]
	if !ok {
		return fallback
	}
	if name, exists := options[strconv.Itoa(value)]; exists && strings.TrimSpace(name) != "" {
		return name
	}
	return fallback
}

func formatTimePointer(value *time.Time) string {
	if value == nil || value.IsZero() {
		return ""
	}
	return value.Format("2006-01-02 15:04:05")
}

func formatIntValue(value int) string {
	if value == 0 {
		return "0"
	}
	return strconv.Itoa(value)
}

func defaultString(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func sanitizeCSVCell(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	switch trimmed[0] {
	case '=', '+', '-', '@':
		return "'" + value
	default:
		return value
	}
}
