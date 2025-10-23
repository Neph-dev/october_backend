package feed

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/news"
)

// ProcessorService handles RSS feed processing and article creation
type ProcessorService struct {
	rssService     *RSSService
	newsService    *news.Service
	companyService company.Service
	logger         *slog.Logger
}

// NewProcessorService creates a new feed processor service
func NewProcessorService(
	rssService *RSSService,
	newsService *news.Service,
	companyService company.Service,
	logger *slog.Logger,
) *ProcessorService {
	return &ProcessorService{
		rssService:     rssService,
		newsService:    newsService,
		companyService: companyService,
		logger:         logger,
	}
}

// ProcessCompanyFeed fetches and processes RSS feed for a specific company
func (s *ProcessorService) ProcessCompanyFeed(ctx context.Context, companyName string) error {
	s.logger.Info("Processing RSS feed for company", "company", companyName)

	// Get company information
	compResp, err := s.companyService.GetCompanyByName(ctx, companyName)
	if err != nil {
		s.logger.Error("Failed to get company", "error", err, "company", companyName)
		return fmt.Errorf("failed to get company %s: %w", companyName, err)
	}

	if compResp.FeedURL == "" {
		s.logger.Warn("Company has no feed URL", "company", companyName)
		return fmt.Errorf("company %s has no feed URL", companyName)
	}

	// Fetch RSS feed
	items, err := s.rssService.FetchFeed(ctx, compResp.FeedURL)
	if err != nil {
		s.logger.Error("Failed to fetch RSS feed", "error", err, "company", companyName, "url", compResp.FeedURL)
		return fmt.Errorf("failed to fetch RSS feed for %s: %w", companyName, err)
	}

	s.logger.Info("Fetched RSS items", "company", companyName, "items", len(items))

	// Process each item
	processed := 0
	skipped := 0
	for _, item := range items {
		article, err := s.newsService.ProcessRSSFeedItem(ctx, item, companyName, compResp.FeedURL)
		if err != nil {
			s.logger.Error("Failed to process RSS item", "error", err, "title", item.Title)
			continue
		}

		// Create article
		err = s.newsService.CreateArticle(ctx, article)
		if err != nil {
			if err == news.ErrDuplicateArticle {
				skipped++
				s.logger.Debug("Skipping duplicate article", "title", item.Title, "guid", item.GUID)
				continue
			}
			s.logger.Error("Failed to create article", "error", err, "title", item.Title)
			continue
		}

		processed++
		s.logger.Debug("Created article", "title", article.Title, "id", article.ID.Hex())
	}

	s.logger.Info("Completed RSS feed processing", 
		"company", companyName, 
		"processed", processed, 
		"skipped", skipped, 
		"total", len(items))

	return nil
}

// ProcessAllCompanyFeeds processes RSS feeds for all companies
func (s *ProcessorService) ProcessAllCompanyFeeds(ctx context.Context) error {
	s.logger.Info("Processing RSS feeds for all companies")

	// Get all companies with a reasonable limit
	companies, err := s.companyService.ListCompanies(ctx, 100, 0)
	if err != nil {
		s.logger.Error("Failed to list companies", "error", err)
		return fmt.Errorf("failed to list companies: %w", err)
	}

	totalProcessed := 0
	totalErrors := 0

	for _, comp := range companies {
		if comp.FeedURL == "" {
			s.logger.Debug("Skipping company with no feed URL", "company", comp.Name)
			continue
		}

		err := s.ProcessCompanyFeed(ctx, comp.Name)
		if err != nil {
			totalErrors++
			s.logger.Error("Failed to process company feed", "error", err, "company", comp.Name)
			continue
		}
		totalProcessed++
	}

	s.logger.Info("Completed processing all company feeds", 
		"processed", totalProcessed, 
		"errors", totalErrors, 
		"total_companies", len(companies))

	return nil
}