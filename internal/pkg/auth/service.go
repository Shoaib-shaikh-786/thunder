package auth

import (
	"context"
	"fmt"
)

// Claims is the data attached to the gin context after token validation.
// Mirrors user.Claims but kept here so middleware has no import cycle.
type Claims struct {
	UserID       string `json:"user_id"`
	Phone        string `json:"phone"`
	Type         string `json:"type"`
	WholesalerID string `json:"wholesaler_id"`
	DealerID     string `json:"dealer_id"`
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
