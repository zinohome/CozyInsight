package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseUintParam(c *gin.Context, param string) (uint64, bool) {
	v, err := strconv.ParseUint(c.Param(param), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid " + param})
		return 0, false
	}
	return v, true
}

func parseUintQuery(c *gin.Context, key string, defaultValue uint64) (uint64, bool) {
	q := c.Query(key)
	if q == "" {
		return defaultValue, true
	}
	v, err := strconv.ParseUint(q, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"code": 400, "error": "invalid " + key})
		return 0, false
	}
	return v, true
}
