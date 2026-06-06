package domain

type MediaType string

const (
	MediaTypeUnspecified MediaType = "unspecified"
	MediaTypeImage       MediaType = "image"
	MediaTypeVideo       MediaType = "video"
	MediaTypeAudio       MediaType = "audio"
	MediaTypeDocument    MediaType = "document"
	MediaTypeother       MediaType = "other"
)

type Media struct {
	Id        string
	Url       string
	Type      MediaType
	CreatedAt int64
	UpdatedAt int64
}

type Unit string

const (
	UnitKG Unit = "kg"
	UnitG  Unit = "g"
	UnitMG Unit = "mg"
	UnitL  Unit = "l"
	UnitML Unit = "ml"
	UnitKL Unit = "kl"
)

type PhysicalAttributes struct {
	Color string
	Type  string
}

type Product struct {
	Id                 string
	Name               string
	Quantity           int64
	Category           string
	Unit               *Unit
	Price              int64
	Description        string
	Image              []*Media
	CreatedAt          int64
	UpdatedAt          int64
	PhysicalAttributes *PhysicalAttributes
}
