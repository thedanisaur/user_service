package main

import (
	"log"
	"user_service/db"
	"user_service/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	log.Println("Starting User Auth...")
	app := fiber.New()
	defer db.GetInstance().Close()

	// Add CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Basic Authentication
	app.Post("/login", handlers.Login)

	// JWT Authentication
	app.Post("/refresh/:username", handlers.RefreshToken)
	app.Get("/users", handlers.GetUsers)
	app.Get("/user/:username", handlers.GetUser)
	app.Post("/user", handlers.CreateUser)

	app.Listen(":4321")
}
