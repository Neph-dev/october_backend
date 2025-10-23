# October Backend

A robust Go server built following NASA's "Power of 10" rules for clean and safe code.

## NASA Clean Code Compliance

This application follows NASA's coding standards for critical systems:

1. **Avoid complex flow constructs** - No goto statements, setjmp, or recursion
2. **No dynamic memory allocation** - All memory allocations are at startup
3. **No functions larger than 60 lines** - All functions are kept simple and focused
4. **Return value checking** - All function return values are checked
5. **Limited scope** - Variables have minimal scope
6. **Runtime assertions** - Critical assumptions are verified
7. **Restricted preprocessor use** - Minimal macro usage
8. **Limited pointer use** - Careful pointer management
9. **Compile with warnings** - All warnings treated as errors
10. **Static analysis** - Code is regularly analyzed for issues

## Architecture

```
cmd/api/main.go           # Application entry point
├── config/               # Configuration management
├── pkg/logger/           # Structured logging
├── internal/
│   ├── interfaces/http/  # HTTP handlers and routing
│   │   └── middleware/   # HTTP middleware
│   ├── domain/           # Business logic
│   └── infra/            # External dependencies
```

## Features

- **Graceful Shutdown**: Proper signal handling and resource cleanup
- **Structured Logging**: JSON-formatted logs with context
- **Configuration Management**: Environment-based configuration with validation
- **Error Handling**: Comprehensive error handling and recovery
- **Health Checks**: Built-in health monitoring endpoints
- **Middleware**: Request logging, recovery, and security middleware
- **Timeouts**: Proper timeout handling for all operations

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