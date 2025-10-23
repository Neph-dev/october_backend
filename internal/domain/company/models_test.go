package company

import (
	"testing"
	"time"
)

func TestIndustryValidation(t *testing.T) {
	tests := []struct {
		name     string
		industry Industry
		want     bool
	}{
		{"Valid Defense", IndustryDefense, true},
		{"Valid Aerospace", IndustryAerospace, true},
		{"Invalid Industry", Industry("Invalid"), false},
		{"Empty Industry", Industry(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.industry.IsValid(); got != tt.want {
				t.Errorf("Industry.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateCompanyRequestToCompany(t *testing.T) {
	founded := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	
	req := &CreateCompanyRequest{
		Name:          "Test Company",
		Country:       "USA",
		Ticker:        "TEST",
		StockExchange: "NYSE",
		Industry:      IndustryDefense,
		FeedURL:       "https://example.com/feed",
		CompanyWebsite: "https://example.com",
		KeyPeople: []KeyPerson{
			{FullName: "John Doe", Position: "CEO"},
		},
		Founded:      founded,
		NumEmployees: 1000,
	}

	company := req.ToCompany()

	// Verify basic fields
	if company.Name != req.Name {
		t.Errorf("Expected name %s, got %s", req.Name, company.Name)
	}
	if company.Country != req.Country {
		t.Errorf("Expected country %s, got %s", req.Country, company.Country)
	}
	if company.Ticker != req.Ticker {
		t.Errorf("Expected ticker %s, got %s", req.Ticker, company.Ticker)
	}
	if company.Industry != req.Industry {
		t.Errorf("Expected industry %s, got %s", req.Industry, company.Industry)
	}

	// Verify metadata defaults
	if !company.Metadata.IsActive {
		t.Error("Expected metadata.IsActive to be true")
	}
	if len(company.Metadata.Tags) != 0 {
		t.Error("Expected empty tags slice")
	}

	// Verify timestamps are set
	if company.CreatedAt.IsZero() {
		t.Error("Expected CreatedAt to be set")
	}
	if company.UpdatedAt.IsZero() {
		t.Error("Expected UpdatedAt to be set")
	}
}

func TestCompanyToResponse(t *testing.T) {
	founded := time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)
	
	company := &Company{
		Name:          "Test Company",
		Country:       "USA",
		Ticker:        "TEST",
		StockExchange: "NYSE",
		Industry:      IndustryDefense,
		FeedURL:       "https://example.com/feed",
		CompanyWebsite: "https://example.com",
		KeyPeople: []KeyPerson{
			{FullName: "John Doe", Position: "CEO"},
		},
		Founded:      founded,
		NumEmployees: 1000,
		Metadata: CompanyMetadata{
			IsActive: true,
			Tags:     []string{"defense", "technology"},
		},
	}

	response := company.ToResponse()

	if response.Name != company.Name {
		t.Errorf("Expected name %s, got %s", company.Name, response.Name)
	}
	if response.Country != company.Country {
		t.Errorf("Expected country %s, got %s", company.Country, response.Country)
	}
	if len(response.KeyPeople) != len(company.KeyPeople) {
		t.Errorf("Expected %d key people, got %d", len(company.KeyPeople), len(response.KeyPeople))
	}
}