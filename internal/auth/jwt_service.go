package auth

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GetSecretKey returns the JWT secret key from environment variables
func GetSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return []byte("your-default-secret-key-change-this")
	}
	return []byte(secret)
}

func GenerateToken(userID uint) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(), // 1 day
	})

	return token.SignedString(GetSecretKey())
}
