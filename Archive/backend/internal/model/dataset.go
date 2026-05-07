package model

// DatasetGroup 数据集分组 (文件夹)
type DatasetGroup struct {
	ID         string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name       string `gorm:"type:varchar(255);not null" json:"name"`
	PID        string `gorm:"column:pid;type:varchar(50)" json:"pid"`
	Level      int    `gorm:"type:int" json:"level"`
	NodeType   string `gorm:"type:varchar(50)" json:"nodeType"` // folder, dataset
	Type       string `gorm:"type:varchar(50)" json:"type"`     // sql, db, etc
	CreateTime int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	CreateBy   string `gorm:"type:varchar(50)" json:"createBy"`
}

func (DatasetGroup) TableName() string {
	return "core_dataset_group"
}

// DatasetTable 数据集表 (实际的数据集定义)
type DatasetTable struct {
	ID                string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name              string `gorm:"type:varchar(255);not null" json:"name"`
	PhysicalTableName string `gorm:"column:table_name;type:varchar(255)" json:"tableName"` // 物理表名或生成的表名
	DatasourceID      string `gorm:"type:varchar(50)" json:"datasourceId"`
	DatasetGroupID    string `gorm:"type:varchar(50)" json:"datasetGroupId"`
	Type              string `gorm:"type:varchar(50)" json:"type"` // db, sql
	Info              string `gorm:"type:longtext" json:"info"`    // JSON配置: SQL语句等
	CreateTime        int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime        int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy          string `gorm:"type:varchar(50)" json:"createBy"`
}

func (DatasetTable) TableName() string {
	return "core_dataset_table"
}

// DatasetTableField 数据集字段
type DatasetTableField struct {
	ID             string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DatasourceID   string `gorm:"type:varchar(50)" json:"datasourceId"`
	DatasetTableID string `gorm:"type:varchar(50)" json:"datasetTableId"`
	DatasetGroupID string `gorm:"type:varchar(50)" json:"datasetGroupId"`
	OriginName     string `gorm:"type:varchar(255)" json:"originName"`
	Name           string `gorm:"type:varchar(255)" json:"name"`
	Type           string `gorm:"type:varchar(50)" json:"type"`      // 原始类型
	DeType         int    `gorm:"type:int" json:"deType"`            // 0:txt, 1:time, 2:int, 3:float, 4:bool, 5:geo
	GroupType      string `gorm:"type:varchar(50)" json:"groupType"` // d: dimension, q: measure
	Checked        bool   `gorm:"type:boolean" json:"checked"`
	ColumnIndex    int    `gorm:"type:int" json:"columnIndex"`
}

func (DatasetTableField) TableName() string {
	return "core_dataset_table_field"
}
