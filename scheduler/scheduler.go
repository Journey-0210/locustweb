package scheduler

import (
	"fmt"
	"loadtest_project/models"
	"loadtest_project/services"
	"time"
)

func StartScheduler() {
	for {
		tasks, err := models.GetApprovedTasksReadyToRun() // 现在 models 有了
		if err != nil {
			fmt.Println("调度器查询任务失败:", err)
		} else {
			now := time.Now()
			for _, task := range tasks {
				// start_time <= now <= end_time
				if now.After(task.StartTime) && now.Before(task.EndTime) {
					fmt.Printf("触发任务 id:%d, target:%s\n", task.ID, task.TargetURL)
					models.UpdateLoadTestStatus(task.ID, "running")
					go services.StartLoadTest(task)
				}
			}
		}
		time.Sleep(30 * time.Second)
	}
}
