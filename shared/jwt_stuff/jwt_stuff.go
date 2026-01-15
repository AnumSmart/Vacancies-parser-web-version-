package jwt_stuff

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 🔐 коструктор для создания экземпляра JWT токена (возвращаем указатель)
func NewJWT(secretAcc string, secretRef string, accessTokenExp, refreshTokenExp time.Duration) *JWT {
	return &JWT{
		SecretAccKey:    secretAcc,
		SecretRefKey:    secretRef,
		AccessTokenExp:  accessTokenExp,
		RefreshTokenExp: refreshTokenExp,
	}
}

// метод структуры JWT для генерации токенов (access и refresh)
func (j *JWT) GenerateTokens(email string) (string, string, error) {
	// Access токен
	accessClaims := NewClaims(j.AccessTokenExp, email, "access", "my_app")
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(j.SecretAccKey))
	if err != nil {
		return "", "", err
	}

	// Refresh токен
	refreshClaims := NewClaims(j.RefreshTokenExp, email, "refresh", "my_app")
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(j.SecretRefKey))
	if err != nil {
		return "", "", err
	}

	return accessTokenString, refreshTokenString, nil
}
