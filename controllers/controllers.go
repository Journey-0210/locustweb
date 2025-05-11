// controllers/controllers.go
package controllers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"loadtest_project/models"
	"loadtest_project/services"
	"loadtest_project/utils"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Register 用户注册接口，角色强制为普通用户
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	// 密码加密
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user.Password = string(hash)
	// 强制角色为 user
	user.Role = "user"
	_, err = models.DB.Exec("INSERT INTO users(username, password, role) VALUES(?,?,?)", user.Username, user.Password, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// Login 用户登录接口，返回 JWT 和角色
func Login(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var user models.User
	err := models.DB.QueryRow(
		"SELECT id, username, password, role FROM users WHERE username = ?", req.Username,
	).Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}
	// 生成包含角色的 Token
	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token生成失败"})
		return
	}
	// 返回 token 和 role
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"role":  user.Role,
	})
}

// AdminOnlyMiddleware 验证仅管理员可访问
func AdminOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		// 支持“Bearer <token>”或直接"<token>"
		token := authHeader
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
		claims, err := utils.ParseToken(token)
		if err != nil || claims.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可访问"})
			c.Abort()
			return
		}
		c.Next()
	}
}

type PendingTaskItem struct {
	ID        int       `json:"id"`
	Username  string    `json:"username"` // 新增：用户名字段
	NumUsers  int       `json:"num_users"`
	RampUp    int       `json:"ramp_up"`
	TargetURL string    `json:"target_url"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

// GetTasksByStatus 根据 query 参数 ?status=xxx 拉任务
func GetTasksByStatus(c *gin.Context) {
	// 1. 校验管理员身份
	authHeader := c.GetHeader("Authorization")
	tok := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ParseToken(tok)
	if err != nil || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可访问"})
		return
	}

	// 2. 读取 status 参数，默认 "pending"
	status := c.Query("status")
	if status == "" {
		status = "pending"
	}
	valid := map[string]bool{"pending": true, "approved": true, "rejected": true, "all": true}
	if !valid[status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的 status 参数"})
		return
	}

	// 3. 构造 SQL
	var rows *sql.Rows
	if status == "all" {
		rows, err = models.DB.Query(`
            SELECT lt.id, u.username, lt.num_users, lt.ramp_up,
                   lt.target_url, lt.start_time, lt.end_time, lt.status
              FROM load_tests lt
              JOIN users u ON lt.user_id = u.id
          ORDER BY lt.start_time ASC
        `)
	} else {
		rows, err = models.DB.Query(`
            SELECT lt.id, u.username, lt.num_users, lt.ramp_up,
                   lt.target_url, lt.start_time, lt.end_time, lt.status
              FROM load_tests lt
              JOIN users u ON lt.user_id = u.id
             WHERE lt.status = ?
          ORDER BY lt.start_time ASC
        `, status)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	// 4. Scan 结果
	var tasks []PendingTaskItem
	for rows.Next() {
		var t PendingTaskItem
		if err := rows.Scan(
			&t.ID, &t.Username, &t.NumUsers, &t.RampUp,
			&t.TargetURL, &t.StartTime, &t.EndTime, &t.Status,
		); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// SubmitRequest 接收前端 JSON，自动把 start_time/end_time 解析成 time.Time
type SubmitRequest struct {
	NumUsers  int       `json:"num_users"`
	RampUp    int       `json:"ramp_up"`
	TargetURL string    `json:"target_url"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// UnmarshalJSON 自定义反序列化，兼容多种输入格式
func (s *SubmitRequest) UnmarshalJSON(data []byte) error {
	// 先用一个中间结构拿到 raw 值
	var raw struct {
		NumUsers  int         `json:"num_users"`
		RampUp    int         `json:"ramp_up"`
		TargetURL string      `json:"target_url"`
		StartRaw  interface{} `json:"start_time"`
		EndRaw    interface{} `json:"end_time"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	s.NumUsers = raw.NumUsers
	s.RampUp = raw.RampUp
	s.TargetURL = raw.TargetURL

	// 统一解析函数：尝试多种常见格式
	parseTime := func(v interface{}) (time.Time, error) {
		str, ok := v.(string)
		if !ok {
			return time.Time{}, fmt.Errorf("时间字段不是字符串: %T", v)
		}
		for _, layout := range []string{
			time.RFC3339,           // 2006-01-02T15:04:05Z07:00
			"2006-01-02T15:04:05Z", // 2006-01-02T15:04:05Z
			"2006-01-02T15:04",     // 2006-01-02T15:04
		} {
			if tm, err := time.Parse(layout, str); err == nil {
				return tm, nil
			}
		}
		return time.Time{}, fmt.Errorf("无法解析时间: %q", str)
	}

	var err error
	if s.StartTime, err = parseTime(raw.StartRaw); err != nil {
		return fmt.Errorf("start_time 解析失败: %w", err)
	}
	if s.EndTime, err = parseTime(raw.EndRaw); err != nil {
		return fmt.Errorf("end_time 解析失败: %w", err)
	}
	return nil
}

// SubmitLoadTest 用户提交压测任务
func SubmitLoadTest(c *gin.Context) {
	// —— 1. 验证 Token，取出用户 ID ——
	authHeader := c.GetHeader("Authorization")
	tok := strings.TrimPrefix(authHeader, "Bearer ")
	claims, err := utils.ParseToken(tok)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的 Token"})
		return
	}
	userID := claims.UserID

	// —— 2. 直接 Bind 到 SubmitRequest，UnmarshalJSON 会做时间解析 ——
	var req SubmitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("请求数据格式错误:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "提交数据格式错误", "detail": err.Error()})
		return
	}

	// —— 3. 构造 LoadTest 并保存 ——
	task := models.LoadTest{
		UserID:    userID,
		NumUsers:  req.NumUsers,
		RampUp:    req.RampUp,
		TargetURL: req.TargetURL,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Status:    "pending",
	}
	if err := models.CreateLoadTest(&task); err != nil {
		log.Println("任务提交失败:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "任务提交失败", "detail": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "任务提交成功，等待审批"})
}

// ApproveLoadTest 管理员审批并启动压测
func ApproveLoadTest(c *gin.Context) {
	AdminOnlyMiddleware()(c)
	if c.IsAborted() {
		return
	}
	idStr := c.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}
	// 更新状态
	models.DB.Exec("UPDATE load_tests SET status='approved' WHERE id=?", id)
	// 查询详情
	var task models.LoadTest
	models.DB.QueryRow("SELECT id, target_url, num_users FROM load_tests WHERE id=?", id).
		Scan(&task.ID, &task.TargetURL, &task.NumUsers)
	// 异步启动压测
	go services.StartLoadTest(task)
	c.JSON(http.StatusOK, gin.H{"message": "任务审批通过，压测已启动"})
}

// RejectLoadTest
func RejectLoadTest(c *gin.Context) {
	// 服用中间件逻辑，剥离 Bearer 前缀
	authHeader := c.GetHeader("Authorization")
	token := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		token = strings.TrimPrefix(authHeader, "Bearer ")
	}
	claims, err := utils.ParseToken(token)
	if err != nil || claims.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "仅管理员可访问"})
		c.Abort()
		return
	}

	// 解析表单 id
	idStr := c.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}
	// 更新状态为 rejected
	if err := models.UpdateLoadTestStatus(id, "rejected"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "拒绝任务失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务已拒绝"})
}

// 获取当前用户所有任务（按 start_time 升序）
func GetUserTasks(c *gin.Context) {
	// 从 Header 拿到 token，去掉 Bearer 前缀
	authHeader := c.GetHeader("Authorization")
	tok := authHeader
	if strings.HasPrefix(authHeader, "Bearer ") {
		tok = strings.TrimPrefix(authHeader, "Bearer ")
	}
	claims, err := utils.ParseToken(tok)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "无效的Token"})
		return
	}
	userID := claims.UserID

	// 查询该用户的所有任务
	rows, err := models.DB.Query(`
        SELECT id, num_users, ramp_up, target_url, start_time, end_time, status
          FROM load_tests
         WHERE user_id = ?
      ORDER BY start_time ASC
    `, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询失败"})
		return
	}
	defer rows.Close()

	var tasks []map[string]interface{}
	for rows.Next() {
		var t models.LoadTest
		if err := rows.Scan(
			&t.ID, &t.NumUsers, &t.RampUp,
			&t.TargetURL, &t.StartTime, &t.EndTime, &t.Status,
		); err != nil {
			continue
		}
		// 转成 JSON 友好结构
		tasks = append(tasks, map[string]interface{}{
			"id":         t.ID,
			"num_users":  t.NumUsers,
			"ramp_up":    t.RampUp,
			"target_url": t.TargetURL,
			"start_time": t.StartTime,
			"end_time":   t.EndTime,
			"status":     t.Status,
		})
	}
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// SaveTestResult 保存压测结果
func SaveTestResult(c *gin.Context) {
	var result models.TestResult
	if err := c.ShouldBindJSON(&result); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	if err := models.CreateTestResult(&result); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存测试结果失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "测试结果保存成功"})
}

// DownloadReport 下载报告
func DownloadReport(c *gin.Context) {
	testIDStr := c.Query("test_id")
	format := c.Query("format")
	testID, err := strconv.Atoi(testIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的 test_id"})
		return
	}
	results, err := models.GetTestResultsByTestID(testID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询测试结果失败"})
		return
	}
	records := [][]string{{"Test ID", "TPS", "Avg Response Time", "Success Count", "Failure Count"}}
	for _, res := range results {
		record := []string{
			strconv.Itoa(res.TestID),
			fmt.Sprintf("%.2f", res.TPS),
			fmt.Sprintf("%.2f", res.AvgResponseTime),
			strconv.Itoa(res.SuccessCount),
			strconv.Itoa(res.FailureCount),
		}
		records = append(records, record)
	}
	if format == "csv" {
		filename := "report_" + testIDStr + ".csv"
		if err := utils.GenerateCSV(records, filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "CSV生成失败"})
			return
		}
		c.FileAttachment(filename, filename)
	} else if format == "pdf" {
		filename := "report_" + testIDStr + ".pdf"
		if err := utils.GeneratePDFReport(records, filename); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "PDF生成失败"})
			return
		}
		c.FileAttachment(filename, filename)
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的格式"})
	}
}
