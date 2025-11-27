package dto

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Stock       int     `json:"stock" binding:"min=0"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"omitempty,min=0"`
	Stock       *int    `json:"stock" binding:"omitempty,min=0"`
}

