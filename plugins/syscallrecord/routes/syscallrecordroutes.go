package routes

import (
	"gin-fast/app/global/app"
	"gin-fast/app/middleware"
	"gin-fast/plugins/syscallrecord/controllers"
	"gin-fast/app/utils/ginhelper"
	"github.com/gin-gonic/gin"
)

func init() {
    // RegisterRoutes 注册sys_call_record插件路由
    var sysCallRecordControllers = controllers.NewSysCallRecordController()
	ginhelper.RegisterPluginRoutes(func (engine *gin.Engine) {
        // sys_call_record插件路由组
        sysCallRecord := engine.Group("/api/plugins/syscallrecord/syscallrecord")
        sysCallRecord.Use(middleware.JWTAuthMiddleware())     // 认证中间件
        sysCallRecord.Use(middleware.DemoAccountMiddleware()) // 添加演示账号中间件
        sysCallRecord.Use(middleware.CasbinMiddleware())      // 权限中间件
        {
            // 创建sys_call_record
            sysCallRecord.POST("/add", sysCallRecordControllers.Create)
            // 更新sys_call_record
            sysCallRecord.PUT("/edit", sysCallRecordControllers.Update)
            // 删除sys_call_record
            sysCallRecord.DELETE("/delete", sysCallRecordControllers.Delete)
            // sys_call_record列表（分页查询）
            sysCallRecord.GET("/list", sysCallRecordControllers.List)
            // 根据ID获取sys_call_record信息
            sysCallRecord.GET("/:id", sysCallRecordControllers.GetByID)
        }

        app.ZapLog.Info("sys_call_record插件路由注册成功")
    })
}