package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/dto"
	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type DatasourceHandler struct {
	service *service.DatasourceService
}

func NewDatasourceHandler(service *service.DatasourceService) *DatasourceHandler {
	return &DatasourceHandler{service: service}
}

func (h *DatasourceHandler) Create(c *gin.Context) {
	var req dto.CreateDatasourceRequest
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

func (h *DatasourceHandler) Get(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	ds, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}

func (h *DatasourceHandler) List(c *gin.Context) {
	list, err := h.service.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *DatasourceHandler) Update(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateDatasourceRequest
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

func (h *DatasourceHandler) Delete(c *gin.Context) {
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *DatasourceHandler) TestConnection(c *gin.Context) {
	var req dto.TestConnectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	if err := h.service.TestConnection(c.Request.Context(), &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "connection ok"})
}

func (h *DatasourceHandler) UploadFile(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "missing file"})
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext != ".xlsx" && ext != ".xls" && ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "unsupported file type, only .xlsx, .xls, .csv are allowed"})
		return
	}

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "create upload dir failed"})
		return
	}

	savePath := filepath.Join(uploadDir, fmt.Sprintf("%d%s", header.Size, ext))
	if err := c.SaveUploadedFile(header, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "save file failed"})
		return
	}

	name := c.PostForm("name")
	if name == "" {
		name = header.Filename
	}

	fileType := "excel"
	if ext == ".csv" {
		fileType = "csv"
	}

	config := map[string]interface{}{
		"file_path": savePath,
		"file_type": fileType,
	}
	configJSON, _ := json.Marshal(config)

	userID := middleware.GetUserID(c)
	ds, err := h.service.Create(c.Request.Context(), &dto.CreateDatasourceRequest{
		Name:   name,
		Type:   fileType,
		Config: string(configJSON),
	}, userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": ds})
}
