package auth

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"backend/internal/common/user"

	"github.com/golang-jwt/jwt/v5"
)

type JWTAuthService struct {
	secretKey []byte
	ttl       time.Duration
}

func NewJWTAuthService(secretKey string, ttl time.Duration) (BaseAuthService, error) {
	secretKey = strings.TrimSpace(secretKey)
	if secretKey == "" {
		return nil, errors.New("auth: JWT secret key is required")
	}
	if ttl <= 0 {
		return nil, errors.New("auth: JWT token ttl must be greater than zero")
	}

	return &JWTAuthService{
		secretKey: []byte(secretKey),
		ttl:       ttl,
	}, nil
}

func (j *JWTAuthService) GenerateToken(u *user.User) (string, error) {
	if u == nil {
		return "", errors.New("auth: user is required")
	}
	if strings.TrimSpace(u.ID) == "" {
		return "", errors.New("auth: user id is required")
	}

	now := time.Now()
	claims := &Claim{
		UserId:   u.ID,
		Email:    u.Email,
		TenantID: u.TenantID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString(j.secretKey)
	if err != nil {
		return "", fmt.Errorf("auth: failed to sign token: %w", err)
	}

	return signedToken, nil
}

// ValidateToken parses and verifies the token, returning its claims on success.
func (j *JWTAuthService) ValidateToken(tokenStr string) (*Claim, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&Claim{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("auth: unexpected signing method")
			}
			return j.secretKey, nil
		},
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
	)
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claim)
	if !ok || !token.Valid {
		return nil, errors.New("auth: invalid token claims")
	}

	return claims, nil
}
