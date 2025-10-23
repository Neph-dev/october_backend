package company

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/pkg/logger"
)

type CompanyService struct {
	repo   Repository
	logger logger.Logger
}

func NewCompanyService(repo Repository, logger logger.Logger) Service {
	return &CompanyService{
		repo:   repo,
		logger: logger,
	}
}

func (s *CompanyService) CreateCompany(ctx context.Context, req *CreateCompanyRequest) (*CompanyResponse, error) {
	if err := s.validateCreateRequest(req); err != nil {
		s.logger.Error("Invalid company creation request", "error", err)
		return nil, fmt.Errorf("%w: %s", ErrInvalidCompanyData, err.Error())
	}

	// Convert request to domain object
	company := req.ToCompany()

	// Check if company already exists by name
	existing, err := s.repo.GetByName(ctx, company.Name)
	if err != nil && err != ErrCompanyNotFound {
		s.logger.Error("Failed to check existing company", "error", err, "name", company.Name)
		return nil, fmt.Errorf("failed to check existing company: %w", err)
	}
	if existing != nil {
		return nil, ErrCompanyExists
	}

	existing, err = s.repo.GetByTicker(ctx, company.Ticker)
	if err != nil && err != ErrCompanyNotFound {
		s.logger.Error("Failed to check existing ticker", "error", err, "ticker", company.Ticker)
		return nil, fmt.Errorf("failed to check existing ticker: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("%w: ticker %s already exists", ErrCompanyExists, company.Ticker)
	}

	if err := s.repo.Create(ctx, company); err != nil {
		s.logger.Error("Failed to create company", "error", err, "name", company.Name)
		return nil, fmt.Errorf("failed to create company: %w", err)
	}

	s.logger.Info("Company created successfully", "id", company.ID.Hex(), "name", company.Name)
	return company.ToResponse(), nil
}

// (case-insensitive)
func (s *CompanyService) GetCompanyByName(ctx context.Context, name string) (*CompanyResponse, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("%w: company name cannot be empty", ErrInvalidCompanyData)
	}

	company, err := s.repo.GetByName(ctx, strings.TrimSpace(name))
	if err != nil {
		if err == ErrCompanyNotFound {
			s.logger.Info("Company not found", "name", name)
		} else {
			s.logger.Error("Failed to get company by name", "error", err, "name", name)
		}
		return nil, err
	}

	return company.ToResponse(), nil
}

func (s *CompanyService) GetCompanyByTicker(ctx context.Context, ticker string) (*CompanyResponse, error) {
	if strings.TrimSpace(ticker) == "" {
		return nil, fmt.Errorf("%w: ticker cannot be empty", ErrInvalidCompanyData)
	}

	company, err := s.repo.GetByTicker(ctx, strings.ToUpper(strings.TrimSpace(ticker)))
	if err != nil {
		if err == ErrCompanyNotFound {
			s.logger.Info("Company not found", "ticker", ticker)
		} else {
			s.logger.Error("Failed to get company by ticker", "error", err, "ticker", ticker)
		}
		return nil, err
	}

	return company.ToResponse(), nil
}

// ListCompanies retrieves a paginated list of companies
func (s *CompanyService) ListCompanies(ctx context.Context, limit, offset int) ([]*CompanyResponse, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	companies, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		s.logger.Error("Failed to list companies", "error", err, "limit", limit, "offset", offset)
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}

	// Convert to response objects
	responses := make([]*CompanyResponse, len(companies))
	for i, company := range companies {
		responses[i] = company.ToResponse()
	}

	return responses, nil
}

// validateCreateRequest validates a company creation request
func (s *CompanyService) validateCreateRequest(req *CreateCompanyRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Validate required fields
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("company name is required")
	}
	if strings.TrimSpace(req.Country) == "" {
		return fmt.Errorf("country is required")
	}
	if strings.TrimSpace(req.Ticker) == "" {
		return fmt.Errorf("ticker is required")
	}
	if strings.TrimSpace(req.StockExchange) == "" {
		return fmt.Errorf("stock exchange is required")
	}
	if !req.Industry.IsValid() {
		return fmt.Errorf("invalid industry: must be %s or %s", IndustryDefense, IndustryAerospace)
	}
	if strings.TrimSpace(req.FeedURL) == "" {
		return fmt.Errorf("feed URL is required")
	}
	if strings.TrimSpace(req.CompanyWebsite) == "" {
		return fmt.Errorf("company website is required")
	}
	if len(req.KeyPeople) == 0 {
		return fmt.Errorf("at least one key person is required")
	}
	if req.NumEmployees <= 0 {
		return fmt.Errorf("number of employees must be greater than 0")
	}
	if req.Founded.IsZero() {
		return fmt.Errorf("founded date is required")
	}
	if req.Founded.After(time.Now()) {
		return fmt.Errorf("founded date cannot be in the future")
	}

	// Validate key people
	for i, person := range req.KeyPeople {
		if strings.TrimSpace(person.FullName) == "" {
			return fmt.Errorf("key person %d: full name is required", i+1)
		}
		if strings.TrimSpace(person.Position) == "" {
			return fmt.Errorf("key person %d: position is required", i+1)
		}
	}

	return nil
}