// описание общих стркутур для всего auth_service
package domain

import (
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists = errors.New("User already exists in base")
)

// структура пользователя
type User struct {
	ID           int64  // ID пользователя
	Email        string // адресс электронной почты
	PasswordHash string // хэш пароля
	CreatedAt    time.Time
	UpdatedAt    *time.Time // может быть nil
}

// TokenPair - пара токенов для клиента
type TokenPair struct {
	AccessToken  string // ← JWT токен (клиенту)
	RefreshToken string // ← случайная строка (клиенту)
}

// RefreshToken - структура для хранения В БАЗЕ ДАННЫХ
type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string // bcrypt хэш
	ExpiresAt time.Time
	CreatedAt time.Time
}
