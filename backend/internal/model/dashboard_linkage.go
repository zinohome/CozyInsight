package model

// DashboardLinkage 仪表板联动配置
type DashboardLinkage struct {
	ID                string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DashboardID       string `gorm:"type:varchar(50);not null;index" json:"dashboardId"`
	SourceComponentID string `gorm:"type:varchar(50);not null" json:"sourceComponentId"` // 源组件
	TargetComponentID string `gorm:"type:varchar(50);not null" json:"targetComponentId"` // 目标组件
	SourceFieldID     string `gorm:"type:varchar(50)" json:"sourceFieldId"`
	TargetFieldID     string `gorm:"type:varchar(50)" json:"targetFieldId"`
	LinkageType       string `gorm:"type:varchar(50)" json:"linkageType"` // click, hover, select
	UpdateType        string `gorm:"type:varchar(50)" json:"updateType"`  // replace, append
	Enable            bool   `gorm:"default:true" json:"enable"`
	CreateTime        int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime        int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (DashboardLinkage) TableName() string {
	return "dashboard_linkage"
}

// DashboardParameter 仪表板参数
type DashboardParameter struct {
	ID            string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DashboardID   string `gorm:"type:varchar(50);not null;index" json:"dashboardId"`
	Name          string `gorm:"type:varchar(200)" json:"name"`
	ParamType     string `gorm:"type:varchar(50)" json:"paramType"` // text, date, select, multiselect
	DefaultValue  string `gorm:"type:varchar(500)" json:"defaultValue"`
	Options       string `gorm:"type:text" json:"options"` // JSON数组,下拉选项
	Required      bool   `gorm:"default:false" json:"required"`
	Enable        bool   `gorm:"default:true" json:"enable"`
	CreateTime    int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime    int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (DashboardParameter) TableName() string {
	return "dashboard_parameter"
}
