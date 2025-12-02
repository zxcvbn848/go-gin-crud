package routes

import (
	"go-gin-crud/internal/controller"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterAccountRoutes 註冊帳戶相關路由
func RegisterAccountRoutes(r *gin.Engine) {
	// 創建帳戶服務（初始餘額 100）
	accountService := service.NewAccountService(100)
	accountController := controller.NewAccountController(accountService)

	api := r.Group("/accounts")
	{
		// 獲取餘額
		api.GET("/balance", accountController.GetBalance)

		// 存款
		api.POST("/deposit", accountController.Deposit)

		// 取款
		api.POST("/withdraw", accountController.Withdraw)

		// 設置餘額
		api.POST("/balance", accountController.SetBalance)

		// 重置帳戶
		api.POST("/reset", accountController.Reset)

		// 批量執行交易
		api.POST("/batch", accountController.ExecuteBatchTransactions)

		// 執行隨機批量交易（演示用）
		api.POST("/batch/random", accountController.ExecuteRandomBatchTransactions)
	}
}
