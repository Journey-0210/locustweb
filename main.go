package main

import (
	"log"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"

	"loadtest_project/config"
	"loadtest_project/models"
	"loadtest_project/routes"
	"loadtest_project/scheduler"
)

func main() {
	// 1. 连接数据库
	config.ConnectDB()
	models.DB = config.DB

	// 2. 创建表结构（首次运行时开启）
	if err := models.CreateTables(); err != nil {
		log.Fatal("建表失败:", err)
	}

	// 3. 启动调度器
	go scheduler.StartScheduler()

	// 4. 初始化 Gin
	r := gin.Default()
	r.Use(cors.Default())

	// 5. 注册 API 路由
	routes.SetupRoutes(r)
	routes.SetupAdminRoutes(r)

	// 6. 提供前端静态资源（所有 .html, .js, .css 等）
	r.Static("/static", "./frontend")

	// 7. 页面入口（注意这些 HTML 文件需放在 frontend 根目录下）
	r.GET("/", func(c *gin.Context) {
		c.File("./frontend/index.html")
	})
	r.GET("/dashboard", func(c *gin.Context) {
		c.File("./frontend/dashboard.html")
	})
	r.GET("/admin", func(c *gin.Context) {
		c.File("./frontend/admin.html")
	})

	// 8. 启动服务
	log.Println("服务器启动于 :8080")
	r.Run(":8080")
}
