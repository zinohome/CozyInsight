package model

// DashboardTab Tab容器配置
type DashboardTab struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DashboardID string `gorm:"type:varchar(50);not null;index" json:"dashboardId"`
	Name        string `gorm:"type:varchar(200)" json:"name"`
	Order       int    `gorm:"default:0" json:"order"`
	ComponentIDs string `gorm:"type:text" json:"componentIds"` // JSON数组,包含的组件IDs
	CreateTime  int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime  int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (DashboardTab) TableName() string {
	return "dashboard_tab"
}

// ChartTemplate 图表模板
type ChartTemplate struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name        string `gorm:"type:varchar(200)" json:"name"`
	Type        string `gorm:"type:varchar(50)" json:"type"` // 图表类型
	Category    string `gorm:"type:varchar(50)" json:"category"` // 分类: business, stat, custom
	Config      string `gorm:"type:text" json:"config"` // 模板配置JSON
	Preview     string `gorm:"type:varchar(500)" json:"preview"` // 预览图URL
	IsSystem    bool   `gorm:"default:false" json:"isSystem"` // 是否系统模板
	IsPublic    bool   `gorm:"default:false" json:"isPublic"` // 是否公开
	CreateTime  int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime  int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy    string `gorm:"type:varchar(50)" json:"createBy"`
}

func (ChartTemplate) TableName() string {
	return "chart_template"
}

// DatasetGroup 数据集分组(完善)
type DatasetGroupTree struct {
	ID          string               `json:"id"`
	Name        string               `json:"name"`
	Pid         string               `json:"pid"`
	Level       int                  `json:"level"`
	Type        string               `json:"type"`
	Children    []*DatasetGroupTree  `json:"children,omitempty"`
}
