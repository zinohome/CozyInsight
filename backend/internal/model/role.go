package model

import "time"

type Role struct {
	ID          uint64     `db:"id" json:"id"`
	Name        string     `db:"name" json:"name"`
	Code        string     `db:"code" json:"code"`
	Description string     `db:"description" json:"description"`
	Status      int8       `db:"status" json:"status"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deleted_at" json:"-"`
}

type Menu struct {
	ID        uint64    `db:"id" json:"id"`
	ParentID  uint64    `db:"parent_id" json:"parentId"`
	Name      string    `db:"name" json:"name"`
	Path      string    `db:"path" json:"path"`
	Component string    `db:"component" json:"component"`
	Icon      string    `db:"icon" json:"icon"`
	Sort      int       `db:"sort" json:"sort"`
	Status    int8      `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}

type RoleMenu struct {
	RoleID uint64 `db:"role_id" json:"roleId"`
	MenuID uint64 `db:"menu_id" json:"menuId"`
}

type UserRole struct {
	UserID uint64 `db:"user_id" json:"userId"`
	RoleID uint64 `db:"role_id" json:"roleId"`
}
