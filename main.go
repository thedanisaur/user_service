package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"user_service/db"
	"user_service/handlers"
	"user_service/security"
	"user_service/types"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func loadConfig(config_path string) types.Config {
	var config types.Config
	config_file, err := os.Open(config_path)
	defer config_file.Close()
	if err != nil {
		fmt.Println(err.Error())
	}
	jsonParser := json.NewDecoder(config_file)
	jsonParser.Decode(&config)
	return config
}

func main() {
	log.Println("Starting User Service...")
	config := loadConfig("./config.json")
	app := fiber.New()
	defer db.GetInstance().Close()
	security.GenerateSigningKey()

	// Add CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Join(config.App.Cors.AllowOrigins, ","),
		AllowHeaders:     strings.Join(config.App.Cors.AllowHeaders, ","),
		AllowCredentials: config.App.Cors.AllowCredentials,
	}))

	// Add Rate Limiter
	var middleware limiter.LimiterHandler
	if config.App.Limiter.LimiterSlidingMiddleware {
		middleware = limiter.SlidingWindow{}
	} else {
		middleware = limiter.FixedWindow{}
	}
	app.Use(limiter.New(limiter.Config{
		Max:                    config.App.Limiter.Max,
		Expiration:             time.Duration(config.App.Limiter.Expiration),
		LimiterMiddleware:      middleware,
		SkipSuccessfulRequests: config.App.Limiter.SkipSuccessfulRequests,
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

	port := fmt.Sprintf(":%d", config.App.Host.Port)
	err := app.ListenTLS(port, config.App.Host.CertificatePath, config.App.Host.KeyPath)
	if err != nil {
		log.Fatal(err.Error())
	}
}
