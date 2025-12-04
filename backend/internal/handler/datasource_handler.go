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

// Validate 测试数据源连接
func (h *DatasourceHandler) Validate(c *gin.Context) {
	id := c.Param("id")
	err := h.svc.Validate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"status": "error", "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "连接成功"})
}

// GetTables 获取数据源的表列表
func (h *DatasourceHandler) GetTables(c *gin.Context) {
	id := c.Param("id")
	tables, err := h.svc.GetTables(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tables)
}

// GetFields 获取表的字段列表
func (h *DatasourceHandler) GetFields(c *gin.Context) {
	id := c.Param("id")
	tableName := c.Query("table")

	if tableName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "table name is required"})
		return
	}

	fields, err := h.svc.GetTableFields(c.Request.Context(), id, tableName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, fields)
}
