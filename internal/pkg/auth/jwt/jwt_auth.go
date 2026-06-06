package jwt

import (
	"backend/internal/pkg/auth"
	"backend/internal/pkg/config"
)

func NewJWTAuth(cfg *config.Config) (auth.BaseAuthService, error) {
	return auth.NewJWTAuthService(
		cfg.ApplicationConfig.JWTSecretKey,
		cfg.ApplicationConfig.JWTTokenTTL,
	)
}
