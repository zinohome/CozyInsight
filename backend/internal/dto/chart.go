package dto

// ChartDimension 维度配置
type ChartDimension struct {
	Field string `json:"field"`
	Sort  string `json:"sort"`
}

// ChartMetric 指标配置
type ChartMetric struct {
	Field       string `json:"field"`
	Aggregation string `json:"aggregation"`
	Alias       string `json:"alias"`
}

// ChartFilter 过滤条件
type ChartFilter struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// ChartOrder 排序配置
type ChartOrder struct {
	Field     string `json:"field"`
	Direction string `json:"direction"`
}

// ChartConfig 图表完整配置
type ChartConfig struct {
	Dimensions []ChartDimension `json:"dimensions"`
	Metrics    []ChartMetric    `json:"metrics"`
	Filters    []ChartFilter    `json:"filters"`
	Orders     []ChartOrder     `json:"orders"`
	Limit      uint64           `json:"limit"`
}

// ChartDataResponse 图表数据响应
type ChartDataResponse struct {
	Dimensions []string                   `json:"dimensions"`
	Metrics    []string                   `json:"metrics"`
	Data       []map[string]interface{} `json:"data"`
}

type CreateChartRequest struct {
	Title     string `json:"title" binding:"required"`
	Type      string `json:"type" binding:"required"`
	DatasetID uint64 `json:"datasetId" binding:"required"`
	Config    string `json:"config"`
}

// UpdateChartRequest 更新图表请求
type UpdateChartRequest struct {
	Title     string  `json:"title"`
	Type      string  `json:"type"`
	DatasetID *uint64 `json:"datasetId"`
	Config    string  `json:"config"`
	Status    *int8   `json:"status"`
}
