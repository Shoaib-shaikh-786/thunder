package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Claims is the data attached to the gin context after token validation.
// Mirrors user.Claims but kept here so middleware has no import cycle.
type Claims struct {
	UserID      int64    `json:"user_id"`
	Phone       string   `json:"phone"`
	TenantID    string   `json:"tenant_id"`
	Role        string   `json:"role"`
	Permissions []string `json:"permissions,omitempty"`
}

// SessionValidator is the function signature the auth service calls to
// look up a token. Injected from user.Service to avoid import cycles.
type SessionValidator func(ctx context.Context, token string) (*Claims, error)

// BaseAuthService is the interface the middleware depends on.
type BaseAuthService interface {
	ValidateToken(token string) (*Claims, error)
}

// SessionAuthService implements BaseAuthService using a DB session lookup.
type SessionAuthService struct {
	validate SessionValidator
}

func NewSessionAuthService(validator SessionValidator) *SessionAuthService {
	return &SessionAuthService{validate: validator}
}

func (s *SessionAuthService) ValidateToken(token string) (*Claims, error) {
	claims, err := s.validate(context.Background(), token)
	if err != nil {
		return nil, err
	}
	if claims == nil {
		return nil, fmt.Errorf("invalid or expired token")
	}
	return claims, nil
}

// EncodeClaimsToToken encodes claims into a base64-encoded JSON string.
// This can be used as a token without needing API validation.
func EncodeClaimsToToken(claims *Claims) (string, error) {
	if claims == nil {
		return "", fmt.Errorf("claims cannot be nil")
	}

	// Marshal claims to JSON
	jsonData, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claims: %w", err)
	}

	// Encode to base64
	token := base64.StdEncoding.EncodeToString(jsonData)
	return token, nil
}

// DecodeClaimsFromToken decodes a base64-encoded claims token back to Claims struct.
// This can be used to validate tokens without hitting the database.
func DecodeClaimsFromToken(token string) (*Claims, error) {
	if token == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	// Decode from base64
	jsonData, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token: %w", err)
	}

	// Unmarshal JSON to claims
	var claims Claims
	err = json.Unmarshal(jsonData, &claims)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal claims: %w", err)
	}

	// Validate required fields
	if claims.UserID == 0 || claims.TenantID == "" {
		return nil, fmt.Errorf("invalid claims: missing required fields")
	}

	return &claims, nil
}
