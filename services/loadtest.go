// services/loadtest.go
package services

import (
	"loadtest_project/models"
)

// StartLoadTest 由调度器调用，触发无 UI 模式的 Locust 压测
func StartLoadTest(task models.LoadTest) {
	// RunLocustForTask 会更新状态并保存结果
	RunLocustForTask(task)
}
