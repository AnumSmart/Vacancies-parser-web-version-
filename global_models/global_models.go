package globalmodels

import (
	"errors"
	"time"
)

var (
	ErrUserAlreadyExists    = errors.New("User already exists in base")
	ErrUserWrongCredentials = errors.New("Wrong user papameters for search in base")
)

// структура опций для работы с куки
type CookieOptions struct {
	Name     string // имя куки
	Value    string // значение
	MaxAge   int    // в секундах
	Path     string // путь
	HttpOnly *bool  // nil = использовать дефолт (true)
}

// структура пользователя для авторизации и регистрации
type User struct {
	ID           int64  // ID пользователя
	Email        string // адресс электронной почты
	PasswordHash string // хэш пароля
	CreatedAt    time.Time
	UpdatedAt    *time.Time // может быть nil
}
