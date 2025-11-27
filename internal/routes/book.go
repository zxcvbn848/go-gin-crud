package routes

import (
	"go-gin-crud/internal/database"
	"go-gin-crud/internal/database/models"

	"github.com/gin-gonic/gin"
)

func RegisterBookRoutes(r *gin.Engine) {

	r.POST("/books", CreateBook)
	r.GET("/books", GetBooks)
	r.GET("/books/:id", GetBook)
	r.PUT("/books/:id", UpdateBook)
	r.DELETE("/books/:id", DeleteBook)
}

func CreateBook(c *gin.Context) {
	var book models.Book
	if err := c.ShouldBindJSON(&book); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	database.DB.Create(&book)
	c.JSON(200, book)
}

func GetBooks(c *gin.Context) {
	var books []models.Book
	database.DB.Find(&books)
	c.JSON(200, books)
}

func GetBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book
	result := database.DB.First(&book, id)

	if result.Error != nil {
		c.JSON(404, gin.H{"error": "找不到資料"})
		return
	}

	c.JSON(200, book)
}

func UpdateBook(c *gin.Context) {
	id := c.Param("id")
	var book models.Book

	if err := database.DB.First(&book, id).Error; err != nil {
		c.JSON(404, gin.H{"error": "找不到資料"})
		return
	}

	var input models.Book
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&book).Updates(input)
	c.JSON(200, book)
}

func DeleteBook(c *gin.Context) {
	id := c.Param("id")
	database.DB.Delete(&models.Book{}, id)
	c.JSON(200, gin.H{"message": "刪除成功"})
}
