package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Neph-dev/october_backend/pkg/logger"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
	logger   logger.Logger
}

type Config struct {
	URI            string
	DatabaseName   string
	ConnectTimeout time.Duration
	PingTimeout    time.Duration
	MaxPoolSize    uint64
	MinPoolSize    uint64
}

// NewClient creates a new MongoDB client with proper configuration
func NewClient(config Config, logger logger.Logger) (*Client, error) {
	if config.URI == "" {
		return nil, fmt.Errorf("MongoDB URI cannot be empty")
	}
	if config.DatabaseName == "" {
		return nil, fmt.Errorf("database name cannot be empty")
	}

	// Set default timeouts if not provided
	if config.ConnectTimeout == 0 {
		config.ConnectTimeout = 10 * time.Second
	}
	if config.PingTimeout == 0 {
		config.PingTimeout = 5 * time.Second
	}
	if config.MaxPoolSize == 0 {
		config.MaxPoolSize = 100
	}
	if config.MinPoolSize == 0 {
		config.MinPoolSize = 5
	}

	clientOptions := options.Client().
		ApplyURI(config.URI).
		SetConnectTimeout(config.ConnectTimeout).
		SetMaxPoolSize(config.MaxPoolSize).
		SetMinPoolSize(config.MinPoolSize)

	// Create connection context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), config.ConnectTimeout)
	defer cancel()

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("Failed to connect to MongoDB", "error", err, "uri", config.URI)
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Test the connection
	pingCtx, pingCancel := context.WithTimeout(context.Background(), config.PingTimeout)
	defer pingCancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		logger.Error("Failed to ping MongoDB", "error", err)
		// Close the client if ping fails
		if closeErr := client.Disconnect(context.Background()); closeErr != nil {
			logger.Error("Failed to close MongoDB client after ping failure", "error", closeErr)
		}
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.DatabaseName)

	logger.Info("Successfully connected to MongoDB",
		"database", config.DatabaseName,
		"maxPoolSize", config.MaxPoolSize,
		"minPoolSize", config.MinPoolSize,
	)

	return &Client{
		client:   client,
		database: database,
		logger:   logger,
	}, nil
}

func (c *Client) Database() *mongo.Database {
	return c.database
}

// Close gracefully closes the MongoDB connection
func (c *Client) Close(ctx context.Context) error {
	if c.client == nil {
		return nil
	}

	c.logger.Info("Closing MongoDB connection")

	if err := c.client.Disconnect(ctx); err != nil {
		c.logger.Error("Failed to disconnect from MongoDB", "error", err)
		return fmt.Errorf("failed to disconnect from MongoDB: %w", err)
	}

	c.logger.Info("MongoDB connection closed successfully")
	return nil
}

// HealthCheck performs a health check on the MongoDB connection
func (c *Client) HealthCheck(ctx context.Context) error {
	if c.client == nil {
		return fmt.Errorf("MongoDB client is not initialized")
	}

	// Ping with a short timeout
	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := c.client.Ping(pingCtx, nil); err != nil {
		return fmt.Errorf("MongoDB health check failed: %w", err)
	}

	return nil
}