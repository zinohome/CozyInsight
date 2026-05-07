package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestValidateAggregate(t *testing.T) {
	tests := []struct {
		name     string
		agg      string
		expected bool
	}{
		{"Valid SUM", "SUM", true},
		{"Valid AVG", "AVG", true},
		{"Valid COUNT", "COUNT", true},
		{"Valid MAX", "MAX", true},
		{"Valid MIN", "MIN", true},
		{"Invalid TOTAL", "TOTAL", false},
		{"Invalid empty", "", false},
		{"Invalid lowercase", "sum", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateAggregate(tt.agg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildAggregateSQL(t *testing.T) {
	tests := []struct {
		name     string
		field    FieldConfig
		expected string
	}{
		{
			name:     "No aggregate",
			field:    FieldConfig{Name: "category"},
			expected: "category",
		},
		{
			name:     "SUM aggregate",
			field:    FieldConfig{Name: "amount", Aggregate: "SUM"},
			expected: "SUM(amount) AS amount",
		},
		{
			name:     "AVG aggregate",
			field:    FieldConfig{Name: "price", Aggregate: "AVG"},
			expected: "AVG(price) AS price",
		},
		{
			name:     "COUNT aggregate",
			field:    FieldConfig{Name: "id", Aggregate: "COUNT"},
			expected: "COUNT(id) AS id",
		},
		{
			name:     "COUNT(*) special case",
			field:    FieldConfig{Name: "*", Aggregate: "COUNT"},
			expected: "COUNT(*) AS count",
		},
		{
			name:     "Invalid aggregate",
			field:    FieldConfig{Name: "value", Aggregate: "INVALID"},
			expected: "value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildAggregateSQL(tt.field)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildSortClause(t *testing.T) {
	tests := []struct {
		name     string
		fields   []FieldConfig
		expected string
	}{
		{
			name:     "No sort",
			fields:   []FieldConfig{{Name: "category"}},
			expected: "",
		},
		{
			name: "Single sort ASC",
			fields: []FieldConfig{
				{Name: "amount", Sort: "ASC"},
			},
			expected: " ORDER BY amount ASC",
		},
		{
			name: "Single sort DESC",
			fields: []FieldConfig{
				{Name: "price", Sort: "DESC"},
			},
			expected: " ORDER BY price DESC",
		},
		{
			name: "Multiple sorts",
			fields: []FieldConfig{
				{Name: "category", Sort: "ASC"},
				{Name: "amount", Sort: "DESC"},
			},
			expected: " ORDER BY category ASC, amount DESC",
		},
		{
			name: "Invalid sort direction",
			fields: []FieldConfig{
				{Name: "value", Sort: "RANDOM"},
			},
			expected: " ORDER BY value ASC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildSortClause(tt.fields)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestChartQueryConfig_GetDimensions(t *testing.T) {
	config := &ChartQueryConfig{
		XAxis: AxisConfig{
			Fields: []FieldConfig{
				{Name: "category"},
				{Name: "region"},
			},
		},
	}

	dimensions := config.GetDimensions()
	assert.Equal(t, 2, len(dimensions))
	assert.Contains(t, dimensions, "category")
	assert.Contains(t, dimensions, "region")
}

func TestChartQueryConfig_HasAggregates(t *testing.T) {
	tests := []struct {
		name     string
		config   *ChartQueryConfig
		expected bool
	}{
		{
			name: "Has aggregates",
			config: &ChartQueryConfig{
				YAxis: AxisConfig{
					Fields: []FieldConfig{
						{Name: "amount", Aggregate: "SUM"},
					},
				},
			},
			expected: true,
		},
		{
			name: "No aggregates",
			config: &ChartQueryConfig{
				YAxis: AxisConfig{
					Fields: []FieldConfig{
						{Name: "value"},
					},
				},
			},
			expected: false,
		},
		{
			name: "Empty fields",
			config: &ChartQueryConfig{
				YAxis: AxisConfig{
					Fields: []FieldConfig{},
				},
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.HasAggregates()
			assert.Equal(t, tt.expected, result)
		})
	}
}
