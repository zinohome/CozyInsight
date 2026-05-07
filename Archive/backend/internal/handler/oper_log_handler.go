package handler

import (
	"cozy-insight-backend/internal/service"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type OperLogHandler struct {
	service service.OperLogService
}

func NewOperLogHandler(service service.OperLogService) *OperLogHandler {
	return &OperLogHandler{service: service}
}

// List 查询操作日志列表
func (h *OperLogHandler) List(c *gin.Context) {
	userID := c.Query("userId")
	module := c.Query("module")
	
	startTime, _ := strconv.ParseInt(c.Query("startTime"), 10, 64)
	endTime, _ := strconv.ParseInt(c.Query("endTime"), 10, 64)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	logs, total, err := h.service.List(c.Request.Context(), userID, module, startTime, endTime, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  logs,
		"total": total,
		"page":  page,
		"pageSize": pageSize,
	})
}

// CleanOld 清理旧日志
func (h *OperLogHandler) CleanOld(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("beforeDays", "90"))

	err := h.service.CleanOldLogs(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "cleaned"})
}
