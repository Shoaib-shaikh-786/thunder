package domain

type Address struct {
	AddressId string
	Line1     string
	Line2     string
	City      string
	State     string
	Country   string
	Zip       string
	Latitude  string
	Longitude string
	IsDefault string
	Recipient *Recipient
}

type Recipient struct {
	FirstName string
	LastName  string
	Email     string
	Phone     string
}

type Customer struct {
	ID        string
	FirstName string
	LastName  string
	Phone     string
	Email     string
	CreatedAt string
	UpdatedAt string
	LegalInfo string
	Deleted   bool
	DeletedAt int64
	Address   []*Address
}
