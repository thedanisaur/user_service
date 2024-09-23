package security

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
	"user_service/types"
	"user_service/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var SIGNING_KEY []byte
var CURRENT_JWTS = make(map[string]string)

func decrypt(ciphertext []byte, config types.Config) ([]byte, error) {
	key_str, err := os.ReadFile(config.App.Host.KeyPath)
	if err != nil {
		log.Printf("Error reading key file: %s\n", err.Error())
		return nil, err
	}
	block, _ := pem.Decode([]byte(key_str))
	parsed_key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		log.Printf("Error parsing key: %s\n", err.Error())
		return nil, err
	}
	private_key := parsed_key.(*rsa.PrivateKey)
	if err != nil {
		log.Printf("Error parsing private key: %s\n", err.Error())
		return nil, err
	}
	plaintext, err := private_key.Decrypt(rand.Reader, ciphertext, &rsa.OAEPOptions{Hash: crypto.SHA512})
	if err != nil {
		log.Printf("Error decrypting ciphertext: %s\n", err.Error())
		return nil, err
	}
	return plaintext, nil
}

func encrypt(plaintext []byte, config types.Config) ([]byte, error) {
	key_str, err := os.ReadFile(config.App.Host.CertificatePath)
	if err != nil {
		log.Printf("Error reading cert file: %s\n", err.Error())
		return nil, err
	}
	block, _ := pem.Decode([]byte(key_str))
	parsed_key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Printf("Error parsing key: %s\n", err.Error())
		return nil, err
	}
	public_key := parsed_key.(*rsa.PublicKey)
	if err != nil {
		log.Printf("Error parsing public key: %s\n", err.Error())
		return nil, err
	}
	ciphertext, err := rsa.EncryptOAEP(sha512.New(), rand.Reader, public_key, plaintext, nil)
	if err != nil {
		log.Printf("Error encrypting plaintext: %s\n", err.Error())
		return nil, err
	}
	return ciphertext, nil
}

func GenerateJWT(txid uuid.UUID, username string) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().UTC().Unix()
	claims["exp"] = time.Now().Add(2 * time.Minute).UTC().Unix()
	claims["iss"] = txid
	// TODO tie to user agent as well
	claims["user"] = username
	signed_token, err := token.SignedString(SIGNING_KEY)
	if err != nil {
		return "", err
	}

	// Store JTW
	CURRENT_JWTS[username] = signed_token

	return signed_token, nil
}

func GenerateSigningKey() {
	SIGNING_KEY = []byte(util.RandomString(64))
}

func GetBasicAuth(auth string, config types.Config) (string, string, bool, error) {
	// Basically copied from gofiber/basicauth/main.go
	// Check if header is valid
	if len(auth) > 6 && strings.ToLower(auth[:5]) == "basic" {
		// Try to decode
		raw, err := base64.StdEncoding.DecodeString(auth[6:])
		if err != nil {
			return "", "", false, err
		}
		credentials := string(raw)
		// First check if we are using TLS
		// TODO [drd] We really don't need this
		// if !config.App.Host.UseTLS {
		// 	// We aren't using TLS so try to decrypt the auth
		// 	plaintext, err := decrypt([]byte(credentials), config)
		// 	if err != nil {
		// 		return "", "", false, err
		// 	}
		// 	credentials = string(plaintext)
		// }
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

func Logout(c *fiber.Ctx) error {
	username := c.Get("Username")
	token, ok := CURRENT_JWTS[username]
	if !ok {
		return errors.New(fmt.Sprintf("User not found: %s", username))
	}
	if token != strings.TrimPrefix(c.Get(fiber.HeaderAuthorization), "Bearer ") {
		return errors.New("Token sent doesn't match user token")
	}
	delete(CURRENT_JWTS, username)
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
