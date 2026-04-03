package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/syscustomerexporttasks/controllers"

	"github.com/gin-gonic/gin"
)

func init() {
	var sysCustomerExportTasksControllers = controllers.NewSysCustomerExportTasksController()
	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		sysCustomerExportTasks := engine.Group("/api/plugins/syscustomerexporttasks/syscustomerexporttasks")
		sysCustomerExportTasks.Use(middleware.JWTAuthMiddleware())
		sysCustomerExportTasks.Use(middleware.DemoAccountMiddleware())
		sysCustomerExportTasks.Use(middleware.CasbinMiddleware())
		{
			sysCustomerExportTasks.GET("/list", sysCustomerExportTasksControllers.List)
			sysCustomerExportTasks.GET("/:id", sysCustomerExportTasksControllers.GetByID)
		}

		app.ZapLog.Info("sys_customer_export_tasks插件路由注册成功")
	})
}
