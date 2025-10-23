package feed

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/mmcdole/gofeed"
)

// RSSService handles RSS feed operations
type RSSService struct {
	parser *gofeed.Parser
	logger *slog.Logger
}

// NewRSSService creates a new RSS service
func NewRSSService(logger *slog.Logger) *RSSService {
	return &RSSService{
		parser: gofeed.NewParser(),
		logger: logger,
	}
}

// FetchFeed fetches and parses an RSS feed from the given URL
func (s *RSSService) FetchFeed(ctx context.Context, feedURL string) ([]*news.RSSFeedItem, error) {
	s.logger.Info("Fetching RSS feed", "url", feedURL)

	feed, err := s.parser.ParseURLWithContext(feedURL, ctx)
	if err != nil {
		s.logger.Error("Failed to parse RSS feed", "error", err, "url", feedURL)
		return nil, fmt.Errorf("failed to parse RSS feed from %s: %w", feedURL, err)
	}

	s.logger.Info("Successfully parsed RSS feed", 
		"url", feedURL, 
		"title", feed.Title, 
		"items", len(feed.Items))

	items := make([]*news.RSSFeedItem, 0, len(feed.Items))
	for _, item := range feed.Items {
		rssItem := s.convertToRSSFeedItem(item)
		if rssItem != nil {
			items = append(items, rssItem)
		}
	}

	return items, nil
}

// convertToRSSFeedItem converts a gofeed.Item to our RSSFeedItem
func (s *RSSService) convertToRSSFeedItem(item *gofeed.Item) *news.RSSFeedItem {
	if item == nil {
		return nil
	}

	// Parse published date
	var publishDate time.Time
	if item.PublishedParsed != nil {
		publishDate = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		publishDate = *item.UpdatedParsed
	} else {
		publishDate = time.Now()
	}

	// Extract content
	content := ""
	if item.Content != "" {
		content = item.Content
	} else if item.Description != "" {
		content = item.Description
	}

	// Use GUID or link as unique identifier
	guid := item.GUID
	if guid == "" {
		guid = item.Link
	}

	// Extract summary (prefer description over content for summary)
	summary := item.Description
	if summary == "" && len(content) > 200 {
		summary = content[:200] + "..."
	}

	return &news.RSSFeedItem{
		Title:       item.Title,
		Summary:     summary,
		Link:        item.Link,
		PublishDate: publishDate,
		GUID:        guid,
		Content:     content,
	}
}