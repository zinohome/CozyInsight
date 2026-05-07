package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/service"
)

type OperationLogHandler struct {
	service *service.OperationLogService
}

func NewOperationLogHandler(service *service.OperationLogService) *OperationLogHandler {
	return &OperationLogHandler{service: service}
}

func (h *OperationLogHandler) List(c *gin.Context) {
	limit, _ := strconv.Atoi(c.Query("limit"))
	list, err := h.service.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}
