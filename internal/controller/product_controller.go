package controller

import (
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/service"
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

func (ctrl *ProductController) CreateProduct(c *gin.Context) {
	var req dto.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "資料格式錯誤"})
		return
	}

	product, err := ctrl.productService.CreateProduct(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "創建失敗"})
		return
	}

	c.JSON(http.StatusOK, product)
}

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

func (ctrl *ProductController) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "無效的 ID"})
		return
	}

	var req dto.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "資料格式錯誤"})
		return
	}

	product, err := ctrl.productService.UpdateProduct(uint(id), req)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(http.StatusOK, product)
}

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

