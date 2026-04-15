package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/app/utils/ginhelper"
	"gin-fast/plugins/syscustomer/controllers"

	"github.com/gin-gonic/gin"
)

func init() {
	sysCustomerControllers := controllers.NewSysCustomerController()

	ginhelper.RegisterPluginRoutes(func(engine *gin.Engine) {
		sysCustomer := engine.Group("/api/plugins/syscustomer/syscustomer")
		sysCustomer.Use(middleware.JWTAuthMiddleware())
		sysCustomer.Use(middleware.DemoAccountMiddleware())
		sysCustomer.Use(middleware.CasbinMiddleware())
		{
			sysCustomer.POST("/add", sysCustomerControllers.Create)
			sysCustomer.POST("/import", sysCustomerControllers.Import)
			sysCustomer.PUT("/edit", sysCustomerControllers.Update)
			sysCustomer.PUT("/updateCustomerStatusTrace", sysCustomerControllers.UpdateCustomerStatusTrace)
			sysCustomer.DELETE("/delete", sysCustomerControllers.Delete)
			sysCustomer.GET("/list", sysCustomerControllers.List)
			sysCustomer.GET("/exportTask/:taskId", sysCustomerControllers.GetExportTask)
			sysCustomer.GET("/importBatch/latest", sysCustomerControllers.GetLatestImportBatch)
			sysCustomer.GET("/importBatch/:batchId", sysCustomerControllers.GetImportBatch)
			sysCustomer.GET("/:id", sysCustomerControllers.GetByID)
		}

		app.ZapLog.Info("sys_customer插件路由注册成功")
	})
}
