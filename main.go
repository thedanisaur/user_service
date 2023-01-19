package main

import (
	"log"
	"user_service/db"
	"user_service/handlers"

	"github.com/gofiber/fiber/v2"
)

func main() {
	log.Println("Starting User Auth...")
	app := fiber.New()
	defer db.GetInstance().Close()

	// Basic Authentication
	app.Post("/login", handlers.Login)

	// JWT Authentication
	app.Post("/refresh/:username", handlers.RefreshToken)
	app.Get("/users", handlers.GetUsers)
	app.Get("/user/:username", handlers.GetUser)
	app.Post("/user", handlers.CreateUser)

	app.Listen(":4321")
}
