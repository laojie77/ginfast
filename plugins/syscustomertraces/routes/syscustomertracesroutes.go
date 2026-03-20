package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/plugins/syscustomertraces/controllers"
	"gin-fast/app/utils/ginhelper"
	"github.com/gin-gonic/gin"
)

func init() {
    // RegisterRoutes 注册sys_customer_traces插件路由
    var sysCustomerTracesControllers = controllers.NewSysCustomerTracesController()
	ginhelper.RegisterPluginRoutes(func (engine *gin.Engine) {
        // sys_customer_traces插件路由组
        sysCustomerTraces := engine.Group("/api/plugins/syscustomertraces/syscustomertraces")
        sysCustomerTraces.Use(middleware.JWTAuthMiddleware())     // 认证中间件
        sysCustomerTraces.Use(middleware.DemoAccountMiddleware()) // 添加演示账号中间件
        sysCustomerTraces.Use(middleware.CasbinMiddleware())      // 权限中间件
        {
            // 创建sys_customer_traces
            sysCustomerTraces.POST("/add", sysCustomerTracesControllers.Create)
            // 更新sys_customer_traces
            sysCustomerTraces.PUT("/edit", sysCustomerTracesControllers.Update)
            // 删除sys_customer_traces
            sysCustomerTraces.DELETE("/delete", sysCustomerTracesControllers.Delete)
            // sys_customer_traces列表（分页查询）
            sysCustomerTraces.GET("/list", sysCustomerTracesControllers.List)
            // 根据ID获取sys_customer_traces信息
            sysCustomerTraces.GET("/:id", sysCustomerTracesControllers.GetByID)
        }

        app.ZapLog.Info("sys_customer_traces插件路由注册成功")
    })
}