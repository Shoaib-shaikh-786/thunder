package user

import (
    auth "backend/internal/pkg/auth"
    "backend/internal/domain"
    "time"
)

type UserType string

const (
    UserTypeWholesaler UserType = "wholesaler"
    UserTypeDealer     UserType = "dealer"
    UserTypeSalesman   UserType = "salesman"
    UserTypeStaff      UserType = "staff"
)

const (
    UserStatusActive          = "active"
    UserStatusPendingApproval = "pending_approval"
    UserStatusRejected        = "rejected"
)

type User struct {
    ID           string    `json:"id"`
    TenantID     string    `json:"tenant_id"`
    Name         string    `json:"name"`
    Phone        string    `json:"phone"`
    PinHash      string    `json:"-"`
    Type         UserType  `json:"type"`
    WholesalerID string    `json:"wholesaler_id"`
    DealerID     string    `json:"dealer_id"`
    Status       string    `json:"status"`
    Metadata     []byte    `json:"metadata"`
    Address      []byte    `json:"address"`
    CreatedAt    time.Time `json:"created_at"`
}

type Claims = auth.Claims

type Session struct {
    Token     string    `json:"token"`
    UserID    string    `json:"user_id"`
    ExpiresAt time.Time `json:"expires_at"`
}

type LoginRequest struct {
    Phone    string `json:"phone" binding:"required"`
    PIN      string `json:"pin" binding:"required"`
    TenantID string `json:"tenant_id"`
}

type LoginResponse struct {
    Token       string   `json:"token"`
    UserType    UserType `json:"user_type"`
    UserID      string   `json:"user_id"`
    Permissions []string `json:"permissions"`
}

type AuthCheckRequest struct {
    Phone    string `json:"phone" binding:"required"`
    TenantID string `json:"tenant_id" binding:"required"`
}

type AuthCheckResponse struct {
    AuthPath string   `json:"auth_path"`
    Role     UserType `json:"role"`
    Status   string   `json:"status"`
}

type AuthVerifyRequest struct {
    Phone    string `json:"phone" binding:"required"`
    Password string `json:"password" binding:"required"`
    TenantID string `json:"tenant_id" binding:"required"`
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
    Role  UserType `json:"role" binding:"required,oneof=staff salesman"`
    PIN   string   `json:"pin" binding:"required"`
}

type InviteRequest struct {
    // wholesaler_id is taken from the token claims, not the body
}

type InviteResponse struct {
    InviteToken string `json:"invite_token"`
    ExpiresAt   string `json:"expires_at"`
}

type DealerJoinRequest struct {
    InviteToken string `json:"invite_token" binding:"required"`
    Phone       string `json:"phone" binding:"required"`
    PIN         string `json:"pin" binding:"required"`
}

type CreateStaffRequest struct {
    Phone string `json:"phone" binding:"required"`
    PIN   string `json:"pin" binding:"required"`
}

// Reuse CreateStaffRequest for salesman too — same shape.
type CreateSalesmanRequest = CreateStaffRequest

type UpdateDealerRequest struct {
    Phone *string `json:"phone"`
    PIN   *string `json:"pin"`
}

type RoleLookupResult struct {
    Role           UserType `json:"role"`
    WholesalerName string   `json:"wholesaler_name"`
}

type RoleLookupRequest struct {
    Phone string `json:"phone" binding:"required"`
}
