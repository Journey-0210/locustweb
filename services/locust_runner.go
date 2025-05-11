package services

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"

	"loadtest_project/models"
)

// round4 保留 4 位小数
func round4(f float64) float64 {
	return math.Round(f*1e4) / 1e4
}

// RunLocustForTask 运行 Locust（无 UI），解析 CSV 结果并写入数据库
func RunLocustForTask(task models.LoadTest) {
	resultsDir := "results"
	prefix := fmt.Sprintf("task_%d_%d", task.ID, time.Now().Unix())
	_ = os.MkdirAll(resultsDir, 0755)

	// 获取 locustfile.py 的绝对路径
	locustPath, err := filepath.Abs("locust/locustfile.py")
	if err != nil {
		fmt.Println("无法获取 locustfile 路径:", err)
		models.UpdateLoadTestStatus(task.ID, "failed")
		return
	}

	// 构造命令
	cmd := exec.Command(
		`C:\Users\Pua Wei Jian\AppData\Local\Programs\Python\Python313\python.exe`,
		"-m", "locust",
		"-f", locustPath,
		"--headless",
		"-u", strconv.Itoa(task.NumUsers),
		"-r", strconv.Itoa(task.RampUp),
		"--host", task.TargetURL,
		"--run-time", fmt.Sprintf("%ds", int(task.EndTime.Sub(task.StartTime).Seconds())),
		"--csv", filepath.Join(resultsDir, prefix),
		"--only-summary",
	)

	if output, err := cmd.CombinedOutput(); err != nil {
		fmt.Printf("Locust 运行失败: %v\n日志:\n%s\n", err, string(output))
		models.UpdateLoadTestStatus(task.ID, "failed")
		return
	}

	// 打开 stats CSV
	csvFile := filepath.Join(resultsDir, prefix+"_stats.csv")
	f, err := os.Open(csvFile)
	if err != nil {
		fmt.Println("打开 CSV 失败:", err)
		models.UpdateLoadTestStatus(task.ID, "failed")
		return
	}
	defer f.Close()

	reader := csv.NewReader(f)
	// 跳过表头
	if _, err := reader.Read(); err != nil {
		fmt.Println("读取 CSV 表头失败:", err)
		models.UpdateLoadTestStatus(task.ID, "failed")
		return
	}

	// 准备变量
	var (
		totalRequests int
		failures      int
		avgResp       float64
		minResp       float64
		maxResp       float64
		rps           float64
	)

	// 解析“Aggregated”或“Total”行
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			fmt.Println("读取 CSV 行错误:", err)
			continue
		}
		// 有些版本第一列是空字符串、第二列是名称；新版本直接第一列是名称
		name := record[0]
		if name == "" {
			name = record[1]
		}
		if name == "Aggregated" || name == "Total" {
			// 参考你的样例：
			// record[2] = Request Count
			// record[3] = Failure Count
			// record[5] = Average response time
			// record[6] = Min response time
			// record[7] = Max response time
			// record[9] = Requests/s
			totalRequests, _ = strconv.Atoi(record[2])
			failures, _ = strconv.Atoi(record[3])
			avgResp, _ = strconv.ParseFloat(record[5], 64)
			minResp, _ = strconv.ParseFloat(record[6], 64)
			maxResp, _ = strconv.ParseFloat(record[7], 64)
			rps, _ = strconv.ParseFloat(record[9], 64)
			break
		}
	}

	// 四舍五入到 4 位小数
	avgResp = round4(avgResp)
	minResp = round4(minResp)
	maxResp = round4(maxResp)
	rps = round4(rps)

	// 计算 errorRate 和 availability
	var errorRate float64
	if totalRequests > 0 {
		errorRate = round4(float64(failures) / float64(totalRequests))
	}
	availability := round4(1 - errorRate)

	// 构造 TestResult 并写入数据库
	result := models.TestResult{
		TestID:              task.ID,
		TPS:                 rps,
		AvgResponseTime:     avgResp,
		SuccessCount:        totalRequests - failures,
		FailureCount:        failures,
		ErrorRate:           errorRate,
		MaxResponseTime:     maxResp,
		MinResponseTime:     minResp,
		RPS:                 rps,
		DownloadSpeed:       0, // 如需，可由 Python JSON 中解析并 round4
		DownloadSize:        0,
		DownloadDuration:    round4(task.EndTime.Sub(task.StartTime).Seconds()),
		DNSTime:             0,
		ConnectTime:         0,
		TTFB:                0,
		ContentDownloadTime: 0,
		Availability:        availability,
	}

	if err := models.CreateTestResult(&result); err != nil {
		fmt.Println("写入测试结果失败:", err)
		models.UpdateLoadTestStatus(task.ID, "failed")
		return
	}

	models.UpdateLoadTestStatus(task.ID, "completed")
	fmt.Printf("任务 %d 已完成，结果已保存\n", task.ID)
}
