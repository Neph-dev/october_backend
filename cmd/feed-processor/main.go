package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Neph-dev/october_backend/config"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	aiInfra "github.com/Neph-dev/october_backend/internal/infra/ai"
	"github.com/Neph-dev/october_backend/internal/infra/database/mongodb"
	"github.com/Neph-dev/october_backend/internal/infra/feed"
	"github.com/Neph-dev/october_backend/internal/infra/vector"
	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

func main() {
	companyName := flag.String("company", "", "Company name to process RSS feed for (leave empty for all companies)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	appLogger := logger.NewLogger(slog.LevelInfo, os.Stdout)

	// Initialize MongoDB client
	dbConfig := mongodb.Config{
		URI:            cfg.Database.URI,
		DatabaseName:   "october",
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    5,
	}

	dbClient, err := mongodb.NewClient(dbConfig, appLogger)
	if err != nil {
		appLogger.Error("Failed to connect to MongoDB", "error", err)
		os.Exit(1)
	}
	defer dbClient.Close(context.Background())

	// Initialize repositories and services
	companyRepo := mongodb.NewCompanyRepository(dbClient.Database(), appLogger)
	newsRepo := mongodb.NewNewsRepository(dbClient.Database())

	// Initialize OpenAI client for embedding service
	openaiClient := openai.NewClient(cfg.AI.OpenAIAPIKey)
	embeddingService := aiInfra.NewEmbeddingService(openaiClient, appLogger)

	// Initialize Qdrant vector repository (optional - will log warning if fails)
	qdrantRepo, err := vector.NewQdrantRepository(cfg.Qdrant.URL, cfg.Qdrant.APIKey, appLogger)
	if err != nil {
		appLogger.Warn("Failed to initialize Qdrant repository, continuing without vector search", "error", err)
		qdrantRepo = nil
	}

	companyService := company.NewCompanyService(companyRepo, appLogger)
	newsService := news.NewService(newsRepo, qdrantRepo, embeddingService, appLogger.Unwrap())
	rssService := feed.NewRSSService(appLogger.Unwrap())
	webScraperService := feed.NewWebScraperService(appLogger.Unwrap())
	processorService := feed.NewProcessorService(rssService, webScraperService, newsService, companyService, appLogger.Unwrap())

	// Create indexes if needed
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := newsRepo.CreateIndexes(ctx); err != nil {
		appLogger.Error("Failed to create news indexes", "error", err)
	}

	// Process feeds
	if *companyName != "" {
		appLogger.Info("Processing RSS feed for specific company", "company", *companyName)
		err = processorService.ProcessCompanyFeed(ctx, *companyName)
	} else {
		appLogger.Info("Processing RSS feeds for all companies")
		err = processorService.ProcessAllCompanyFeeds(ctx)
	}

	if err != nil {
		appLogger.Error("Failed to process RSS feeds", "error", err)
		os.Exit(1)
	}

	appLogger.Info("RSS feed processing completed successfully")
}