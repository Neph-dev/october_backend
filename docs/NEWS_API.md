# News API Documentation

## Overview

The News API provides endpoints to fetch and filter news articles collected from company RSS feeds. Articles are automatically processed and stored with relevance scoring.

## Data Model

### Article Structure

```json
{
  "id": "507f1f77bcf86cd799439011",
  "title": "Raytheon Technologies Announces New Partnership",
  "summary": "RTX Corporation announced a strategic partnership to develop advanced radar systems...",
  "source_url": "https://www.rtx.com/news/2024/10/23/new-partnership",
  "companies": ["Raytheon Technologies"],
  "published_date": "2024-10-23T10:30:00Z",
  "relevance_score": 0.85,
  "processed_date": "2024-10-23T10:35:00Z",
  "feed_source": "https://www.rtx.com/rss-feeds/news"
}
```

### Fields Description

- **id**: Unique MongoDB ObjectID for the article
- **title**: Article headline
- **summary**: Brief description or excerpt from the article
- **source_url**: Direct link to the original article
- **companies**: Array of company names mentioned in the article
- **published_date**: When the article was originally published
- **relevance_score**: Relevance score (0.0 to 1.0) indicating how relevant the article is to the company
- **processed_date**: When the article was processed and stored in our system
- **feed_source**: URL of the RSS feed where the article was found

## API Endpoints

### GET /news

Retrieve a list of news articles with optional filtering and pagination.

#### Query Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `company` | string | Filter by company name | `?company=Raytheon Technologies` |
| `start_date` | string | Filter articles from this date (YYYY-MM-DD) | `?start_date=2024-10-01` |
| `end_date` | string | Filter articles until this date (YYYY-MM-DD) | `?end_date=2024-10-31` |
| `min_relevance` | float | Minimum relevance score (0.0 to 1.0) | `?min_relevance=0.7` |
| `limit` | integer | Number of articles to return (default: 50, max: 1000) | `?limit=20` |
| `offset` | integer | Number of articles to skip for pagination | `?offset=100` |

#### Response

```json
{
  "articles": [
    {
      "id": "507f1f77bcf86cd799439011",
      "title": "Raytheon Technologies Announces New Partnership",
      "summary": "RTX Corporation announced a strategic partnership...",
      "source_url": "https://www.rtx.com/news/2024/10/23/new-partnership",
      "companies": ["Raytheon Technologies"],
      "published_date": "2024-10-23T10:30:00Z",
      "relevance_score": 0.85,
      "processed_date": "2024-10-23T10:35:00Z",
      "feed_source": "https://www.rtx.com/rss-feeds/news"
    }
  ],
  "total": 1250,
  "limit": 50,
  "offset": 0
}
```

### GET /news/{id}

Retrieve a specific news article by its ID.

#### Path Parameters

- `id`: The MongoDB ObjectID of the article

#### Response

```json
{
  "id": "507f1f77bcf86cd799439011",
  "title": "Raytheon Technologies Announces New Partnership",
  "summary": "RTX Corporation announced a strategic partnership to develop advanced radar systems...",
  "source_url": "https://www.rtx.com/news/2024/10/23/new-partnership",
  "companies": ["Raytheon Technologies"],
  "published_date": "2024-10-23T10:30:00Z",
  "relevance_score": 0.85,
  "processed_date": "2024-10-23T10:35:00Z",
  "feed_source": "https://www.rtx.com/rss-feeds/news"
}
```

## Example Requests

### Get Recent News for Raytheon Technologies

```bash
curl "http://localhost:8080/news?company=Raytheon%20Technologies&limit=10"
```

### Get News from Last Month

```bash
curl "http://localhost:8080/news?start_date=2024-09-23&end_date=2024-10-23"
```

### Get High Relevance News with Pagination

```bash
curl "http://localhost:8080/news?min_relevance=0.8&limit=20&offset=40"
```

### Get Specific Article

```bash
curl "http://localhost:8080/news/507f1f77bcf86cd799439011"
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Bad Request",
  "message": "Invalid filter parameters: start_date must be before end_date"
}
```

### HTTP Status Codes

- `200 OK`: Successful request
- `400 Bad Request`: Invalid parameters or request format
- `404 Not Found`: Article not found (for /news/{id})
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error

## Rate Limiting

The News API is protected by rate limiting:
- **Rate**: 10 requests per second
- **Burst**: Up to 20 requests in a burst
- **Response Header**: `X-RateLimit-Remaining` indicates remaining requests

## Data Processing

### RSS Feed Processing

Articles are automatically collected from company RSS feeds and processed every 2 hours as follows:

1. **Fetching**: RSS feeds are parsed using a robust RSS parser
2. **Deduplication**: Articles are deduplicated using GUID or URL
3. **Company Association**: Articles are associated with the relevant company
4. **Relevance Scoring**: Relevance to the company is calculated based on content analysis
5. **Storage**: Articles are stored in MongoDB with proper indexing

### Processing Commands

```bash
# Process all company feeds (manual trigger)
make process-feeds

# Process specific company feed
make process-feed COMPANY="Raytheon Technologies"

# Direct command line usage
./bin/feed-processor
./bin/feed-processor -company="Raytheon Technologies"
```

### Automatic Processing

The system automatically processes RSS feeds every 2 hours when the API server is running. This ensures fresh content is regularly updated without manual intervention.

## MongoDB Indexes

The following indexes are automatically created for optimal performance:

- `guid` (unique): Ensures no duplicate articles
- `companies`: Fast filtering by company
- `published_date`: Date range queries
- `relevance_score`: Relevance filtering
- `feed_source`: Source filtering
- `companies + published_date`: Compound index for common queries

## Monitoring and Health

Articles processing can be monitored through application logs:

```bash
# View processing logs
tail -f /var/log/october-backend.log | grep "RSS feed processing"
```

Health check endpoint remains available at `/health` for overall application status.

## Future Enhancements

- **Enhanced Content Analysis**: More sophisticated relevance scoring algorithms
- **Content Extraction**: Full article content extraction and analysis
- **Real-time Processing**: WebSocket or Server-Sent Events for real-time updates
- **Search**: Full-text search capabilities
- **Categorization**: Automatic article categorization (earnings, contracts, partnerships, etc.)
- **Webhooks**: Notifications for important news based on criteria
- **Additional Companies**: Easy expansion to include more aerospace and defense companies