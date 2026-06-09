package service

import (
	"context"
	"go-gin-crud/internal/cache"
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/logger"
	"go-gin-crud/internal/repository"
	"math"
)

type ProductService interface {
	CreateProduct(req dto.CreateProductRequest) (*dto.ProductResponse, error)
	GetProductByID(id uint) (*dto.ProductResponse, error)
	UpdateProduct(id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error)
	DeleteProduct(id uint) error
	GetProductsWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error)
}

type productService struct {
	productRepo  repository.ProductRepository
	productCache cache.ProductCache
}

func NewProductService(productRepo repository.ProductRepository, productCache cache.ProductCache) ProductService {
	return &productService{
		productRepo:  productRepo,
		productCache: productCache,
	}
}

func (s *productService) CreateProduct(req dto.CreateProductRequest) (*dto.ProductResponse, error) {
	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	resp := &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	if s.productCache != nil {
		_ = s.productCache.SetProduct(context.Background(), product.ID, resp)
		logger.Log.WithField("product_id", product.ID).Info("Product 快取已寫入（Create）")
	}
	return resp, nil
}

func (s *productService) GetProductByID(id uint) (*dto.ProductResponse, error) {
	ctx := context.Background()
	if s.productCache != nil {
		if cached, err := s.productCache.GetProduct(ctx, id); err == nil && cached != nil {
			logger.Log.WithField("product_id", id).Info("Product 從 Redis 快取取得")
			return cached, nil
		}
	}

	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	resp := &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	if s.productCache != nil {
		_ = s.productCache.SetProduct(ctx, id, resp)
		logger.Log.WithField("product_id", id).Info("Product 從 DB 取得並寫入快取")
	} else {
		logger.Log.WithField("product_id", id).Info("Product 從 DB 取得（未啟用快取）")
	}
	return resp, nil
}

func (s *productService) UpdateProduct(id uint, req dto.UpdateProductRequest) (*dto.ProductResponse, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Price > 0 {
		product.Price = req.Price
	}
	if req.Stock != nil {
		product.Stock = *req.Stock
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	resp := &dto.ProductResponse{
		ID:          product.ID,
		Name:        product.Name,
		Description: product.Description,
		Price:       product.Price,
		Stock:       product.Stock,
		CreatedAt:   product.CreatedAt,
		UpdatedAt:   product.UpdatedAt,
	}
	if s.productCache != nil {
		_ = s.productCache.SetProduct(context.Background(), product.ID, resp)
		logger.Log.WithField("product_id", product.ID).Info("Product 快取已更新（Update）")
	}
	return resp, nil
}

func (s *productService) DeleteProduct(id uint) error {
	if err := s.productRepo.Delete(id); err != nil {
		return err
	}
	if s.productCache != nil {
		_ = s.productCache.DeleteProduct(context.Background(), id)
		logger.Log.WithField("product_id", id).Info("Product 快取已刪除（Delete）")
	}
	return nil
}

func (s *productService) GetProductsWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error) {
	// 設定預設值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	products, total, err := s.productRepo.FindAllWithPagination(req.Page, req.PageSize, req.Search)
	if err != nil {
		return nil, err
	}

	// 轉換為 ProductResponse
	productResponses := make([]dto.ProductResponse, len(products))
	for i, product := range products {
		productResponses[i] = dto.ProductResponse{
			ID:          product.ID,
			Name:        product.Name,
			Description: product.Description,
			Price:       product.Price,
			Stock:       product.Stock,
			CreatedAt:   product.CreatedAt,
			UpdatedAt:   product.UpdatedAt,
		}
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       productResponses,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

