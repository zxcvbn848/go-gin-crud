package database

import (
	"log"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	dsn := "gogin:a3935522@tcp(127.0.0.1:3307)/goGinCRUD?charset=utf8mb4&parseTime=True&loc=Local"
	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("❌ 無法連線到資料庫: ", err)
	}

	log.Println("✅ 成功連上資料庫")
}
