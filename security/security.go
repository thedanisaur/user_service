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
	// "fmt"
	"log"
	"os"
	"strings"
	"time"

	"user_service/db"
	"user_service/types"
	"user_service/util"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

var SIGNING_KEY []byte

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

func deleteJWT(claims jwt.MapClaims) error {
	database, err := db.GetInstance()
	if err != nil {
		log.Printf("Failed to connect to DB\n%s\n", err.Error())
	}

	issued_string := claims["iss"].(string)
	issued_uuid, _ := uuid.Parse(issued_string)
	delete_query := `
		DELETE FROM sessions
		WHERE session_issued_uuid = UUID_TO_BIN(?)
	`
	_, err = database.Exec(delete_query, issued_uuid)
	if err != nil {
		log.Printf("Failed to delete session\n%s\n", err.Error())
		return err
	}
	return nil
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

func GenerateJWT(txid uuid.UUID, username string, config types.Config) (string, error) {
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["iat"] = time.Now().UTC().Unix()
	claims["exp"] = time.Now().Add(time.Duration(config.App.LoginExpirationMs) * time.Millisecond).UTC().Unix()
	claims["iss"] = txid
	// TODO tie to user agent as well
	claims["user"] = username
	signed_token, err := token.SignedString(SIGNING_KEY)
	if err != nil {
		return "", err
	}

	// Store JTW
	storeJWT(claims, config)
	log.Printf("token | %s\n", signed_token)

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

func getJWT(issued_uuid uuid.UUID) (jwt.MapClaims, error) {
	database, err := db.GetInstance()
	if err != nil {
		log.Printf("Failed to connect to DB\n%s\n", err.Error())
		return nil, err
	}

	var session_issued_at time.Time
	var session_expiration time.Time
	var session_issued_uuid uuid.UUID
	var person_username string
	select_query := `
		SELECT session_issued_at
			, session_expiration
			, BIN_TO_UUID(session_issued_uuid)
			, person_username
		FROM sessions
		WHERE session_issued_uuid = UUID_TO_BIN(?)
	`
	err = database.QueryRow(select_query, issued_uuid).Scan(&session_issued_at, &session_expiration, &session_issued_uuid, &person_username)
	if err != nil {
		log.Printf("Session not found\n%s\n", err.Error())
		return nil, err
	}
	claims_dbo := jwt.MapClaims{
		"exp": float64(session_expiration.Unix()),
		"iat": float64(session_issued_at.Unix()),
		"iss": session_issued_uuid,
		"user": person_username,
	}
	return claims_dbo, nil
}

func Logout(username string, token string) error {
	claims, _ := parseToken(token)
	return deleteJWT(claims)
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

func storeJWT(claims jwt.MapClaims, config types.Config) error {
	database, err := db.GetInstance()
	if err != nil {
		log.Printf("Failed to connect to DB\n%s\n", err.Error())
	}

	username := claims["user"].(string)
	count_query := `SELECT COUNT(*) FROM sessions WHERE person_username = ?`
	var session_count int
	err = database.QueryRow(count_query, username).Scan(&session_count)
	if err != nil {
		log.Printf("Failed to count sessions\n%s\n", err.Error())
		return err
	}

	if session_count >= config.App.MaxSessions {
		delete_query := `
			DELETE FROM sessions
			WHERE session_id = (
				SELECT session_id FROM (
					SELECT session_id
					FROM sessions
					WHERE person_username = ?
					ORDER BY session_issued_at ASC
					LIMIT 1
				) AS oldest
			)
		`
		_, err = database.Exec(delete_query, username)
		if err != nil {
			log.Printf("Failed to delete session\n%s\n", err.Error())
			return err
		}
	}

	insert_query := `
		INSERT INTO sessions
		(
			session_issued_at
			, session_expiration
			, session_issued_uuid
			, person_username
		) VALUES (?, ?, ?, ?)
	`
	issued_at := time.Unix(claims["iat"].(int64), 0).UTC()
	expiration := time.Unix(claims["exp"].(int64), 0).UTC()
	issued_uuid := claims["iss"].(uuid.UUID)
	result, err := database.Exec(insert_query, issued_at, expiration, issued_uuid[:], username)
	if err != nil {
		log.Printf("Failed user insert\n%s\n", err.Error())
		return err
	}
	_, err = result.LastInsertId()
	if err != nil {
		log.Printf("Failed retrieve inserted id\n%s\n", err.Error())
		return err
	}
	return nil
}

func ValidateJWT(c *fiber.Ctx) (jwt.MapClaims, error) {
	token := c.Get(fiber.HeaderAuthorization)
	if strings.HasPrefix(token, "Bearer ") {
		passed_claims, err := parseToken(token)
		if err != nil {
			log.Printf(err.Error())
			return nil, errors.New("Failed to parse current token")
		}
		issued_string := passed_claims["iss"].(string)
		issued_uuid, _ := uuid.Parse(issued_string)
		stored_claims, err := getJWT(issued_uuid)
		if err != nil {
			log.Printf(err.Error())
			return nil, errors.New("Failed to parse stored token")
		}
		// Make sure the token is valid
		if !passed_claims.VerifyExpiresAt(time.Now().UTC().Unix(), true) ||
			!passed_claims.VerifyIssuedAt(time.Now().UTC().Unix(), true) ||
			!passed_claims.VerifyIssuer(stored_claims["iss"].(uuid.UUID).String(), true) ||
			passed_claims["username"] != stored_claims["username"] {
			return nil, errors.New("Invalid credentials")
		}
		return stored_claims, nil
	}
	return nil, errors.New("Invalid credentials")
}

func GetUser(username string) (types.User, error) {
	txid := uuid.New()
	log.Printf("GetUser | %s\n", txid.String())
	database, err := db.GetInstance()
	if err != nil {
		log.Printf("Failed to connect to DB\n%s\n", err.Error())
		return types.User{}, errors.New("Failed to connect to DB")
	}
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
	err = row.Scan(&user.ID, &user.Username, &user.Password, &user.Email, &user.CreatedOn)
	if err != nil {
		log.Printf("Failed to retrieve user:\n%s\n", err.Error())
		return types.User{}, errors.New("Failed to retrieve user")
	}
	return user, nil
}