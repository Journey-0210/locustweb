package models

import (
	"database/sql"
	"fmt"
	"time"
)

// 全局数据库连接实例
var DB *sql.DB

// User 用户模型
type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

// LoadTest 表示一条压测任务记录
// StartTime 和 EndTime 使用 time.Time 类型，以便于调度器比较
type LoadTest struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	NumUsers  int       `json:"num_users"`
	RampUp    int       `json:"ramp_up"`
	TargetURL string    `json:"target_url"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"` // 任务状态: pending, approved, running, completed, failed
}

// TestResult 存储单次压测的结果
type TestResult struct {
	ID              int     `json:"id"`
	TestID          int     `json:"test_id"`
	TPS             float64 `json:"tps"`
	AvgResponseTime float64 `json:"avg_response_time"`
	SuccessCount    int     `json:"success_count"`
	FailureCount    int     `json:"failure_count"`
}

// CreateTables 初始化数据库表结构
func CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'user'
		);`,
		`CREATE TABLE IF NOT EXISTS load_tests (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT NOT NULL,
			num_users INT NOT NULL,
			ramp_up INT NOT NULL,
			target_url VARCHAR(255) NOT NULL,
			start_time DATETIME NOT NULL,
			end_time DATETIME NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending',
			FOREIGN KEY (user_id) REFERENCES users(id)
		);`,
		`CREATE TABLE IF NOT EXISTS test_results (
			id INT AUTO_INCREMENT PRIMARY KEY,
			test_id INT NOT NULL,
			tps DOUBLE,
			avg_response_time DOUBLE,
			success_count INT,
			failure_count INT,
			FOREIGN KEY (test_id) REFERENCES load_tests(id)
		);`,
	}
	for _, q := range queries {
		if _, err := DB.Exec(q); err != nil {
			return fmt.Errorf("建表失败: %v", err)
		}
	}
	return nil
}

// CreateLoadTest 插入一条新的压测任务
func CreateLoadTest(t *LoadTest) error {
	res, err := DB.Exec(
		"INSERT INTO load_tests(user_id, num_users, ramp_up, target_url, start_time, end_time, status) VALUES(?,?,?,?,?,?,?)",
		t.UserID, t.NumUsers, t.RampUp, t.TargetURL, t.StartTime, t.EndTime, t.Status,
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err == nil {
		t.ID = int(id)
	}
	return nil
}

// UpdateLoadTestStatus 更新指定任务的状态
func UpdateLoadTestStatus(id int, status string) error {
	_, err := DB.Exec("UPDATE load_tests SET status=? WHERE id=?", status, id)
	return err
}

// GetApprovedTasksReadyToRun 查询所有已经审批且到执行时间的任务
func GetApprovedTasksReadyToRun() ([]LoadTest, error) {
	now := time.Now()
	rows, err := DB.Query(
		"SELECT id, user_id, num_users, ramp_up, target_url, start_time, end_time FROM load_tests WHERE status='approved' AND start_time <= ?", now,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []LoadTest
	for rows.Next() {
		var t LoadTest
		if err := rows.Scan(&t.ID, &t.UserID, &t.NumUsers, &t.RampUp, &t.TargetURL, &t.StartTime, &t.EndTime); err != nil {
			continue
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// CreateTestResult 保存一次压测结果
func CreateTestResult(r *TestResult) error {
	_, err := DB.Exec(
		"INSERT INTO test_results(test_id, tps, avg_response_time, success_count, failure_count) VALUES(?,?,?,?,?)",
		r.TestID, r.TPS, r.AvgResponseTime, r.SuccessCount, r.FailureCount,
	)
	return err
}

// GetTestResultsByTestID 获取指定任务的所有测试结果
func GetTestResultsByTestID(testID int) ([]TestResult, error) {
	rows, err := DB.Query(
		"SELECT id, test_id, tps, avg_response_time, success_count, failure_count FROM test_results WHERE test_id = ?", testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TestResult
	for rows.Next() {
		var r TestResult
		if err := rows.Scan(&r.ID, &r.TestID, &r.TPS, &r.AvgResponseTime, &r.SuccessCount, &r.FailureCount); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}
