# October Backend

A robust Go server built following NASA's "Power of 10" rules for clean and safe code, featuring MongoDB integration, company data management, and rate-limited APIs.

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

## Features

### Core Infrastructure
- **Graceful Shutdown**: Proper signal handling and resource cleanup
- **Structured Logging**: JSON-formatted logs with context
- **Configuration Management**: Environment-based configuration with validation
- **Error Handling**: Comprehensive error handling and recovery
- **Health Checks**: Built-in health monitoring endpoints
- **Middleware**: Request logging, recovery, and security middleware
- **Timeouts**: Proper timeout handling for all operations

### Database Integration
- **MongoDB Support**: Full MongoDB integration with connection pooling
- **Company Management**: CRUD operations for defense/aerospace companies
- **News Processing**: Automated RSS feed processing and article storage
- **Data Validation**: Comprehensive input validation and sanitization
- **Indexing**: Optimized database indexes for performance

### News & RSS Features
- **RSS Feed Processing**: Automated collection from company feeds
- **Article Management**: Storage with deduplication and validation
- **Sentiment Analysis**: Basic sentiment scoring for articles
- **Relevance Scoring**: Company relevance calculation
- **Filtering**: Advanced filtering by company, date, sentiment, and relevance
- **Pagination**: Efficient pagination for large datasets

### API Features
- **Rate Limiting**: Token bucket algorithm with per-IP tracking
- **RESTful Endpoints**: Clean REST API design
- **Error Responses**: Consistent error response format
- **Request Logging**: Detailed request/response logging

### AI/RAG Features
- **Natural Language Queries**: Ask questions in plain English about companies
- **OpenAI Integration**: Powered by GPT-4o-mini for cost-effective AI responses
- **Retrieval-Augmented Generation**: Responses backed by real news articles
- **Query Analysis**: Intelligent parsing of user intent and entities
- **Source Attribution**: See which articles were used for each response
- **Confidence Scoring**: Reliability assessment for AI-generated answers
- **Company Context**: Focus queries on specific defense contractors

## Quick Start

### Prerequisites

- Go 1.21 or later
- MongoDB 4.4 or later
- OpenAI API key (for AI/RAG features)

### Installation

1. Clone the repository
2. Install dependencies:
   ```bash
   go mod download
   ```

3. Set up environment:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration, including:
   # OPENAI_API_KEY=your_openai_api_key_here
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

### News API

#### Get News Articles
```bash
GET /news
```

**Rate Limited**: 10 requests/second, burst of 20

**Query Parameters:**
- `company`: Filter by company name
- `start_date`: Filter from date (YYYY-MM-DD)
- `end_date`: Filter until date (YYYY-MM-DD)
- `sentiment`: Filter by sentiment (-2 to 2)
- `min_relevance`: Minimum relevance score (0.0 to 1.0)
- `limit`: Number of results (default: 50, max: 1000)
- `offset`: Pagination offset

**Examples:**
```bash
# Get recent news for Lockheed Martin
curl "http://localhost:8080/news?company=Lockheed%20Martin&limit=10"

# Get positive news from last month
curl "http://localhost:8080/news?sentiment=1&start_date=2024-09-23&end_date=2024-10-23"

# Get high relevance news with pagination
curl "http://localhost:8080/news?min_relevance=0.8&limit=20&offset=40"
```

#### Get Specific Article
```bash
GET /news/{id}
```

**Example:**
```bash
curl http://localhost:8080/news/507f1f77bcf86cd799439011
```

### AI/RAG API

#### Ask AI Questions
```bash
POST /ai/query
```

**Rate Limited**: 10 requests/second, burst of 20

**Request Body:**
```json
{
  "question": "How did RTX perform this quarter?",
  "company_context": ["Raytheon Technologies"]
}
```

**Examples:**
```bash
# Financial performance query
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{"question": "How did RTX perform this quarter?"}'

# Defense contracts query
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{"question": "What defense contracts did RTX recently win?"}'

# Military developments query
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{"question": "What are the latest military training developments?", "company_context": ["US War Department"]}'
```

#### Analyze Query Intent
```bash
POST /ai/analyze
```

**Request Body:**
```json
{
  "question": "What were RTX earnings this quarter?"
}
```

**Example:**
```bash
curl -X POST http://localhost:8080/ai/analyze \
  -H "Content-Type: application/json" \
  -d '{"question": "What were RTX earnings this quarter?"}'
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

## RSS Feed Processing

The application includes automated RSS feed processing to collect and store news articles:

### Processing Commands

```bash
# Seed company data first
make seed-data

# Process all company RSS feeds
make process-feeds

# Process specific company feed
make process-feed COMPANY="Lockheed Martin"

# Direct command usage
./bin/feed-processor
./bin/feed-processor -company="Raytheon Technologies"
```

### Article Processing Features

- **Deduplication**: Articles are deduplicated using GUID or URL
- **Sentiment Analysis**: Basic sentiment scoring (-2 to +2)
- **Relevance Scoring**: Company relevance calculation (0.0 to 1.0)
- **Automatic Indexing**: Database indexes for optimal query performance

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