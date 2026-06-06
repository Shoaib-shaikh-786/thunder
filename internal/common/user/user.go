package user

type UserState string

const (
	UserStateActive  UserState = "active"
	UserStateBlocked UserState = "blocked"
	UserStatePending UserState = "pending"
)

type PermissionType string

const (
	PermissionTypeRead  PermissionType = "read"
	PermissionTypeWrite PermissionType = "write"
	PermissionTypeAdmin PermissionType = "admin"
)

type Permission struct {
	ID             string         `json:"id"`
	Name           string         `json:"name"`
	PermissionType PermissionType `json:"permission_type"`
	Module         string         `json:"module"`
	Resource       string         `json:"resource"`
	Action         string         `json:"action"`
}

type User struct {
	DbID                    string       `json:"db_id"`
	ID                      string       `json:"id"`
	Name                    string       `json:"name"`
	Email                   string       `json:"email"`
	Phone                   string       `json:"phone"`
	Password                string       `json:"password"`
	Role                    string       `json:"role"`
	Permissions             []Permission `json:"permissions"`
	State                   UserState    `json:"state"`
	TenantID                string       `json:"tenant_id"`
	VerificationToken       string       `json:"verification_token"`
	VerificationTokenExpiry int64        `json:"verification_token_expiry"`
}
