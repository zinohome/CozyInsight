package model

// DashboardComponent 仪表板组件
type DashboardComponent struct {
	ID          string                 `json:"id"`
	DashboardID string                 `json:"dashboardId"`
	ChartID     string                 `json:"chartId"`
	Type        string                 `json:"type"` // chart, text, image, etc.
	X           int                    `json:"x"`
	Y           int                    `json:"y"`
	W           int                    `json:"w"` // width
	H           int                    `json:"h"` // height
	Config      map[string]interface{} `json:"config,omitempty"`
	CreateTime  int64                  `json:"createTime"`
	UpdateTime  int64                  `json:"updateTime"`
}

func (DashboardComponent) TableName() string {
	return "core_dashboard_component"
}

// DashboardLayout 仪表板布局配置
type DashboardLayout struct {
	Components []DashboardComponent `json:"components"`
	Config     LayoutConfig         `json:"config"`
}

// LayoutConfig 布局配置
type LayoutConfig struct {
	BackgroundColor string `json:"backgroundColor,omitempty"`
	Padding         int    `json:"padding,omitempty"`
	Cols            int    `json:"cols,omitempty"`     // 网格列数
	RowHeight       int    `json:"rowHeight,omitempty"` // 行高
}
