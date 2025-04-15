// main.go
package main

import (
	"loadtest_project/config"
	"loadtest_project/models"
	"loadtest_project/routes"
	"loadtest_project/scheduler"
	"log"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

func main() {
	// 连接数据库
	config.ConnectDB()
	models.DB = config.DB

	// 首次运行时创建数据表（根据需要可以注释掉后续重复建表）
	if err := models.CreateTables(); err != nil {
		log.Fatal("建表失败:", err)
	}

	// 启动任务调度器（自动扫描任务并触发 Locust 压测）
	go scheduler.StartScheduler()

	// 初始化 Gin 引擎
	r := gin.Default()

	// 注册 API 路由
	routes.SetupRoutes(r)

	// 提供静态文件服务（确保你的前端代码位于项目根目录下的 frontend 文件夹中）
	r.Static("/static", "./frontend")

	// 设置默认首页路由，直接返回 index.html
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})

	log.Println("服务器启动于 :8080")
	// 这里 r.Run 必须放在最后，因为它会阻塞后续代码执行
	r.Run(":8080")
}
