package dto

import "time"

type CreateBookRequest struct {
	Title  string `json:"title" binding:"required"`
	Author string `json:"author" binding:"required"`
}

type UpdateBookRequest struct {
	Title  string `json:"title"`
	Author string `json:"author"`
}

type BookResponse struct {
	ID        uint      `json:"id"`
	Title     string    `json:"title"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PaginationRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Search   string `form:"search"`
}

type PaginationResponse struct {
	Data       interface{} `json:"data"`
	Page       int          `json:"page"`
	PageSize   int          `json:"page_size"`
	Total      int64        `json:"total"`
	TotalPages int          `json:"total_pages"`
}

