package models

type User struct {
	ID       uint   `gorm:"primaryKey"`
	Email    string `gorm:"unique" json:"email"`
	Password string `json:"password"`
	Role     string `gorm:"type:varchar(20);default:'user'" json:"role"`
}
