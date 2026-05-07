package model

// Dashboard 仪表板模型
type Dashboard struct {
	ID              string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name            string `gorm:"type:varchar(255);not null" json:"name"`
	PID             string `gorm:"type:varchar(50);default:'0'" json:"pid"`   // 父ID，0表示根节点
	NodeType        string `gorm:"type:varchar(50);not null" json:"nodeType"` // folder | dashboard
	Type            string `gorm:"type:varchar(50)" json:"type"`              // dashboard | dataV (仪表板 | 数据大屏)
	CanvasStyleData string `gorm:"type:longtext" json:"canvasStyleData"`      // 画布样式 JSON
	ComponentData   string `gorm:"type:longtext" json:"componentData"`        // 组件数据 JSON
	Status          int    `gorm:"default:0" json:"status"`                   // 0=未发布 1=已发布
	PublishTime     int64  `gorm:"default:0" json:"publishTime"`              // 发布时间
	Sort            int    `gorm:"default:0" json:"sort"`                     // 排序
	CreateTime      int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime      int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy        string `gorm:"type:varchar(50)" json:"createBy"`
}

func (Dashboard) TableName() string {
	return "dashboard"
}
