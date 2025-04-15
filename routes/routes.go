package routes

import (
	"github.com/gin-gonic/gin"
	"loadtest_project/controllers"
)

func SetupRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)
		api.POST("/submit", controllers.SubmitLoadTest)
		api.POST("/approve", controllers.ApproveLoadTest)
		api.POST("/upload_result", controllers.SaveTestResult)
		api.GET("/download_report", controllers.DownloadReport)
	}
}
