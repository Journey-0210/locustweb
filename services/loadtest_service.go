package services

import (
	"loadtest_project/models"
	"log"
	"os/exec"
	"strconv"
)

func StartLoadTest(task models.LoadTest) {
	cmd := exec.Command("locust", "-f", "test.py",
		"--host", task.TargetURL,
		"-u", strconv.Itoa(task.NumUsers),
		"--headless")
	err := cmd.Run()
	if err != nil {
		log.Println("压测失败:", err)
		models.DB.Exec("UPDATE load_tests SET status='failed' WHERE id=?", task.ID)
		return
	}
	models.DB.Exec("UPDATE load_tests SET status='completed' WHERE id=?", task.ID)
}
