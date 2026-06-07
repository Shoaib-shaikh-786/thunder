package user

import "time"

type UserType string

const (
	UserTypeWholesaler UserType = "wholesaler"
	UserTypeDealer     UserType = "dealer"
	UserTypeSalesman   UserType = "salesman"
	UserTypeStaff      UserType = "staff"
)

// User represents any user in the system regardless of type.
type User struct {
	ID           string    `json:"id"`
	Phone        string    `json:"phone"`
	PinHash      string    `json:"-"` // never expose
	Type         UserType  `json:"type"`
	WholesalerID string    `json:"wholesaler_id"` // always set
	DealerID     string    `json:"dealer_id"`     // set for salesman/staff
	CreatedAt    time.Time `json:"created_at"`
}

// Session is stored in DB; the token is what the client holds.
type Session struct {
	Token     string    `json:"token"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Claims is attached to gin context after middleware validates the token.
type Claims struct {
	UserID       string   `json:"user_id"`
	Phone        string   `json:"phone"`
	Type         UserType `json:"type"`
	WholesalerID string   `json:"wholesaler_id"`
	DealerID     string   `json:"dealer_id"`
}

// ── Request / Response DTOs ───────────────────────────────────────────────────

type LoginRequest struct {
	Phone string `json:"phone" binding:"required"`
	PIN   string `json:"pin"   binding:"required"`
}

type LoginResponse struct {
	Token    string   `json:"token"`
	UserType UserType `json:"user_type"`
	UserID   string   `json:"user_id"`
}

// Used by wholesaler to invite a dealer via QR code.
type InviteRequest struct {
	// wholesaler_id is taken from the token claims, not the body
}

type InviteResponse struct {
	InviteToken string `json:"invite_token"`
	ExpiresAt   string `json:"expires_at"`
}

// Dealer self-registers using the QR invite token.
type DealerJoinRequest struct {
	InviteToken string `json:"invite_token" binding:"required"`
	Phone       string `json:"phone"        binding:"required"`
	PIN         string `json:"pin"          binding:"required"`
}

type CreateStaffRequest struct {
	Phone string `json:"phone" binding:"required"`
	PIN   string `json:"pin"   binding:"required"`
}

// Reuse CreateStaffRequest for salesman too — same shape.
type CreateSalesmanRequest = CreateStaffRequest

type UpdateDealerRequest struct {
	Phone *string `json:"phone"` // nil = no change
	PIN   *string `json:"pin"`
}

type RoleLookupResult struct {
	Role           UserType `json:"role"`
	WholesalerName string   `json:"wholesaler_name"`
}

type RoleLookupRequest struct {
	Phone string `json:"phone" binding:"required"`
}
