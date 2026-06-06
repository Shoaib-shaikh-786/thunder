package auth

import (
	"testing"
	"time"

	"backend/internal/common/user"
)

func TestJWTAuthServiceGenerateAndValidateToken(t *testing.T) {
	service, err := NewJWTAuthService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTAuthService returned error: %v", err)
	}

	token, err := service.GenerateToken(&user.User{
		ID:       "user-123",
		Email:    "shoaib@example.com",
		TenantID: "tenant-456",
	})
	if err != nil {
		t.Fatalf("GenerateToken returned error: %v", err)
	}
	if token == "" {
		t.Fatal("GenerateToken returned an empty token")
	}

	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken returned error: %v", err)
	}

	if claims.UserId != "user-123" {
		t.Fatalf("claims.UserId = %q, want %q", claims.UserId, "user-123")
	}
	if claims.Email != "shoaib@example.com" {
		t.Fatalf("claims.Email = %q, want %q", claims.Email, "shoaib@example.com")
	}
	if claims.TenantID != "tenant-456" {
		t.Fatalf("claims.TenantID = %q, want %q", claims.TenantID, "tenant-456")
	}
}

func TestJWTAuthServiceRejectsInvalidSetup(t *testing.T) {
	if _, err := NewJWTAuthService("", time.Hour); err == nil {
		t.Fatal("NewJWTAuthService accepted an empty secret")
	}

	if _, err := NewJWTAuthService("test-secret", 0); err == nil {
		t.Fatal("NewJWTAuthService accepted a zero ttl")
	}
}

func TestJWTAuthServiceRejectsInvalidUser(t *testing.T) {
	service, err := NewJWTAuthService("test-secret", time.Hour)
	if err != nil {
		t.Fatalf("NewJWTAuthService returned error: %v", err)
	}

	if _, err := service.GenerateToken(nil); err == nil {
		t.Fatal("GenerateToken accepted a nil user")
	}

	if _, err := service.GenerateToken(&user.User{}); err == nil {
		t.Fatal("GenerateToken accepted a user without an id")
	}
}
