package model

type User struct {
	ID         string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Username   string `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"`
	Email      string `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"`
	Password   string `gorm:"type:varchar(255);not null" json:"-"`         // json:"-" 不输出到 JSON
	Role       string `gorm:"type:varchar(20);default:'user'" json:"role"` // admin | user
	CreateTime int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
}

func (User) TableName() string {
	return "user"
}
