package handler

import (
	"cozy-insight-backend/internal/model"
	"cozy-insight-backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ScheduleHandler struct {
	service service.ScheduleService
}

func NewScheduleHandler(service service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

// Create 创建定时任务
func (h *ScheduleHandler) Create(c *gin.Context) {
	var task model.ScheduleTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, _ := c.Get("userID")
	task.CreateBy = userID.(string)

	if err := h.service.CreateTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Update 更新定时任务
func (h *ScheduleHandler) Update(c *gin.Context) {
	id := c.Param("id")
	
	var task model.ScheduleTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	task.ID = id

	if err := h.service.UpdateTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// Delete 删除定时任务
func (h *ScheduleHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DeleteTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}

// Get 获取定时任务详情
func (h *ScheduleHandler) Get(c *gin.Context) {
	id := c.Param("id")

	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// List 获取定时任务列表
func (h *ScheduleHandler) List(c *gin.Context) {
	tasks, err := h.service.ListTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, tasks)
}

// Enable 启用定时任务
func (h *ScheduleHandler) Enable(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.EnableTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "enabled"})
}

// Disable 禁用定时任务
func (h *ScheduleHandler) Disable(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.DisableTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "disabled"})
}

// Execute 立即执行定时任务
func (h *ScheduleHandler) Execute(c *gin.Context) {
	id := c.Param("id")

	if err := h.service.ExecuteTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "executed"})
}
