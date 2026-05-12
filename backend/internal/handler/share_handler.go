package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/middleware"
	"cozy-insight/internal/service"
)

type ShareHandler struct {
	service *service.ShareLinkService
}

func NewShareHandler(service *service.ShareLinkService) *ShareHandler {
	return &ShareHandler{service: service}
}

func (h *ShareHandler) GetDashboard(c *gin.Context) {
	token := c.Param("token")
	password := c.Query("password")
	dashboard, err := h.service.GetDashboard(c.Request.Context(), token, password)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": dashboard})
}

func (h *ShareHandler) List(c *gin.Context) {
	userID := middleware.GetUserID(c)
	links, err := h.service.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": links})
}
