package model

import "time"

type ShareLink struct {
	ID           uint64     `db:"id" json:"id"`
	Token        string     `db:"token" json:"token"`
	ResourceType string     `db:"resource_type" json:"resourceType"`
	ResourceID   uint64     `db:"resource_id" json:"resourceId"`
	CreatedBy    uint64     `db:"created_by" json:"createdBy"`
	ExpireAt     *time.Time `db:"expire_at" json:"expireAt"`
	Password     string     `db:"password" json:"password,omitempty"`
	Status       int8       `db:"status" json:"status"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
}
