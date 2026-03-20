package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/plugins/syschannelcompany/controllers"
	"gin-fast/app/utils/ginhelper"
	"github.com/gin-gonic/gin"
)

func init() {
    // RegisterRoutes 注册sys_channel_company插件路由
    var sysChannelCompanyControllers = controllers.NewSysChannelCompanyController()
	ginhelper.RegisterPluginRoutes(func (engine *gin.Engine) {
        // sys_channel_company插件路由组
        sysChannelCompany := engine.Group("/api/plugins/syschannelcompany/syschannelcompany")
        sysChannelCompany.Use(middleware.JWTAuthMiddleware())     // 认证中间件
        sysChannelCompany.Use(middleware.DemoAccountMiddleware()) // 添加演示账号中间件
        sysChannelCompany.Use(middleware.CasbinMiddleware())      // 权限中间件
        {
            // 创建sys_channel_company
            sysChannelCompany.POST("/add", sysChannelCompanyControllers.Create)
            // 更新sys_channel_company
            sysChannelCompany.PUT("/edit", sysChannelCompanyControllers.Update)
            // 删除sys_channel_company
            sysChannelCompany.DELETE("/delete", sysChannelCompanyControllers.Delete)
            // sys_channel_company列表（分页查询）
            sysChannelCompany.GET("/list", sysChannelCompanyControllers.List)
            // 根据ID获取sys_channel_company信息
            sysChannelCompany.GET("/:id", sysChannelCompanyControllers.GetByID)
        }

        app.ZapLog.Info("sys_channel_company插件路由注册成功")
    })
}