package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type DatasetHandler struct {
	service *service.DatasetService
}

func NewDatasetHandler(service *service.DatasetService) *DatasetHandler {
	return &DatasetHandler{service: service}
}

func (h *DatasetHandler) Create(c *gin.Context) {
	var req dto.CreateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	userID := middleware.GetUserID(c)

	ds, err := h.service.Create(c.Request.Context(), &req, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasetHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	ds, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasetHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DatasetHandler) Update(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	var req dto.UpdateDatasetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.Update(c.Request.Context(), id, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasetHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasetHandler) SyncFields(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	if err := h.service.SyncFields(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasetHandler) Preview(c *gin.Context) {
	id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
	limit, _ := strconv.ParseUint(c.Query("limit"), 10, 64)
	if limit == 0 {
		limit = 100
	}

	resp, err := h.service.PreviewData(c.Request.Context(), id, limit)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": resp})
}
