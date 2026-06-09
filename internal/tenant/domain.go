package tenant

import "time"

type Tenant struct {
	ID             string    `gorm:"primaryKey;type:varchar(50)"` // The unique subdomain slug e.g., "factory-a"
	CompanyName    string    `gorm:"type:varchar(150);not null"`
	BusinessType   string    `gorm:"type:varchar(30);not null"`   // factory | wholesaler | corporate
	BrandingConfig []byte    `gorm:"type:jsonb"`                  // Stores frontend primary/accent hex colors, logo URLs
	IsActive       bool      `gorm:"default:true;index"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}


