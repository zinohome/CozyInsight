package model

// Share 分享模型
type Share struct {
	ID           string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	ResourceType string `gorm:"type:varchar(50);not null" json:"resourceType"` // dashboard, chart
	ResourceID   string `gorm:"type:varchar(50);not null" json:"resourceId"`
	Token        string `gorm:"type:varchar(50);uniqueIndex" json:"token"`
	Password     string `gorm:"type:varchar(100)" json:"password,omitempty"`
	ExpireTime   int64  `gorm:"default:0" json:"expireTime"` // 过期时间,0表示永不过期
	CreateTime   int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	CreateBy     string `gorm:"type:varchar(50)" json:"createBy"`
}

func (Share) TableName() string {
	return "sys_share"
}
