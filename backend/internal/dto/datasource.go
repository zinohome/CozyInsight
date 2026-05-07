package dto

// CreateDatasourceRequest 创建数据源请求
type CreateDatasourceRequest struct {
	Name   string `json:"name" binding:"required"`
	Type   string `json:"type" binding:"required"`
	Config string `json:"config" binding:"required"`
}

// UpdateDatasourceRequest 更新数据源请求
type UpdateDatasourceRequest struct {
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config string `json:"config"`
	Status int8   `json:"status"`
}

// DatasourceResponse 数据源响应
type DatasourceResponse struct {
	ID        uint64 `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Config    string `json:"config"`
	Status    int8   `json:"status"`
	CreatedBy uint64 `json:"createdBy"`
	CreatedAt string `json:"createdAt"`
}

// TestConnectionRequest 测试连接请求
type TestConnectionRequest struct {
	Type   string                 `json:"type" binding:"required"`
	Config map[string]interface{} `json:"config" binding:"required"`
}
