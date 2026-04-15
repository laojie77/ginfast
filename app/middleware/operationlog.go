package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"gin-fast/app/global/app"
	"gin-fast/app/models"
	"gin-fast/app/utils/common"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// OperationLogMiddleware 操作日志中间件
func OperationLogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 跳过不需要记录日志的请求
		if shouldSkipLog(c) {
			c.Next()
			return
		}

		startTime := time.Now()

		// 复制请求体用于记录
		var requestBody []byte
		if c.Request.Body != nil && !isMultipartRequest(c) {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}
		//创建自定义的ResponseWriter来捕获响应
		writer := &responseWriter{
			body:           bytes.NewBuffer(nil),
			ResponseWriter: c.Writer,
		}
		c.Writer = writer
		//记录操作日志
		defer func() {
			log := buildOperationLog(c, startTime, requestBody, writer.body.Bytes())
			if log != nil {
				go saveOperationLog(log)
			}
		}()

		c.Next()
	}
}

// responseWriter 自定义ResponseWriter用于捕获响应数据
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// shouldSkipLog 判断是否需要跳过日志记录
func shouldSkipLog(c *gin.Context) bool {
	// 跳过静态文件、健康检查等请求
	skipPaths := []string{
		"/swagger/",
		"/favicon.ico",
		"/health",
		"/metrics",
		"/api/refreshToken",  // 刷新token
		"/api/captcha/id",    // 生成验证码ID
		"/api/captcha/image", // 获取验证码图片
		"/api/config/get",    // 获取配置信息
	}

	path := c.Request.URL.Path
	for _, skipPath := range skipPaths {
		if strings.Contains(path, skipPath) {
			return true
		}
	}

	return false
}

// recordOperationLog 记录操作日志
func buildOperationLog(c *gin.Context, startTime time.Time, requestBody, responseBody []byte) *models.SysOperationLog {
	if c == nil || c.Request == nil {
		return nil
	}

	duration := time.Since(startTime).Milliseconds()
	// 获取用户信息
	var userID uint
	var username string
	var tenantID uint
	operationType := getOperationType(c)
	// 尝试从JWT token获取用户信息
	claims := common.GetClaims(c)
	if claims != nil {
		userID = claims.UserID
		username = claims.Username
		tenantID = claims.TenantID
	} else if c.Request.URL.Path == "/api/login" && c.Request.Method == "POST" && len(requestBody) > 0 {
		var loginReq struct {
			Username string `json:"username"`
		}
		if err := json.Unmarshal(requestBody, &loginReq); err == nil && loginReq.Username != "" {
			username = loginReq.Username
			operationType = models.OperationLogin
		}
	}
	// 构建操作日志
	return &models.SysOperationLog{
		UserID:      userID,
		Username:    username,
		Module:      getOperationModule(c),
		Operation:   operationType,
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		IP:          c.ClientIP(),
		UserAgent:   c.Request.UserAgent(),
		RequestData: sanitizeRequestData(c, requestBody),
		StatusCode:  c.Writer.Status(),
		Duration:    duration,
		ErrorMsg:    getErrorMessage(c, responseBody),
		Location:    getLocationByIP(c.ClientIP()),
		TenantID:    tenantID,
	}
}

func saveOperationLog(log *models.SysOperationLog) {
	if log == nil {
		return
	}

	if err := app.DB().Create(log).Error; err != nil {
		app.ZapLog.Error("记录操作日志失败", zap.Error(err))
	}
}

// getOperation 获取操作类型
func getOperationModule(c *gin.Context) string {
	path := c.Request.URL.Path
	if strings.Contains(path, "/users") {
		return "用户管理"
	} else if strings.Contains(path, "/sysMenu") {
		return "菜单管理"
	} else if strings.Contains(path, "/sysRole") {
		return "角色管理"
	} else if strings.Contains(path, "/sysDepartment") {
		return "部门管理"
	} else if strings.Contains(path, "/sysDict") {
		return "字典管理"
	} else if strings.Contains(path, "/sysApi") {
		return "API管理"
	} else if strings.Contains(path, "/sysAffix") {
		return "文件管理"
	} else if strings.Contains(path, "/config") {
		return "系统配置"
	} else if strings.Contains(path, "/sysOperationLog") {
		return "操作日志管理"
	}
	return "其他"
}

func getOperationType(c *gin.Context) string {
	path := strings.ToLower(c.Request.URL.Path)
	if strings.Contains(path, "/import") {
		return models.OperationImport
	}
	if strings.Contains(path, "/export") {
		return models.OperationExport
	}

	switch c.Request.Method {
	case "POST":
		return models.OperationCreate
	case "PUT", "PATCH":
		return models.OperationUpdate
	case "DELETE":
		return models.OperationDelete
	case "GET":
		return models.OperationQuery
	default:
		return "unknown"
	}
}

func getErrorMessage(c *gin.Context, responseBody []byte) string {
	if c.Writer.Status() < 400 {
		return ""
	}

	if err, exists := c.Get("error"); exists {
		if target, ok := err.(error); ok && target != nil {
			return target.Error()
		}
	}

	if len(responseBody) > 0 {
		var response map[string]interface{}
		if err := json.Unmarshal(responseBody, &response); err == nil {
			if msg, ok := response["message"].(string); ok && msg != "" {
				return msg
			}
		}
	}

	return "请求处理失败"
}

// getLocationByIP 根据IP获取地理位置（简化实现）
func getLocationByIP(ip string) string {
	// 这里可以集成第三方IP地理位置服务
	// 简化实现：返回空字符串
	return ""
}

func isMultipartRequest(c *gin.Context) bool {
	if c == nil || c.Request == nil {
		return false
	}

	contentType := strings.ToLower(strings.TrimSpace(c.Request.Header.Get("Content-Type")))
	return strings.HasPrefix(contentType, "multipart/form-data")
}

// sanitizeRequestData 对请求数据进行脱敏处理
func sanitizeRequestData(c *gin.Context, data []byte) string {
	if isMultipartRequest(c) {
		return sanitizeMultipartRequestData(c)
	}

	if len(data) == 0 {
		return ""
	}

	if json.Valid(data) {
		var jsonData map[string]interface{}
		if err := json.Unmarshal(data, &jsonData); err == nil {
			if _, exists := jsonData["password"]; exists {
				jsonData["password"] = "***"
			}
			if _, exists := jsonData["Password"]; exists {
				jsonData["Password"] = "***"
			}
			if _, exists := jsonData["newPassword"]; exists {
				jsonData["newPassword"] = "***"
			}

			if sanitized, err := json.Marshal(jsonData); err == nil {
				return string(sanitized)
			}
		}
	}

	return truncateAndNormalizeText(data, 10000)
}

func sanitizeResponseData(data []byte) string {
	if len(data) == 0 {
		return ""
	}

	if len(data) > 5000 {
		return truncateAndNormalizeText(data, 5000)
	}
	return truncateAndNormalizeText(data, 0)
}

func sanitizeMultipartRequestData(c *gin.Context) string {
	summary := struct {
		ContentType string                   `json:"contentType"`
		Fields      map[string][]string      `json:"fields,omitempty"`
		Files       []map[string]interface{} `json:"files,omitempty"`
	}{
		ContentType: "multipart/form-data",
	}

	if c == nil || c.Request == nil || c.Request.MultipartForm == nil {
		if raw, err := json.Marshal(summary); err == nil {
			return string(raw)
		}
		return "{\"contentType\":\"multipart/form-data\"}"
	}

	form := c.Request.MultipartForm
	if len(form.Value) > 0 {
		summary.Fields = make(map[string][]string, len(form.Value))
		for key, values := range form.Value {
			items := make([]string, 0, len(values))
			for _, value := range values {
				items = append(items, truncateAndNormalizeText([]byte(value), 500))
			}
			summary.Fields[key] = items
		}
	}

	if len(form.File) > 0 {
		files := make([]map[string]interface{}, 0)
		for field, headers := range form.File {
			for _, header := range headers {
				files = append(files, map[string]interface{}{
					"field":    field,
					"filename": header.Filename,
					"size":     header.Size,
				})
			}
		}
		summary.Files = files
	}

	if raw, err := json.Marshal(summary); err == nil {
		return string(raw)
	}
	return "{\"contentType\":\"multipart/form-data\"}"
}

func truncateAndNormalizeText(data []byte, limit int) string {
	if len(data) == 0 {
		return ""
	}

	truncated := false
	if limit > 0 && len(data) > limit {
		data = data[:limit]
		truncated = true
	}

	text := strings.ToValidUTF8(string(data), "?")
	if truncated {
		text += "...(truncated)"
	}
	return text
}
