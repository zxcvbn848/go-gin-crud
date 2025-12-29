package controller

import (
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
	"go-gin-crud/internal/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService service.UserService
}

func NewUserController(userService service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

// CreateUser 創建用戶
// @Summary 創建用戶
// @Description 創建一個新用戶（需要管理員權限）
// @Tags user
// @Security BearerAuth
// @Param request body dto.CreateUserRequest true "用戶資訊"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} gin.H
// @Failure 403 {object} gin.H
// @Router /users [post]
func (ctrl *UserController) CreateUser(c *gin.Context) {
	var req dto.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	user, err := ctrl.userService.CreateUser(req)
	if err != nil {
		if err.Error() == "email 已存在" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email 已存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建失敗"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUsers 獲取用戶列表
// @Summary 獲取用戶列表
// @Description 獲取用戶列表，支援分頁和搜尋（需要管理員權限）
// @Tags user
// @Security BearerAuth
// @Param page query int false "頁碼" default(1)
// @Param page_size query int false "每頁數量" default(10)
// @Param search query string false "搜尋關鍵字"
// @Success 200 {object} dto.PaginationResponse
// @Failure 403 {object} gin.H
// @Router /users [get]
func (ctrl *UserController) GetUsers(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 如果沒有提供參數，使用預設值
		req.Page = 1
		req.PageSize = 10
		req.Search = ""
	}

	response, err := ctrl.userService.GetUsersWithPagination(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetUser 獲取單一用戶
// @Summary 獲取單一用戶
// @Description 根據 ID 獲取用戶詳細資訊（需要管理員權限）
// @Tags user
// @Security BearerAuth
// @Param id path int true "用戶 ID"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /users/{id} [get]
func (ctrl *UserController) GetUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	user, err := ctrl.userService.GetUserByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateUser 更新用戶
// @Summary 更新用戶
// @Description 更新指定用戶的資訊（需要管理員權限）
// @Tags user
// @Security BearerAuth
// @Param id path int true "用戶 ID"
// @Param request body dto.UpdateUserRequest true "更新資訊"
// @Success 200 {object} dto.UserResponse
// @Failure 400 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /users/{id} [put]
func (ctrl *UserController) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	var req dto.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	user, err := ctrl.userService.UpdateUser(uint(id), req)
	if err != nil {
		if err.Error() == "email 已存在" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "email 已存在"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser 刪除用戶
// @Summary 刪除用戶
// @Description 刪除指定用戶（需要管理員權限）
// @Tags user
// @Security BearerAuth
// @Param id path int true "用戶 ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /users/{id} [delete]
func (ctrl *UserController) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	if err := ctrl.userService.DeleteUser(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}
