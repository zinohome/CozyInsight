package model

type DatasetTableField struct {
	ID           string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	TableID      string `gorm:"type:varchar(50);not null;index" json:"tableId"`
	Name         string `gorm:"type:varchar(255);not null" json:"name"`
	OriginName   string `gorm:"type:varchar(255)" json:"originName"`   // 原始字段名
	Type         string `gorm:"type:varchar(50);not null" json:"type"` // INTEGER, VARCHAR, DECIMAL, DATE, etc.
	DisplayName  string `gorm:"type:varchar(255)" json:"displayName"`  // 别名/显示名称
	GroupType    string `gorm:"type:varchar(20)" json:"groupType"`     // d=维度, q=指标
	IsCalculated bool   `gorm:"default:false" json:"isCalculated"`     // 是否计算字段
	Expression   string `gorm:"type:text" json:"expression"`           // 计算字段表达式
	Description  string `gorm:"type:varchar(500)" json:"description"`
	SortIndex    int    `gorm:"default:0" json:"sortIndex"` // 排序
	CreateTime   int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime   int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (DatasetTableField) TableName() string {
	return "dataset_table_field"
}
