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

