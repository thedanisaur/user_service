package handlers

import (
	"fmt"
	"log"
	"user_service/db"
	"user_service/security"
	"user_service/types"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

func CreateUser(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("CreateUser | %s\n", txid.String())
	if security.ValidateJWT(c) != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(fmt.Sprintf("Unauthorized: %s\n", txid.String()))
	}
	var user types.User
	err := c.BodyParser(&user)
	if err != nil {
		log.Printf("Failed to parse user data\n%s\n", err.Error())
		return c.Status(fiber.StatusBadRequest).SendString(fmt.Sprintf("Failed to parse user data: %s\n", txid.String()))
	}
	hashed_password, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		log.Printf("Failed to hash password\n%s\n", err.Error())
		err_string := fmt.Sprintf("Internal Server Error: %s\n", txid.String())
		return c.Status(fiber.StatusInternalServerError).SendString(err_string)
	}
	database := db.GetInstance()
	query_string := `
		INSERT INTO people
		(
			person_username
			, person_password
			, person_email
			, person_created_on
		) VALUES (?, ?, ?, ?)
	`
	result, err := database.Exec(query_string, user.Username, hashed_password, user.Email, user.CreatedOn)
	if err != nil {
		log.Printf("Failed user insert\n%s\n", err.Error())
		err_string := fmt.Sprintf("Database Error: %s\n", txid.String())
		return c.Status(fiber.StatusServiceUnavailable).SendString(err_string)
	}
	id, err := result.LastInsertId()
	if err != nil {
		log.Printf("Failed retreive inserted id\n%s\n", err.Error())
		err_string := fmt.Sprintf("Database Error: %s\n", txid.String())
		return c.Status(fiber.StatusServiceUnavailable).SendString(err_string)
	}
	return c.Status(fiber.StatusOK).JSON(id)
}

func GetUser(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("GetUser | %s\n", txid.String())
	if security.ValidateJWT(c) != nil {
		return c.Status(fiber.StatusUnauthorized).SendString(fmt.Sprintf("Unauthorized: %s\n", txid.String()))
	}
	username := c.Params("username")
	database := db.GetInstance()
	query_string := `
		SELECT BIN_TO_UUID(person_id) person_id
			, person_username
			, person_password
			, person_email
			, person_created_on
		FROM people
		WHERE person_username = LOWER(?)
	`
	row := database.QueryRow(query_string, username)
	var user types.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedOn)
	if err != nil {
		log.Printf("Database Error:\n%s\n", err.Error())
		err_string := fmt.Sprintf("Database Error: %s\n", txid.String())
		return c.Status(fiber.StatusServiceUnavailable).SendString(err_string)
	}

	return c.Status(fiber.StatusOK).JSON(user)
}

func GetUsers(c *fiber.Ctx) error {
	txid := uuid.New()
	log.Printf("GetUsers | %s\n", txid.String())
	if security.ValidateJWT(c) != nil {
		err_string := fmt.Sprintf("Unauthorized: %s\n", txid.String())
		return c.Status(fiber.StatusUnauthorized).SendString(err_string)
	}
	database := db.GetInstance()
	query_string := `
		SELECT BIN_TO_UUID(person_id) person_id
			, person_username
			, person_password
			, person_email
			, person_created_on
		FROM people
	`
	rows, err := database.Query(query_string)
	if err != nil {
		log.Printf("Database Error:\n%s\n", err.Error())
		err_string := fmt.Sprintf("Database Error: %s\n", txid.String())
		return c.Status(fiber.StatusServiceUnavailable).SendString(err_string)
	}

	var users []types.User
	for rows.Next() {
		var user types.User
		err = rows.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedOn)
		if err != nil {
			log.Printf("Failed to scan row\n%s\n", err.Error())
			continue
		}
		users = append(users, user)
	}

	err = rows.Err()
	if err != nil {
		log.Println("Error scanning rows")
		err_string := fmt.Sprintf("Internal Server Error: %s\n", txid.String())
		return c.Status(fiber.StatusInternalServerError).SendString(err_string)
	}
	return c.Status(fiber.StatusOK).JSON(users)
}
