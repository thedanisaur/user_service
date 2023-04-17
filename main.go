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

func loadConfig(config_path string) (types.Config, error) {
	var config types.Config
	config_file, err := os.Open(config_path)
	if err != nil {
		return config, err
	}
	defer config_file.Close()
	jsonParser := json.NewDecoder(config_file)
	jsonParser.Decode(&config)
	return config, nil
}

func main() {
	log.Println("Starting User Service...")
	config, err := loadConfig("./config.json")
	if err != nil {
		log.Printf("Error opening config, cannot continue: %s\n", err.Error())
		return
	}
	app := fiber.New()
	database, err := db.GetInstance()
	if err != nil {
		log.Printf(err.Error())
	} else {
		defer database.Close()
	}
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
	app.Post("/login", handlers.Login(config))
	app.Get("/publickey", handlers.PublicKey(config))

	// JWT Authentication
	app.Get("/users", handlers.GetUsers)
	app.Get("/user/:username", handlers.GetUser)
	app.Get("/validate", handlers.Validate)

	app.Post("/logout", handlers.Logout)
	app.Post("/refresh/:username", handlers.RefreshToken)
	app.Post("/user", handlers.CreateUser)

	port := fmt.Sprintf(":%d", config.App.Host.Port)
	if config.App.Host.UseTLS {
		err = app.ListenTLS(port, config.App.Host.CertificatePath, config.App.Host.KeyPath)
	} else {
		log.Println("Warning - not using TLS")
		err = app.Listen(port)
	}
	if err != nil {
		log.Fatal(err.Error())
	}
}
