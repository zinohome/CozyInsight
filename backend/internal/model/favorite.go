package model

import "time"

type Favorite struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"userId"`
	ResourceType string    `db:"resource_type" json:"resourceType"`
	ResourceID   uint64    `db:"resource_id" json:"resourceId"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}
