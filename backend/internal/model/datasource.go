package model

import "time"

type Datasource struct {
	ID         uint64     `db:"id" json:"id"`
	Name       string     `db:"name" json:"name"`
	Type       string     `db:"type" json:"type"`
	Config     string     `db:"config" json:"config"`
	FilePath   string     `db:"file_path" json:"filePath"`
	FileType   string     `db:"file_type" json:"fileType"`
	Status     int8       `db:"status" json:"status"`
	CreatedBy  uint64     `db:"created_by" json:"createdBy"`
	CreatedAt  time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt  *time.Time `db:"deleted_at" json:"-"`
}
