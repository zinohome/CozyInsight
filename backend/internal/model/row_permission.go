package model

import "time"

type RowPermission struct {
	ID        uint64    `db:"id" json:"id"`
	DatasetID uint64    `db:"dataset_id" json:"datasetId"`
	FieldName string    `db:"field_name" json:"fieldName"`
	Operator  string    `db:"operator" json:"operator"`
	Value     string    `db:"value" json:"value"`
	UserAttr  string    `db:"user_attr" json:"userAttr"`
	Status    int8      `db:"status" json:"status"`
	CreatedAt time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt time.Time `db:"updated_at" json:"updatedAt"`
}
