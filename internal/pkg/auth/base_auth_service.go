package auth

import (
	"backend/internal/common/user"
)

type BaseAuthService interface {
	ValidateToken(token string) (*Claim, error)
	GenerateToken(user *user.User) (string, error)
}
