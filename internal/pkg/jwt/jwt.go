package jwt

package user

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// jwtClaims is the internal registered-claims wrapper used by the jwt library.
// The public Claims struct is embedded so fields are serialised at the top level.
type jwtClaims struct {
	Claims
	jwt.RegisteredClaims
}

// GenerateJWT mints a signed HS256 token containing the user's claims.
// tokenTTL controls how long the token is valid (e.g. 30 * 24 * time.Hour).
func GenerateJWT(user *User, secret string, tokenTTL time.Duration) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt secret must not be empty")
	}

	now := time.Now()
	c := jwtClaims{
		Claims: Claims{
			UserID:      user.ID,
			Phone:       user.Phone,
			TenantID:    user.TenantID,
			Role:        user.Role,
			Permissions: user.Permissions,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   fmt.Sprintf("%d", user.ID),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return token.SignedString([]byte(secret))
}

// ValidateJWT parses and verifies a signed JWT, returning the embedded Claims.
// Returns an error if the token is expired, tampered with, or otherwise invalid.
func ValidateJWT(tokenStr, secret string) (*Claims, error) {
	if tokenStr == "" {
		return nil, fmt.Errorf("token is empty")
	}

	parsed, err := jwt.ParseWithClaims(
		tokenStr,
		&jwtClaims{},
		func(t *jwt.Token) (interface{}, error) {
			// Guard against algorithm-switching attacks.
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
			}
			return []byte(secret), nil
		},
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	c, ok := parsed.Claims.(*jwtClaims)
	if !ok || !parsed.Valid {
		return nil, fmt.Errorf("invalid token claims")
	}

	claims := c.Claims // copy out the public Claims
	return &claims, nil
}