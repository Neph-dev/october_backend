package news

import (
	"context"
)

// Repository defines the interface for news data access
type Repository interface {
	// Create saves a new article to the repository
	Create(ctx context.Context, article *Article) error
	
	// GetByID retrieves an article by its ID
	GetByID(ctx context.Context, id string) (*Article, error)
	
	// GetByGUID retrieves an article by its GUID
	GetByGUID(ctx context.Context, guid string) (*Article, error)
	
	// List retrieves articles with optional filtering
	List(ctx context.Context, filter *NewsFilter) ([]*Article, error)
	
	// Count returns the total number of articles matching the filter
	Count(ctx context.Context, filter *NewsFilter) (int64, error)
	
	// Update updates an existing article
	Update(ctx context.Context, article *Article) error
	
	// Delete removes an article by ID
	Delete(ctx context.Context, id string) error
	
	// ExistsByGUID checks if an article with the given GUID exists
	ExistsByGUID(ctx context.Context, guid string) (bool, error)
}