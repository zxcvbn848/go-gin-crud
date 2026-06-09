package controller

import (
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
	"go-gin-crud/internal/validator"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProductController struct {
	productService service.ProductService
}

func NewProductController(productService service.ProductService) *ProductController {
	return &ProductController{
		productService: productService,
	}
}

// CreateProduct 創建產品
// @Summary 創建產品
// @Description 創建一個新產品（需要管理員權限）
// @Tags product
// @Security BearerAuth
// @Param request body dto.CreateProductRequest true "產品資訊"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Router /products [post]
func (ctrl *ProductController) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	product, err := ctrl.productService.CreateProduct(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建失敗"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// GetProducts 獲取產品列表
// @Summary 獲取產品列表
// @Description 獲取產品列表，支援分頁和搜尋（需要認證）
// @Tags product
// @Security BearerAuth
// @Param page query int false "頁碼" default(1)
// @Param page_size query int false "每頁數量" default(10)
// @Param search query string false "搜尋關鍵字"
// @Success 200 {object} dto.PaginationResponse
// @Failure 401 {object} map[string]interface{}
// @Router /products [get]
func (ctrl *ProductController) GetProducts(c *gin.Context) {
	var req dto.PaginationRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		// 如果沒有提供參數，使用預設值
		req.Page = 1
		req.PageSize = 10
		req.Search = ""
	}

	response, err := ctrl.productService.GetProductsWithPagination(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查詢失敗"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// GetProduct 獲取單一產品
// @Summary 獲取單一產品
// @Description 根據 ID 獲取產品詳細資訊（需要認證）
// @Tags product
// @Security BearerAuth
// @Param id path int true "產品 ID"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /products/{id} [get]
func (ctrl *ProductController) GetProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	product, err := ctrl.productService.GetProductByID(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// UpdateProduct 更新產品
// @Summary 更新產品
// @Description 更新指定產品的資訊（需要管理員權限）
// @Tags product
// @Security BearerAuth
// @Param id path int true "產品 ID"
// @Param request body dto.UpdateProductRequest true "更新資訊"
// @Success 200 {object} dto.ProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /products/{id} [put]
func (ctrl *ProductController) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		validator.HandleValidationError(c, err)
		return
	}

	product, err := ctrl.productService.UpdateProduct(uint(id), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, product)
}

// DeleteProduct 刪除產品
// @Summary 刪除產品
// @Description 刪除指定產品（需要管理員權限）
// @Tags product
// @Security BearerAuth
// @Param id path int true "產品 ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /products/{id} [delete]
func (ctrl *ProductController) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	if err := ctrl.productService.DeleteProduct(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "刪除成功"})
}
