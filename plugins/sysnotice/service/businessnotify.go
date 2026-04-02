package service

import (
	"context"
	"fmt"
	"strings"
	"time"
)

const (
	NoticeBusinessSceneSystemInternal = "system_internal"
	NoticeBusinessSceneExportDownload = "export_download"
	NoticeBusinessSceneCustomerAssign = "customer_assigned"

	NoticeBusinessActionKindRoute    = "route"
	NoticeBusinessActionKindURL      = "url"
	NoticeBusinessActionKindDownload = "download"

	NoticeBusinessActionOpenModeSelf  = "self"
	NoticeBusinessActionOpenModeBlank = "blank"
)

type NoticeBusinessAction struct {
	Label    string `json:"label,omitempty"`
	Kind     string `json:"kind,omitempty"`
	Value    string `json:"value,omitempty"`
	OpenMode string `json:"openMode,omitempty"`
}

type NoticeBusinessPayload struct {
	ID        string                 `json:"id"`
	Scene     string                 `json:"scene"`
	Title     string                 `json:"title"`
	Content   string                 `json:"content"`
	Level     string                 `json:"level,omitempty"`
	NoticeID  uint                   `json:"noticeId,omitempty"`
	Action    *NoticeBusinessAction  `json:"action,omitempty"`
	Extra     map[string]interface{} `json:"extra,omitempty"`
	Timestamp int64                  `json:"timestamp"`
}

type NoticeBusinessNotifyRequest struct {
	TenantID uint
	UserIDs  []uint
	Payload  NoticeBusinessPayload
}

type NoticeBusinessService struct {
	hub *NoticeRealtimeHub
}

func NewNoticeBusinessService() *NoticeBusinessService {
	return &NoticeBusinessService{
		hub: GetNoticeRealtimeHub(),
	}
}

func (s *NoticeBusinessService) NotifyUsers(ctx context.Context, req NoticeBusinessNotifyRequest) error {
	if s == nil || s.hub == nil {
		return nil
	}

	userIDs := uniqueUintSlice(req.UserIDs)
	if req.TenantID == 0 || len(userIDs) == 0 {
		return nil
	}

	payload := req.Payload
	if payload.Timestamp <= 0 {
		payload.Timestamp = time.Now().Unix()
	}
	if strings.TrimSpace(payload.ID) == "" {
		payload.ID = fmt.Sprintf("%s-%d", strings.TrimSpace(payload.Scene), time.Now().UnixNano())
	}

	s.hub.PushBusinessNotice(ctx, req.TenantID, userIDs, payload)
	return nil
}

func (s *NoticeBusinessService) NotifyCurrentUserSystemMessage(ctx context.Context, tenantID, userID uint, title, content string) error {
	return s.NotifyUsers(ctx, NoticeBusinessNotifyRequest{
		TenantID: tenantID,
		UserIDs:  []uint{userID},
		Payload: NoticeBusinessPayload{
			Scene:   NoticeBusinessSceneSystemInternal,
			Title:   title,
			Content: content,
			Level:   "info",
		},
	})
}

func (s *NoticeBusinessService) NotifyExportReady(
	ctx context.Context,
	tenantID, userID uint,
	title, content, downloadURL string,
	extra map[string]interface{},
) error {
	return s.NotifyUsers(ctx, NoticeBusinessNotifyRequest{
		TenantID: tenantID,
		UserIDs:  []uint{userID},
		Payload: NoticeBusinessPayload{
			Scene:   NoticeBusinessSceneExportDownload,
			Title:   title,
			Content: content,
			Level:   "success",
			Action: &NoticeBusinessAction{
				Label:    "下载文件",
				Kind:     NoticeBusinessActionKindDownload,
				Value:    strings.TrimSpace(downloadURL),
				OpenMode: NoticeBusinessActionOpenModeBlank,
			},
			Extra: extra,
		},
	})
}

func (s *NoticeBusinessService) NotifyCustomerAssigned(
	ctx context.Context,
	tenantID, userID uint,
	title, content, route string,
	extra map[string]interface{},
) error {
	return s.NotifyUsers(ctx, NoticeBusinessNotifyRequest{
		TenantID: tenantID,
		UserIDs:  []uint{userID},
		Payload: NoticeBusinessPayload{
			Scene:   NoticeBusinessSceneCustomerAssign,
			Title:   title,
			Content: content,
			Level:   "info",
			Action: &NoticeBusinessAction{
				Label:    "查看客户",
				Kind:     NoticeBusinessActionKindRoute,
				Value:    strings.TrimSpace(route),
				OpenMode: NoticeBusinessActionOpenModeSelf,
			},
			Extra: extra,
		},
	})
}
