# Company API Documentation

## Overview

The Company API provides endpoints to manage and retrieve defense, aerospace, and government entity information. It includes rate limiting, MongoDB integration, and comprehensive data validation following NASA's clean code principles.

## Company Data Model

Each company includes the following information:

```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "Lockheed Martin",
  "country": "United States",
  "ticker": "LMT",
  "stockExchange": "NYSE",
  "industry": "Defense",
  "feedUrl": "https://news.lockheedmartin.com/rss",
  "companyWebsite": "https://www.lockheedmartin.com",
  "keyPeople": [
    {
      "fullName": "James Taiclet",
      "position": "Chairman, President and CEO"
    }
  ],
  "founded": "1995-03-15T00:00:00Z",
  "numEmployees": 116000,
  "metadata": {
    "lastFeedUpdate": "2025-10-23T12:00:00Z",
    "isActive": true,
    "tags": []
  }
}
```

## API Endpoints

### Get Company by Name

Retrieve company information by company name.

**Endpoint:** `GET /company/{company-name}`

**Rate Limiting:** 10 requests per second, burst of 20

**Parameters:**
- `company-name` (path): The company name (case-insensitive, URL encoded)

**Examples:**
```bash
# Get Lockheed Martin
curl http://localhost:8080/company/Lockheed%20Martin

# Get Raytheon Technologies
curl http://localhost:8080/company/Raytheon%20Technologies
```

**Responses:**

**200 OK:**
```json
{
  "id": "507f1f77bcf86cd799439011",
  "name": "Lockheed Martin",
  "country": "United States",
  "ticker": "LMT",
  "stockExchange": "NYSE",
  "industry": "Defense",
  "feedUrl": "https://news.lockheedmartin.com/rss",
  "companyWebsite": "https://www.lockheedmartin.com",
  "keyPeople": [
    {
      "fullName": "James Taiclet",
      "position": "Chairman, President and CEO"
    }
  ],
  "founded": "1995-03-15T00:00:00Z",
  "numEmployees": 116000,
  "metadata": {
    "lastFeedUpdate": "2025-10-23T12:00:00Z",
    "isActive": true,
    "tags": []
  }
}
```

**404 Not Found:**
```json
{
  "error": true,
  "message": "company not found",
  "status": 404
}
```

**429 Too Many Requests:**
```json
{
  "error": true,
  "message": "rate limit exceeded",
  "status": 429
}
```

### Create Company (Internal Use)

Create a new company record. This endpoint is primarily used for data seeding.

**Endpoint:** `POST /companies`

**Request Body:**
```json
{
  "name": "Example Corp",
  "country": "United States",
  "ticker": "EXAM",
  "stockExchange": "NASDAQ",
  "industry": "Aerospace",
  "feedUrl": "https://example.com/rss",
  "companyWebsite": "https://example.com",
  "keyPeople": [
    {
      "fullName": "John Doe",
      "position": "CEO"
    }
  ],
  "founded": "2000-01-01T00:00:00Z",
  "numEmployees": 5000
}
```

## Pre-loaded Companies

The system comes with three pre-configured companies:

### 1. Lockheed Martin
- **Name:** Lockheed Martin
- **Ticker:** LMT
- **Industry:** Defense
- **Feed URL:** https://news.lockheedmartin.com/rss
- **Key People:** James Taiclet (Chairman, President and CEO)
- **Employees:** 116,000

### 2. Raytheon Technologies (RTX Corporation)
- **Name:** Raytheon Technologies
- **Ticker:** RTX
- **Industry:** Aerospace
- **Feed URL:** https://www.rtx.com/rss-feeds/news
- **Key People:** Gregory J. Hayes (Chairman and CEO)
- **Employees:** 185,000

### 3. US War Department
- **Name:** US War Department
- **Ticker:** N/A (Government Entity)
- **Industry:** Government
- **Feed URL:** https://www.war.gov/DesktopModules/ArticleCS/RSS.ashx?ContentType=1&Site=945&max=10
- **Key People:** Pete Hegseth (Secretary of War)
- **Employees:** 2,870,000

## Supported Industries

The system supports three industry types:

- **Defense**: Private defense contractors (requires ticker and stock exchange)
- **Aerospace**: Aerospace companies (requires ticker and stock exchange)  
- **Government**: Government entities (ticker and stock exchange are optional)

### Validation Rules

- For **Defense** and **Aerospace** industries: Company ticker and stock exchange are required
- For **Government** industry: Company ticker and stock exchange are optional (can be empty)
- All other fields (name, country, feed URL, website, key people, etc.) are required for all industries

## Database Setup

### MongoDB Configuration

The application uses MongoDB for data persistence. Configure the database connection via environment variables:

```bash
# Database Configuration
DATABASE_URI=mongodb://localhost:27017/october
```

### Data Seeding

To populate the database with initial company data:

```bash
# Using Make
make seed-data

# Direct execution
go run ./cmd/seed
```

### Database Schema

The application automatically creates the following indexes for optimal performance:

- **Unique index on company name** - Ensures no duplicate company names
- **Unique index on ticker** - Ensures no duplicate ticker symbols
- **Compound index on country and industry** - Optimizes filtering queries

## Rate Limiting

The company endpoints implement rate limiting to prevent abuse:

- **Rate:** 10 requests per second
- **Burst:** 20 requests
- **Scope:** Per client IP address
- **Cleanup:** Inactive clients removed after 30 minutes

Rate limiting headers are not currently exposed but can be added if needed.

## Error Handling

The API implements comprehensive error handling:

- **400 Bad Request:** Invalid input data or malformed requests
- **404 Not Found:** Company not found
- **409 Conflict:** Company already exists (during creation)
- **429 Too Many Requests:** Rate limit exceeded
- **500 Internal Server Error:** Server-side errors

All errors follow a consistent format:
```json
{
  "error": true,
  "message": "descriptive error message",
  "status": 404
}
```

## Security Features

- **Input Validation:** All inputs are validated for type, length, and format
- **Rate Limiting:** Prevents abuse and ensures fair usage
- **SQL Injection Prevention:** MongoDB queries use proper parameterization
- **Error Sanitization:** Internal errors are not exposed to clients

## Monitoring and Logging

All API operations are logged with structured JSON logging:

```json
{
  "time": "2025-10-23T12:00:00Z",
  "level": "INFO",
  "msg": "Getting company by name",
  "name": "Lockheed Martin",
  "client_ip": "192.168.1.100"
}
```

## Development

### Running the Application

```bash
# Start the server
make run

# Start with debug logging
make debug

# Build binaries
make build
```

### Testing the API

```bash
# Health check
curl http://localhost:8080/health

# Get company data
curl http://localhost:8080/company/Lockheed%20Martin

# Test rate limiting (run multiple times quickly)
for i in {1..25}; do curl http://localhost:8080/company/Lockheed%20Martin; done
```

### Environment Variables

```bash
# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration
DATABASE_URI=mongodb://localhost:27017/october

# Logging
LOG_LEVEL=info
```

## NASA Clean Code Compliance

The company API implementation follows NASA's clean code rules:

- **Functions under 60 lines:** All functions are kept simple and focused
- **Comprehensive error handling:** All return values are checked
- **Input validation:** All inputs are validated before processing
- **Resource cleanup:** Database connections are properly managed
- **No recursion:** All algorithms use iterative approaches
- **Defensive programming:** Assumptions are validated with assertions

## Future Enhancements

- **Pagination:** Add pagination support for company listings
- **Search:** Implement full-text search capabilities
- **Caching:** Add Redis caching for frequently accessed data
- **Authentication:** Add API key or OAuth authentication
- **Audit Logging:** Track all data modifications
- **Metrics:** Expose Prometheus metrics for monitoring