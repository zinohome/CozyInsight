package dto

// CreateDashboardRequest 创建仪表板请求
type CreateDashboardRequest struct {
	Title  string `json:"title" binding:"required"`
	Config string `json:"config"`
}

// UpdateDashboardRequest 更新仪表板请求
type UpdateDashboardRequest struct {
	Title  string `json:"title"`
	Config string `json:"config"`
	Status int8   `json:"status"`
}

// AddChartToDashboardRequest 添加图表到仪表板请求
type AddChartToDashboardRequest struct {
	ChartID   uint64 `json:"chartId" binding:"required"`
	PositionX int    `json:"positionX"`
	PositionY int    `json:"positionY"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Config    string `json:"config"`
}
