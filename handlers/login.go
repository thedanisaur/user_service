package handlers

import (
	"fmt"
	"log"
	"os"
	"strings"
	"user_service/db"
	"user_service/security"
	"user_service/types"
	"user_service/util"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func Login(config types.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		txid := uuid.New()
		log.Printf("%s | %s\n", util.GetFunctionName(Login), txid.String())
		err_string := fmt.Sprintf("Unauthorized: %s\n", txid.String())
		username, password, has_auth, err := security.GetBasicAuth(c.Get(fiber.HeaderAuthorization), config)
		if has_auth && err == nil {
			database, err := db.GetInstance()
			if err != nil {
				log.Printf("Failed to connect to DB\n%s\n", err.Error())
				err_string := fmt.Sprintf("Database Error: %s\n", txid.String())
				return c.Status(fiber.StatusInternalServerError).SendString(err_string)
			}
			query_string := `
				SELECT person_password
				FROM people
				WHERE person_username = LOWER(?)
			`
			row := database.QueryRow(query_string, username)
			var user_password string
			err = row.Scan(&user_password)
			if err != nil {
				log.Printf("Invalid username: %s\n", err.Error())
				return c.Status(fiber.StatusUnauthorized).SendString(err_string)
			}
			err = bcrypt.CompareHashAndPassword([]byte(user_password), []byte(password))
			if err != nil {
				log.Printf("Invalid password: %s\n", err.Error())
				return c.Status(fiber.StatusUnauthorized).SendString(err_string)
			}
			token, err := security.GenerateJWT(txid, username, config)
			if err != nil {
				log.Printf("Error generating jwt: %s\n", err.Error())
				return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Internal Server Error: %s\n", txid.String()))
			}
			// Return Authorized
			response := fiber.Map{
				"txid":     txid.String(),
				"username": username,
				"token":    fmt.Sprintf("Bearer %s", token),
			}
			return c.Status(fiber.StatusOK).JSON(response)
		} else {
			log.Println("Invalid credentials")
			return c.Status(fiber.StatusUnauthorized).SendString(err_string)
		}
	}
}

func Logout(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("%s | %s\n", util.GetFunctionName(Logout), txid.String())

	username := c.Get("Username")
	token := strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ")
	err := security.Logout(username, token)
	if err != nil {
		log.Println(err.Error())
		err_string := fmt.Sprintf("Unauthorized: %s\n", txid.String())
		return c.Status(fiber.StatusUnauthorized).SendString(err_string)
	}

	response := fiber.Map{
		"txid":     txid.String(),
		"username": username,
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func PublicKey(config types.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		txid := uuid.New()
		log.Printf("%s | %s\n", util.GetFunctionName(PublicKey), txid.String())
		if config.App.Host.UseTLS {
			log.Printf("Service is running TLS, no need to send public key.\n")
			return c.Status(fiber.StatusServiceUnavailable).SendString(fmt.Sprintf("Service is running TLS, just log in: %s\n", txid.String()))
		}

		key, err := os.ReadFile(config.App.Host.CertificatePath)
		if err != nil {
			log.Printf("Error reading cert file: %s\n", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Internal Server Error: %s\n", txid.String()))
		}
		response := fiber.Map{
			"txid":       txid.String(),
			"public_key": string(key),
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}
}

// func RefreshToken(c *fiber.Ctx) error {
func RefreshToken(config types.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		txid := uuid.New()
		log.Printf("%s | %s\n", util.GetFunctionName(RefreshToken), txid.String())
		username := c.Get("Username")
		token, err := security.GenerateJWT(txid, username, config)
		if err != nil {
			log.Printf("Error generating jwt: %s\n", err.Error())
			return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Internal Server Error: %s\n", txid.String()))
		}
		response := fiber.Map{
			"txid":     txid.String(),
			"username": username,
			"token":    fmt.Sprintf("Bearer %s", token),
		}
		return c.Status(fiber.StatusOK).JSON(response)
	}
}

func Validate(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("%s | %s\n", util.GetFunctionName(Validate), txid.String())
	username := c.Get("Username")
	response := fiber.Map{
		"txid":     txid.String(),
		"username": username,
		"token":    c.Get(fiber.HeaderAuthorization),
	}
	return c.Status(fiber.StatusOK).JSON(response)
}
