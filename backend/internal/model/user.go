package model

import "time"

type User struct {
	ID           uint64     `db:"id" json:"id"`
	Username     string     `db:"username" json:"username"`
	PasswordHash string     `db:"password_hash" json:"-"`
	Email        string     `db:"email" json:"email"`
	NickName     string     `db:"nick_name" json:"nickName"`
	Avatar       string     `db:"avatar" json:"avatar"`
	Phone        string     `db:"phone" json:"phone"`
	Status       int8       `db:"status" json:"status"`
	IsAdmin      int8       `db:"is_admin" json:"isAdmin"`
	LastLoginAt  *time.Time `db:"last_login_at" json:"lastLoginAt"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"-"`
}
