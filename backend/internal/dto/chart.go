package dto

// CreateChartRequest 创建图表请求
type CreateChartRequest struct {
	Title     string `json:"title" binding:"required"`
	Type      string `json:"type" binding:"required"`
	DatasetID uint64 `json:"datasetId" binding:"required"`
	Config    string `json:"config"`
}

// UpdateChartRequest 更新图表请求
type UpdateChartRequest struct {
	Title     string `json:"title"`
	Type      string `json:"type"`
	DatasetID uint64 `json:"datasetId"`
	Config    string `json:"config"`
	Status    int8   `json:"status"`
}
