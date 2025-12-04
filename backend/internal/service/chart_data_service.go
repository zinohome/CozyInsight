package service

import (
	"context"
	"cozy-insight-backend/internal/engine"
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/repository"
	"cozy-insight-backend/pkg/logger"
	"encoding/json"
	"fmt"

	"go.uber.org/zap"
)

type ChartDataService interface {
	GetChartData(ctx context.Context, chartId string) ([]map[string]interface{}, error)
}

type chartDataService struct {
	chartRepo   repository.ChartRepository
	datasetRepo repository.DatasetRepository
}

func NewChartDataService(chartRepo repository.ChartRepository, datasetRepo repository.DatasetRepository) ChartDataService {
	return &chartDataService{
		chartRepo:   chartRepo,
		datasetRepo: datasetRepo,
	}
}

// GetChartData 获取图表数据
func (s *chartDataService) GetChartData(ctx context.Context, chartId string) ([]map[string]interface{}, error) {
	// 1. 获取图表配置
	chart, err := s.chartRepo.Get(ctx, chartId)
	if err != nil {
		logger.Log.Error("failed to get chart", zap.String("chartId", chartId), zap.Error(err))
		return nil, fmt.Errorf("chart not found")
	}

	// 2. 获取关联的数据集表
	if chart.TableID == "" {
		return nil, fmt.Errorf("chart has no associated table")
	}

	table, err := s.datasetRepo.GetTable(ctx, chart.TableID)
	if err != nil {
		logger.Log.Error("failed to get table", zap.String("tableId", chart.TableID), zap.Error(err))
		return nil, fmt.Errorf("table not found")
	}

	// 3. 解析图表配置
	xAxisConfig, yAxisConfig, err := s.parseChartConfig(chart)
	if err != nil {
		return nil, err
	}

	// 4. 构建 SQL
	sql, err := s.buildSQL(table, xAxisConfig, yAxisConfig)
	if err != nil {
		logger.Log.Error("failed to build SQL", zap.Error(err))
		return nil, fmt.Errorf("failed to build SQL: %w", err)
	}

	logger.Log.Info("executing chart query", zap.String("sql", sql))

	// 5. 执行查询
	data, err := s.executeQuery(ctx, sql)
	if err != nil {
		logger.Log.Error("failed to execute query", zap.Error(err))
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return data, nil
}

// parseChartConfig 解析图表配置
func (s *chartDataService) parseChartConfig(chart *model.ChartView) (xAxis, yAxis map[string]interface{}, err error) {
	// 解析 xAxis
	if chart.XAxis != "" {
		if err := json.Unmarshal([]byte(chart.XAxis), &xAxis); err != nil {
			return nil, nil, fmt.Errorf("failed to parse xAxis: %w", err)
		}
	}

	// 解析 yAxis
	if chart.YAxis != "" {
		if err := json.Unmarshal([]byte(chart.YAxis), &yAxis); err != nil {
			return nil, nil, fmt.Errorf("failed to parse yAxis: %w", err)
		}
	}

	return xAxis, yAxis, nil
}

// buildSQL 构建 SQL 查询语句
func (s *chartDataService) buildSQL(table *model.DatasetTable, xAxisConfig, yAxisConfig map[string]interface{}) (string, error) {
	// 获取字段配置
	xFields := s.extractFields(xAxisConfig)
	yFields := s.extractFields(yAxisConfig)

	if len(xFields) == 0 && len(yFields) == 0 {
		return "", fmt.Errorf("no fields configured")
	}

	// 构建 SELECT 子句
	var selectFields []string
	var groupByFields []string

	// 添加维度字段 (xAxis)
	for _, field := range xFields {
		selectFields = append(selectFields, field)
		groupByFields = append(groupByFields, field)
	}

	// 添加指标字段 (yAxis) - 通常需要聚合
	for _, field := range yFields {
		// 默认使用 SUM 聚合
		selectFields = append(selectFields, fmt.Sprintf("SUM(%s) as %s", field, field))
	}

	// 构建完整 SQL
	sql := fmt.Sprintf("SELECT %s FROM %s", joinFields(selectFields), table.TableName)

	// 添加 GROUP BY (如果有维度字段)
	if len(groupByFields) > 0 {
		sql += fmt.Sprintf(" GROUP BY %s", joinFields(groupByFields))
	}

	// 限制返回行数 (避免大量数据)
	sql += " LIMIT 1000"

	return sql, nil
}

// extractFields 从配置中提取字段名
func (s *chartDataService) extractFields(config map[string]interface{}) []string {
	var fields []string

	if config == nil {
		return fields
	}

	// 尝试从 fields 数组中提取
	if fieldsData, ok := config["fields"].([]interface{}); ok {
		for _, f := range fieldsData {
			if fieldMap, ok := f.(map[string]interface{}); ok {
				if name, ok := fieldMap["name"].(string); ok {
					fields = append(fields, name)
				}
			}
		}
	}

	return fields
}

// executeQuery 执行 SQL 查询
func (s *chartDataService) executeQuery(ctx context.Context, sql string) ([]map[string]interface{}, error) {
	if engine.Client == nil {
		return nil, fmt.Errorf("calcite engine not initialized")
	}

	return engine.Client.ExecuteQuery(ctx, sql)
}

// joinFields 连接字段名
func joinFields(fields []string) string {
	if len(fields) == 0 {
		return "*"
	}

	result := ""
	for i, field := range fields {
		if i > 0 {
			result += ", "
		}
		result += field
	}
	return result
}
