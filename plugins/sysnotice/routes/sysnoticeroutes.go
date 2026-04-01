package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/sysnotice/controllers"
	"github.com/gin-gonic/gin"
)

func init() {
	// RegisterRoutes 注册sys_notice插件路由
	var sysNoticeControllers = controllers.NewSysNoticeController()
	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		// sys_notice插件路由组
		sysNotice := engine.Group("/api/plugins/sysnotice/sysnotice")
		sysNotice.Use(middleware.JWTAuthMiddleware())     // 认证中间件
		sysNotice.Use(middleware.DemoAccountMiddleware()) // 添加演示账号中间件
		sysNotice.Use(middleware.CasbinMiddleware())      // 权限中间件
		{
			// 管理端通知列表
			sysNotice.GET("/list", sysNoticeControllers.List)
			// 管理端通知详情
			sysNotice.GET("/:id", sysNoticeControllers.GetByID)
			// 新增通知
			sysNotice.POST("/add", sysNoticeControllers.Add)
			// 编辑通知
			sysNotice.PUT("/edit", sysNoticeControllers.Update)
			// 发布通知
			sysNotice.POST("/publish", sysNoticeControllers.Publish)
			// 撤回通知
			sysNotice.POST("/revoke", sysNoticeControllers.Revoke)
			// 删除通知
			sysNotice.DELETE("/delete", sysNoticeControllers.Delete)
			// 当前用户收件箱列表
			sysNotice.GET("/inbox/list", sysNoticeControllers.InboxList)
			// 当前用户收件箱详情
			sysNotice.GET("/inbox/:id", sysNoticeControllers.InboxDetail)
			// 当前用户标记已读
			sysNotice.POST("/inbox/read", sysNoticeControllers.MarkRead)
			// 当前用户全部已读
			sysNotice.POST("/inbox/readAll", sysNoticeControllers.MarkAllRead)
			// 当前用户未读数量
			sysNotice.GET("/inbox/unreadCount", sysNoticeControllers.UnreadCount)
		}

		app.ZapLog.Info("sys_notice插件路由注册成功")
	})
}
