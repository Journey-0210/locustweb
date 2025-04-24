package routes

import (
	"github.com/gin-gonic/gin"
	"loadtest_project/controllers"
)

func SetupRoutes(r *gin.Engine) {
	// 注册 / 登录
	r.POST("/api/register", controllers.Register)
	r.POST("/api/login", controllers.Login)
	// 用户提交任务
	r.POST("/api/submit", controllers.SubmitLoadTest)
	// Locust 回调存结果
	r.POST("/api/upload_result", controllers.SaveTestResult)
	// 用户下载报告
	r.GET("/api/download_report", controllers.DownloadReport)
}

func SetupAdminRoutes(r *gin.Engine) {
	admin := r.Group("/admin")
	{
		admin.GET("/tasks", controllers.AdminOnlyMiddleware(), controllers.GetPendingTasks)
		admin.POST("/approve", controllers.AdminOnlyMiddleware(), controllers.ApproveLoadTest)
	}
}
