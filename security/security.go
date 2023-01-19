package security

import (
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
)

var SIGNING_KEY = []byte("samplekey")

func GenerateJWT(username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	// TODO shorten time
	claims["expires_at"] = time.Now().Add(24 * time.Hour)
	claims["issued_at"] = time.Now()
	claims["authorized"] = true
	claims["user"] = username
	// TODO replace the sample key ya dingus
	signed_token, err := token.SignedString(SIGNING_KEY)
	if err != nil {
		return "", err
	}

	return signed_token, nil
}

func GetBasicAuth(auth string) (string, string, bool, error) {
	// Basically copied from gofiber/basicauth/main.go
	// Check if header is valid
	if len(auth) > 6 && strings.ToLower(auth[:5]) == "basic" {
		// Try to decode
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		if err == nil {
			// Convert to string
			credentials := string(raw)
			// Find semicolumn
			for i := 0; i < len(credentials); i++ {
				if credentials[i] == ':' {
					// Split into user & pass
					username := credentials[:i]
					password := credentials[i+1:]
					return username, password, true, nil
				}
			}
		}
		return "", "", false, err
	}
	return "", "", false, errors.New("Invalid header")
}

func ValidateJWT(c *fiber.Ctx) error {
	token := c.Get(fiber.HeaderAuthorization)
	if strings.HasPrefix(token, "Bearer ") {
		token = strings.TrimPrefix(token, "Bearer ")
		token, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
			_, ok := t.Method.(*jwt.SigningMethodHMAC)
			if !ok {
				return "", errors.New("Invalid signing method")
			}
			return SIGNING_KEY, nil
		})
		if err != nil || !token.Valid {
			log.Printf(err.Error())
			return errors.New("Invalid JWT")
		}
	} else {
		return errors.New("Invalid credentials")
	}
	return nil
}
