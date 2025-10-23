package news

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Service handles business logic for news operations
type Service struct {
	repo   Repository
	logger *slog.Logger
}

// NewService creates a new news service
func NewService(repo Repository, logger *slog.Logger) *Service {
	return &Service{
		repo:   repo,
		logger: logger,
	}
}

// CreateArticle creates a new article after validation
func (s *Service) CreateArticle(ctx context.Context, article *Article) error {
	if err := article.Validate(); err != nil {
		s.logger.Error("Invalid article data", "error", err)
		return err
	}

	// Check if article already exists by GUID
	exists, err := s.repo.ExistsByGUID(ctx, article.GUID)
	if err != nil {
		s.logger.Error("Error checking article existence", "error", err, "guid", article.GUID)
		return err
	}
	if exists {
		s.logger.Debug("Article already exists", "guid", article.GUID)
		return ErrDuplicateArticle
	}

	// Set processed date
	article.ProcessedDate = time.Now()

	// Generate ID if not set
	if article.ID.IsZero() {
		article.ID = primitive.NewObjectID()
	}

	if err := s.repo.Create(ctx, article); err != nil {
		s.logger.Error("Failed to create article", "error", err, "title", article.Title)
		return err
	}

	s.logger.Info("Article created successfully", "id", article.ID.Hex(), "title", article.Title)
	return nil
}

// GetArticleByID retrieves an article by its ID
func (s *Service) GetArticleByID(ctx context.Context, id string) (*Article, error) {
	article, err := s.repo.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get article by ID", "error", err, "id", id)
		return nil, err
	}
	return article, nil
}

// ListArticles retrieves articles with filtering
func (s *Service) ListArticles(ctx context.Context, filter *NewsFilter) ([]*Article, int64, error) {
	if filter == nil {
		filter = &NewsFilter{}
	}

	// Set default limit if not specified
	if filter.Limit <= 0 {
		filter.Limit = 50
	}

	// Validate filter
	if err := s.validateFilter(filter); err != nil {
		s.logger.Error("Invalid filter", "error", err)
		return nil, 0, err
	}

	articles, err := s.repo.List(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to list articles", "error", err)
		return nil, 0, err
	}

	count, err := s.repo.Count(ctx, filter)
	if err != nil {
		s.logger.Error("Failed to count articles", "error", err)
		return articles, 0, err
	}

	return articles, count, nil
}

// ProcessRSSFeedItem processes an RSS feed item into an article
func (s *Service) ProcessRSSFeedItem(ctx context.Context, item *RSSFeedItem, companyName, feedSource string) (*Article, error) {
	article := &Article{
		Title:          item.Title,
		Summary:        item.Summary,
		SourceURL:      item.Link,
		Companies:      []string{companyName},
		PublishedDate:  item.PublishDate,
		RelevanceScore: s.calculateRelevanceScore(item, companyName),
		FeedSource:     feedSource,
		Content:        item.Content,
		GUID:           item.GUID,
	}

	return article, nil
}

// calculateRelevanceScore calculates a basic relevance score for the article
func (s *Service) calculateRelevanceScore(item *RSSFeedItem, companyName string) float64 {
	// Simple relevance scoring - can be enhanced with ML/NLP
	score := 0.5 // Base score
	
	// Check if company name appears in title (higher weight)
	if containsIgnoreCase(item.Title, companyName) {
		score += 0.3
	}
	
	// Check if company name appears in summary
	if containsIgnoreCase(item.Summary, companyName) {
		score += 0.2
	}
	
	// Ensure score is within bounds
	if score > 1.0 {
		score = 1.0
	}
	
	return score
}

// validateFilter validates the news filter parameters
func (s *Service) validateFilter(filter *NewsFilter) error {
	if filter.StartDate != nil && filter.EndDate != nil {
		if filter.StartDate.After(*filter.EndDate) {
			return ErrInvalidFilter
		}
	}
	
	if filter.MinRelevance != nil && (*filter.MinRelevance < 0 || *filter.MinRelevance > 1) {
		return ErrInvalidFilter
	}
	
	if filter.Limit < 0 || filter.Limit > 1000 {
		return ErrInvalidFilter
	}
	
	if filter.Offset < 0 {
		return ErrInvalidFilter
	}
	
	return nil
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	// Simple implementation - can be enhanced with better text matching
	return len(s) >= len(substr) && 
		   (s == substr || 
		    len(s) > len(substr) && 
		    (s[:len(substr)] == substr || 
		     s[len(s)-len(substr):] == substr))
}