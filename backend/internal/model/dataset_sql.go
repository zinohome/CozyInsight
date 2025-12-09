package model

// DatasetTableUnion 数据集关联(SQL模式)
type DatasetTableUnion struct {
	ID               string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DatasetTableID   string `gorm:"type:varchar(50);not null" json:"datasetTableId"`
	ParentTableID    string `gorm:"type:varchar(50)" json:"parentTableId"`
	ParentFieldID    string `gorm:"type:varchar(50)" json:"parentFieldId"`
	CurrentTableID   string `gorm:"type:varchar(50)" json:"currentTableId"`
	CurrentFieldID   string `gorm:"type:varchar(50)" json:"currentFieldId"`
	UnionType        string `gorm:"type:varchar(50)" json:"unionType"` // left, right, inner, full
	CreateTime       int64  `gorm:"autoCreateTime:milli" json:"createTime"`
}

func (DatasetTableUnion) TableName() string {
	return "dataset_table_union"
}

// DatasetTableTask 数据集任务(抽取任务)
type DatasetTableTask struct {
	ID             string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	TableID        string `gorm:"type:varchar(50);not null;index" json:"tableId"`
	Type           string `gorm:"type:varchar(50)" json:"type"` // all_scope, incremental
	Status         string `gorm:"type:varchar(50)" json:"status"` // pending, running, success, failed
	StartTime      int64  `gorm:"default:0" json:"startTime"`
	EndTime        int64  `gorm:"default:0" json:"endTime"`
	Info           string `gorm:"type:text" json:"info"`
	LastExecStatus string `gorm:"type:varchar(50)" json:"lastExecStatus"`
	CreateTime     int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (DatasetTableTask) TableName() string {
	return "dataset_table_task"
}
