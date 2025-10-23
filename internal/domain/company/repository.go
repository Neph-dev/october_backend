package company

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	ErrCompanyNotFound    = errors.New("company not found")
	ErrCompanyExists      = errors.New("company already exists")
	ErrInvalidCompanyData = errors.New("invalid company data")
)

// Repository defines the interface for company data operations
type Repository interface {
	Create(ctx context.Context, company *Company) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*Company, error)
	GetByName(ctx context.Context, name string) (*Company, error)
	GetByTicker(ctx context.Context, ticker string) (*Company, error)
	Update(ctx context.Context, company *Company) error
	Delete(ctx context.Context, id primitive.ObjectID) error
	List(ctx context.Context, limit, offset int) ([]*Company, error)
}

// Service defines the business logic interface for company operations
type Service interface {
	CreateCompany(ctx context.Context, req *CreateCompanyRequest) (*CompanyResponse, error)
	GetCompanyByName(ctx context.Context, name string) (*CompanyResponse, error)
	GetCompanyByTicker(ctx context.Context, ticker string) (*CompanyResponse, error)
	ListCompanies(ctx context.Context, limit, offset int) ([]*CompanyResponse, error)
}