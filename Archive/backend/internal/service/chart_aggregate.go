package service

import (
	"cozy-insight-backend/internal/model"
	"encoding/json"
	"fmt"
)

// AggregateType 聚合类型
type AggregateType string

const (
	AggregateSum   AggregateType = "SUM"
	AggregateAvg   AggregateType = "AVG"
	AggregateCount AggregateType = "COUNT"
	AggregateMax   AggregateType = "MAX"
	AggregateMin   AggregateType = "MIN"
)

// ValidateAggregate 验证聚合类型
func ValidateAggregate(agg string) bool {
	validAggregates := []AggregateType{
		AggregateSum,
		AggregateAvg,
		AggregateCount,
		AggregateMax,
		AggregateMin,
	}

	for _, valid := range validAggregates {
		if AggregateType(agg) == valid {
			return true
		}
	}
	return false
}

// buildAggregateSQL 构建聚合SQL表达式
func buildAggregateSQL(field FieldConfig) string {
	if field.Aggregate == "" {
		return field.Name
	}

	// 验证聚合类型
	if !ValidateAggregate(field.Aggregate) {
		return field.Name
	}

	// COUNT(*)特殊处理
	if field.Aggregate == string(AggregateCount) && field.Name == "*" {
		return "COUNT(*) AS count"
	}

	// 构建聚合表达式
	return fmt.Sprintf("%s(%s) AS %s", field.Aggregate, field.Name, field.Name)
}

// buildSortClause 构建ORDER BY子句
func buildSortClause(fields []FieldConfig) string {
	var sortFields []string

	for _, field := range fields {
		if field.Sort != "" {
			// 验证排序方向
			sort := field.Sort
			if sort != "ASC" && sort != "DESC" {
				sort = "ASC"
			}
			sortFields = append(sortFields, fmt.Sprintf("%s %s", field.Name, sort))
		}
	}

	if len(sortFields) == 0 {
		return ""
	}

	return " ORDER BY " + joinStrings(sortFields, ", ")
}

// joinStrings 连接字符串
func joinStrings(strs []string, sep string) string {
	if len(strs) == 0 {
		return ""
	}

	result := strs[0]
	for i := 1; i < len(strs); i++ {
		result += sep + strs[i]
	}
	return result
}

// ParseChartConfig 解析图表配置
func ParseChartConfig(chart *model.ChartView) (*ChartQueryConfig, error) {
	config := &ChartQueryConfig{
		ChartID: chart.ID,
		TableID: chart.TableID,
	}

	// 解析X轴
	if chart.XAxis != "" {
		var xAxisConfig AxisConfig
		if err := json.Unmarshal([]byte(chart.XAxis), &xAxisConfig); err != nil {
			return nil, fmt.Errorf("invalid xAxis config: %w", err)
		}
		config.XAxis = xAxisConfig
	}

	// 解析Y轴
	if chart.YAxis != "" {
		var yAxisConfig AxisConfig
		if err := json.Unmarshal([]byte(chart.YAxis), &yAxisConfig); err != nil {
			return nil, fmt.Errorf("invalid yAxis config: %w", err)
		}
		config.YAxis = yAxisConfig
	}

	return config, nil
}

// ChartQueryConfig 图表查询配置
type ChartQueryConfig struct {
	ChartID string
	TableID string
	XAxis   AxisConfig
	YAxis   AxisConfig
}

// GetDimensions 获取维度字段(X轴)
func (c *ChartQueryConfig) GetDimensions() []string {
	var dimensions []string
	for _, field := range c.XAxis.Fields {
		dimensions = append(dimensions, field.Name)
	}
	return dimensions
}

// GetMetrics 获取指标字段(Y轴)
func (c *ChartQueryConfig) GetMetrics() []FieldConfig {
	return c.YAxis.Fields
}

// HasAggregates 是否有聚合字段
func (c *ChartQueryConfig) HasAggregates() bool {
	for _, field := range c.YAxis.Fields {
		if field.Aggregate != "" {
			return true
		}
	}
	return false
}
