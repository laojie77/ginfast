package service

import (
	"context"
	"encoding/json"
	"html"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"gin-fast/app/global/app"
	models2 "gin-fast/plugins/sysnotice/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

const (
	noticeWSWriteWait      = 10 * time.Second
	noticeWSPongWait       = 60 * time.Second
	noticeWSPingPeriod     = (noticeWSPongWait * 9) / 10
	noticeWSMaxMessageSize = 4096
	// WebSocket 同步给前端弹窗的“最近通知预览”最大条数。
	// 这个值决定右上角通知弹窗最多能拿到多少条站内通知预览数据。
	noticeRecentLimit = 15
)

var (
	noticeRealtimeHubOnce sync.Once
	noticeRealtimeHubInst *NoticeRealtimeHub
	noticeHTMLTagRegexp   = regexp.MustCompile(`<[^>]+>`)
)

type noticeAudienceKey struct {
	UserID   uint
	TenantID uint
}

type noticeWSClient struct {
	hub      *NoticeRealtimeHub
	conn     *websocket.Conn
	send     chan []byte
	userID   uint
	tenantID uint
	username string
}

type NoticeRealtimeHub struct {
	upgrader websocket.Upgrader
	mu       sync.RWMutex
	clients  map[noticeAudienceKey]map[*noticeWSClient]struct{}
}

type noticeRealtimeMessage struct {
	Type      string      `json:"type"`
	Timestamp int64       `json:"timestamp"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
}

type noticeClientCommand struct {
	Type string `json:"type"`
}

type NoticeRealtimePreview struct {
	RecipientID uint       `json:"recipientId"`
	NoticeID    uint       `json:"noticeId"`
	Title       string     `json:"title"`
	Content     string     `json:"content"`
	Type        int8       `json:"type"`
	Level       string     `json:"level"`
	PublishTime *time.Time `json:"publishTime,omitempty"`
	ReadStatus  int8       `json:"readStatus"`
}

type NoticeRealtimeSyncPayload struct {
	UserID      uint                    `json:"userId"`
	TenantID    uint                    `json:"tenantId"`
	UnreadCount int64                   `json:"unreadCount"`
	Recent      []NoticeRealtimePreview `json:"recent"`
}

type NoticeRealtimePublishedPayload struct {
	Notice      NoticeRealtimePreview `json:"notice"`
	UnreadCount int64                 `json:"unreadCount"`
}

type NoticeRealtimeUnreadPayload struct {
	UnreadCount int64  `json:"unreadCount"`
	Reason      string `json:"reason,omitempty"`
	NoticeID    uint   `json:"noticeId,omitempty"`
}

type NoticeRealtimeRevokedPayload struct {
	NoticeID    uint  `json:"noticeId"`
	UnreadCount int64 `json:"unreadCount"`
}

type websocketNoticeDispatcher struct {
	hub *NoticeRealtimeHub
}

type noticePreviewRow struct {
	RecipientID uint       `gorm:"column:recipient_id"`
	NoticeID    uint       `gorm:"column:notice_id"`
	Title       string     `gorm:"column:title"`
	Content     string     `gorm:"column:content"`
	Type        int8       `gorm:"column:type"`
	Level       string     `gorm:"column:level"`
	PublishTime *time.Time `gorm:"column:publish_time"`
	ReadStatus  int8       `gorm:"column:read_status"`
}

type noticeUnreadCountRow struct {
	UserID uint  `gorm:"column:user_id"`
	Count  int64 `gorm:"column:count"`
}

func newDefaultNoticeDispatcher() NoticeDispatcher {
	return websocketNoticeDispatcher{hub: GetNoticeRealtimeHub()}
}

func (d websocketNoticeDispatcher) Dispatch(ctx context.Context, notice *models2.SysNotice, recipients models2.SysNoticeRecipientList) error {
	if d.hub == nil {
		return nil
	}
	return d.hub.BroadcastPublishedNotice(ctx, notice, recipients)
}

func GetNoticeRealtimeHub() *NoticeRealtimeHub {
	noticeRealtimeHubOnce.Do(func() {
		noticeRealtimeHubInst = &NoticeRealtimeHub{
			upgrader: websocket.Upgrader{
				ReadBufferSize:  1024,
				WriteBufferSize: 1024,
				CheckOrigin: func(r *http.Request) bool {
					return true
				},
			},
			clients: make(map[noticeAudienceKey]map[*noticeWSClient]struct{}),
		}
	})
	return noticeRealtimeHubInst
}

func (h *NoticeRealtimeHub) Serve(c *gin.Context, claims *app.Claims) error {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return err
	}

	client := &noticeWSClient{
		hub:      h,
		conn:     conn,
		send:     make(chan []byte, 16),
		userID:   claims.UserID,
		tenantID: claims.TenantID,
		username: claims.Username,
	}

	h.register(client)

	go client.writePump()
	go client.readPump()

	client.sendJSON("connected", map[string]interface{}{
		"userId":    client.userID,
		"tenantId":  client.tenantID,
		"username":  client.username,
		"connected": true,
	}, "")

	if err := h.sendSyncSnapshot(c.Request.Context(), client); err != nil {
		app.ZapLog.Warn("初始化通知同步快照失败",
			zap.Error(err),
			zap.Uint("user_id", client.userID),
			zap.Uint("tenant_id", client.tenantID),
		)
	}

	return nil
}

func (h *NoticeRealtimeHub) register(client *noticeWSClient) {
	key := noticeAudienceKey{UserID: client.userID, TenantID: client.tenantID}

	h.mu.Lock()
	defer h.mu.Unlock()

	if _, exists := h.clients[key]; !exists {
		h.clients[key] = make(map[*noticeWSClient]struct{})
	}
	h.clients[key][client] = struct{}{}
}

func (h *NoticeRealtimeHub) unregister(client *noticeWSClient) {
	key := noticeAudienceKey{UserID: client.userID, TenantID: client.tenantID}

	h.mu.Lock()
	defer h.mu.Unlock()

	clientSet, exists := h.clients[key]
	if !exists {
		return
	}
	if _, exists = clientSet[client]; !exists {
		return
	}

	delete(clientSet, client)
	close(client.send)

	if len(clientSet) == 0 {
		delete(h.clients, key)
	}
}

func (h *NoticeRealtimeHub) sendToAudience(userID, tenantID uint, message noticeRealtimeMessage) {
	key := noticeAudienceKey{UserID: userID, TenantID: tenantID}
	payload, err := json.Marshal(message)
	if err != nil {
		app.ZapLog.Warn("序列化 WebSocket 消息失败",
			zap.Error(err),
			zap.String("type", message.Type),
			zap.Uint("user_id", userID),
			zap.Uint("tenant_id", tenantID),
		)
		return
	}

	h.mu.RLock()
	clientSet := h.clients[key]
	clients := make([]*noticeWSClient, 0, len(clientSet))
	for client := range clientSet {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- payload:
		default:
			go func(target *noticeWSClient) {
				h.unregister(target)
				_ = target.conn.Close()
			}(client)
		}
	}
}

func (h *NoticeRealtimeHub) BroadcastPublishedNotice(ctx context.Context, notice *models2.SysNotice, recipients models2.SysNoticeRecipientList) error {
	if notice == nil || len(recipients) == 0 {
		return nil
	}

	recipientMap := make(map[uint]*models2.SysNoticeRecipient)
	userIDs := make([]uint, 0, len(recipients))
	for _, recipient := range recipients {
		if recipient == nil || recipient.UserID == 0 {
			continue
		}
		if _, exists := recipientMap[recipient.UserID]; exists {
			continue
		}
		recipientMap[recipient.UserID] = recipient
		userIDs = append(userIDs, recipient.UserID)
	}

	if len(userIDs) == 0 {
		return nil
	}

	countMap, err := h.loadUnreadCountMap(ctx, notice.TenantID, userIDs)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		recipient := recipientMap[userID]
		payload := NoticeRealtimePublishedPayload{
			Notice:      buildRealtimePreviewFromNotice(notice, recipient),
			UnreadCount: countMap[userID],
		}
		h.sendToAudience(userID, notice.TenantID, noticeRealtimeMessage{
			Type:      "notice_published",
			Timestamp: time.Now().Unix(),
			Data:      payload,
		})
	}

	return nil
}

func (h *NoticeRealtimeHub) PushNoticeRevoked(ctx context.Context, tenantID, noticeID uint) error {
	userIDs, err := h.loadRecipientUserIDs(ctx, tenantID, noticeID)
	if err != nil {
		return err
	}
	if len(userIDs) == 0 {
		return nil
	}

	countMap, err := h.loadUnreadCountMap(ctx, tenantID, userIDs)
	if err != nil {
		return err
	}

	for _, userID := range userIDs {
		h.sendToAudience(userID, tenantID, noticeRealtimeMessage{
			Type:      "notice_revoked",
			Timestamp: time.Now().Unix(),
			Data: NoticeRealtimeRevokedPayload{
				NoticeID:    noticeID,
				UnreadCount: countMap[userID],
			},
		})
	}

	return nil
}

func (h *NoticeRealtimeHub) PushNoticeRead(ctx context.Context, tenantID, userID, noticeID uint) error {
	count, err := h.loadUnreadCount(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	h.sendToAudience(userID, tenantID, noticeRealtimeMessage{
		Type:      "notice_read",
		Timestamp: time.Now().Unix(),
		Data: NoticeRealtimeUnreadPayload{
			UnreadCount: count,
			Reason:      "read",
			NoticeID:    noticeID,
		},
	})
	return nil
}

func (h *NoticeRealtimeHub) PushAllRead(ctx context.Context, tenantID, userID uint) error {
	count, err := h.loadUnreadCount(ctx, tenantID, userID)
	if err != nil {
		return err
	}

	h.sendToAudience(userID, tenantID, noticeRealtimeMessage{
		Type:      "notice_all_read",
		Timestamp: time.Now().Unix(),
		Data: NoticeRealtimeUnreadPayload{
			UnreadCount: count,
			Reason:      "read_all",
		},
	})
	return nil
}

func (h *NoticeRealtimeHub) PushBusinessNotice(ctx context.Context, tenantID uint, userIDs []uint, payload NoticeBusinessPayload) {
	uniqueUserIDs := uniqueUintSlice(userIDs)
	if tenantID == 0 || len(uniqueUserIDs) == 0 {
		return
	}

	message := noticeRealtimeMessage{
		Type:      "business_notice",
		Timestamp: time.Now().Unix(),
		Data:      payload,
	}

	for _, userID := range uniqueUserIDs {
		h.sendToAudience(userID, tenantID, message)
	}
}

func (h *NoticeRealtimeHub) sendSyncSnapshot(ctx context.Context, client *noticeWSClient) error {
	unreadCount, err := h.loadUnreadCount(ctx, client.tenantID, client.userID)
	if err != nil {
		return err
	}

	// recent 是给右上角弹窗用的预览数据，不是完整收件箱列表。
	recent, err := h.loadRecentNotices(ctx, client.tenantID, client.userID, noticeRecentLimit)
	if err != nil {
		return err
	}

	client.sendJSON("notice_sync", NoticeRealtimeSyncPayload{
		UserID:      client.userID,
		TenantID:    client.tenantID,
		UnreadCount: unreadCount,
		Recent:      recent,
	}, "")

	return nil
}

func (h *NoticeRealtimeHub) loadRecentNotices(ctx context.Context, tenantID, userID uint, limit int) ([]NoticeRealtimePreview, error) {
	queryCtx := normalizeContext(ctx)
	rows := make([]noticePreviewRow, 0, limit)

	// 右上角弹窗优先展示未读，再按发布时间倒序，避免已读记录把未读挤出预览区。
	err := app.DB().WithContext(queryCtx).
		Table("sys_notice_recipient").
		Select(`
			sys_notice_recipient.id AS recipient_id,
			sys_notice_recipient.notice_id,
			sys_notice.title,
			sys_notice.content,
			sys_notice.type,
			sys_notice.level,
			sys_notice.publish_time,
			sys_notice_recipient.read_status
		`).
		Joins("JOIN sys_notice ON sys_notice.id = sys_notice_recipient.notice_id").
		Where(`
			sys_notice_recipient.user_id = ?
			AND sys_notice_recipient.tenant_id = ?
			AND sys_notice.publish_status = ?
		`, userID, tenantID, models2.SysNoticePublishStatusPublished).
		Order("sys_notice_recipient.read_status ASC, sys_notice_recipient.publish_time DESC, sys_notice_recipient.id DESC").
		Limit(limit).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	result := make([]NoticeRealtimePreview, 0, len(rows))
	for _, row := range rows {
		result = append(result, NoticeRealtimePreview{
			RecipientID: row.RecipientID,
			NoticeID:    row.NoticeID,
			Title:       row.Title,
			Content:     trimNoticeContent(row.Content),
			Type:        row.Type,
			Level:       row.Level,
			PublishTime: row.PublishTime,
			ReadStatus:  row.ReadStatus,
		})
	}

	return result, nil
}

func (h *NoticeRealtimeHub) loadUnreadCount(ctx context.Context, tenantID, userID uint) (int64, error) {
	countMap, err := h.loadUnreadCountMap(ctx, tenantID, []uint{userID})
	if err != nil {
		return 0, err
	}
	return countMap[userID], nil
}

func (h *NoticeRealtimeHub) loadUnreadCountMap(ctx context.Context, tenantID uint, userIDs []uint) (map[uint]int64, error) {
	result := make(map[uint]int64)
	uniqueIDs := uniqueUintSlice(userIDs)
	if len(uniqueIDs) == 0 {
		return result, nil
	}

	queryCtx := normalizeContext(ctx)
	rows := make([]noticeUnreadCountRow, 0, len(uniqueIDs))

	err := app.DB().WithContext(queryCtx).
		Model(&models2.SysNoticeRecipient{}).
		Select("sys_notice_recipient.user_id, COUNT(*) AS count").
		Joins("JOIN sys_notice ON sys_notice.id = sys_notice_recipient.notice_id").
		Where(`
			sys_notice_recipient.tenant_id = ?
			AND sys_notice_recipient.user_id IN ?
			AND sys_notice_recipient.read_status = ?
			AND sys_notice.publish_status = ?
		`, tenantID, uniqueIDs, models2.SysNoticeReadStatusUnread, models2.SysNoticePublishStatusPublished).
		Group("sys_notice_recipient.user_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, userID := range uniqueIDs {
		result[userID] = 0
	}
	for _, row := range rows {
		result[row.UserID] = row.Count
	}

	return result, nil
}

func (h *NoticeRealtimeHub) loadRecipientUserIDs(ctx context.Context, tenantID, noticeID uint) ([]uint, error) {
	queryCtx := normalizeContext(ctx)
	var userIDs []uint

	err := app.DB().WithContext(queryCtx).
		Model(&models2.SysNoticeRecipient{}).
		Where("tenant_id = ? AND notice_id = ?", tenantID, noticeID).
		Distinct().
		Pluck("user_id", &userIDs).Error
	if err != nil {
		return nil, err
	}

	return uniqueUintSlice(userIDs), nil
}

func (c *noticeWSClient) readPump() {
	defer func() {
		c.hub.unregister(c)
		_ = c.conn.Close()
	}()

	c.conn.SetReadLimit(noticeWSMaxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(noticeWSPongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(noticeWSPongWait))
	})

	for {
		messageType, payload, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				app.ZapLog.Warn("WebSocket 连接异常关闭",
					zap.Error(err),
					zap.Uint("user_id", c.userID),
					zap.Uint("tenant_id", c.tenantID),
				)
			}
			return
		}

		if messageType != websocket.TextMessage {
			continue
		}

		c.handleCommand(payload)
	}
}

func (c *noticeWSClient) writePump() {
	ticker := time.NewTicker(noticeWSPingPeriod)
	defer func() {
		ticker.Stop()
		_ = c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			_ = c.conn.SetWriteDeadline(time.Now().Add(noticeWSWriteWait))
			if !ok {
				_ = c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			writer, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			if _, err = writer.Write(message); err != nil {
				_ = writer.Close()
				return
			}

			if err = writer.Close(); err != nil {
				return
			}

		case <-ticker.C:
			_ = c.conn.SetWriteDeadline(time.Now().Add(noticeWSWriteWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *noticeWSClient) handleCommand(payload []byte) {
	command := noticeClientCommand{}
	if err := json.Unmarshal(payload, &command); err != nil {
		c.sendJSON("error", nil, "WebSocket 消息必须是 JSON 格式")
		return
	}

	switch strings.ToLower(strings.TrimSpace(command.Type)) {
	case "ping":
		c.sendJSON("pong", map[string]interface{}{"ok": true}, "")
	case "sync", "get_unread_count":
		if err := c.hub.sendSyncSnapshot(context.Background(), c); err != nil {
			app.ZapLog.Warn("主动同步通知快照失败",
				zap.Error(err),
				zap.Uint("user_id", c.userID),
				zap.Uint("tenant_id", c.tenantID),
			)
			c.sendJSON("error", nil, "同步通知数据失败")
		}
	default:
		c.sendJSON("error", nil, "不支持的 WebSocket 消息类型")
	}
}

func (c *noticeWSClient) sendJSON(messageType string, data interface{}, message string) {
	payload, err := json.Marshal(noticeRealtimeMessage{
		Type:      messageType,
		Timestamp: time.Now().Unix(),
		Message:   message,
		Data:      data,
	})
	if err != nil {
		app.ZapLog.Warn("序列化客户端消息失败",
			zap.Error(err),
			zap.String("type", messageType),
			zap.Uint("user_id", c.userID),
			zap.Uint("tenant_id", c.tenantID),
		)
		return
	}

	select {
	case c.send <- payload:
	default:
		go func() {
			c.hub.unregister(c)
			_ = c.conn.Close()
		}()
	}
}

func buildRealtimePreviewFromNotice(notice *models2.SysNotice, recipient *models2.SysNoticeRecipient) NoticeRealtimePreview {
	preview := NoticeRealtimePreview{
		NoticeID:    notice.ID,
		Title:       notice.Title,
		Content:     trimNoticeContent(notice.Content),
		Type:        notice.Type,
		Level:       notice.Level,
		PublishTime: notice.PublishTime,
		ReadStatus:  models2.SysNoticeReadStatusUnread,
	}
	if recipient != nil {
		preview.RecipientID = recipient.ID
		preview.ReadStatus = recipient.ReadStatus
		if recipient.PublishTime != nil {
			preview.PublishTime = recipient.PublishTime
		}
	}
	return preview
}

func trimNoticeContent(content string) string {
	plain := html.UnescapeString(content)
	plain = noticeHTMLTagRegexp.ReplaceAllString(plain, " ")
	plain = strings.Join(strings.Fields(strings.TrimSpace(plain)), " ")
	if plain == "" {
		return ""
	}

	contentRunes := []rune(plain)
	if len(contentRunes) <= 80 {
		return plain
	}

	return string(contentRunes[:80]) + "..."
}

func normalizeContext(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}
