package service

import (
"fmt"
"strings"
)

// buildWhereConditions 构建 WHERE 条件
func (s *chartDataService) buildWhereConditions(filters []FilterCondition) []string {
	var conditions []string

	for _, f := range filters {
		switch f.Operator {
		case "=", "!=", ">", "<", ">=", "<=":
			conditions = append(conditions, fmt.Sprintf("%s %s '%v'", f.Field, f.Operator, f.Value))
		case "LIKE":
			conditions = append(conditions, fmt.Sprintf("%s LIKE '%%%v%%'", f.Field, f.Value))
		case "IN":
			if values, ok := f.Value.([]interface{}); ok {
				var strValues []string
				for _, v := range values {
					strValues = append(strValues, fmt.Sprintf("'%v'", v))
				}
				conditions = append(conditions, fmt.Sprintf("%s IN (%s)", f.Field, strings.Join(strValues, ", ")))
			}
		}
	}

	return conditions
}
