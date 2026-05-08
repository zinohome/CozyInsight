package dto

// CreateDatasetRequest 创建数据集请求
type CreateDatasetRequest struct {
	Name         string `json:"name" binding:"required"`
	DatasourceID uint64 `json:"datasourceId" binding:"required"`
	DatabaseName string `json:"databaseName"`
	TableName    string `json:"tableName" binding:"required"`
	SQL          string `json:"sql"`
	Type         string `json:"type" binding:"required"`
	Mode         int8   `json:"mode"`
}

// UpdateDatasetRequest 更新数据集请求
type UpdateDatasetRequest struct {
	Name         string  `json:"name"`
	DatasourceID *uint64 `json:"datasourceId"`
	DatabaseName string  `json:"databaseName"`
	TableName    string  `json:"tableName"`
	SQL          string  `json:"sql"`
	Type         string  `json:"type"`
	Mode         *int8   `json:"mode"`
	Status       *int8   `json:"status"`
}

// DatasetFieldResponse 字段响应
type DatasetFieldResponse struct {
	ID         uint64 `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	DeType     int8   `json:"deType"`
	Length     int    `json:"length"`
	Precision  int    `json:"precision"`
	Scale      int    `json:"scale"`
	OriginName string `json:"originName"`
}

// PreviewDataRequest 预览数据请求
type PreviewDataRequest struct {
	Limit uint64 `json:"limit"`
}

// PreviewDataResponse 预览数据响应
type PreviewDataResponse struct {
	Fields []DatasetFieldResponse     `json:"fields"`
	Data   []map[string]interface{} `json:"data"`
}
