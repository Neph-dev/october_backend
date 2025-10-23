# AI/RAG API Documentation

The AI API provides intelligent question-answering capabilities using Retrieval-Augmented Generation (RAG) powered by OpenAI. Users can ask natural language questions about defense companies and get AI-generated responses based on the latest news articles and company information.

## Features

- **Natural Language Processing**: Ask questions in plain English
- **Retrieval-Augmented Generation**: Responses backed by real news articles
- **Company Context**: Focus queries on specific companies
- **Source Attribution**: See which articles were used to generate responses
- **Query Analysis**: Understand how questions are interpreted
- **Confidence Scoring**: Assess the reliability of responses

## Prerequisites

- OpenAI API key configured in environment variables
- News articles processed in the database
- Server running with AI endpoints enabled

## API Endpoints

### Process AI Query

Ask a natural language question and get an AI-generated response with sources.

**Endpoint:** `POST /ai/query`

**Rate Limiting:** 10 requests per second, burst of 20

**Request Body:**
```json
{
  "question": "How did RTX perform this quarter?",
  "company_context": ["Raytheon Technologies"]
}
```

**Request Fields:**
- `question` (required): Natural language question (1-1000 characters)
- `company_context` (optional): Array of company names to focus the search

**Response:**
```json
{
  "answer": "Based on recent news articles, RTX Corporation reported strong Q3 2025 results with significant contract wins and technological developments. The company secured a $1.7 billion contract for LTAMDS and completed critical engine testing for Collaborative Combat Aircraft. RTX's various divisions showed continued growth with Pratt & Whitney engines surpassing 600,000 flying hours and Collins Aerospace expanding operations.",
  "sources": [
    {
      "article_id": "68fa57bde91d89e06c4125c1",
      "title": "RTX Reports Q3 2025 Results",
      "summary": "RTX Corporation announced strong third quarter financial results...",
      "company_name": "Raytheon Technologies",
      "published_date": "2025-10-21T00:00:00Z",
      "source_url": "https://www.rtx.com/news/2025/10/21/rtx-reports-q3-2025-results",
      "relevance_score": 0.95
    }
  ],
  "confidence": 0.87,
  "processing_time": "2.5s",
  "companies_referenced": ["Raytheon Technologies"]
}
```

**Response Fields:**
- `answer`: AI-generated response to the question
- `sources`: Array of news articles used as context
- `confidence`: Confidence score (0.0-1.0) indicating response reliability
- `processing_time`: Time taken to process the query
- `companies_referenced`: Companies identified in the query

### Analyze Query

Analyze a question to understand intent and extract entities without generating a full response.

**Endpoint:** `POST /ai/analyze`

**Request Body:**
```json
{
  "question": "What were RTX earnings this quarter?"
}
```

**Response:**
```json
{
  "query_type": "financial",
  "company_names": ["Raytheon Technologies"],
  "keywords": ["earnings", "quarter", "financial"],
  "time_window": {
    "start_date": "2025-07-01T00:00:00Z",
    "end_date": "2025-10-23T00:00:00Z",
    "period": "this_quarter"
  },
  "search_terms": ["RTX", "earnings", "quarter", "financial", "results"]
}
```

## Query Types

The AI system categorizes questions into different types:

- **financial**: Questions about earnings, revenue, financial performance
- **contracts**: Questions about defense contracts, awards, deals
- **general**: General questions about companies or industry
- **comparison**: Comparative questions between companies
- **news**: Questions about recent news or developments

## Supported Companies

Currently supports queries about:

- **Raytheon Technologies (RTX)**: Aerospace and defense corporation
- **US War Department**: Government defense entity
- **Lockheed Martin**: Defense contractor (limited support)

## Example Queries

### Financial Performance
```bash
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How did RTX perform financially this quarter?",
    "company_context": ["Raytheon Technologies"]
  }'
```

### Defense Contracts
```bash
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What defense contracts did RTX recently win?"
  }'
```

### Military Developments
```bash
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the latest military training developments?",
    "company_context": ["US War Department"]
  }'
```

### Industry Comparison
```bash
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "How does RTX compare to other defense contractors?"
  }'
```

### General Technology
```bash
curl -X POST http://localhost:8080/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "What are the recent developments in defense technology?"
  }'
```

## Error Responses

All endpoints return consistent error responses:

```json
{
  "error": "Bad Request",
  "message": "Question is required"
}
```

### HTTP Status Codes

- `200 OK`: Successful request
- `400 Bad Request`: Invalid parameters or request format
- `404 Not Found`: No relevant information found
- `429 Too Many Requests`: Rate limit exceeded
- `500 Internal Server Error`: Server error (AI service unavailable)

## Configuration

### Environment Variables

```bash
# Required: OpenAI API Key
OPENAI_API_KEY=your_openai_api_key_here

# Server Configuration
SERVER_HOST=0.0.0.0
SERVER_PORT=8080

# Database Configuration
DATABASE_URI=mongodb://localhost:27017/october
```

### OpenAI Model Configuration

The system uses **GPT-4o-mini** for cost efficiency while maintaining good performance:

- **Query Analysis**: Uses low temperature (0.1) for consistent parsing
- **Response Generation**: Uses moderate temperature (0.3) for natural responses
- **Token Limits**: 200 tokens for analysis, 500 tokens for responses

## Performance Considerations

### Response Times
- **Query Analysis**: ~1-2 seconds
- **Article Retrieval**: ~0.5-1 seconds
- **AI Response Generation**: ~1-3 seconds
- **Total Processing**: ~2.5-6 seconds

### Rate Limiting
- AI endpoints are rate-limited due to OpenAI API costs
- Recommended: Cache frequent queries on client side
- Consider implementing user-based rate limiting for production

### Cost Optimization
- Uses GPT-4o-mini for cost efficiency
- Limits context to top 10 most relevant articles
- Implements confidence scoring to indicate response quality

## Troubleshooting

### Common Issues

1. **"OpenAI API key cannot be empty"**
   - Set the `OPENAI_API_KEY` environment variable
   - Ensure the key is valid and has sufficient credits

2. **"No relevant information found"**
   - Process RSS feeds first: `make process-feeds`
   - Check if articles exist for the queried companies
   - Try broader or different question phrasing

3. **"Failed to process query"**
   - Check OpenAI API status and quota
   - Verify database connectivity
   - Check server logs for detailed error information

4. **Low confidence scores**
   - May indicate limited relevant articles
   - Consider processing more recent RSS feeds
   - Try more specific questions

### Testing

Use the provided test script:
```bash
# Test all AI endpoints
make test-ai

# Or run directly
./scripts/test_ai_api.sh
```

## Security Considerations

- API keys are never logged or exposed in responses
- Input validation prevents injection attacks
- Rate limiting prevents abuse
- All user inputs are sanitized before processing

## Future Enhancements

- **Semantic Search**: Vector-based article similarity
- **Multi-language Support**: Support for non-English queries
- **Custom Models**: Fine-tuned models for defense industry
- **Real-time Updates**: Stream processing for immediate insights
- **Advanced Analytics**: Query performance and usage metrics