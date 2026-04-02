package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/sysnotice/controllers"

	"github.com/gin-gonic/gin"
)

func init() {
	var sysNoticeControllers = controllers.NewSysNoticeController()
	var sysNoticeWebSocketControllers = controllers.NewSysNoticeWebSocketController()

	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		sysNoticeWS := engine.Group("/api/plugins/sysnotice")
		sysNoticeWS.Use(middleware.JWTAuthMiddleware())
		{
			sysNoticeWS.GET("/ws", sysNoticeWebSocketControllers.Connect)
			sysNoticeWS.POST("/mock-event", sysNoticeWebSocketControllers.MockEvent)
		}

		sysNotice := engine.Group("/api/plugins/sysnotice/sysnotice")
		sysNotice.Use(middleware.JWTAuthMiddleware())
		sysNotice.Use(middleware.DemoAccountMiddleware())
		sysNotice.Use(middleware.CasbinMiddleware())
		{
			sysNotice.GET("/list", sysNoticeControllers.List)
			sysNotice.GET("/:id", sysNoticeControllers.GetByID)
			sysNotice.POST("/add", sysNoticeControllers.Add)
			sysNotice.PUT("/edit", sysNoticeControllers.Update)
			sysNotice.POST("/publish", sysNoticeControllers.Publish)
			sysNotice.POST("/revoke", sysNoticeControllers.Revoke)
			sysNotice.DELETE("/delete", sysNoticeControllers.Delete)
			sysNotice.GET("/inbox/list", sysNoticeControllers.InboxList)
			sysNotice.GET("/inbox/:id", sysNoticeControllers.InboxDetail)
			sysNotice.POST("/inbox/read", sysNoticeControllers.MarkRead)
			sysNotice.POST("/inbox/readAll", sysNoticeControllers.MarkAllRead)
			sysNotice.GET("/inbox/unreadCount", sysNoticeControllers.UnreadCount)
		}

		app.ZapLog.Info("sys_notice插件路由注册成功")
	})
}
