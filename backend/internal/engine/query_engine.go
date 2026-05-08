package engine

import (
	"fmt"
	"strings"
)

// ChartQueryConfig describes what a chart needs from the data layer.
type ChartQueryConfig struct {
	Dimensions []Dimension `json:"dimensions"`
	Metrics    []Metric    `json:"metrics"`
	Filters    []Filter    `json:"filters"`
	Orders     []Order     `json:"orders"`
	Limit      uint64      `json:"limit"`
}

// Dimension represents a grouping column.
type Dimension struct {
	Field string `json:"field"`
	Sort  string `json:"sort"`
}

// Metric represents an aggregated value.
type Metric struct {
	Field       string `json:"field"`
	Aggregation string `json:"aggregation"`
	Alias       string `json:"alias"`
}

// Filter represents a WHERE condition.
type Filter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Order represents a sort specification.
type Order struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

var allowedAggregations = map[string]bool{
	"SUM":   true,
	"COUNT": true,
	"AVG":   true,
	"MAX":   true,
	"MIN":   true,
}

var allowedOperators = map[string]bool{
	"=":        true,
	"!=":       true,
	">":        true,
	"<":        true,
	">=":       true,
	"<=":       true,
	"LIKE":     true,
	"NOT LIKE": true,
	"IN":       true,
	"NOT IN":   true,
}

// BuildSQL generates a parameterized SELECT SQL and its bound arguments.
func BuildSQL(tableName string, dialect string, config ChartQueryConfig) (string, []interface{}, error) {
	var selectCols []string
	var groupByCols []string

	for _, d := range config.Dimensions {
		selectCols = append(selectCols, QuoteIdentifier(d.Field, dialect))
		groupByCols = append(groupByCols, QuoteIdentifier(d.Field, dialect))
	}

	for _, m := range config.Metrics {
		agg := strings.ToUpper(m.Aggregation)
		if !allowedAggregations[agg] {
			return "", nil, fmt.Errorf("unsupported aggregation: %s", m.Aggregation)
		}
		alias := m.Alias
		if alias == "" {
			alias = fmt.Sprintf("%s_%s", strings.ToLower(agg), m.Field)
		}
		selectCols = append(selectCols,
			fmt.Sprintf("%s(%s) AS %s", agg, QuoteIdentifier(m.Field, dialect), QuoteIdentifier(alias, dialect)))
	}

	// Validate filters early so that invalid operators are caught before
	// the "at least one dimension or metric" check.
	var args []interface{}
	var conditions []string
	for _, f := range config.Filters {
		op := strings.ToUpper(f.Operator)
		if !allowedOperators[op] {
			return "", nil, fmt.Errorf("unsupported operator: %s", f.Operator)
		}
		conditions = append(conditions,
			fmt.Sprintf("%s %s ?", QuoteIdentifier(f.Field, dialect), op))
		args = append(args, f.Value)
	}

	if len(selectCols) == 0 {
		return "", nil, fmt.Errorf("at least one dimension or metric required")
	}

	query := fmt.Sprintf("SELECT %s FROM %s",
		strings.Join(selectCols, ", "),
		QuoteIdentifier(tableName, dialect))

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if len(groupByCols) > 0 {
		query += " GROUP BY " + strings.Join(groupByCols, ", ")
	}

	if len(config.Orders) > 0 {
		var orderParts []string
		for _, o := range config.Orders {
			dir := strings.ToLower(o.Direction)
			if dir != "asc" && dir != "desc" {
				dir = "asc"
			}
			orderParts = append(orderParts,
				fmt.Sprintf("%s %s", QuoteIdentifier(o.Field, dialect), dir))
		}
		query += " ORDER BY " + strings.Join(orderParts, ", ")
	}

	if config.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, config.Limit)
	}

	return query, args, nil
}

// QuoteIdentifier wraps a name in database-specific identifier quotes.
func QuoteIdentifier(name string, dialect string) string {
	switch dialect {
	case "postgresql", "postgres":
		return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
	case "sqlite":
		return "\"" + strings.ReplaceAll(name, "\"", "\"\"") + "\""
	default:
		// MySQL and others
		return "`" + strings.ReplaceAll(name, "`", "``") + "`"
	}
}
