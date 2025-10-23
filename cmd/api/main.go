package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Neph-dev/october_backend/config"
	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	aiInfra "github.com/Neph-dev/october_backend/internal/infra/ai"
	"github.com/Neph-dev/october_backend/internal/infra/cache"
	"github.com/Neph-dev/october_backend/internal/infra/database/mongodb"
	"github.com/Neph-dev/october_backend/internal/infra/feed"
	"github.com/Neph-dev/october_backend/internal/infra/search"
	httpHandler "github.com/Neph-dev/october_backend/internal/interfaces/http"
	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

const (
	shutdownTimeout = 30 * time.Second
	exitSuccess     = 0
	exitFailure     = 1
)

type Application struct {
	config         *config.Config
	logger         logger.Logger
	server         *http.Server
	dbClient       *mongodb.Client
	companyService company.Service
	newsService    *news.Service
	aiService      ai.Service
	rssService     *feed.RSSService
	processorService *feed.ProcessorService
}

// main is the entry point of the application
// Following NASA's rules: keep functions simple, handle all errors, no recursion
func main() {
	healthCheck := flag.Bool("health", false, "Perform health check and exit")
	flag.Parse()

	if *healthCheck {
		os.Exit(performHealthCheck())
	}

	os.Exit(run())
}

// run contains the main application logic
// Separated from main() for better testing and error handling
func run() int {
	// Load configuration first - fail fast if invalid
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		return exitFailure
	}

	// Initialize logger early for proper error reporting
	logLevel := parseLogLevel(cfg.Logger.Level)
	appLogger := logger.NewLogger(logLevel, os.Stdout)

	// Create application instance
	app := &Application{
		config: cfg,
		logger: appLogger,
	}

	// Initialize the application
	if err := app.initialize(); err != nil {
		app.logger.Error("Failed to initialize application", "error", err)
		return exitFailure
	}

	// Start the application
	if err := app.start(); err != nil {
		app.logger.Error("Failed to start application", "error", err)
		return exitFailure
	}

	return exitSuccess
}

// initialize sets up all application components
func (app *Application) initialize() error {
	app.logger.Info("Initializing application")

	// Initialize MongoDB client
	dbConfig := mongodb.Config{
		URI:            app.config.Database.URI,
		DatabaseName:   "october",
		ConnectTimeout: 10 * time.Second,
		PingTimeout:    5 * time.Second,
		MaxPoolSize:    100,
		MinPoolSize:    5,
	}

	var err error
	app.dbClient, err = mongodb.NewClient(dbConfig, app.logger)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Initialize repositories
	companyRepo := mongodb.NewCompanyRepository(app.dbClient.Database(), app.logger)
	newsRepo := mongodb.NewNewsRepository(app.dbClient.Database())

	// Initialize services
	app.companyService = company.NewCompanyService(companyRepo, app.logger)
	app.newsService = news.NewService(newsRepo, app.logger.Unwrap())
	app.rssService = feed.NewRSSService(app.logger.Unwrap())
	app.processorService = feed.NewProcessorService(app.rssService, app.newsService, app.companyService, app.logger.Unwrap())
	
	// Initialize Google Custom Search service
	googleSearchService := search.NewGoogleSearchService(
		app.config.AI.CustomSearchAPIKey,
		app.config.AI.CustomSearchEngineID,
		app.logger,
	)
	
	// Initialize summary cache
	summaryCache := cache.NewMemoryCache()
	
	// Initialize AI service with Google Custom Search integration and caching
	openaiClient := openai.NewClient(app.config.AI.OpenAIAPIKey)
	app.aiService = aiInfra.NewOpenAIService(
		openaiClient,
		app.newsService,
		googleSearchService,
		summaryCache,
		app.logger,
	)

	// Create HTTP router with dependencies
	router := httpHandler.NewRouter(app.logger, app.companyService, app.newsService, app.aiService)
	router.SetupRoutes()

	// Create indexes for better performance
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	if err := newsRepo.CreateIndexes(ctx); err != nil {
		app.logger.Error("Failed to create news indexes", "error", err)
		// Don't fail completely, as indexes can be created later
	} else {
		app.logger.Info("Database indexes created successfully")
	}

	// Create HTTP server with timeouts.
	app.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%s", app.config.Server.Host, app.config.Server.Port),
		Handler:      router,
		ReadTimeout:  app.config.Server.ReadTimeout,
		WriteTimeout: app.config.Server.WriteTimeout,
		IdleTimeout:  app.config.Server.IdleTimeout,
	}

	app.logger.Info("Application initialized successfully")
	return nil
}

// start begins the application lifecycle
func (app *Application) start() error {
	app.logger.Info("Starting application", 
		"host", app.config.Server.Host,
		"port", app.config.Server.Port,
	)

	// Channel to listen for interrupt signals
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Channel to listen for server errors
	serverErrors := make(chan error, 1)

	// Start RSS feed refresh in the background
	go app.startRSSFeedRefresh()

	// Start HTTP server in a goroutine
	go func() {
		app.logger.Info("Server listening", "address", app.server.Addr)
		serverErrors <- app.server.ListenAndServe()
	}()

	// Block until we receive a signal or error
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("server error: %w", err)
		}

	case sig := <-interrupt:
		app.logger.Info("Shutdown signal received", "signal", sig.String())

		// Graceful shutdown with timeout
		ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		// Shutdown HTTP server
		if err := app.server.Shutdown(ctx); err != nil {
			app.logger.Error("Server forced to shutdown", "error", err)
			return fmt.Errorf("server shutdown error: %w", err)
		}

		// Close database connection
		if app.dbClient != nil {
			if err := app.dbClient.Close(ctx); err != nil {
				app.logger.Error("Failed to close database connection", "error", err)
				// Don't return error here, as server shutdown was successful
			}
		}

		app.logger.Info("Application shutdown completed")
	}

	return nil
}

// startRSSFeedRefresh starts the background RSS feed refresh process
func (app *Application) startRSSFeedRefresh() {
	app.logger.Info("Starting RSS feed refresh scheduler", "interval", "2 hours")
	
	// Create a ticker for 2 hours
	ticker := time.NewTicker(2 * time.Hour)
	defer ticker.Stop()

	// Process feeds immediately on startup
	app.processRSSFeeds()

	// Process feeds every 2 hours
	for range ticker.C {
		app.processRSSFeeds()
	}
}

// processRSSFeeds processes RSS feeds for all companies
func (app *Application) processRSSFeeds() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	app.logger.Info("Starting scheduled RSS feed processing")
	
	err := app.processorService.ProcessAllCompanyFeeds(ctx)
	if err != nil {
		app.logger.Error("Failed to process RSS feeds", "error", err)
	} else {
		app.logger.Info("Completed scheduled RSS feed processing")
	}
}

// parseLogLevel converts string log level to slog.Level
// Following NASA's rule: validate all inputs
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// performHealthCheck performs a health check against the running application
func performHealthCheck() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config for health check: %v\n", err)
		return exitFailure
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	url := fmt.Sprintf("http://%s:%s/health", cfg.Server.Host, cfg.Server.Port)
	if cfg.Server.Host == "0.0.0.0" {
		url = fmt.Sprintf("http://localhost:%s/health", cfg.Server.Port)
	}

	resp, err := client.Get(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Health check failed: %v\n", err)
		return exitFailure
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "Health check failed with status: %d\n", resp.StatusCode)
		return exitFailure
	}

	fmt.Println("Health check passed")
	return exitSuccess
}
