package main

import (
	"chat-app/internal/app"
	"chat-app/internal/config"
	"chat-app/internal/db"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	// nếu làm jwt thì lưu ở sessionStorage
	err := godotenv.Load() // măc định nó sẽ tìm file .env ở thư mục gốc của project
	if err != nil {
		log.Println("No .env file found")
	}
	// khởi tạo db
	if err := db.InitDB(); err != nil {
		log.Fatal("cannot connect to db:", err)
	}
	// Initialize configuration
	config := config.NewConfig()
	// init application
	application := app.NewApplication(config)

	// start server
	if err :=application.Run(); err != nil {
		panic(err)
	}
}
