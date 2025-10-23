# October Backend

A robust Go server built following NASA's "Power of 10" rules for clean and safe code, featuring MongoDB integration, company data management, and rate-limited APIs.

## Quick Start

### Prerequisites

- Go 1.21 or later
- MongoDB 4.4 or later

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Start MongoDB (using Docker):
   ```bash
   docker run -d --name mongodb -p 27017:27017 mongo:latest
   ```

5. Build and run:
   ```bash
   make build
   make seed-data  # Seed initial company data
   make run
   ```

## API Endpoints

### Company API

#### Get Company by Name
```bash
GET /company/{company-name}
```

**Rate Limited**: 10 requests/second, burst of 20

**Examples:**
```bash
# Get Lockheed Martin
curl http://localhost:8080/company/Lockheed%20Martin

# Get Raytheon Technologies  
curl http://localhost:8080/company/Raytheon%20Technologies
```

#### Health Check
```bash
GET /health
```

**Example:**
```bash
curl http://localhost:8080/health
```

### Pre-loaded Companies

The system includes two defense/aerospace companies:

1. **Lockheed Martin** (LMT)
   - Industry: Defense
   - Feed: https://news.lockheedmartin.com/rss
   - Employees: 116,000

2. **Raytheon Technologies** (RTX) 
   - Industry: Aerospace
   - Feed: https://www.rtx.com/rss-feeds/news
   - Employees: 185,000

## Architecture

```
cmd/
├── api/main.go           # Application entry point
└── seed/main.go          # Database seeding utility
config/                   # Configuration management
pkg/logger/               # Structured logging
internal/
├── domain/company/       # Company business logic
├── infra/database/       # Database implementations
└── interfaces/http/      # HTTP handlers and middleware
```

## Running the Application

### Prerequisites

- Go 1.21 or later
- MongoDB (if using database features)

### Environment Setup

1. Copy environment template:
   ```bash
   cp .env.example .env
   ```

2. Adjust configuration in `.env` as needed

### Build and Run

```bash
# Build the application
go build -o october-server ./cmd/api

# Run the application
./october-server
```

### Development

```bash
# Run directly
go run ./cmd/api

# Run with custom log level
LOG_LEVEL=debug go run ./cmd/api

# Run with custom port
SERVER_PORT=9090 go run ./cmd/api
```

## Health Check

The application provides a health check endpoint:

```bash
curl http://localhost:8080/health
```

Response:
```json
{
  "status": "healthy",
  "timestamp": "2024-10-23T10:30:00Z"
}
```

## Configuration

All configuration is handled through environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `SERVER_READ_TIMEOUT` | `15s` | HTTP read timeout |
| `SERVER_WRITE_TIMEOUT` | `15s` | HTTP write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | HTTP idle timeout |
| `DATABASE_URI` | `mongodb://localhost:27017/october` | Database connection string |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |

## Safety Features

### Error Handling
- All functions check return values
- Comprehensive error logging
- Graceful degradation on failures

### Resource Management
- Proper cleanup on shutdown
- Connection pooling and limits
- Memory usage monitoring

### Signal Handling
- SIGINT and SIGTERM handling
- Graceful shutdown with timeout
- Active connection draining

### Logging
- Structured JSON logging
- Request/response logging
- Error tracking and alerting

## Development Guidelines

1. **Function Size**: Keep functions under 60 lines
2. **Error Handling**: Always check and handle errors
3. **Resource Cleanup**: Use defer for cleanup operations
4. **Testing**: Write tests for all public functions
5. **Documentation**: Document all exported functions and types
6. **Validation**: Validate all inputs and configurations

## Production Deployment

### Docker (Recommended)

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o october-server ./cmd/api

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/october-server .
CMD ["./october-server"]
```

### System Service

Create a systemd service file for production deployment:

```ini
[Unit]
Description=October Backend Service
After=network.target

[Service]
Type=simple
User=october
Group=october
WorkingDirectory=/opt/october
ExecStart=/opt/october/october-server
Restart=always
RestartSec=5
EnvironmentFile=/opt/october/.env

[Install]
WantedBy=multi-user.target
```

## Monitoring

The application exposes metrics and health endpoints:

- Health: `GET /health`
- Metrics: Available through structured logs

## Security

- Input validation on all endpoints
- Proper error message sanitization
- Rate limiting (when configured)
- CORS protection
- Security headers middleware