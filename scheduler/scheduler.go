package scheduler

import (
	"fmt"
	"loadtest_project/models"
	"os/exec"
	"strconv"
	"time"
)

// StartScheduler 每30秒扫描一次 approved 状态的任务，如果当前时间处于任务时间内，则触发 Locust 压测
func StartScheduler() {
	for {
		tasks, err := models.GetApprovedTasksReadyToRun()
		if err != nil {
			fmt.Println("调度器查询任务失败:", err)
		} else {
			now := time.Now()
			for _, task := range tasks {
				if now.After(task.StartTime) && now.Before(task.EndTime) && task.Status == "approved" {
					fmt.Printf("触发任务 id:%d, target:%s\n", task.ID, task.TargetURL)
					models.UpdateLoadTestStatus(task.ID, "running")
					go RunLocustTask(task)
				}
			}
		}
		time.Sleep(30 * time.Second)
	}
}

// RunLocustTask 调用 Locust 脚本，根据任务参数启动压测
func RunLocustTask(task models.LoadTest) {
	duration := int(task.EndTime.Sub(task.StartTime).Seconds())
	cmd := exec.Command("locust", "-f", "locust/test_script.py", "--headless",
		"-u", strconv.Itoa(task.NumUsers),
		"-r", strconv.Itoa(task.RampUp),
		"--host", task.TargetURL,
		"--run-time", fmt.Sprintf("%ds", duration),
	)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("任务 id:%d 压测失败: %v, 输出: %s\n", task.ID, err, string(output))
		models.UpdateLoadTestStatus(task.ID, "failed")
		return
	}
	fmt.Printf("任务 id:%d 压测完成, 输出: %s\n", task.ID, string(output))
	models.UpdateLoadTestStatus(task.ID, "completed")
}
