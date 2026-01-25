package handlers

import (
	"errors"
	"net/http"
)

// Package-level errors
var (
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidEmail    = errors.New("invalid email format")
	ErrUserNotFound    = errors.New("user not found")
	ErrInvalidPassword = errors.New("invalid password")
	ErrTokenExpired    = errors.New("token expired")
)

type APIError struct {
	Code    string `json:"code"`    // для фронтенда: "USER_EXISTS"
	Message string `json:"message"` // для пользователя
	Field   string `json:"field,omitempty"`
}

// функция - маппер для формирования нужного результата в зависимости от типа кастомной ошибки
func ToAPIError(err error) (int, APIError) {
	switch {
	case errors.Is(err, ErrUserExists):
		return http.StatusConflict, APIError{
			Code:    "USER_EXISTS",
			Message: "This email is already registered",
		}
	case errors.Is(err, ErrInvalidEmail):
		return http.StatusBadRequest, APIError{
			Code:    "INVALID_EMAIL",
			Message: "Please provide a valid email address",
			Field:   "email",
		}
	default:
		return http.StatusInternalServerError, APIError{
			Code:    "INTERNAL_ERROR",
			Message: "Something went wrong",
		}
	}
}
