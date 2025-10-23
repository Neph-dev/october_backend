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
	httpHandler "github.com/Neph-dev/october_backend/internal/interfaces/http"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

const (
	shutdownTimeout = 30 * time.Second
	exitSuccess     = 0
	exitFailure     = 1
)

// Application represents our main application
type Application struct {
	config *config.Config
	logger logger.Logger
	server *http.Server
}

// main is the entry point of the application
// Following NASA's rules: keep functions simple, handle all errors, no recursion
func main() {
	// Parse command line flags
	healthCheck := flag.Bool("health", false, "Perform health check and exit")
	flag.Parse()

	// Handle health check
	if *healthCheck {
		os.Exit(performHealthCheck())
	}

	// Exit with proper code
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

	// Create HTTP router
	router := httpHandler.NewRouter(app.logger)
	router.SetupRoutes()

	// Create HTTP server with timeouts (NASA rule: always set timeouts)
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

		if err := app.server.Shutdown(ctx); err != nil {
			app.logger.Error("Server forced to shutdown", "error", err)
			return fmt.Errorf("server shutdown error: %w", err)
		}

		app.logger.Info("Server shutdown completed")
	}

	return nil
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
		// Default to info level if invalid
		return slog.LevelInfo
	}
}

// performHealthCheck performs a health check against the running application
func performHealthCheck() int {
	// Load config to get the server address
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config for health check: %v\n", err)
		return exitFailure
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Make health check request
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
