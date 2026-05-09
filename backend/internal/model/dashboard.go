package model

import "time"

const (
	DashboardTypeDashboard = "dashboard"
	DashboardTypeScreen    = "screen"
)

type Dashboard struct {
	ID           uint64     `db:"id" json:"id"`
	Title        string     `db:"title" json:"title"`
	Type         string     `db:"type" json:"type"`
	Config       string     `db:"config" json:"config"`
	ShareToken   string     `db:"share_token" json:"shareToken"`
	ShareEnabled int8       `db:"share_enabled" json:"shareEnabled"`
	Status       int8       `db:"status" json:"status"`
	CreatedBy    uint64     `db:"created_by" json:"createdBy"`
	CreatedAt    time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt    *time.Time `db:"deleted_at" json:"-"`
}

type DashboardChart struct {
	ID          uint64    `db:"id" json:"id"`
	DashboardID uint64    `db:"dashboard_id" json:"dashboardId"`
	ChartID     uint64    `db:"chart_id" json:"chartId"`
	PositionX   int       `db:"position_x" json:"positionX"`
	PositionY   int       `db:"position_y" json:"positionY"`
	Width       int       `db:"width" json:"width"`
	Height      int       `db:"height" json:"height"`
	Config      string    `db:"config" json:"config"`
	CreatedAt   time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time `db:"updated_at" json:"updatedAt"`
}
