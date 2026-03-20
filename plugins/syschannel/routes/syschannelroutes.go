package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/syschannel/controllers"

	"github.com/gin-gonic/gin"
)

func init() {
	// RegisterRoutes 注册sys_channel插件路由
	var sysChannelControllers = controllers.NewSysChannelController()
	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		// sys_channel插件路由组
		sysChannel := engine.Group("/api/plugins/syschannel/syschannel")
		sysChannel.Use(middleware.JWTAuthMiddleware())     // 认证中间件
		sysChannel.Use(middleware.DemoAccountMiddleware()) // 添加演示账号中间件
		sysChannel.Use(middleware.CasbinMiddleware())      // 权限中间件
		{
			// 创建sys_channel
			sysChannel.POST("/add", sysChannelControllers.Create)
			// 更新sys_channel
			sysChannel.PUT("/edit", sysChannelControllers.Update)
			// 删除sys_channel
			sysChannel.DELETE("/delete", sysChannelControllers.Delete)
			// sys_channel列表（分页查询）
			sysChannel.GET("/list", sysChannelControllers.List)
			// 根据ID获取sys_channel信息
			sysChannel.GET("/:id", sysChannelControllers.GetByID)
		}

		app.ZapLog.Info("sys_channel插件路由注册成功")
	})
}
