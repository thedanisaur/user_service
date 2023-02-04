package main

import (
	"log"
	"time"
	"user_service/db"
	"user_service/handlers"
	"user_service/security"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	log.Println("Starting User Service...")
	app := fiber.New()
	defer db.GetInstance().Close()
	security.GenerateSigningKey()

	// Add CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "https://127.0.0.1:8080, https://localhost:8080, https://127.0.0.1:1234, https://localhost:1234",
		AllowHeaders: `Accept
			, Accept-Encoding
			, Accept-Language
			, Access-Control-Request-Headers
			, Access-Control-Request-Method
			, Connection
			, Host
			, Origin
			, Referer
			, Sec-Fetch-Dest
			, Sec-Fetch-Mode
			, Sec-Fetch-Site
			, User-Agent
			, Content-Type
			, Content-Length
			, Authorization
			, Username`,
		AllowCredentials: true,
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
	app.Get("/users", handlers.GetUsers)
	app.Get("/user/:username", handlers.GetUser)
	app.Get("/validate", handlers.Validate)

	app.Post("/logout", handlers.Logout)
	app.Post("/refresh/:username", handlers.RefreshToken)
	app.Post("/user", handlers.CreateUser)

	log.Fatal(app.ListenTLS(":4321", "./secrets/cert.crt", "./secrets/key.key"))
}
