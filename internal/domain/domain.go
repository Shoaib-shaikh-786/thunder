package domain

// ── Media ─────────────────────────────────────────────────────────────────────

type MediaType string

const (
	MediaTypeUnspecified MediaType = "unspecified"
	MediaTypeImage       MediaType = "image"
	MediaTypeVideo       MediaType = "video"
	MediaTypeAudio       MediaType = "audio"
	MediaTypeDocument    MediaType = "document"
	MediaTypeOther       MediaType = "other"
)

type Media struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	Type      MediaType `json:"type"`
	CreatedAt int64     `json:"created_at"`
	UpdatedAt int64     `json:"updated_at"`
}

// ── Unit ──────────────────────────────────────────────────────────────────────

type Unit string

const (
	UnitKG Unit = "kg"
	UnitG  Unit = "g"
	UnitMG Unit = "mg"
	UnitL  Unit = "l"
	UnitML Unit = "ml"
	UnitKL Unit = "kl"
)

// ── PhysicalAttributes ────────────────────────────────────────────────────────

type PhysicalAttributes struct {
	Color string `json:"color"`
	Type  string `json:"type"`
}

// ── Address ───────────────────────────────────────────────────────────────────
// Shared across user (saved address) and order (snapshot at order time).

type Address struct {
	Line1     string `json:"line1"`
	Line2     string `json:"line2"`
	City      string `json:"city"`
	State     string `json:"state"`
	Country   string `json:"country"`
	Zip       string `json:"zip"`
	Latitude  string `json:"latitude"`
	Longitude string `json:"longitude"`
}
