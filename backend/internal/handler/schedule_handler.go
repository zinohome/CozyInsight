package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"cozy-insight/internal/middleware"
	"cozy-insight/internal/model"
	"cozy-insight/internal/service"
)

type ScheduleHandler struct {
	service service.ScheduleService
}

func NewScheduleHandler(service service.ScheduleService) *ScheduleHandler {
	return &ScheduleHandler{service: service}
}

func (h *ScheduleHandler) Create(c *gin.Context) {
	var task model.ScheduleTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	task.CreatedBy = middleware.GetUserID(c)
	if err := h.service.CreateTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": task})
}

func (h *ScheduleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid id"})
		return
	}
	var task model.ScheduleTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	task.ID = id
	if err := h.service.UpdateTask(c.Request.Context(), &task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": task})
}

func (h *ScheduleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid id"})
		return
	}
	if err := h.service.DeleteTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *ScheduleHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid id"})
		return
	}
	task, err := h.service.GetTask(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"code": 404, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": task})
}

func (h *ScheduleHandler) List(c *gin.Context) {
	list, err := h.service.ListTasks(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": list})
}

func (h *ScheduleHandler) Enable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid id"})
		return
	}
	if err := h.service.EnableTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *ScheduleHandler) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid id"})
		return
	}
	if err := h.service.DisableTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}

func (h *ScheduleHandler) Execute(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid id"})
		return
	}
	if err := h.service.ExecuteTask(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 200, "data": "ok"})
}