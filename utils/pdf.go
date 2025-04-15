// utils/pdf.go
package utils

import (
	"fmt"
	"github.com/phpdave11/gofpdf"
)

// GeneratePDFReport 将二维字符串数组生成 PDF 文件
func GeneratePDFReport(records [][]string, filename string) error {
	if len(records) == 0 {
		return fmt.Errorf("empty records")
	}
	// 创建 PDF 文档，"L" 表示横向，"mm" 是单位，"A4" 是纸张大小
	pdf := gofpdf.New("L", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	// 输出标题
	pdf.Cell(40, 10, "压测结果报告")
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 10)
	colCount := len(records[0])
	colWidth := 270.0 / float64(colCount)
	// 遍历 records 数组生成表格
	for _, row := range records {
		for _, cell := range row {
			pdf.CellFormat(colWidth, 7, cell, "1", 0, "C", false, 0, "")
		}
		pdf.Ln(-1)
	}
	return pdf.OutputFileAndClose(filename)
}
