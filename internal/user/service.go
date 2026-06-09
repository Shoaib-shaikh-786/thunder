package user

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// Login handles salesman, staff, and wholesaler login (phone + PIN).
func (s *Service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	user, err := s.repo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("login: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid phone or PIN")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PinHash), []byte(req.PIN)); err != nil {
		return nil, fmt.Errorf("invalid phone or PIN")
	}

	token, err := generateSecureToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session token: %w", err)
	}

	session := &Session{
		Token:     token,
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days persistent login
	}

	if err := s.repo.CreateSession(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		Token:    token,
		UserType: user.Type,
		UserID:   user.ID,
	}, nil
}

// Logout deletes the session from DB — instant revocation.
func (s *Service) Logout(ctx context.Context, token string) error {
	return s.repo.DeleteSession(ctx, token)
}

// GenerateInvite creates a QR invite token for a wholesaler to share with a dealer.
func (s *Service) GenerateInvite(ctx context.Context, wholesalerID string) (*InviteResponse, error) {
	token, err := generateSecureToken()
	if err != nil {
		return nil, err
	}

	expiresAt := time.Now().Add(24 * time.Hour) // QR valid for 24h
	if err := s.repo.CreateInviteToken(ctx, token, wholesalerID, expiresAt); err != nil {
		return nil, fmt.Errorf("failed to create invite: %w", err)
	}

	return &InviteResponse{
		InviteToken: token,
		ExpiresAt:   expiresAt.Format(time.RFC3339),
	}, nil
}

// DealerJoin registers a dealer using a wholesaler's QR invite token.
func (s *Service) DealerJoin(ctx context.Context, req DealerJoinRequest) (*LoginResponse, error) {
	// Claim the invite token (atomic — marks used=true)
	wholesalerID, err := s.repo.ClaimInviteToken(ctx, req.InviteToken)
	if err != nil {
		return nil, err
	}

	// Check phone not already registered
	existing, err := s.repo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("phone already registered")
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(req.PIN), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash PIN: %w", err)
	}

	dealer := &User{
		ID:           uuid.New().String(),
		Phone:        req.Phone,
		PinHash:      string(pinHash),
		Type:         UserTypeDealer,
		WholesalerID: wholesalerID,
		DealerID:     "",
		CreatedAt:    time.Now(),
	}

	if err := s.repo.CreateUser(ctx, dealer); err != nil {
		return nil, fmt.Errorf("failed to create dealer: %w", err)
	}

	// Auto login after registration
	loginReq := LoginRequest{Phone: req.Phone, PIN: req.PIN}
	return s.Login(ctx, loginReq)
}

// ValidateToken is used by auth middleware — single DB lookup returns full claims.
func (s *Service) ValidateToken(ctx context.Context, token string) (*Claims, error) {
	claims, err := s.repo.GetSessionWithUser(ctx, token)
	if err != nil {
		return nil, err
	}
	if claims == nil {
		return nil, fmt.Errorf("invalid or expired token")
	}
	return claims, nil
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func generateSecureToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// createUser is the shared internal factory for salesman and staff.
func (s *Service) createUserUnderWholesaler(
	ctx context.Context,
	wholesalerID string,
	phone, pin string,
	userType UserType,
) (*LoginResponse, error) {
	existing, err := s.repo.GetUserByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, fmt.Errorf("phone already registered")
	}

	pinHash, err := bcrypt.GenerateFromPassword([]byte(pin), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash PIN: %w", err)
	}

	u := &User{
		ID:           uuid.New().String(),
		Phone:        phone,
		PinHash:      string(pinHash),
		Type:         userType,
		WholesalerID: wholesalerID,
		CreatedAt:    time.Now(),
	}
	if err := s.repo.CreateUser(ctx, u); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &LoginResponse{UserID: u.ID, UserType: u.Type}, nil
}

func (s *Service) CreateSalesman(ctx context.Context, wholesalerID string, req CreateSalesmanRequest) (*LoginResponse, error) {
	return s.createUserUnderWholesaler(ctx, wholesalerID, req.Phone, req.PIN, UserTypeSalesman)
}

func (s *Service) CreateStaff(ctx context.Context, wholesalerID string, req CreateStaffRequest) (*LoginResponse, error) {
	return s.createUserUnderWholesaler(ctx, wholesalerID, req.Phone, req.PIN, UserTypeStaff)
}

// DeleteUser hard-deletes a salesman or staff, enforcing wholesaler ownership.
func (s *Service) DeleteUser(ctx context.Context, wholesalerID, targetID string, expectedType UserType) error {
	return s.repo.DeleteUserByIDAndWholesaler(ctx, targetID, wholesalerID, expectedType)
}

// UpdateDealer updates phone and/or PIN.
// Wholesalers must own the dealer; dealers can only touch themselves (enforced in handler).
func (s *Service) UpdateDealer(ctx context.Context, claims *Claims, targetID string, req UpdateDealerRequest) error {
	// For wholesalers, scope the update to their dealer. For dealers updating
	// themselves, wholesalerID scoping is implicit (targetID == claims.UserID).
	wholesalerID := ""
	if claims.Type == string(UserTypeWholesaler) {
		wholesalerID = claims.UserID
	}

	var newHash *string
	if req.PIN != nil {
		h, err := bcrypt.GenerateFromPassword([]byte(*req.PIN), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("failed to hash PIN: %w", err)
		}
		s := string(h)
		newHash = &s
	}

	return s.repo.UpdateDealer(ctx, targetID, wholesalerID, req.Phone, newHash)
}

func (s *Service) GetRoleByPhone(ctx context.Context, req RoleLookupRequest) (*RoleLookupResult, error) {
	result, err := s.repo.GetRoleByPhone(ctx, req.Phone)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, fmt.Errorf("no account found for this phone number")
	}
	return result, nil
}
