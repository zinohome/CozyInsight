package model

// ChartView 图表视图
type ChartView struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name        string `gorm:"type:varchar(255);not null" json:"name"`
	SceneID     string `gorm:"type:varchar(50)" json:"sceneId"` // 分组ID
	TableID     string `gorm:"type:varchar(50)" json:"tableId"` // 数据集ID
	Type        string `gorm:"type:varchar(50)" json:"type"`    // 图表类型: bar, line, pie...
	Title       string `gorm:"type:varchar(255)" json:"title"`
	XAxis       string `gorm:"type:longtext" json:"xAxis"`       // JSON: 维度配置
	YAxis       string `gorm:"type:longtext" json:"yAxis"`       // JSON: 指标配置
	CustomAttr  string `gorm:"type:longtext" json:"customAttr"`  // JSON: 图表特有属性
	CustomStyle string `gorm:"type:longtext" json:"customStyle"` // JSON: 样式配置
	Snapshot    string `gorm:"type:longtext" json:"snapshot"`    // 快照/缩略图
	CreateTime  int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime  int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy    string `gorm:"type:varchar(50)" json:"createBy"`
}

func (ChartView) TableName() string {
	return "core_chart_view"
}
