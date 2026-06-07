package user

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ── User ─────────────────────────────────────────────────────────────────────

func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (id, phone, pin_hash, type, wholesaler_id, dealer_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, u.ID, u.Phone, u.PinHash, u.Type, u.WholesalerID, u.DealerID, u.CreatedAt)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

func (r *Repository) GetUserByPhone(ctx context.Context, phone string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, phone, pin_hash, type, wholesaler_id, dealer_id, created_at
		FROM users WHERE phone = $1
	`, phone).Scan(&u.ID, &u.Phone, &u.PinHash, &u.Type, &u.WholesalerID, &u.DealerID, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return u, nil
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, phone, pin_hash, type, wholesaler_id, dealer_id, created_at
		FROM users WHERE id = $1
	`, id).Scan(&u.ID, &u.Phone, &u.PinHash, &u.Type, &u.WholesalerID, &u.DealerID, &u.CreatedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// ── Session ───────────────────────────────────────────────────────────────────

func (r *Repository) CreateSession(ctx context.Context, s *Session) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO sessions (token, user_id, expires_at)
		VALUES ($1, $2, $3)
	`, s.Token, s.UserID, s.ExpiresAt)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

// GetSessionWithUser does a single join so middleware only hits DB once per request.
func (r *Repository) GetSessionWithUser(ctx context.Context, token string) (*Claims, error) {
	c := &Claims{}
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.phone, u.type, u.wholesaler_id, u.dealer_id
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1
		  AND s.expires_at > NOW()
	`, token).Scan(&c.UserID, &c.Phone, &c.Type, &c.WholesalerID, &c.DealerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // invalid or expired
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return c, nil
}

func (r *Repository) DeleteSession(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, token)
	return err
}

// ── Invite Tokens ─────────────────────────────────────────────────────────────

func (r *Repository) CreateInviteToken(ctx context.Context, token, wholesalerID string, expiresAt time.Time) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO invite_tokens (token, wholesaler_id, expires_at, used)
		VALUES ($1, $2, $3, false)
	`, token, wholesalerID, expiresAt)
	return err
}

// ClaimInviteToken returns the wholesaler_id and marks token as used atomically.
func (r *Repository) ClaimInviteToken(ctx context.Context, token string) (string, error) {
	var wholesalerID string
	err := r.db.QueryRow(ctx, `
		UPDATE invite_tokens
		SET used = true
		WHERE token = $1
		  AND used = false
		  AND expires_at > NOW()
		RETURNING wholesaler_id
	`, token).Scan(&wholesalerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("invite token invalid or expired")
		}
		return "", fmt.Errorf("claim invite token: %w", err)
	}
	return wholesalerID, nil
}

// DeleteUserByIDAndWholesaler hard-deletes a salesman or staff row, scoped to the wholesaler.
func (r *Repository) DeleteUserByIDAndWholesaler(ctx context.Context, id, wholesalerID string, userType UserType) error {
	tag, err := r.db.Exec(ctx, `
        DELETE FROM users
        WHERE id = $1
          AND wholesaler_id = $2
          AND type = $3
    `, id, wholesalerID, userType)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("not found or not yours")
	}
	return nil
}

// UpdateDealer updates phone and/or pin_hash.
// If wholesalerID is non-empty it is added to the WHERE clause (wholesaler editing a dealer).
// If wholesalerID is empty the update is by ID only (dealer editing themselves).
func (r *Repository) UpdateDealer(ctx context.Context, id, wholesalerID string, phone, pinHash *string) error {
	// Build SET clause dynamically — only touch what was provided.
	setClauses := []string{}
	args := []any{}
	argN := 1

	if phone != nil {
		setClauses = append(setClauses, fmt.Sprintf("phone = $%d", argN))
		args = append(args, *phone)
		argN++
	}
	if pinHash != nil {
		setClauses = append(setClauses, fmt.Sprintf("pin_hash = $%d", argN))
		args = append(args, *pinHash)
		argN++
	}

	// id is always the first WHERE condition
	args = append(args, id)
	where := fmt.Sprintf("id = $%d AND type = 'dealer'", argN)
	argN++

	if wholesalerID != "" {
		args = append(args, wholesalerID)
		where += fmt.Sprintf(" AND wholesaler_id = $%d", argN)
	}

	q := fmt.Sprintf(
		"UPDATE users SET %s WHERE %s",
		strings.Join(setClauses, ", "),
		where,
	)

	tag, err := r.db.Exec(ctx, q, args...)
	if err != nil {
		return fmt.Errorf("update dealer: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("not found or not yours")
	}
	return nil
}

// Add this method to internal/user/repository.go

func (r *Repository) GetRoleByPhone(ctx context.Context, phone string) (*RoleLookupResult, error) {
	result := &RoleLookupResult{}

	err := r.db.QueryRow(ctx, `
		SELECT
			u.type,
			w.name AS wholesaler_name
		FROM users u
		LEFT JOIN users w ON w.id = u.wholesaler_id AND w.type = 'wholesaler'
		WHERE u.phone = $1
	`, phone).Scan(&result.Role, &result.WholesalerName)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // phone not found
		}
		return nil, fmt.Errorf("get role by phone: %w", err)
	}

	return result, nil
}
