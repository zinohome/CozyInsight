package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Email    string `json:"email" binding:"required,email"`
	NickName string `json:"nickName"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   uint64 `json:"userId"`
	Username string `json:"username"`
	NickName string `json:"nickName"`
	IsAdmin  bool   `json:"isAdmin"`
}

type UserInfoResponse struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	NickName string `json:"nickName"`
	Avatar   string `json:"avatar"`
	IsAdmin  bool   `json:"isAdmin"`
}
