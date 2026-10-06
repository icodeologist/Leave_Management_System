package auth

import (
	"errors"
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

// verify jwt token and extract the token claims
func ValidateToken(tokenString, secret string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("invalid signing method")
		}
		return []byte(secret), nil
	}, jwt.WithExpirationRequired())
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || claims.UserID <= 0 {
		return nil, errors.New("invalid token claims")
	}
	if claims.Role != model.RoleEmployee && claims.Role != model.RoleAdmin {
		return nil, errors.New("invalid role")
	}

	return claims, nil
}
