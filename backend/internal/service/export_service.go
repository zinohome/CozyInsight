package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"

	"github.com/xuri/excelize/v2"
)

type ExportService interface {
	ExportToExcel(ctx context.Context, data []map[string]interface{}, filename string) (string, error)
	ExportToCSV(ctx context.Context, data []map[string]interface{}, filename string) (string, error)
}

type exportService struct{}

func NewExportService() ExportService {
	return &exportService{}
}

// ExportToExcel 导出数据为Excel文件
func (s *exportService) ExportToExcel(ctx context.Context, data []map[string]interface{}, filename string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("no data to export")
	}

	f := excelize.NewFile()
	defer f.Close()

	sheetName := "Sheet1"
	
	// 获取列名
	var columns []string
	for key := range data[0] {
		columns = append(columns, key)
	}

	// 写入表头
	for i, col := range columns {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, col)
	}

	// 设置表头样式
	style, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"#E0E0E0"}, Pattern: 1},
	})
	f.SetCellStyle(sheetName, "A1", fmt.Sprintf("%s1", string(rune('A'+len(columns)-1))), style)

	// 写入数据
	for rowIdx, row := range data {
		for colIdx, col := range columns {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			value := row[col]
			f.SetCellValue(sheetName, cell, value)
		}
	}

	// 自动调整列宽
	for i := range columns {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheetName, colName, colName, 15)
	}

	// 保存文件
	filepath := fmt.Sprintf("/tmp/%s.xlsx", filename)
	if err := f.SaveAs(filepath); err != nil {
		return "", fmt.Errorf("failed to save excel: %w", err)
	}

	return filepath, nil
}

// ExportToCSV 导出数据为CSV文件
func (s *exportService) ExportToCSV(ctx context.Context, data []map[string]interface{}, filename string) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("no data to export")
	}

	filepath := fmt.Sprintf("/tmp/%s.csv", filename)
	file, err := os.Create(filepath)
	if err != nil {
		return "", fmt.Errorf("failed to create csv file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 获取列名
	var columns []string
	for key := range data[0] {
		columns = append(columns, key)
	}

	// 写入表头
	if err := writer.Write(columns); err != nil {
		return "", fmt.Errorf("failed to write header: %w", err)
	}

	// 写入数据
	for _, row := range data {
		var record []string
		for _, col := range columns {
			value := row[col]
			record = append(record, fmt.Sprintf("%v", value))
		}
		if err := writer.Write(record); err != nil {
			return "", fmt.Errorf("failed to write row: %w", err)
		}
	}

	return filepath, nil
}
