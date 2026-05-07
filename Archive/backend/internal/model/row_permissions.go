package model

// DatasetRowPermissions 数据集行级权限
type DatasetRowPermissions struct {
	ID            string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	DatasetID     string `gorm:"type:varchar(50);not null;index" json:"datasetId"`
	AuthTargetType string `gorm:"type:varchar(50)" json:"authTargetType"` // user, role, dept
	AuthTargetID   string `gorm:"type:varchar(50);index" json:"authTargetId"`
	WhereCondition string `gorm:"type:text" json:"whereCondition"` // SQL WHERE条件
	ExpressType    string `gorm:"type:varchar(50)" json:"expressType"` // sql, formula
	Enable         bool   `gorm:"default:true" json:"enable"`
	CreateTime     int64  `gorm:"autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64  `gorm:"autoUpdateTime:milli" json:"updateTime"`
	CreateBy       string `gorm:"type:varchar(50)" json:"createBy"`
}

func (DatasetRowPermissions) TableName() string {
	return "dataset_row_permissions"
}

// RowPermissionsTree 行权限配置树
type RowPermissionsTree struct {
	ID             string `gorm:"primaryKey;type:varchar(50)" json:"id"`
	PermissionID   string `gorm:"type:varchar(50);not null;index" json:"permissionId"`
	EnableExpand   bool   `gorm:"default:false" json:"enableExpand"`
	TreeConfig     string `gorm:"type:text" json:"treeConfig"` // JSON配置
	CreateTime     int64  `gorm:"autoCreateTime:milli" json:"createTime"`
}

func (RowPermissionsTree) TableName() string {
	return "row_permissions_tree"
}
