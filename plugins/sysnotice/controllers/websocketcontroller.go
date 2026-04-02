package controllers

import (
	"net/http"
	"strings"

	baseControllers "gin-fast/app/controllers"
	"gin-fast/app/global/app"
	"gin-fast/plugins/sysnotice/models"
	"gin-fast/plugins/sysnotice/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type SysNoticeWebSocketController struct {
	baseControllers.Common
}

func NewSysNoticeWebSocketController() *SysNoticeWebSocketController {
	return &SysNoticeWebSocketController{
		Common: baseControllers.Common{},
	}
}

func (wc *SysNoticeWebSocketController) Connect(c *gin.Context) {
	claims := wc.GetClaims(c)
	if claims == nil {
		wc.Fail(c, "未登录或登录状态已失效", nil, http.StatusUnauthorized)
		return
	}

	if err := service.GetNoticeRealtimeHub().Serve(c, claims); err != nil {
		app.ZapLog.Warn("升级 WebSocket 连接失败",
			zap.Error(err),
			zap.Uint("user_id", claims.UserID),
			zap.Uint("tenant_id", claims.TenantID),
		)
		if !c.Writer.Written() {
			wc.Fail(c, "建立 WebSocket 连接失败", err, http.StatusBadRequest)
		}
	}
}

func (wc *SysNoticeWebSocketController) MockEvent(c *gin.Context) {
	claims := wc.GetClaims(c)
	if claims == nil {
		wc.Fail(c, "未登录或登录状态已失效", nil, http.StatusUnauthorized)
		return
	}

	var req models.SysNoticeRealtimeMockRequest
	if err := req.Validate(c); err != nil {
		wc.FailAndAbort(c, err.Error(), err)
		return
	}

	payload := service.NoticeBusinessPayload{
		Scene:   strings.TrimSpace(req.Scene),
		Title:   strings.TrimSpace(req.Title),
		Content: strings.TrimSpace(req.Content),
		Level:   strings.TrimSpace(req.Level),
	}

	if actionValue := strings.TrimSpace(req.ActionValue); actionValue != "" {
		payload.Action = &service.NoticeBusinessAction{
			Label:    strings.TrimSpace(req.ActionLabel),
			Kind:     strings.TrimSpace(req.ActionKind),
			Value:    actionValue,
			OpenMode: strings.TrimSpace(req.OpenMode),
		}
	}

	if err := service.NewNoticeBusinessService().NotifyUsers(c, service.NoticeBusinessNotifyRequest{
		TenantID: claims.TenantID,
		UserIDs:  []uint{claims.UserID},
		Payload:  payload,
	}); err != nil {
		wc.FailAndAbort(c, "发送实时通知失败", err)
		return
	}

	wc.SuccessWithMessage(c, "实时通知已发送", nil)
}
