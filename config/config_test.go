package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Test with default values
	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify default values
	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected port '8080', got '%s'", config.Server.Port)
	}

	if config.Logger.Level != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.Logger.Level)
	}
}

func TestLoadWithEnvironment(t *testing.T) {
	// Set environment variables
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("SERVER_PORT", "9090")
	os.Setenv("LOG_LEVEL", "debug")
	defer func() {
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("SERVER_PORT")
		os.Unsetenv("LOG_LEVEL")
	}()

	config, err := Load()
	if err != nil {
		t.Fatalf("Load() failed: %v", err)
	}

	// Verify environment values
	if config.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", config.Server.Host)
	}

	if config.Server.Port != "9090" {
		t.Errorf("Expected port '9090', got '%s'", config.Server.Port)
	}

	if config.Logger.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.Logger.Level)
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				Server: ServerConfig{
					Host:         "localhost",
					Port:         "8080",
					ReadTimeout:  15 * time.Second,
					WriteTimeout: 15 * time.Second,
					IdleTimeout:  60 * time.Second,
				},
				Database: DatabaseConfig{
					URI: "mongodb://localhost:27017/test",
				},
				Logger: LoggerConfig{
					Level: "info",
				},
			},
			wantErr: false,
		},
		{
			name: "empty host",
			config: &Config{
				Server: ServerConfig{
					Host: "",
					Port: "8080",
				},
				Database: DatabaseConfig{
					URI: "mongodb://localhost:27017/test",
				},
				Logger: LoggerConfig{
					Level: "info",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid port",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: "invalid",
				},
				Database: DatabaseConfig{
					URI: "mongodb://localhost:27017/test",
				},
				Logger: LoggerConfig{
					Level: "info",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			config: &Config{
				Server: ServerConfig{
					Host: "localhost",
					Port: "8080",
				},
				Database: DatabaseConfig{
					URI: "mongodb://localhost:27017/test",
				},
				Logger: LoggerConfig{
					Level: "invalid",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}