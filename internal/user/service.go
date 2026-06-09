package user

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const defaultTokenTTL = 30 * 24 * time.Hour

type Service struct {
	repo      *Repository
	jwtSecret string
}

func NewService(repo *Repository, jwtSecret string) *Service {
	return &Service{repo: repo, jwtSecret: jwtSecret}
}

// Login handles phone + PIN authentication and returns a signed JWT.
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, req.Phone, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.PIN)); err != nil {
		return nil, fmt.Errorf("invalid phone or PIN")
	}

	token, err := GenerateJWT(user, s.jwtSecret, defaultTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:    token,
		UserType: user.Role,
		UserID:   user.ID,
	}, nil
}

// Logout is a no-op for stateless JWT auth.
// If you need server-side revocation, maintain a denylist here.
func (s *Service) Logout(_ context.Context, _ string) error {
	return nil
}

// CheckAuth checks if a phone number is registered.
func (s *Service) CheckAuth(ctx context.Context, req AuthCheckRequest) (*AuthCheckResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, req.Phone, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("auth check: %w", err)
	}
	return &AuthCheckResponse{IsRegistered: user != nil}, nil
}

// VerifyAuth verifies phone and PIN without issuing a token.
func (s *Service) VerifyAuth(ctx context.Context, req AuthVerifyRequest) (*AuthVerifyResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, req.Phone, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("verify auth: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.PIN)); err != nil {
		return &AuthVerifyResponse{IsValid: false}, nil
	}
	return &AuthVerifyResponse{IsValid: true}, nil
}

// GetRole retrieves the role for a phone number.
func (s *Service) GetRole(ctx context.Context, req GetRoleRequest) (*GetRoleResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, req.Phone, req.TenantID)
	if err != nil {
		return nil, fmt.Errorf("get role: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}
	return &GetRoleResponse{Role: user.Role}, nil
}

// ValidateToken is called by the auth middleware.
// It verifies the JWT signature and expiry — no DB round-trip needed.
func (s *Service) ValidateToken(_ context.Context, token string) (*Claims, error) {
	claims, err := ValidateJWT(token, s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("invalid or expired token: %w", err)
	}
	return claims, nil
}

// UpdateUser updates user information.
func (s *Service) UpdateUser(ctx context.Context, userID int64, u *User) error {
	return s.repo.UpdateUser(ctx, userID, u)
}

// DeleteUser removes a user.
func (s *Service) DeleteUser(ctx context.Context, userID int64) error {
	return s.repo.DeleteUser(ctx, userID)
}