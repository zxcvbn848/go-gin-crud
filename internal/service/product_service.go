package service

import (
	"go-gin-crud/internal/database/models"
	"go-gin-crud/internal/dto"
	"go-gin-crud/internal/repository"
	"math"
)

type ProductService interface {
	CreateProduct(req dto.CreateProductRequest) (*models.Product, error)
	GetProductByID(id uint) (*models.Product, error)
	UpdateProduct(id uint, req dto.UpdateProductRequest) (*models.Product, error)
	DeleteProduct(id uint) error
	GetProductsWithPagination(req dto.PaginationRequest) (*dto.PaginationResponse, error)
}

type productService struct {
	productRepo repository.ProductRepository
}

func NewProductService(productRepo repository.ProductRepository) ProductService {
	return &productService{
		productRepo: productRepo,
	}
}

func (s *productService) CreateProduct(req dto.CreateProductRequest) (*models.Product, error) {
	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		Stock:       req.Stock,
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	return product, nil
}

func (s *productService) GetProductByID(id uint) (*models.Product, error) {
	return s.productRepo.FindByID(id)
}

func (s *productService) UpdateProduct(id uint, req dto.UpdateProductRequest) (*models.Product, error) {
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

	return product, nil
}

func (s *productService) DeleteProduct(id uint) error {
	return s.productRepo.Delete(id)
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

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))

	return &dto.PaginationResponse{
		Data:       products,
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
	}, nil
}

