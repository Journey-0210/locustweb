package models

import (
	"database/sql"
	"time"

	"loadtest_project/config"
)

var DB *sql.DB

// User 用户模型
type User struct {
	ID       int
	Username string
	Password string // 已加密
	Role     string // "user" 或 "admin"
}

// LoadTest 压测任务模型
type LoadTest struct {
	ID        int
	UserID    int
	NumUsers  int
	RampUp    int
	TargetURL string
	StartTime time.Time
	EndTime   time.Time
	Status    string // pending, approved, running, completed, failed
}

// TestResult 压测结果模型
type TestResult struct {
	ID              int
	TestID          int
	TPS             float64
	AvgResponseTime float64
	SuccessCount    int
	FailureCount    int
	CreatedAt       time.Time
}

// CreateTables 建表函数（首次运行时调用）
// 可在 main.go 中临时调用以创建表
func CreateTables() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INT AUTO_INCREMENT PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL,
			role VARCHAR(50)
		);`,
		`CREATE TABLE IF NOT EXISTS load_tests (
			id INT AUTO_INCREMENT PRIMARY KEY,
			user_id INT,
			num_users INT,
			ramp_up INT,
			target_url VARCHAR(255),
			start_time DATETIME,
			end_time DATETIME,
			status VARCHAR(50),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS test_results (
			id INT AUTO_INCREMENT PRIMARY KEY,
			test_id INT,
			tps FLOAT,
			avg_response_time FLOAT,
			success_count INT,
			failure_count INT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`,
	}
	for _, q := range queries {
		_, err := config.DB.Exec(q)
		if err != nil {
			return err
		}
	}
	return nil
}

// CreateLoadTest 将任务写入数据库
func CreateLoadTest(task *LoadTest) error {
	_, err := config.DB.Exec("INSERT INTO load_tests(user_id, num_users, ramp_up, target_url, start_time, end_time, status) VALUES(?,?,?,?,?,?,?)",
		task.UserID, task.NumUsers, task.RampUp, task.TargetURL, task.StartTime, task.EndTime, task.Status)
	return err
}

// GetLoadTestByID 根据任务ID查询任务
func GetLoadTestByID(id int) (*LoadTest, error) {
	var task LoadTest
	err := config.DB.QueryRow("SELECT id, user_id, num_users, ramp_up, target_url, start_time, end_time, status FROM load_tests WHERE id = ?", id).
		Scan(&task.ID, &task.UserID, &task.NumUsers, &task.RampUp, &task.TargetURL, &task.StartTime, &task.EndTime, &task.Status)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateLoadTestStatus 更新任务状态
func UpdateLoadTestStatus(id int, status string) error {
	_, err := config.DB.Exec("UPDATE load_tests SET status = ? WHERE id = ?", status, id)
	return err
}

// GetApprovedTasksReadyToRun 查询所有状态为 approved 的任务
func GetApprovedTasksReadyToRun() ([]LoadTest, error) {
	rows, err := config.DB.Query("SELECT id, user_id, num_users, ramp_up, target_url, start_time, end_time, status FROM load_tests WHERE status = 'approved'")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tasks []LoadTest
	for rows.Next() {
		var t LoadTest
		err := rows.Scan(&t.ID, &t.UserID, &t.NumUsers, &t.RampUp, &t.TargetURL, &t.StartTime, &t.EndTime, &t.Status)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// CreateTestResult 保存测试结果
func CreateTestResult(result *TestResult) error {
	_, err := config.DB.Exec("INSERT INTO test_results(test_id, tps, avg_response_time, success_count, failure_count, created_at) VALUES(?,?,?,?,?,?)",
		result.TestID, result.TPS, result.AvgResponseTime, result.SuccessCount, result.FailureCount, time.Now())
	return err
}

// GetTestResultsByTestID 根据任务ID查询测试结果
func GetTestResultsByTestID(testID int) ([]TestResult, error) {
	rows, err := config.DB.Query("SELECT test_id, tps, avg_response_time, success_count, failure_count, created_at FROM test_results WHERE test_id = ?", testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var results []TestResult
	for rows.Next() {
		var tr TestResult
		err := rows.Scan(&tr.TestID, &tr.TPS, &tr.AvgResponseTime, &tr.SuccessCount, &tr.FailureCount, &tr.CreatedAt)
		if err != nil {
			return nil, err
		}
		results = append(results, tr)
	}
	return results, nil
}
