package handlers

import (
	"fmt"
	"log"
	"user_service/db"
	"user_service/security"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func RefreshToken(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("RefreshToken: %s\n", txid.String())
	username := c.Get("Username")
	if security.ValidateJWT(c) != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(fmt.Sprintf("Unauthorized: %s\n", txid.String()))
	}
	token, err := security.GenerateJWT(txid, username)
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

func Validate(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("Validate: %s\n", txid.String())
	username := c.Get("Username")
	if security.ValidateJWT(c) != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(fmt.Sprintf("Unauthorized: %s\n", txid.String()))
	}
	response := fiber.Map{
		"txid":     txid.String(),
		"username": username,
		"token":    c.Get(fiber.HeaderAuthorization),
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func Login(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("Login: %s\n", txid.String())
	err_string := fmt.Sprintf("Unauthorized: %s\n", txid.String())
	username, password, has_auth, err := security.GetBasicAuth(c.Get(fiber.HeaderAuthorization))
	if has_auth && err == nil {
		database := db.GetInstance()
		query_string := `
			SELECT person_password
			FROM people
			WHERE person_username = LOWER(?)
		`
		row := database.QueryRow(query_string, username)
		var user_password string
		err := row.Scan(&user_password)
		if err != nil {
			log.Printf("Invalid username: %s\n", err.Error())
			return c.Status(fiber.StatusUnauthorized).SendString(err_string)
		}
		err = bcrypt.CompareHashAndPassword([]byte(user_password), []byte(password))
		if err != nil {
			log.Printf("Invalid password: %s\n", err.Error())
			return c.Status(fiber.StatusUnauthorized).SendString(err_string)
		}
		token, err := security.GenerateJWT(txid, username)
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
