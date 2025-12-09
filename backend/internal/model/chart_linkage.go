package model

// ChartViewLinkage 图表联动配置
type ChartViewLinkage struct {
	ID                 string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	SourceViewID       string `gorm:"type:varchar(50);not null;index" json:"sourceViewId"` // 源图表
	TargetViewID       string `gorm:"type:varchar(50);not null;index" json:"targetViewId"` // 目标图表  
	SourceFieldID      string `gorm:"type:varchar(50)" json:"sourceFieldId"`
	TargetFieldID      string `gorm:"type:varchar(50)" json:"targetFieldId"`
	LinkageType        string `gorm:"type:varchar(50)" json:"linkageType"` // click, hover
	UpdateType         string `gorm:"type:varchar(50)" json:"updateType"` // replace, append
	Ext                string `gorm:"type:text" json:"ext"` // JSON扩展配置
	CreateTime         int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime         int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (ChartViewLinkage) TableName() string {
	return "chart_view_linkage"
}

// ChartViewDrill 图表钻取配置
type ChartViewDrill struct {
	ID           string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	ChartViewID  string `gorm:"type:varchar(50);not null;index" json:"chartViewId"`
	FieldID      string `gorm:"type:varchar(50)" json:"fieldId"`
	DrillFields  string `gorm:"type:text" json:"drillFields"` // JSON数组,钻取字段顺序
	DrillType    string `gorm:"type:varchar(50)" json:"drillType"` // down, up
	CreateTime   int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime   int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (ChartViewDrill) TableName() string {
	return "chart_view_drill"
}
