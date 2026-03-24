package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/syscustomer/controllers"
	"github.com/gin-gonic/gin"
)

func init() {
	// RegisterRoutes 注册sys_customer插件路由
	var sysCustomerControllers = controllers.NewSysCustomerController()
	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		// sys_customer插件路由组
		sysCustomer := engine.Group("/api/plugins/syscustomer/syscustomer")
		sysCustomer.Use(middleware.JWTAuthMiddleware())     // 认证中间件
		sysCustomer.Use(middleware.DemoAccountMiddleware()) // 添加演示账号中间件
		sysCustomer.Use(middleware.CasbinMiddleware())      // 权限中间件
		{
			// 创建sys_customer
			sysCustomer.POST("/add", sysCustomerControllers.Create)
			// 更新sys_customer
			sysCustomer.PUT("/edit", sysCustomerControllers.Update)
			sysCustomer.PUT("/updateCustomerStatusTrace", sysCustomerControllers.UpdateCustomerStatusTrace)
			// 删除sys_customer
			sysCustomer.DELETE("/delete", sysCustomerControllers.Delete)
			// sys_customer列表（分页查询）
			sysCustomer.GET("/list", sysCustomerControllers.List)
			// 根据ID获取sys_customer信息
			sysCustomer.GET("/:id", sysCustomerControllers.GetByID)
		}

		app.ZapLog.Info("sys_customer插件路由注册成功")
	})
}
