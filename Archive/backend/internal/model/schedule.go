package model

// ScheduleTask 定时任务模型
type ScheduleTask struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name        string `gorm:"type:varchar(200);not null" json:"name"`
	Type        string `gorm:"type:varchar(50)" json:"type"` // email_report, snapshot, data_sync
	CronExpr    string `gorm:"type:varchar(100)" json:"cronExpr"`
	Enabled     bool   `gorm:"default:false" json:"enabled"`
	Status      string `gorm:"type:varchar(20)" json:"status"` // active, inactive, running
	Config      string `gorm:"type:text" json:"config"`        // JSON配置
	LastRunTime int64  `gorm:"default:0" json:"lastRunTime"`
	CreateTime  int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime  int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy    string `gorm:"type:varchar(50)" json:"createBy"`
}

func (ScheduleTask) TableName() string {
	return "sys_schedule_task"
}
