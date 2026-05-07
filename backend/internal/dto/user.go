package dto

// CreateUserRequest 创建用户请求（管理员用）
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
	Status   int8   `json:"status"`
	IsAdmin  bool   `json:"isAdmin"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Email    string `json:"email"`
	NickName string `json:"nickName"`
	Phone    string `json:"phone"`
	Status   *int8  `json:"status"`
	IsAdmin  *bool  `json:"isAdmin"`
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}
