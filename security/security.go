package security

import (
	"encoding/base64"
	"errors"
	"log"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var SIGNING_KEY = []byte("samplekey")
var CURRENT_JWTS = make(map[string]string)

func GenerateJWT(txid uuid.UUID, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().UTC().Unix()
	claims["exp"] = time.Now().Add(5 * time.Minute).UTC().Unix()
	claims["iss"] = txid
	claims["user"] = username
	// TODO replace the sample key ya dingus
	signed_token, err := token.SignedString(SIGNING_KEY)
	if err != nil {
		return "", err
	}

	// Store JTW
	CURRENT_JWTS[username] = signed_token

	return signed_token, nil
}

func GetBasicAuth(auth string) (string, string, bool, error) {
	// Basically copied from gofiber/basicauth/main.go
	// Check if header is valid
	if len(auth) > 6 && strings.ToLower(auth[:5]) == "basic" {
		// Try to decode
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			return "", "", false, err
		}
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
	return "", "", false, errors.New("Invalid header")
}

func ValidateJWT(c *fiber.Ctx) error {
	token := c.Get(fiber.HeaderAuthorization)
	username := c.Get("Username")
	if strings.HasPrefix(token, "Bearer ") {
		passed_claims, err := parseToken(token)
		if err != nil {
			log.Printf(err.Error())
			return errors.New("Failed to parse current token")
		}
		existing_claims, err := parseToken(CURRENT_JWTS[username])
		if err != nil {
			log.Printf(err.Error())
			return errors.New("Failed to parse existing token")
		}
		// Make sure the token is valid
		if !passed_claims.VerifyExpiresAt(time.Now().UTC().Unix(), true) ||
			!passed_claims.VerifyIssuedAt(time.Now().UTC().Unix(), true) ||
			!passed_claims.VerifyIssuer(existing_claims["iss"].(string), true) ||
			passed_claims["username"] != existing_claims["username"] {
			return errors.New("Invalid credentials")
		}
	} else {
		return errors.New("Invalid credentials")
	}
	return nil
}

func parseToken(token string) (jwt.MapClaims, error) {
	token = strings.TrimPrefix(token, "Bearer ")
	parsed_token, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		_, ok := t.Method.(*jwt.SigningMethodHMAC)
		if !ok {
			return "", errors.New("Invalid signing method")
		}
		return SIGNING_KEY, nil
	})
	if err != nil || !parsed_token.Valid {
		log.Printf(err.Error())
		return nil, errors.New("Invalid JWT")
	}

	claims, ok := parsed_token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("Missing Claims")
	}
	return claims, nil
}
