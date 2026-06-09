package user

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// CreateUser inserts a new user into the database
func (r *Repository) CreateUser(ctx context.Context, u *User) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO users (tenant_id, name, phone, email, password_hash, role, status, metadata, address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW(), NOW())
		RETURNING id
	`, u.TenantID, u.Name, u.Phone, u.Email, u.PasswordHash, u.Role, u.Status, u.Metadata, u.Address)
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}
	return nil
}

// GetUserByPhone retrieves a user by phone number and optional tenant ID
func (r *Repository) GetUserByPhone(ctx context.Context, phone, tenantID string) (*User, error) {
	u := &User{}
	query := `
		SELECT id, tenant_id, name, phone, email, password_hash, role, status, metadata, address, created_at, updated_at
		FROM users WHERE phone = $1`
	args := []any{phone}
	
	if tenantID != "" {
		query += " AND tenant_id = $2"
		args = append(args, tenantID)
	}
	
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&u.ID, &u.TenantID, &u.Name, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Status, &u.Metadata, &u.Address, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return u, nil
}

// GetUserByID retrieves a user by ID
func (r *Repository) GetUserByID(ctx context.Context, id int64) (*User, error) {
	u := &User{}
	err := r.db.QueryRow(ctx, `
		SELECT id, tenant_id, name, phone, email, password_hash, role, status, metadata, address, created_at, updated_at
		FROM users WHERE id = $1
	`, id).Scan(
		&u.ID, &u.TenantID, &u.Name, &u.Phone, &u.Email, &u.PasswordHash,
		&u.Role, &u.Status, &u.Metadata, &u.Address, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

// CreateSession inserts a new session
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

// GetSessionWithUser retrieves session and associated user claims
func (r *Repository) GetSessionWithUser(ctx context.Context, token string) (*Claims, error) {
	c := &Claims{}
	err := r.db.QueryRow(ctx, `
		SELECT u.id, u.phone, u.role, u.tenant_id
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1
		  AND s.expires_at > NOW()
	`, token).Scan(&c.UserID, &c.Phone, &c.Type, &c.TenantID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	return c, nil
}

// DeleteSession removes a session
func (r *Repository) DeleteSession(ctx context.Context, token string) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM sessions WHERE token = $1
	`, token)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

// UpdateUser updates user details
func (r *Repository) UpdateUser(ctx context.Context, id int64, u *User) error {
	_, err := r.db.Exec(ctx, `
		UPDATE users
		SET name = $1, email = $2, metadata = $3, address = $4, updated_at = NOW()
		WHERE id = $5
	`, u.Name, u.Email, u.Metadata, u.Address, id)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	return nil
}

// DeleteUser removes a user
func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `
		DELETE FROM users WHERE id = $1
	`, id)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *Repository) GetRoleByPhone(ctx context.Context, phone string) (*RoleLookupResult, error) {
	result := &RoleLookupResult{}

	err := r.db.QueryRow(ctx, `
		SELECT u.role, '' AS admin_name
		FROM users u
		WHERE u.phone = $1
	`, phone).Scan(&result.Role, &result.AdminName)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // phone not found
		}
		return nil, fmt.Errorf("get role by phone: %w", err)
	}

	return result, nil
}
