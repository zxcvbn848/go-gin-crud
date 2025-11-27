package models

type Product struct {
	ID          uint    `gorm:"primaryKey" json:"id"`
	Name        string  `gorm:"not null" json:"name"`
	Description string  `json:"description"`
	Price       float64 `gorm:"not null" json:"price"`
	Stock       int     `gorm:"default:0" json:"stock"`
}

