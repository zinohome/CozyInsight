package handler

import (
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

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
	header, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "missing file"})
		return
	}

	const maxFileSize = 100 << 20 // 100MB
	if header.Size > maxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "file too large, max 100MB"})
		return
	}

	ext := filepath.Ext(header.Filename)
	if ext != ".xlsx" && ext != ".xls" && ext != ".csv" {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "unsupported file type, only .xlsx, .xls, .csv are allowed"})
		return
	}

	if !validateFileMagic(header, ext) {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "file content does not match extension"})
		return
	}

	uploadDir := "./uploads"
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "create upload dir failed"})
		return
	}

	savePath := filepath.Join(uploadDir, uuid.New().String()+ext)
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
	configJSON, err := json.Marshal(config)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": "marshal config failed"})
		return
	}

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

func validateFileMagic(header *multipart.FileHeader, ext string) bool {
	f, err := header.Open()
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 8)
	if _, err := io.ReadFull(f, buf); err != nil {
		return false
	}

	switch ext {
	case ".xlsx", ".xls":
		// ZIP magic for xlsx, or OLE2 header for xls
		return buf[0] == 0x50 && buf[1] == 0x4B && buf[2] == 0x03 && buf[3] == 0x04 ||
			buf[0] == 0xD0 && buf[1] == 0xCF && buf[2] == 0x11 && buf[3] == 0xE0
	case ".csv":
		// CSV has no magic; accept if first bytes are printable text
		for _, b := range buf {
			if b == 0 {
				return false
			}
		}
		return true
	}
	return false
}
