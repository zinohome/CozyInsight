package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	svc service.DashboardService
}

func NewDashboardHandler(svc service.DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) Create(c *gin.Context) {
	var dashboard model.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.svc.Create(c.Request.Context(), &dashboard); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var dashboard model.Dashboard
	if err := c.ShouldBindJSON(&dashboard); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	dashboard.ID = id

	if err := h.svc.Update(c.Request.Context(), &dashboard); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (h *DashboardHandler) Get(c *gin.Context) {
	id := c.Param("id")
	dashboard, err := h.svc.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}

func (h *DashboardHandler) List(c *gin.Context) {
	list, err := h.svc.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, list)
}

// Publish 发布仪表板
func (h *DashboardHandler) Publish(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Publish(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "发布成功"})
}

// Unpublish 下线仪表板
func (h *DashboardHandler) Unpublish(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.Unpublish(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "message": "下线成功"})
}

// View 查看已发布的仪表板（公开访问）
func (h *DashboardHandler) View(c *gin.Context) {
	id := c.Param("id")

	dashboard, err := h.svc.GetPublished(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "仪表板不存在或未发布"})
		return
	}

	c.JSON(http.StatusOK, dashboard)
}
