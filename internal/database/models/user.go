package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"password"`
}
