package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/Neph-dev/october_backend/internal/infra/feed"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

func main() {
	// Test RSS feeds
	testFeeds := []string{
		"https://feeds.bbci.co.uk/news/rss.xml",
		"https://rss.cnn.com/rss/edition.rss",
		"https://www.defense.gov/DesktopModules/ArticleCS/RSS.ashx?ContentType=1&Site=945&max=10",
	}

	appLogger := logger.NewLogger(slog.LevelInfo, os.Stdout)
	rssService := feed.NewRSSService(appLogger.Unwrap())

	ctx := context.Background()

	for _, feedURL := range testFeeds {
		fmt.Printf("\n=== Testing RSS Feed: %s ===\n", feedURL)
		
		items, err := rssService.FetchFeed(ctx, feedURL)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}

		fmt.Printf("Successfully fetched %d items\n", len(items))
		if len(items) > 0 {
			item := items[0]
			fmt.Printf("First item:\n")
			fmt.Printf("  Title: %s\n", item.Title)
			fmt.Printf("  Summary: %.100s...\n", item.Summary)
			fmt.Printf("  Link: %s\n", item.Link)
			fmt.Printf("  GUID: %s\n", item.GUID)
		}
	}
}