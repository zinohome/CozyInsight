package model

import "time"

type OperationLog struct {
	ID           uint64    `db:"id" json:"id"`
	UserID       uint64    `db:"user_id" json:"userId"`
	Username     string    `db:"username" json:"username"`
	Method       string    `db:"method" json:"method"`
	Path         string    `db:"path" json:"path"`
	Query        string    `db:"query" json:"query"`
	Body         string    `db:"body" json:"body"`
	IP           string    `db:"ip" json:"ip"`
	UserAgent    string    `db:"user_agent" json:"userAgent"`
	StatusCode   int       `db:"status_code" json:"statusCode"`
	Duration     int64     `db:"duration" json:"duration"`
	ErrorMessage string    `db:"error_message" json:"errorMessage"`
	CreatedAt    time.Time `db:"created_at" json:"createdAt"`
}
