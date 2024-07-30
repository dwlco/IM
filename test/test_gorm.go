package main

import (
	"fmt"
	"ginchat/models"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {

	newLogger := logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
		SlowThreshold: 200 * time.Millisecond,
		LogLevel:      logger.Info,
		Colorful:      true,
	})

	DB, _ := gorm.Open(mysql.Open("root:dwl@tcp(127.0.0.1:3306)/ginchat?charset=utf8mb4&parseTime=True&loc=Local"), &gorm.Config{Logger: newLogger})
	fmt.Println("config mysql testing......")

	// Migrate the schema
	//DB.AutoMigrate(&models.UserBasic{})
	// DB.AutoMigrate(&models.Message{})
	DB.AutoMigrate(&models.GroupBasic{})
	DB.AutoMigrate(&models.Contact{})
	// Create

	// user := models.UserBasic{}
	// user.Name = "大龙"
	// fmt.Println(user)
	// fmt.Println(db.Select("name").Create(&user))

	// // Read
	// fmt.Println(db.First(&user, 1))

	// // Update - update product's price to 200
	// db.Model(&user).Update("pass_word", "zxcv")

}
