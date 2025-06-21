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
	"user_service/util"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/google/uuid"
)

func AuthorizationMiddleware(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("%s | %s\n", util.GetFunctionName(AuthorizationMiddleware), txid.String())

	claims, err := security.ValidateJWT(c)
	if err != nil {
		log.Printf("Failed to Validate JWT\n%s\n", err.Error())
		err_string := fmt.Sprintf("Unauthorized: %s\n", txid.String())
		return c.Status(fiber.StatusInternalServerError).SendString(err_string)
	}
	user, err := security.GetUser(claims["user"].(string))
	if err != nil {
		log.Printf("Failed to get User\n%s\n", err.Error())
		err_string := fmt.Sprintf("Internal Server Error: %s\n", txid.String())
		return c.Status(fiber.StatusInternalServerError).SendString(err_string)
	}
	c.Locals("user", user)
	return c.Next()
}

// TODO this belongs somewhere else
func deleteExpiredUserSessions(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		database, err := db.GetInstance()
		if err != nil {
			log.Printf("Failed to connect to DB\n%s\n", err.Error())
		}
		log.Printf("Deleting expired sessions")
		query := `DELETE FROM sessions WHERE session_expiration < UTC_TIMESTAMP()`
		result, err := database.Exec(query)
		if err != nil {
			log.Print(err.Error())
		}
		rows_affected, err := result.RowsAffected()
		if err != nil {
			log.Print(err.Error())
		}
		log.Printf("Deleted %d expired sessions\n", rows_affected)
	}
}

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
		log.Print(err.Error())
	} else {
		defer database.Close()
	}
	security.GenerateSigningKey()

	// Start Workers
	// We can just use the login expiration time here because we
	// refresh tokens twice as fast as they expire
	go deleteExpiredUserSessions(time.Duration(config.App.LoginExpirationMs) * time.Millisecond)

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
	app.Get("/users", AuthorizationMiddleware, handlers.GetUsers)
	app.Get("/user/:username", AuthorizationMiddleware, handlers.GetUser)
	app.Get("/validate", AuthorizationMiddleware, handlers.Validate)

	app.Post("/logout", AuthorizationMiddleware, handlers.Logout)
	app.Post("/refresh/:username", AuthorizationMiddleware, handlers.RefreshToken(config))
	app.Post("/user", AuthorizationMiddleware, handlers.CreateUser)
	app.Put("/user", AuthorizationMiddleware, handlers.UpdatePassword)

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
