package company

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Company represents a company entity with comprehensive business information
type Company struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name           string             `bson:"name" json:"name" validate:"required,min=1,max=200"`
	Country        string             `bson:"country" json:"country" validate:"required,min=2,max=100"`
	Ticker         string             `bson:"ticker" json:"ticker" validate:"omitempty,min=1,max=10"`
	StockExchange  string             `bson:"stockExchange" json:"stockExchange" validate:"omitempty,min=1,max=50"`
	Industry       Industry           `bson:"industry" json:"industry" validate:"required"`
	FeedURL        string             `bson:"feedUrl" json:"feedUrl" validate:"required,url"`
	CompanyWebsite string             `bson:"companyWebsite" json:"companyWebsite" validate:"required,url"`
	KeyPeople      []KeyPerson        `bson:"keyPeople" json:"keyPeople" validate:"required,min=1"`
	Founded        time.Time          `bson:"founded" json:"founded" validate:"required"`
	NumEmployees   int                `bson:"numEmployees" json:"numEmployees" validate:"required,min=1"`
	Metadata       CompanyMetadata    `bson:"metadata" json:"metadata"`
	CreatedAt      time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt      time.Time          `bson:"updatedAt" json:"updatedAt"`
}

type Industry string

const (
	IndustryDefense    Industry = "Defense"
	IndustryAerospace  Industry = "Aerospace"
	IndustryGovernment Industry = "Government"
)

func (i Industry) IsValid() bool {
	switch i {
	case IndustryDefense, IndustryAerospace, IndustryGovernment:
		return true
	default:
		return false
	}
}

// KeyPerson represents a key person in the company
type KeyPerson struct {
	FullName string `bson:"fullName" json:"fullName" validate:"required,min=2,max=100"`
	Position string `bson:"position" json:"position" validate:"required,min=2,max=100"`
}

type CompanyMetadata struct {
	LastFeedUpdate time.Time `bson:"lastFeedUpdate" json:"lastFeedUpdate"`
	IsActive       bool      `bson:"isActive" json:"isActive"`
	Tags           []string  `bson:"tags" json:"tags"`
}

// CreateCompanyRequest represents the request payload for creating a company
type CreateCompanyRequest struct {
	Name           string      `json:"name" validate:"required,min=1,max=200"`
	Country        string      `json:"country" validate:"required,min=2,max=100"`
	Ticker         string      `json:"ticker" validate:"omitempty,min=1,max=10"`
	StockExchange  string      `json:"stockExchange" validate:"omitempty,min=1,max=50"`
	Industry       Industry    `json:"industry" validate:"required"`
	FeedURL        string      `json:"feedUrl" validate:"required,url"`
	CompanyWebsite string      `json:"companyWebsite" validate:"required,url"`
	KeyPeople      []KeyPerson `json:"keyPeople" validate:"required,min=1"`
	Founded        time.Time   `json:"founded" validate:"required"`
	NumEmployees   int         `json:"numEmployees" validate:"required,min=1"`
}

// ToCompany converts a CreateCompanyRequest to a Company domain object
func (req *CreateCompanyRequest) ToCompany() *Company {
	now := time.Now()
	return &Company{
		Name:           req.Name,
		Country:        req.Country,
		Ticker:         req.Ticker,
		StockExchange:  req.StockExchange,
		Industry:       req.Industry,
		FeedURL:        req.FeedURL,
		CompanyWebsite: req.CompanyWebsite,
		KeyPeople:      req.KeyPeople,
		Founded:        req.Founded,
		NumEmployees:   req.NumEmployees,
		Metadata: CompanyMetadata{
			IsActive: true,
			Tags:     []string{},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// CompanyResponse represents the response payload for company data
type CompanyResponse struct {
	ID             string      `json:"id"`
	Name           string      `json:"name"`
	Country        string      `json:"country"`
	Ticker         string      `json:"ticker"`
	StockExchange  string      `json:"stockExchange"`
	Industry       Industry    `json:"industry"`
	FeedURL        string      `json:"feedUrl"`
	CompanyWebsite string      `json:"companyWebsite"`
	KeyPeople      []KeyPerson `json:"keyPeople"`
	Founded        time.Time   `json:"founded"`
	NumEmployees   int         `json:"numEmployees"`
	Metadata       CompanyMetadata `json:"metadata"`
}

// ToResponse converts a Company to a CompanyResponse
func (c *Company) ToResponse() *CompanyResponse {
	return &CompanyResponse{
		ID:             c.ID.Hex(),
		Name:           c.Name,
		Country:        c.Country,
		Ticker:         c.Ticker,
		StockExchange:  c.StockExchange,
		Industry:       c.Industry,
		FeedURL:        c.FeedURL,
		CompanyWebsite: c.CompanyWebsite,
		KeyPeople:      c.KeyPeople,
		Founded:        c.Founded,
		NumEmployees:   c.NumEmployees,
		Metadata:       c.Metadata,
	}
}