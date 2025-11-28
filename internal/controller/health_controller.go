package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// HealthController 提供服務健康檢查
type HealthController struct{}

func NewHealthController() *HealthController {
	return &HealthController{}
}

// GetHealth 回傳服務狀態
func (ctrl *HealthController) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}


