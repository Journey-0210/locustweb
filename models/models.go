package models

import (
	"database/sql"
	"fmt"
	"time"
)

var DB *sql.DB

type User struct {
	ID       int
	Username string
	Password string
	Role     string
}

type LoadTest struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	NumUsers  int       `json:"num_users"`
	RampUp    int       `json:"ramp_up"`
	TargetURL string    `json:"target_url"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}

type TestResult struct {
	ID                  int     `json:"id"`
	TestID              int     `json:"test_id"`
	TPS                 float64 `json:"tps"`
	AvgResponseTime     float64 `json:"avg_response_time"`
	SuccessCount        int     `json:"success_count"`
	FailureCount        int     `json:"failure_count"`
	ErrorRate           float64 `json:"error_rate"`
	MaxResponseTime     float64 `json:"max_response_time"`
	MinResponseTime     float64 `json:"min_response_time"`
	RPS                 float64 `json:"rps"`
	DownloadSpeed       float64 `json:"download_speed"`
	DownloadSize        float64 `json:"download_size"`
	DownloadDuration    float64 `json:"download_duration"`
	DNSTime             float64 `json:"dns_time"`
	ConnectTime         float64 `json:"connect_time"`
	TTFB                float64 `json:"ttfb"`
	ContentDownloadTime float64 `json:"content_download_time"`
	Availability        float64 `json:"availability"`
}

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
			error_rate DOUBLE,
			max_response_time DOUBLE,
			min_response_time DOUBLE,
			rps DOUBLE,
			download_speed DOUBLE,
			download_size DOUBLE,
			download_duration DOUBLE,
			dns_time DOUBLE,
			connect_time DOUBLE,
			ttfb DOUBLE,
			content_download_time DOUBLE,
			availability DOUBLE,
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

func UpdateLoadTestStatus(id int, status string) error {
	_, err := DB.Exec("UPDATE load_tests SET status=? WHERE id=?", status, id)
	return err
}

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

func CreateTestResult(r *TestResult) error {
	_, err := DB.Exec(`
		INSERT INTO test_results (
			test_id, tps, avg_response_time, success_count, failure_count,
			error_rate, max_response_time, min_response_time, rps, download_speed,
			download_size, download_duration, dns_time, connect_time, ttfb,
			content_download_time, availability
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.TestID, r.TPS, r.AvgResponseTime, r.SuccessCount, r.FailureCount,
		r.ErrorRate, r.MaxResponseTime, r.MinResponseTime, r.RPS, r.DownloadSpeed,
		r.DownloadSize, r.DownloadDuration, r.DNSTime, r.ConnectTime, r.TTFB,
		r.ContentDownloadTime, r.Availability,
	)
	return err
}

func GetTestResultsByTestID(testID int) ([]TestResult, error) {
	rows, err := DB.Query(`
		SELECT id, test_id, tps, avg_response_time, success_count, failure_count,
		       error_rate, max_response_time, min_response_time, rps, download_speed,
		       download_size, download_duration, dns_time, connect_time, ttfb,
		       content_download_time, availability
		FROM test_results WHERE test_id = ?`, testID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TestResult
	for rows.Next() {
		var r TestResult
		if err := rows.Scan(
			&r.ID, &r.TestID, &r.TPS, &r.AvgResponseTime, &r.SuccessCount, &r.FailureCount,
			&r.ErrorRate, &r.MaxResponseTime, &r.MinResponseTime, &r.RPS, &r.DownloadSpeed,
			&r.DownloadSize, &r.DownloadDuration, &r.DNSTime, &r.ConnectTime, &r.TTFB,
			&r.ContentDownloadTime, &r.Availability,
		); err != nil {
			continue
		}
		results = append(results, r)
	}
	return results, nil
}
