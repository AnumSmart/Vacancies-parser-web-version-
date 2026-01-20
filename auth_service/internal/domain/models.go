// описание общих стркутур для всего auth_service
package domain

import "time"

// структура пользователя
type User struct {
	ID           string // ID пользователя
	Email        string // адресс электронной почты
	PasswordHash string // хэш пароля
	CreatedAt    time.Time
	UpdatedAt    *time.Time // может быть nil
}

// структура для jwt токенов
type TokenPair struct {
	AccessToken  string
	RefreshToken string
}
