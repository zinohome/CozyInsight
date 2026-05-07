package database

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"cozy-insight/pkg/config"
)

func TestNew_ConnectFail(t *testing.T) {
	cfg := config.DatabaseConfig{
		Driver:   "mysql",
		Host:     "127.0.0.1",
		Port:     13306, // wrong port to force failure
		Username: "root",
		Password: "wrong",
		Database: "test",
		Charset:  "utf8mb4",
		ParseTime: true,
		Loc:       "Local",
	}
	db, err := New(cfg)
	assert.Error(t, err)
	assert.Nil(t, db)
}
