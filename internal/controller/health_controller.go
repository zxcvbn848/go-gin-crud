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

// GetHealth 健康檢查
// @Summary 健康檢查
// @Description 檢查服務是否正常運行
// @Tags health
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (ctrl *HealthController) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}
