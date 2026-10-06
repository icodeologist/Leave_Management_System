package auth

import (
	"time"

	"leave-management/internal/model"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID int64      `json:"user_id"`
	Role   model.Role `json:"role"`
	jwt.RegisteredClaims
}

func GenerateToken(user model.User, secret string, expiry time.Duration) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID: user.ID,
		Role:   user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}
