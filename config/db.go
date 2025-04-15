// config/db.go
package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

var DB *sql.DB

func ConnectDB() {
	var err error
	// 修改为你的 MySQL 用户名、密码、数据库名
	dsn := "root:WeiJian0210!@tcp(127.0.0.1:3306)/loadtest?parseTime=true"
	DB, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}
	if err = DB.Ping(); err != nil {
		log.Fatal("数据库不可用:", err)
	}
	fmt.Println("✅ Connected to MySQL successfully!")
}
