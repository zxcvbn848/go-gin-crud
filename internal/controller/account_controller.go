package controller

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/service"

	"github.com/gin-gonic/gin"
)

// AccountController 帳戶控制器
type AccountController struct {
	accountService service.AccountService
}

// NewAccountController 創建帳戶控制器
func NewAccountController(accountService service.AccountService) *AccountController {
	return &AccountController{
		accountService: accountService,
	}
}

// GetBalance 獲取餘額
// @Summary 獲取帳戶餘額
// @Description 獲取當前帳戶餘額
// @Tags account
// @Success 200 {object} dto.AccountResponse
// @Router /api/accounts/balance [get]
func (ctrl *AccountController) GetBalance(c *gin.Context) {
	balance := ctrl.accountService.GetBalance()
	c.JSON(http.StatusOK, dto.AccountResponse{
		Balance: balance,
	})
}

// Deposit 存款
// @Summary 存款
// @Description 向帳戶存款
// @Tags account
// @Param request body dto.DepositRequest true "存款金額"
// @Success 200 {object} dto.TransactionResponse
// @Router /api/accounts/deposit [post]
func (ctrl *AccountController) Deposit(c *gin.Context) {
	var req dto.DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := ctrl.accountService.Deposit(req.Amount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Withdraw 取款
// @Summary 取款
// @Description 從帳戶取款
// @Tags account
// @Param request body dto.WithdrawRequest true "取款金額"
// @Success 200 {object} dto.TransactionResponse
// @Router /api/accounts/withdraw [post]
func (ctrl *AccountController) Withdraw(c *gin.Context) {
	var req dto.WithdrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := ctrl.accountService.Withdraw(req.Amount)
	if err != nil {
		// 即使餘額不足，也返回交易響應（包含錯誤信息）
		if resp != nil {
			c.JSON(http.StatusOK, resp)
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SetBalance 設置餘額
// @Summary 設置餘額
// @Description 設置帳戶初始餘額
// @Tags account
// @Param balance body object{balance=int64} true "餘額"
// @Success 200 {object} dto.AccountResponse
// @Router /api/accounts/balance [post]
func (ctrl *AccountController) SetBalance(c *gin.Context) {
	var req struct {
		Balance int64 `json:"balance" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctrl.accountService.SetBalance(req.Balance)
	c.JSON(http.StatusOK, dto.AccountResponse{
		Balance: req.Balance,
	})
}

// Reset 重置帳戶
// @Summary 重置帳戶
// @Description 重置帳戶餘額為 0
// @Tags account
// @Success 200 {object} dto.AccountResponse
// @Router /api/accounts/reset [post]
func (ctrl *AccountController) Reset(c *gin.Context) {
	ctrl.accountService.Reset()
	c.JSON(http.StatusOK, dto.AccountResponse{
		Balance: 0,
	})
}

// ExecuteBatchTransactions 批量執行交易
// @Summary 批量執行交易
// @Description 並發執行多筆交易（存款、取款、查詢）
// @Tags account
// @Param request body dto.BatchTransactionRequest true "交易操作列表"
// @Success 200 {object} dto.BatchTransactionResponse
// @Router /api/accounts/batch [post]
func (ctrl *AccountController) ExecuteBatchTransactions(c *gin.Context) {
	var req dto.BatchTransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := ctrl.accountService.ExecuteBatchTransactions(req.Operations)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// ExecuteRandomBatchTransactions 執行隨機批量交易（演示用）
// @Summary 執行隨機批量交易
// @Description 生成隨機交易操作並執行（用於演示併發）
// @Tags account
// @Param count query int false "交易數量" default(10)
// @Param delay query bool false "是否添加延遲（用於演示）" default(false)
// @Success 200 {object} dto.BatchTransactionResponse
// @Router /api/accounts/batch/random [post]
func (ctrl *AccountController) ExecuteRandomBatchTransactions(c *gin.Context) {
	countStr := c.DefaultQuery("count", "10")
	count := 10
	if c, err := parseInt(countStr); err == nil && c > 0 {
		count = c
	}

	// 生成隨機操作：-1 (取款), 0 (查詢), 1-2 (存款)
	// 使用本地隨機數生成器（Go 1.20+ 推薦方式）
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	operations := make([]int, count)
	for i := 0; i < count; i++ {
		operations[i] = r.Intn(4) - 1 // -1, 0, 1, 2
	}

	// 檢查是否使用延遲版本
	useDelay := c.Query("delay") == "true"

	// 輸出 useDelay 的值到日誌
	logger.Log.WithFields(map[string]interface{}{
		"useDelay":   useDelay,
		"count":      count,
		"operations": operations,
	}).Info("執行隨機批量交易")

	var resp *dto.BatchTransactionResponse
	var err error

	if useDelay {
		// 使用帶延遲的版本（演示用）
		if accountServiceWithDelay, ok := ctrl.accountService.(interface {
			ExecuteBatchTransactionsWithDelay(operations []int, delay time.Duration) (*dto.BatchTransactionResponse, error)
		}); ok {
			resp, err = accountServiceWithDelay.ExecuteBatchTransactionsWithDelay(operations, 1000*time.Millisecond)
		} else {
			resp, err = ctrl.accountService.ExecuteBatchTransactions(operations)
		}
	} else {
		resp, err = ctrl.accountService.ExecuteBatchTransactions(operations)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

// parseInt 輔助函數：解析整數
func parseInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}
