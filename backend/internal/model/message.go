package model

import "time"

type Message struct {
	ID        uint64    `db:"id" json:"id"`
	UserID    uint64    `db:"user_id" json:"userId"`
	Title     string    `db:"title" json:"title"`
	Content   string    `db:"content" json:"content"`
	Type      string    `db:"type" json:"type"`
	IsRead    int8      `db:"is_read" json:"isRead"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
}
