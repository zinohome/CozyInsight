package model

import "time"

type Chart struct {
	ID          uint64     `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Type        string     `db:"type" json:"type"`
	DatasetID   uint64     `db:"dataset_id" json:"datasetId"`
	Config      string     `db:"config" json:"config"`
	Status      int8       `db:"status" json:"status"`
	CreatedBy   uint64     `db:"created_by" json:"createdBy"`
	CreatedAt   time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt   *time.Time `db:"deleted_at" json:"-"`
}
