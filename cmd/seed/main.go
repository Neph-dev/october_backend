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
	raytheonFounded, _ := time.Parse("2006-01-02", "2020-04-03") // RTX Corporation formed
	warDeptFounded, _ := time.Parse("2006-01-02", "1947-09-18")  // Department of Defense established
	northropFounded, _ := time.Parse("2006-01-02", "1939-08-18") // Northrop Grumman founded
	boeingFounded, _ := time.Parse("2006-01-02", "1916-07-15")   // Boeing founded
	generalDynamicsFounded, _ := time.Parse("2006-01-02", "1952-02-21") // General Dynamics founded
	l3harrisFounded, _ := time.Parse("2006-01-02", "2019-06-29") // L3Harris merger completed
	huntingtonFounded, _ := time.Parse("2006-01-02", "1886-01-01") // Huntington Ingalls founded
	textronFounded, _ := time.Parse("2006-01-02", "1923-01-01")  // Textron founded
	lockheedFounded, _ := time.Parse("2006-01-02", "1995-03-15") // Lockheed Martin merger completed

	return []*company.CreateCompanyRequest{
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
		{
			Name:          "US War Department",
			Country:       "United States",
			Ticker:        "",
			StockExchange: "",
			Industry:      company.IndustryGovernment,
			FeedURL:       "https://www.war.gov/DesktopModules/ArticleCS/RSS.ashx?ContentType=1&Site=945&max=10",
			CompanyWebsite: "https://www.war.gov",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Pete Hegseth",
					Position: "Secretary of War",
				},
				{
					FullName: "General C.Q. Brown Jr.",
					Position: "Chairman of the Joint Chiefs of Staff",
				},
				{
					FullName: "Kathleen Hicks",
					Position: "Deputy Secretary of War",
				},
			},
			Founded:      warDeptFounded,
			NumEmployees: 2870000, // Active military + civilian personnel
		},
		{
			Name:          "Northrop Grumman",
			Country:       "United States", 
			Ticker:        "NOC",
			StockExchange: "NYSE",
			Industry:      company.IndustryDefense,
			FeedURL:       "https://news.northropgrumman.com/rss/news-releases",
			CompanyWebsite: "https://www.northropgrumman.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Kathy J. Warden",
					Position: "Chairman, CEO and President",
				},
				{
					FullName: "David F. Keffer",
					Position: "Chief Financial Officer",
				},
				{
					FullName: "Mark A. Caylor",
					Position: "Corporate Vice President and Chief Operating Officer",
				},
			},
			Founded:      northropFounded,
			NumEmployees: 95000,
		},
		{
			Name:          "Boeing",
			Country:       "United States",
			Ticker:        "BA",
			StockExchange: "NYSE", 
			Industry:      company.IndustryAerospace,
			FeedURL:       "https://boeing.mediaroom.com/rss",
			CompanyWebsite: "https://www.boeing.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "David L. Calhoun",
					Position: "President and CEO",
				},
				{
					FullName: "Brian J. West",
					Position: "Executive Vice President and Chief Financial Officer",
				},
				{
					FullName: "Stephanie F. Pope",
					Position: "Executive Vice President and Chief Operating Officer",
				},
			},
			Founded:      boeingFounded,
			NumEmployees: 171000,
		},
		{
			Name:          "General Dynamics",
			Country:       "United States",
			Ticker:        "GD",
			StockExchange: "NYSE",
			Industry:      company.IndustryDefense,
			FeedURL:       "https://www.gd.com/news",
			CompanyWebsite: "https://www.generaldynamics.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Phebe N. Novakovic",
					Position: "Chairman and Chief Executive Officer",
				},
				{
					FullName: "Jason W. Aiken",
					Position: "Senior Vice President and Chief Financial Officer",
				},
				{
					FullName: "Mark C. Roualet",
					Position: "Executive Vice President and Chief Operating Officer",
				},
			},
			Founded:      generalDynamicsFounded,
			NumEmployees: 106500,
		},
		{
			Name:          "L3Harris Technologies",
			Country:       "United States",
			Ticker:        "LHX",
			StockExchange: "NYSE",
			Industry:      company.IndustryDefense,
			FeedURL:       "https://www.l3harris.com/newsroom",
			CompanyWebsite: "https://www.l3harris.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Christopher E. Kubasik",
					Position: "Chair, Chief Executive Officer and President",
				},
				{
					FullName: "Michelle Turner",
					Position: "Senior Vice President and Chief Financial Officer",
				},
				{
					FullName: "Sean Stackley",
					Position: "President, Integrated Mission Systems",
				},
			},
			Founded:      l3harrisFounded,
			NumEmployees: 50000,
		},
		{
			Name:          "Huntington Ingalls Industries",
			Country:       "United States",
			Ticker:        "HII",
			StockExchange: "NYSE",
			Industry:      company.IndustryDefense,
			FeedURL:       "https://hii.com/newsroom",
			CompanyWebsite: "https://www.huntingtoningalls.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Christopher D. Kastner",
					Position: "President and Chief Executive Officer",
				},
				{
					FullName: "Thomas E. Stiehle",
					Position: "Senior Vice President and Chief Financial Officer",
				},
				{
					FullName: "Andy Green",
					Position: "Executive Vice President and Chief Operating Officer",
				},
			},
			Founded:      huntingtonFounded,
			NumEmployees: 44000,
		},
		{
			Name:          "Textron Inc",
			Country:       "United States",
			Ticker:        "TXT",
			StockExchange: "NYSE",
			Industry:      company.IndustryAerospace,
			FeedURL:       "https://investor.textron.com/news-releases/default.aspx",
			CompanyWebsite: "https://www.textron.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "Scott C. Donnelly",
					Position: "Chairman and Chief Executive Officer",
				},
				{
					FullName: "Frank T. Connor",
					Position: "Executive Vice President and Chief Financial Officer",
				},
				{
					FullName: "E. Robert Lupone",
					Position: "Executive Vice President, General Counsel, Secretary and Chief Compliance Officer",
				},
			},
			Founded:      textronFounded,
			NumEmployees: 35000,
		},
		{
			Name:          "Lockheed Martin Corporation",
			Country:       "United States",
			Ticker:        "LMT",
			StockExchange: "NYSE",
			Industry:      company.IndustryDefense,
			FeedURL:       "https://news.lockheedmartin.com/news-releases?pagetemplate=rss",
			CompanyWebsite: "https://www.lockheedmartin.com",
			KeyPeople: []company.KeyPerson{
				{
					FullName: "James D. Taiclet",
					Position: "Chairman, President and Chief Executive Officer",
				},
				{
					FullName: "Jay D. Malave",
					Position: "Executive Vice President and Chief Financial Officer",
				},
				{
					FullName: "Frank A. St. John",
					Position: "Chief Operating Officer",
				},
			},
			Founded:      lockheedFounded,
			NumEmployees: 122000,
		},
	}
}