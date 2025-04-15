// controllers/controllers.go
package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"loadtest_project/models"
	"loadtest_project/utils"
	"log"
	"net/http"
	"strconv"
	"time"
)

// Register 用户注册接口，接收 JSON 格式 { "username": "...", "password": "..." }
func Register(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "密码加密失败"})
		return
	}
	user.Password = string(hash)
	user.Role = "user" // 默认注册为普通用户
	_, err = models.DB.Exec("INSERT INTO users(username, password, role) VALUES(?,?,?)", user.Username, user.Password, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "注册成功"})
}

// Login 用户登录接口，验证后返回 JWT
func Login(c *gin.Context) {
	var req models.User
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}
	var user models.User
	err := models.DB.QueryRow("SELECT id, username, password, role FROM users WHERE username = ?", req.Username).
		Scan(&user.ID, &user.Username, &user.Password, &user.Role)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户不存在"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "密码错误"})
		return
	}
	token, err := utils.GenerateToken(user.ID, user.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Token生成失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

// SubmitLoadTest 用户提交压测任务接口
// 期望接收 JSON 格式数据，包含：user_id, num_users, ramp_up, target_url, start_time, end_time
type SubmitRequest struct {
	UserID    int    `json:"user_id"`
	NumUsers  int    `json:"num_users"`
	RampUp    int    `json:"ramp_up"`
	TargetURL string `json:"target_url"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// 自定义的 UnmarshalJSON 方法来处理字段类型
func (s *SubmitRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}

	// 先将 JSON 数据解析为一个 map 类型
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// 处理 user_id 字段，确保它是整数类型
	if userID, ok := raw["user_id"].(string); ok {
		// 如果是字符串，尝试转换为 int
		parsedID, err := strconv.Atoi(userID)
		if err != nil {
			return fmt.Errorf("无法将 user_id 转换为整数: %v", err)
		}
		s.UserID = parsedID
	} else if userID, ok := raw["user_id"].(float64); ok {
		// 如果是数字，直接转换为 int
		s.UserID = int(userID)
	}

	// 继续处理其他字段
	if numUsers, ok := raw["num_users"].(float64); ok {
		s.NumUsers = int(numUsers)
	}
	if rampUp, ok := raw["ramp_up"].(float64); ok {
		s.RampUp = int(rampUp)
	}
	if targetURL, ok := raw["target_url"].(string); ok {
		s.TargetURL = targetURL
	}
	if startTime, ok := raw["start_time"].(string); ok {
		s.StartTime = startTime
	}
	if endTime, ok := raw["end_time"].(string); ok {
		s.EndTime = endTime
	}

	return nil
}

// SubmitLoadTest 用户提交压测任务接口
func SubmitLoadTest(c *gin.Context) {
	var req SubmitRequest
	// 绑定 JSON 数据到 SubmitRequest 结构体
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("请求数据格式错误:", err) // 打印具体的错误信息
		c.JSON(http.StatusBadRequest, gin.H{"error": "提交数据格式错误", "detail": err.Error()})
		return
	}

	// 解析 start_time 和 end_time
	const layoutWithoutSeconds = "2006-01-02T15:04"
	startTime, err := time.Parse(layoutWithoutSeconds, req.StartTime)
	if err != nil {
		log.Println("解析 start_time 错误，输入:", req.StartTime, "错误信息:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "start_time 格式错误", "detail": err.Error()})
		return
	}

	endTime, err := time.Parse(layoutWithoutSeconds, req.EndTime)
	if err != nil {
		log.Println("解析 end_time 错误，输入:", req.EndTime, "错误信息:", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": "end_time 格式错误", "detail": err.Error()})
		return
	}

	// 创建任务对象
	task := models.LoadTest{
		UserID:    req.UserID,
		NumUsers:  req.NumUsers,
		RampUp:    req.RampUp,
		TargetURL: req.TargetURL,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "pending",
	}

	// 保存任务到数据库
	if err := models.CreateLoadTest(&task); err != nil {
		log.Println("任务提交失败:", err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "任务提交失败", "detail": err.Error()})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, gin.H{"message": "任务提交成功，等待审批"})
}

// ApproveLoadTest 管理员审批任务接口，参数：通过 PostForm 提交字段 id（任务ID）
func ApproveLoadTest(c *gin.Context) {
	idStr := c.PostForm("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的任务ID"})
		return
	}
	if err := models.UpdateLoadTestStatus(id, "approved"); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "审批失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "任务审批成功"})
}

// SaveTestResult Locust 压测结束后调用此接口上传测试结果，参数以 JSON 格式传入：
// test_id, tps, avg_response_time, success_count, failure_count
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

// DownloadReport 根据请求参数生成报告并返回下载文件
// URL 参数：test_id 与 format（csv 或 pdf）
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
	// 构造二维字符串数组：第一行为表头，其余行为数据
	records := [][]string{
		{"Test ID", "TPS", "Avg Response Time", "Success Count", "Failure Count"},
	}
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
