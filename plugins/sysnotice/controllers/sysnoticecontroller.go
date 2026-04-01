package controllers

import (
	"gin-fast/app/controllers"
	"gin-fast/plugins/sysnotice/models"
	"strconv"

	"gin-fast/plugins/sysnotice/service"

	"github.com/gin-gonic/gin"
)

// SysNoticeController 系统通知控制器
type SysNoticeController struct {
	controllers.Common
	SysNoticeService *service.SysNoticeService
}

func NewSysNoticeController() *SysNoticeController {
	return &SysNoticeController{
		Common:           controllers.Common{},
		SysNoticeService: service.NewSysNoticeService(),
	}
}

func (sc *SysNoticeController) List(c *gin.Context) {
	var req models.SysNoticeListRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	list, total, err := sc.SysNoticeService.List(c, &req)
	if err != nil {
		sc.FailAndAbort(c, "获取通知列表失败", err)
	}

	sc.Success(c, gin.H{
		"list":  list,
		"total": total,
	})
}

func (sc *SysNoticeController) GetByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		sc.FailAndAbort(c, "通知ID格式错误", err)
	}

	notice, err := sc.SysNoticeService.GetByID(c, uint(id))
	if err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.Success(c, notice)
}

func (sc *SysNoticeController) Add(c *gin.Context) {
	var req models.SysNoticeAddRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	notice, err := sc.SysNoticeService.Create(c, &req)
	if err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.SuccessWithMessage(c, "通知保存成功", notice)
}

func (sc *SysNoticeController) Update(c *gin.Context) {
	var req models.SysNoticeUpdateRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	notice, err := sc.SysNoticeService.Update(c, &req)
	if err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.SuccessWithMessage(c, "通知更新成功", notice)
}

func (sc *SysNoticeController) Publish(c *gin.Context) {
	var req models.SysNoticeActionRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	notice, err := sc.SysNoticeService.Publish(c, req.ID)
	if err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.SuccessWithMessage(c, "通知发布成功", notice)
}

func (sc *SysNoticeController) Revoke(c *gin.Context) {
	var req models.SysNoticeActionRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	notice, err := sc.SysNoticeService.Revoke(c, req.ID)
	if err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.SuccessWithMessage(c, "通知撤回成功", notice)
}

func (sc *SysNoticeController) Delete(c *gin.Context) {
	var req models.SysNoticeActionRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	if err := sc.SysNoticeService.Delete(c, req.ID); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.SuccessWithMessage(c, "通知删除成功", nil)
}

func (sc *SysNoticeController) InboxList(c *gin.Context) {
	var req models.SysNoticeInboxListRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	list, total, unreadCount, err := sc.SysNoticeService.InboxList(c, &req)
	if err != nil {
		sc.FailAndAbort(c, "获取收件箱失败", err)
	}

	sc.Success(c, gin.H{
		"list":        list,
		"total":       total,
		"unreadCount": unreadCount,
	})
}

func (sc *SysNoticeController) InboxDetail(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		sc.FailAndAbort(c, "通知ID格式错误", err)
	}

	recipient, err := sc.SysNoticeService.GetInboxDetail(c, uint(id))
	if err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.Success(c, recipient)
}

func (sc *SysNoticeController) MarkRead(c *gin.Context) {
	var req models.SysNoticeActionRequest
	if err := req.Validate(c); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	if err := sc.SysNoticeService.MarkRead(c, req.ID); err != nil {
		sc.FailAndAbort(c, err.Error(), err)
	}

	sc.SuccessWithMessage(c, "通知已标记为已读", nil)
}

func (sc *SysNoticeController) MarkAllRead(c *gin.Context) {
	rows, err := sc.SysNoticeService.MarkAllRead(c)
	if err != nil {
		sc.FailAndAbort(c, "批量已读失败", err)
	}

	sc.Success(c, gin.H{
		"updated": rows,
	}, "操作成功")
}

func (sc *SysNoticeController) UnreadCount(c *gin.Context) {
	count, err := sc.SysNoticeService.UnreadCount(c)
	if err != nil {
		sc.FailAndAbort(c, "获取未读数量失败", err)
	}

	sc.Success(c, gin.H{
		"count": count,
	})
}
