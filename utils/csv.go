// utils/csv.go
package utils

import (
	"encoding/csv"
	"os"
)

// GenerateCSV 将二维字符串数组写入 CSV 文件
func GenerateCSV(records [][]string, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	for _, record := range records {
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	return nil
}
