package controller

import (
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
	"go-gin-crud/internal/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BookController struct {
	bookService service.BookService
}

func NewBookController(bookService service.BookService) *BookController {
	return &BookController{
		bookService: bookService,
	}
}

// CreateBook 創建書籍
// @Summary 創建書籍
// @Description 創建一本新書籍（需要管理員權限）
// @Tags book
// @Security BearerAuth
// @Param request body dto.CreateBookRequest true "書籍資訊"
// @Success 200 {object} dto.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /books [post]
func (ctrl *BookController) CreateBook(c *gin.Context) {
	var req dto.CreateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	book, err := ctrl.bookService.CreateBook(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建失敗"})
		return
	}

	c.JSON(http.StatusOK, book)
}

// GetBooks 獲取書籍列表
// @Summary 獲取書籍列表
// @Description 獲取書籍列表，支援分頁和搜尋（需要認證）
// @Tags book
// @Security BearerAuth
// @Param page query int false "頁碼" default(1)
// @Param page_size query int false "每頁數量" default(10)
// @Param search query string false "搜尋關鍵字"
// @Success 200 {object} dto.PaginationResponse
// @Failure 401 {object} map[string]interface{}
// @Router /books [get]
func (ctrl *BookController) GetBooks(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 如果沒有提供參數，使用預設值
		req.Page = 1
		req.PageSize = 10
		req.Search = ""
	}

	response, err := ctrl.bookService.GetBooksWithPagination(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetBook 獲取單一書籍
// @Summary 獲取單一書籍
// @Description 根據 ID 獲取書籍詳細資訊（需要認證）
// @Tags book
// @Security BearerAuth
// @Param id path int true "書籍 ID"
// @Success 200 {object} dto.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /books/{id} [get]
func (ctrl *BookController) GetBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	book, err := ctrl.bookService.GetBookByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, book)
}

// UpdateBook 更新書籍
// @Summary 更新書籍
// @Description 更新指定書籍的資訊（需要管理員權限）
// @Tags book
// @Security BearerAuth
// @Param id path int true "書籍 ID"
// @Param request body dto.UpdateBookRequest true "更新資訊"
// @Success 200 {object} dto.BookResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /books/{id} [put]
func (ctrl *BookController) UpdateBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	var req dto.UpdateBookRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	book, err := ctrl.bookService.UpdateBook(uint(id), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, book)
}

// DeleteBook 刪除書籍
// @Summary 刪除書籍
// @Description 刪除指定書籍（需要管理員權限）
// @Tags book
// @Security BearerAuth
// @Param id path int true "書籍 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /books/{id} [delete]
func (ctrl *BookController) DeleteBook(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	if err := ctrl.bookService.DeleteBook(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}

