package model

type Datasource struct {
	ID            string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name          string `gorm:"type:varchar(255);not null" json:"name"`
	Description   string `gorm:"type:varchar(255)" json:"description"`
	Type          string `gorm:"type:varchar(50);not null" json:"type"`
	PID           string `gorm:"column:pid;type:varchar(50)" json:"pid"`
	EditType      int    `gorm:"default:0" json:"editType"` // 0: replace, 1: append
	Configuration string `gorm:"type:longtext" json:"configuration"`
	Status        string `gorm:"type:varchar(50)" json:"status"`
	CreateTime    int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime    int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy      string `gorm:"type:varchar(50)" json:"createBy"`
	QrtzInstance  string `gorm:"type:varchar(50)" json:"qrtzInstance"`
	TaskStatus    string `gorm:"type:varchar(50)" json:"taskStatus"`
}

func (Datasource) TableName() string {
	return "core_datasource"
}
