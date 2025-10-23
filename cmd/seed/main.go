package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Neph-dev/october_backend/config"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/infra/database/mongodb"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

// seedCompanies seeds the database with initial company data
func main() {
	fmt.Println("Starting data seeding...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	appLogger := logger.NewLogger(slog.LevelInfo, os.Stdout)

	// Initialize MongoDB client
	dbConfig := mongodb.Config{
		URI:            cfg.Database.URI,
		DatabaseName:   "october",
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
		MaxPoolSize:    10,
		MinPoolSize:    2,
	}

	dbClient, err := mongodb.NewClient(dbConfig, appLogger)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		dbClient.Close(ctx)
	}()

	// Initialize repository and service
	companyRepo := mongodb.NewCompanyRepository(dbClient.Database(), appLogger)
	companyService := company.NewCompanyService(companyRepo, appLogger)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Seed companies
	companies := getCompaniesToSeed()
	for _, comp := range companies {
		fmt.Printf("Seeding company: %s\n", comp.Name)
		
		existing, err := companyService.GetCompanyByName(ctx, comp.Name)
		if err == nil && existing != nil {
			fmt.Printf("Company %s already exists, skipping...\n", comp.Name)
			continue
		}

		result, err := companyService.CreateCompany(ctx, comp)
		if err != nil {
			fmt.Printf("Failed to create company %s: %v\n", comp.Name, err)
			continue
		}

		fmt.Printf("Successfully created company: %s (ID: %s)\n", result.Name, result.ID)
	}

	fmt.Println("Data seeding completed!")
}

// getCompaniesToSeed returns the companies to seed in the database
func getCompaniesToSeed() []*company.CreateCompanyRequest {
	// Parse dates for founding dates
	lockheedFounded, _ := time.Parse("2006-01-02", "1995-03-15") // Lockheed Martin Corporation formed
	raytheonFounded, _ := time.Parse("2006-01-02", "2020-04-03") // RTX Corporation formed

	return []*company.CreateCompanyRequest{
		{
			Name:          "Lockheed Martin",
			Country:       "United States",
			Ticker:        "LMT",
			StockExchange: "NYSE",
			Industry:      company.IndustryDefense,
			FeedURL:       "https://news.lockheedmartin.com/rss",
			CompanyWebsite: "https://www.lockheedmartin.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "James Taiclet",
					Position: "Chairman, President and CEO",
				},
				{
					FullName: "Jesus Malave",
					Position: "Chief Financial Officer",
				},
				{
					FullName: "Gregory M. Ulmer",
					Position: "Executive Vice President, Aeronautics",
				},
			},
			Founded:      lockheedFounded,
			NumEmployees: 116000,
		},
		{
			Name:          "Raytheon Technologies",
			Country:       "United States",
			Ticker:        "RTX",
			StockExchange: "NYSE",
			Industry:      company.IndustryAerospace,
			FeedURL:       "https://www.rtx.com/rss-feeds/news",
			CompanyWebsite: "https://www.rtx.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Gregory J. Hayes",
					Position: "Chairman and CEO",
				},
				{
					FullName: "Neil Mitchill",
					Position: "Chief Financial Officer",
				},
				{
					FullName: "Christopher T. Calio",
					Position: "President and Chief Operating Officer",
				},
			},
			Founded:      raytheonFounded,
			NumEmployees: 185000,
		},
	}
}