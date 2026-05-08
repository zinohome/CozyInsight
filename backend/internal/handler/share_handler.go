package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

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
	dashboard, err := h.service.GetDashboard(c.Request.Context(), token)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": dashboard})
}
