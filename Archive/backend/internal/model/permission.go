package model

// Role 角色模型
type Role struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name        string `gorm:"type:varchar(100);not null;uniqueIndex" json:"name"`
	Description string `gorm:"type:varchar(500)" json:"description"`
	Type        string `gorm:"type:varchar(50)" json:"type"` // system, custom
	CreateTime  int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime  int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy    string `gorm:"type:varchar(50)" json:"createBy"`
}

func (Role) TableName() string {
	return "sys_role"
}

// Permission 权限模型
type Permission struct {
	ID          string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	Name        string `gorm:"type:varchar(100);not null" json:"name"`
	Resource    string `gorm:"type:varchar(100)" json:"resource"`    // datasource, dataset, chart, dashboard
	ResourceID  string `gorm:"type:varchar(50)" json:"resourceId"`   // 资源ID
	Action      string `gorm:"type:varchar(50)" json:"action"`       // read, write, delete, manage
	Description string `gorm:"type:varchar(500)" json:"description"`
	CreateTime  int64  `gorm:"autoCreateTime:milli" json:"createTime"`
}

func (Permission) TableName() string {
	return "sys_permission"
}

// RolePermission 角色权限关联
type RolePermission struct {
	ID           string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	RoleID       string `gorm:"type:varchar(50);not null;index" json:"roleId"`
	PermissionID string `gorm:"type:varchar(50);not null;index" json:"permissionId"`
	CreateTime   int64  `gorm:"autoCreateTime:milli" json:"createTime"`
}

func (RolePermission) TableName() string {
	return "sys_role_permission"
}

// UserRole 用户角色关联
type UserRole struct {
	ID         string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	UserID     string `gorm:"type:varchar(50);not null;index" json:"userId"`
	RoleID     string `gorm:"type:varchar(50);not null;index" json:"roleId"`
	CreateTime int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	CreateBy   string `gorm:"type:varchar(50)" json:"createBy"`
}

func (UserRole) TableName() string {
	return "sys_user_role"
}

// ResourcePermission 资源权限配置
type ResourcePermission struct {
	ID         string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	ResourceType string `gorm:"type:varchar(50);not null" json:"resourceType"` // datasource, dataset, chart, dashboard
	ResourceID string `gorm:"type:varchar(50);not null;index" json:"resourceId"`
	TargetType string `gorm:"type:varchar(20);not null" json:"targetType"` // user, role
	TargetID   string `gorm:"type:varchar(50);not null;index" json:"targetId"`
	Permission string `gorm:"type:varchar(20);not null" json:"permission"` // read, write, manage
	CreateTime int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	CreateBy   string `gorm:"type:varchar(50)" json:"createBy"`
}

func (ResourcePermission) TableName() string {
	return "sys_resource_permission"
}
