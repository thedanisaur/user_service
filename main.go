package main

import (
	"log"
	"time"
	"user_service/db"
	"user_service/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	log.Println("Starting User Service...")
	app := fiber.New()
	defer db.GetInstance().Close()

	// Add CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "127.0.0.1",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	// Add Rate Limiter
	app.Use(limiter.New(limiter.Config{
		Max:                    5,
		Expiration:             5 * time.Minute,
		LimiterMiddleware:      limiter.SlidingWindow{},
		SkipSuccessfulRequests: true,
	}))

	// Basic Authentication
	app.Post("/login", handlers.Login)

	// JWT Authentication
	app.Post("/refresh/:username", handlers.RefreshToken)
	app.Get("/users", handlers.GetUsers)
	app.Get("/user/:username", handlers.GetUser)
	app.Post("/user", handlers.CreateUser)
	app.Get("/validate", handlers.Validate)

	log.Fatal(app.ListenTLS(":4321", "./certs/cert.crt", "./keys/key.key"))
}
