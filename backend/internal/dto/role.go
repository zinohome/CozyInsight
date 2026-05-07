package dto

// CreateRoleRequest 创建角色请求
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Status      int8   `json:"status"`
}

// UpdateRoleRequest 更新角色请求
type UpdateRoleRequest struct {
	Name        string `json:"name"`
	Code        string `json:"code"`
	Description string `json:"description"`
	Status      *int8  `json:"status"`
}

// SetRoleMenusRequest 设置角色菜单请求
type SetRoleMenusRequest struct {
	MenuIDs []uint64 `json:"menuIds"`
}

// SetUserRolesRequest 设置用户角色请求
type SetUserRolesRequest struct {
	RoleIDs []uint64 `json:"roleIds"`
}
