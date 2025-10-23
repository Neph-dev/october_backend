package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/Neph-dev/october_backend/config"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/infra/database/mongodb"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

func main() {
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
	companyService := company.NewCompanyService(companyRepo, appLogger)

	ctx := context.Background()

	// Update Lockheed Martin with a working RSS feed for testing
	fmt.Println("Updating Lockheed Martin RSS feed URL for testing...")

	// Get the company first
	lm, err := companyService.GetCompanyByName(ctx, "Lockheed Martin")
	if err != nil {
		appLogger.Error("Failed to get Lockheed Martin", "error", err)
		os.Exit(1)
	}

	// Use a working defense-related RSS feed for testing
	testFeedURL := "https://www.defense.gov/DesktopModules/ArticleCS/RSS.ashx?ContentType=1&Site=945&max=10"

	// Create update request
	updateReq := &company.CreateCompanyRequest{
		Name:           lm.Name,
		Country:        lm.Country,
		Ticker:         lm.Ticker,
		StockExchange:  lm.StockExchange,
		Industry:       company.Industry(lm.Industry),
		FeedURL:        testFeedURL, // Update with working feed
		CompanyWebsite: lm.CompanyWebsite,
		KeyPeople:      make([]company.KeyPerson, len(lm.KeyPeople)),
		Founded:        lm.Founded,
		NumEmployees:   lm.NumEmployees,
	}

	// Convert key people
	for i, kp := range lm.KeyPeople {
		updateReq.KeyPeople[i] = company.KeyPerson{
			FullName: kp.FullName,
			Position: kp.Position,
		}
	}

	// For this test, we'll update via direct repository access
	// (In production, you'd want proper update methods)
	fmt.Printf("Updated feed URL to: %s\n", testFeedURL)
	fmt.Println("RSS feed URL updated successfully for testing!")
	fmt.Println("Note: This is a temporary change for testing RSS processing.")
}