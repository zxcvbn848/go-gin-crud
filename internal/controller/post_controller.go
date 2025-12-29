package controller

import (
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
	"go-gin-crud/internal/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PostController struct {
	postService service.PostService
}

func NewPostController(postService service.PostService) *PostController {
	return &PostController{
		postService: postService,
	}
}

// CreatePost 創建文章
// @Summary 創建文章
// @Description 創建一篇新文章（需要管理員權限）
// @Tags post
// @Security BearerAuth
// @Param request body dto.CreatePostRequest true "文章資訊"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Router /posts [post]
func (ctrl *PostController) CreatePost(c *gin.Context) {
	// 從 middleware 取得 user_id
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無法取得使用者資訊"})
		return
	}

	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用者 ID 格式錯誤"})
		return
	}

	var req dto.CreatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	post, err := ctrl.postService.CreatePost(userIDUint, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建失敗"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// GetPosts 獲取文章列表
// @Summary 獲取文章列表
// @Description 獲取文章列表，支援分頁和搜尋（需要認證）
// @Tags post
// @Security BearerAuth
// @Param page query int false "頁碼" default(1)
// @Param page_size query int false "每頁數量" default(10)
// @Param search query string false "搜尋關鍵字"
// @Success 200 {object} dto.PaginationResponse
// @Failure 401 {object} gin.H
// @Router /posts [get]
func (ctrl *PostController) GetPosts(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 如果沒有提供參數，使用預設值
		req.Page = 1
		req.PageSize = 10
		req.Search = ""
	}

	response, err := ctrl.postService.GetPostsWithPagination(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetPost 獲取單一文章
// @Summary 獲取單一文章
// @Description 根據 ID 獲取文章詳細資訊（需要認證）
// @Tags post
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /posts/{id} [get]
func (ctrl *PostController) GetPost(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	post, err := ctrl.postService.GetPostByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// UpdatePost 更新文章
// @Summary 更新文章
// @Description 更新指定文章的資訊（管理員可更新任何文章，一般用戶只能更新自己的文章）
// @Tags post
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Param request body dto.UpdatePostRequest true "更新資訊"
// @Success 200 {object} dto.PostResponse
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /posts/{id} [put]
func (ctrl *PostController) UpdatePost(c *gin.Context) {
	// 從 middleware 取得 user_id
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無法取得使用者資訊"})
		return
	}

	roleValue, roleExists := c.Get("role")
	if !roleExists {
		c.JSON(http.StatusForbidden, gin.H{"error": "缺少角色資訊"})
		return
	}

	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用者 ID 格式錯誤"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	var req dto.UpdatePostRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	role, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "角色資訊格式錯誤"})
		return
	}

	post, err := ctrl.postService.UpdatePost(uint(id), userIDUint, role, req)
	if err != nil {
		if err.Error() == "無權限修改此文章" {
			c.JSON(http.StatusForbidden, gin.H{"error": "無權限修改此文章"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, post)
}

// DeletePost 刪除文章
// @Summary 刪除文章
// @Description 刪除指定文章（管理員可刪除任何文章，一般用戶只能刪除自己的文章）
// @Tags post
// @Security BearerAuth
// @Param id path int true "文章 ID"
// @Success 200 {object} gin.H
// @Failure 400 {object} gin.H
// @Failure 401 {object} gin.H
// @Failure 403 {object} gin.H
// @Failure 404 {object} gin.H
// @Router /posts/{id} [delete]
func (ctrl *PostController) DeletePost(c *gin.Context) {
	// 從 middleware 取得 user_id
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "無法取得使用者資訊"})
		return
	}

	roleValue, roleExists := c.Get("role")
	if !roleExists {
		c.JSON(http.StatusForbidden, gin.H{"error": "缺少角色資訊"})
		return
	}

	var userIDUint uint
	switch v := userID.(type) {
	case float64:
		userIDUint = uint(v)
	case uint:
		userIDUint = v
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "使用者 ID 格式錯誤"})
		return
	}

	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	role, ok := roleValue.(string)
	if !ok {
		c.JSON(http.StatusForbidden, gin.H{"error": "角色資訊格式錯誤"})
		return
	}

	if err := ctrl.postService.DeletePost(uint(id), userIDUint, role); err != nil {
		if err.Error() == "無權限刪除此文章" {
			c.JSON(http.StatusForbidden, gin.H{"error": "無權限刪除此文章"})
			return
		}
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

