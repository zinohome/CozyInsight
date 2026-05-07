package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatasourceHandler struct {
	svc service.DatasourceService
}

func NewDatasourceHandler(svc service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{svc: svc}
}

func (h *DatasourceHandler) Create(c *gin.Context) {
	var ds model.Datasource
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Create(c.Request.Context(), &ds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ds)
}

func (h *DatasourceHandler) Update(c *gin.Context) {
	var ds model.Datasource
	if err := c.ShouldBindJSON(&ds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Update(c.Request.Context(), &ds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ds)
}

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success"})
}

func (h *DatasourceHandler) Get(c *gin.Context) {
	id := c.Param("id")
	ds, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, ds)
}

func (h *DatasourceHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

// TestConnection 测试数据源连接（使用已保存的数据源 ID）
func (h *DatasourceHandler) TestConnection(c *gin.Context) {
	id := c.Param("id")
	result, err := h.svc.TestConnection(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// TestConnectionByConfig 测试数据源连接（使用配置 JSON，无需保存）
func (h *DatasourceHandler) TestConnectionByConfig(c *gin.Context) {
	type TestRequest struct {
		Configuration string `json:"configuration"`
	}

	var req TestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.TestConnectionByConfig(c.Request.Context(), req.Configuration)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDatabases 获取数据源的数据库列表
func (h *DatasourceHandler) GetDatabases(c *gin.Context) {
	id := c.Param("id")
	databases, err := h.svc.GetDatabases(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"databases": databases})
}

// GetTables 获取数据源的表列表
func (h *DatasourceHandler) GetTables(c *gin.Context) {
	id := c.Param("id")
	database := c.Query("database")

	tables, err := h.svc.GetTables(c.Request.Context(), id, database)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tables": tables})
}

// GetTableSchema 获取表结构
func (h *DatasourceHandler) GetTableSchema(c *gin.Context) {
	id := c.Param("id")
	database := c.Query("database")
	table := c.Query("table")

	if table == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table name is required"})
		return
	}

	schema, err := h.svc.GetTableSchema(c.Request.Context(), id, database, table)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"schema": schema})
}
