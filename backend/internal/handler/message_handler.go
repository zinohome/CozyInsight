package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type MessageHandler struct {
	service *service.MessageService
}

func NewMessageHandler(service *service.MessageService) *MessageHandler {
	return &MessageHandler{service: service}
}

func (h *MessageHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	unreadOnly := c.Query("unreadOnly") == "true"
	msgs, err := h.service.List(c.Request.Context(), userID, unreadOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": msgs})
}

func (h *MessageHandler) CountUnread(c *gin.Context) {
	userID := middleware.GetUserID(c)
	count, err := h.service.CountUnread(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": count})
}

func (h *MessageHandler) MarkAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.MarkAsRead(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *MessageHandler) MarkAllAsRead(c *gin.Context) {
	userID := middleware.GetUserID(c)
	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	userID := middleware.GetUserID(c)
	id, ok := parseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}
