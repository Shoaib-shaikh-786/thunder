package auth

import "github.com/golang-jwt/jwt/v5"

type Claim struct {
	UserId   string `json:"user_id"`
	Email    string `json:"email"`
	TenantID string `json:"tenant_id,omitempty"`
	jwt.RegisteredClaims
}
