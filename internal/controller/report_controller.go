package controller

import (
	"net/http"

	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

type ReportController struct {
	reportService service.ReportService
}

func NewReportController(reportService service.ReportService) *ReportController {
	return &ReportController{reportService: reportService}
}

// GetOverview 總覽統計
// @Summary 總覽統計
// @Description 各表筆數與商品庫存總值（需要管理員權限）
// @Tags report
// @Security BearerAuth
// @Success 200 {object} dto.ReportOverviewResponse
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /reports/overview [get]
func (ctrl *ReportController) GetOverview(c *gin.Context) {
	// 傳 c.Request.Context() 而不是 c —— 請求層逾時的 deadline 掛在
	// request 的 context 上，帶下去 GORM 才會在逾時的時候中斷查詢
	resp, err := ctrl.reportService.Overview(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得總覽統計失敗"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetDaily 每日新增統計
// @Summary 每日新增統計
// @Description 區間內每日新增的使用者/文章/商品數。未指定區間時為最近 30 天，區間上限 365 天（需要管理員權限）
// @Tags report
// @Security BearerAuth
// @Param from query string false "起始日期 YYYY-MM-DD"
// @Param to query string false "結束日期 YYYY-MM-DD"
// @Success 200 {object} dto.ReportDailyResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /reports/daily [get]
func (ctrl *ReportController) GetDaily(c *gin.Context) {
	var req dto.ReportDailyRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "日期格式錯誤，請使用 YYYY-MM-DD"})
		return
	}

	resp, err := ctrl.reportService.Daily(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得每日統計失敗"})
		return
	}
	c.JSON(http.StatusOK, resp)
}

// GetTopAuthors 作者發文排行
// @Summary 作者發文排行
// @Description 依發文數排序的作者排行，預設前 10 名，上限 100（需要管理員權限）
// @Tags report
// @Security BearerAuth
// @Param limit query int false "取前幾名，1~100，預設 10"
// @Success 200 {object} dto.ReportAuthorsResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /reports/authors [get]
func (ctrl *ReportController) GetTopAuthors(c *gin.Context) {
	var req dto.ReportAuthorsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "limit 需為 1~100 的整數"})
		return
	}

	resp, err := ctrl.reportService.TopAuthors(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "取得作者排行失敗"})
		return
	}
	c.JSON(http.StatusOK, resp)
}
