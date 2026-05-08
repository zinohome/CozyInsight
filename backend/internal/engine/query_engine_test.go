package engine

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildSQL_BasicAggregation(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "SUM", Alias: "total"}},
	}
	sql, args, err := BuildSQL("orders", "mysql", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT `dept`, SUM(`amount`) AS `total` FROM `orders` GROUP BY `dept`", sql)
	assert.Len(t, args, 0)
}

func TestBuildSQL_WithFilter(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "COUNT"}},
		Filters:    []Filter{{Field: "status", Operator: "=", Value: "1"}},
	}
	sql, args, err := BuildSQL("orders", "mysql", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT `dept`, COUNT(`amount`) AS `count_amount` FROM `orders` WHERE `status` = ? GROUP BY `dept`", sql)
	assert.Equal(t, []interface{}{"1"}, args)
}

func TestBuildSQL_WithOrder(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "month"}},
		Metrics:    []Metric{{Field: "revenue", Aggregation: "SUM"}},
		Orders:     []Order{{Field: "revenue", Direction: "desc"}},
		Limit:      10,
	}
	sql, args, err := BuildSQL("sales", "mysql", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT `month`, SUM(`revenue`) AS `sum_revenue` FROM `sales` GROUP BY `month` ORDER BY `revenue` desc LIMIT 10", sql)
	assert.Len(t, args, 0)
}

func TestBuildSQL_InvalidAggregation(t *testing.T) {
	config := ChartQueryConfig{
		Metrics: []Metric{{Field: "x", Aggregation: "DROP"}},
	}
	_, _, err := BuildSQL("t", "mysql", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported aggregation")
}

func TestBuildSQL_InvalidOperator(t *testing.T) {
	config := ChartQueryConfig{
		Filters: []Filter{{Field: "x", Operator: "DROP", Value: "1"}},
	}
	_, _, err := BuildSQL("t", "mysql", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported operator")
}

func TestBuildSQL_EmptyDimensionsAndMetrics(t *testing.T) {
	config := ChartQueryConfig{}
	_, _, err := BuildSQL("orders", "mysql", config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "at least one dimension or metric required")
}

func TestBuildSQL_DefaultMetricAlias(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "SUM"}},
	}
	sql, _, err := BuildSQL("orders", "mysql", config)
	require.NoError(t, err)
	assert.Contains(t, sql, "SUM(`amount`) AS `sum_amount`")
}

func TestBuildSQL_InvalidOrderDirection(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "SUM"}},
		Orders:     []Order{{Field: "amount", Direction: "invalid"}},
	}
	sql, _, err := BuildSQL("orders", "mysql", config)
	require.NoError(t, err)
	assert.Contains(t, sql, "ORDER BY `amount` asc")
}

func TestBuildSQL_PostgreSQL(t *testing.T) {
	config := ChartQueryConfig{
		Dimensions: []Dimension{{Field: "dept"}},
		Metrics:    []Metric{{Field: "amount", Aggregation: "SUM", Alias: "total"}},
		Filters:    []Filter{{Field: "status", Operator: "=", Value: "1"}},
		Orders:     []Order{{Field: "amount", Direction: "desc"}},
	}
	sql, _, err := BuildSQL("orders", "postgresql", config)
	require.NoError(t, err)
	assert.Equal(t, "SELECT \"dept\", SUM(\"amount\") AS \"total\" FROM \"orders\" WHERE \"status\" = ? GROUP BY \"dept\" ORDER BY \"amount\" desc", sql)
}

func TestQuoteIdentifier_Dialects(t *testing.T) {
	assert.Equal(t, "`field`", QuoteIdentifier("field", "mysql"))
	assert.Equal(t, "`field`", QuoteIdentifier("field", ""))
	assert.Equal(t, "`fi``eld`", QuoteIdentifier("fi`eld", "mysql"))
	assert.Equal(t, "\"field\"", QuoteIdentifier("field", "postgresql"))
	assert.Equal(t, "\"field\"", QuoteIdentifier("field", "postgres"))
	assert.Equal(t, "\"fi\"\"eld\"", QuoteIdentifier("fi\"eld", "postgresql"))
}
