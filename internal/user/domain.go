package user

import (
	"backend/internal/domain"
	auth "backend/internal/pkg/auth"
	"time"
)

type UserType string

const (
	UserTypeAdmin      UserType = "admin"
	UserTypeBuyer      UserType = "buyer"
	UserTypeFieldAgent UserType = "field_agent"
	UserTypeStaff      UserType = "staff"
)

const (
	UserStatusActive          = "active"
	UserStatusPendingApproval = "pending_approval"
	UserStatusRejected        = "rejected"
)

type User struct {
	ID           int64     `json:"id"`
	TenantID     string    `json:"tenant_id"`
	Name         string    `json:"name"`
	Phone        string    `json:"phone"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Role         UserType  `json:"role"`
	Status       string    `json:"status"`
	Metadata     []byte    `json:"metadata"`
	Address      []byte    `json:"address"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Claims = auth.Claims

type Session struct {
	Token     string    `json:"token"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	PIN      string `json:"pin" binding:"required"`
	TenantID string `json:"tenant_id"`
}

type LoginResponse struct {
	Token    string   `json:"token"`
	UserType UserType `json:"user_type"`
	UserID   int64    `json:"user_id"`
}

type AuthCheckRequest struct {
	Phone    string `json:"phone" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

type AuthCheckResponse struct {
	IsRegistered bool `json:"is_registered"`
}

type AuthVerifyRequest struct {
	Phone    string `json:"phone" binding:"required"`
	PIN      string `json:"pin" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

type AuthVerifyResponse struct {
	IsValid bool `json:"is_valid"`
}

type GetRoleRequest struct {
	Phone    string `json:"phone" binding:"required"`
	TenantID string `json:"tenant_id" binding:"required"`
}

type GetRoleResponse struct {
	Role UserType `json:"role"`
}

type SignupRequest struct {
	Name     string          `json:"name" binding:"required"`
	Phone    string          `json:"phone" binding:"required"`
	ShopName string          `json:"shop_name" binding:"required"`
	Address  *domain.Address `json:"address" binding:"required"`
	TenantID string          `json:"tenant_id" binding:"required"`
}

type SignupResponse struct {
	UserID string `json:"user_id"`
}

type CreateInternalUserRequest struct {
	Name  string   `json:"name" binding:"required"`
	Phone string   `json:"phone" binding:"required"`
	Role  UserType `json:"role" binding:"required,oneof=staff field_agent"`
	PIN   string   `json:"pin" binding:"required"`
}

type UpdateUserRequest struct {
	Name  *string `json:"name"`
	Phone *string `json:"phone"`
	PIN   *string `json:"pin"`
}

type RoleLookupResult struct {
	Role      UserType `json:"role"`
	AdminName string   `json:"admin_name"`
}

type RoleLookupRequest struct {
	Phone string `json:"phone" binding:"required"`
}
