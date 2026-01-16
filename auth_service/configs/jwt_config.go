package configs

import (
	"time"
)

type JWTConfig struct {
	SecretAccKey    string
	SecretRefKey    string
	AccessTokenExp  time.Duration
	RefreshTokenExp time.Duration
}
