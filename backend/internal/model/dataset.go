package model

import "time"

type Dataset struct {
	ID             uint64     `db:"id" json:"id"`
	Name           string     `db:"name" json:"name"`
	DatasourceID   uint64     `db:"datasource_id" json:"datasourceId"`
	DatabaseName   string     `db:"database_name" json:"databaseName"`
	TableName      string     `db:"table_name" json:"tableName"`
	Type           string     `db:"type" json:"type"`
	Mode           int8       `db:"mode" json:"mode"`
	Status         int8       `db:"status" json:"status"`
	CreatedBy      uint64     `db:"created_by" json:"createdBy"`
	CreatedAt      time.Time  `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time  `db:"updated_at" json:"updatedAt"`
	DeletedAt      *time.Time `db:"deleted_at" json:"-"`
}

type DatasetField struct {
	ID         uint64    `db:"id" json:"id"`
	DatasetID  uint64    `db:"dataset_id" json:"datasetId"`
	Name       string    `db:"name" json:"name"`
	Type       string    `db:"type" json:"type"`
	DeType     int8      `db:"de_type" json:"deType"`
	Length     int       `db:"length" json:"length"`
	Precision  int       `db:"precision" json:"precision"`
	Scale      int       `db:"scale" json:"scale"`
	OriginName string    `db:"origin_name" json:"originName"`
	CreatedAt  time.Time `db:"created_at" json:"createdAt"`
	UpdatedAt  time.Time `db:"updated_at" json:"updatedAt"`
}
