package model

// SysOperLog 操作日志
type SysOperLog struct {
	ID         string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	UserID     string `gorm:"type:varchar(50);index" json:"userId"`
	Username   string `gorm:"type:varchar(100)" json:"username"`
	Module     string `gorm:"type:varchar(100);index" json:"module"` // datasource, dataset, chart, dashboard
	Action     string `gorm:"type:varchar(50)" json:"action"` // create, update, delete, view, export, share
	Detail     string `gorm:"type:text" json:"detail"` // 操作详情JSON
	ResourceID string `gorm:"type:varchar(50);index" json:"resourceId"` // 资源ID
	IP         string `gorm:"type:varchar(50)" json:"ip"`
	UserAgent  string `gorm:"type:varchar(255)" json:"userAgent"`
	Status     int    `gorm:"default:1" json:"status"` // 1=成功 0=失败
	ErrorMsg   string `gorm:"type:text" json:"errorMsg"` // 错误信息
	CreateTime int64  `gorm:"autoCreateTime:milli;index" json:"createTime"`
}

func (SysOperLog) TableName() string {
	return "sys_oper_log"
}

// SysSetting 系统设置
type SysSetting struct {
	ID         string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Type       string `gorm:"type:varchar(50);index" json:"type"` // email, auth, system, display
	SettingKey string `gorm:"type:varchar(100);uniqueIndex" json:"key"`
	Value      string `gorm:"type:text" json:"value"`
	UpdateTime int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	UpdateBy   string `gorm:"type:varchar(50)" json:"updateBy"`
}

func (SysSetting) TableName() string {
	return "sys_setting"
}

// DatasetTableFieldCalculated 数据集计算字段
type DatasetTableFieldCalculated struct {
	ID             string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DatasetTableID string `gorm:"type:varchar(50);not null;index" json:"datasetTableId"`
	FieldName      string `gorm:"type:varchar(255)" json:"fieldName"`
	DisplayName    string `gorm:"type:varchar(255)" json:"displayName"`
	Expression     string `gorm:"type:text" json:"expression"` // 计算公式
	DataType       string `gorm:"type:varchar(50)" json:"dataType"` // string, long, double, date
	CreateTime     int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (DatasetTableFieldCalculated) TableName() string {
	return "dataset_table_field_calculated"
}
