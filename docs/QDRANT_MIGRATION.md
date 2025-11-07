# Qdrant Vector Database Migration Guide

## Overview
This document outlines the migration to Qdrant vector database for semantic search capabilities in the October backend application.

## What We've Completed

### 1. Infrastructure Setup ✅
- **Docker Compose**: Added Qdrant service to `docker-compose.yml`
  - Qdrant REST API on port 6333
  - gRPC API on port 6334  
  - Persistent volume for data storage
  - Configured with API key authentication

- **Dependencies**: Installed Go packages
  - `github.com/qdrant/go-client` v1.15.2
  - Required gRPC and protobuf dependencies

### 2. Configuration ✅
- **Config Structure**: Added `QdrantConfig` to `config/config.go`
  - `QDRANT_URL`: Qdrant server URL (default: http://localhost:6333)
  - `QDRANT_API_KEY`: API key for authentication

- **Environment Variables**: Updated `.env` file
  ```
  QDRANT_URL=http://localhost:6333
  QDRANT_API_KEY=eyJhbGci...
  ```

### 3. Embedding Service ✅
- **File**: `internal/infra/ai/embedding_service.go`
- **Features**:
  - `EmbedText()`: Convert any text to 1536-dimensional vector
  - `EmbedArticle()`: Optimized embedding for articles (title + summary + content)
  - `EmbedQuery()`: Enhanced embedding for search queries
  - `EmbedBatch()`: Batch processing for multiple texts
  - Uses OpenAI `text-embedding-3-small` model

### 4. Qdrant Repository ✅
- **File**: `internal/infra/vector/qdrant_repository.go`
- **Features**:
  - `EnsureCollection()`: Create collection if not exists
  - `UpsertArticle()`: Store/update single article vector
  - `UpsertBatch()`: Batch upsert for bulk operations
  - `Search()`: Semantic search with filters
  - `DeleteArticle()`: Remove article from vector DB
  - `GetCollectionInfo()`: Get collection stats

- **Collection Schema**:
  - Name: `defense_articles`
  - Vector dimensions: 1536
  - Distance metric: Cosine similarity
  - Metadata fields:
    - `article_id`: MongoDB article ID
    - `company`: Company name
    - `title`: Article title
    - `published_date`: Unix timestamp
    - `relevance_score`: Original relevance score

## Next Steps

### Step 4: Update Article Creation Flow
Modify the news service to generate embeddings when creating articles:

```go
// In internal/domain/news/service.go
func (s *Service) CreateArticle(ctx context.Context, article *Article) error {
    // 1. Save to MongoDB (existing)
    err := s.repository.Create(ctx, article)
    
    // 2. Generate embedding
    embedding, err := s.embeddingService.EmbedArticle(
        ctx,
        article.Title,
        article.Summary,
        article.Content,
    )
    
    // 3. Store in Qdrant
    metadata := map[string]interface{}{
        "article_id":      article.ID.Hex(),
        "company":         article.Companies[0], // Primary company
        "title":           article.Title,
        "published_date":  article.PublishedDate.Unix(),
        "relevance_score": article.RelevanceScore,
    }
    
    err = s.qdrantRepo.UpsertArticle(ctx, article.ID.Hex(), embedding, metadata)
    
    return nil
}
```

### Step 5: Update AI Query Processing
Replace keyword-based retrieval with semantic search:

```go
// In internal/infra/ai/openai_service.go
func (s *OpenAIService) retrieveRelevantArticles(...) {
    // 1. Generate query embedding
    queryVector, err := s.embeddingService.EmbedQuery(ctx, analysis.Question)
    
    // 2. Semantic search in Qdrant
    searchFilter := &vector.SearchFilter{
        Companies: analysis.CompanyNames,
        StartDate: analysis.TimeWindow.StartDate,
        EndDate:   analysis.TimeWindow.EndDate,
    }
    
    results, err := s.qdrantRepo.Search(ctx, queryVector, 10, searchFilter)
    
    // 3. Fetch full articles from MongoDB using IDs
    articleIDs := make([]string, len(results))
    for i, result := range results {
        articleIDs[i] = result.ArticleID
    }
    
    articles := s.newsService.GetArticlesByIDs(ctx, articleIDs)
    
    // 4. Convert to source references
    return convertToSourceReferences(articles, results)
}
```

### Step 6: Create Migration Script
Build a script to index existing MongoDB articles:

```go
// cmd/migrate/qdrant_migration.go
func main() {
    // 1. Connect to MongoDB and Qdrant
    // 2. Fetch all articles from MongoDB
    // 3. Generate embeddings in batches
    // 4. Upsert to Qdrant in batches
    // 5. Log progress and errors
}
```

### Step 7: Integration
Wire up services in `cmd/api/main.go`:

```go
// Initialize embedding service
embeddingService := ai.NewEmbeddingService(openaiClient, logger)

// Initialize Qdrant repository
qdrantRepo, err := vector.NewQdrantRepository(
    cfg.Qdrant.URL,
    cfg.Qdrant.APIKey,
    logger,
)

// Ensure collection exists
err = qdrantRepo.EnsureCollection(ctx)

// Pass to news service and AI service
newsService := news.NewService(mongoRepo, qdrantRepo, embeddingService, logger)
aiService := ai.NewOpenAIService(openaiClient, newsService, googleSearch, embeddingService, qdrantRepo, cache, logger)
```

## Starting Qdrant

### Local Development
```bash
# Start Qdrant with docker-compose
docker-compose --profile local-db up qdrant -d

# Check Qdrant is running
curl http://localhost:6333/
```

### Docker Compose Full Stack
```bash
# Start all services including Qdrant
docker-compose --profile local-db up -d
```

## Testing Semantic Search

### 1. Index a Test Article
```bash
curl -X POST http://localhost:8080/api/v1/news/articles \
  -H "Content-Type: application/json" \
  -d '{
    "title": "RTX CEO Christopher Calio discusses Q3 results",
    "summary": "Defense contractor reports strong earnings...",
    "companies": ["Raytheon Technologies"]
  }'
```

### 2. Test Semantic Query
```bash
curl -X POST http://localhost:8080/api/v1/ai/query \
  -H "Content-Type: application/json" \
  -d '{
    "question": "Who is the CEO of RTX?"
  }'
```

Expected: Should find the article about Christopher Calio even though the query doesn't contain exact keywords.

## Monitoring

### Qdrant Web UI
Visit: http://localhost:6333/dashboard

### Collection Stats
```bash
curl http://localhost:6333/collections/defense_articles
```

### Search Test
```bash
curl -X POST http://localhost:6333/collections/defense_articles/points/search \
  -H "Content-Type: application/json" \
  -d '{
    "vector": [0.1, 0.2, ...], # 1536 dimensions
    "limit": 5
  }'
```

## Performance Expectations

### Embedding Generation
- Single article: ~100-200ms
- Batch (100 articles): ~2-3 seconds

### Search Performance
- Semantic search (10 results): ~10-50ms
- With filters: ~15-60ms

### Storage
- Per article: ~6KB (1536 float32 + metadata)
- 10,000 articles: ~60MB

## Troubleshooting

### Qdrant Connection Issues
```bash
# Check if Qdrant is running
docker ps | grep qdrant

# Check Qdrant logs
docker logs october_backend-qdrant-1

# Restart Qdrant
docker-compose restart qdrant
```

### Collection Issues
```bash
# Delete and recreate collection
curl -X DELETE http://localhost:6333/collections/defense_articles
# Collection will be auto-created on next article creation
```

### OpenAI Embedding Errors
- Check API key is valid
- Monitor rate limits (3,000 RPM for text-embedding-3-small)
- Implement retry logic with exponential backoff

## Cost Analysis

### OpenAI Embeddings
- Model: text-embedding-3-small
- Cost: $0.02 per 1M tokens
- Avg article: ~500 tokens
- 10,000 articles: ~$0.10

### Qdrant Hosting
- Self-hosted (Docker): Free
- Qdrant Cloud: ~$25/month for starter tier
- AWS/GCP deployment: ~$30-50/month

## Migration Checklist

- [x] Add Qdrant to docker-compose.yml
- [x] Install go-client dependencies
- [x] Create embedding service
- [x] Create Qdrant repository
- [x] Update configuration
- [x] Add environment variables
- [ ] Update news service for article creation
- [ ] Update AI service for semantic search
- [ ] Create migration script
- [ ] Wire up services in main.go
- [ ] Test article indexing
- [ ] Test semantic search
- [ ] Run migration for existing articles
- [ ] Monitor performance
- [ ] Update documentation

## References

- [Qdrant Documentation](https://qdrant.tech/documentation/)
- [Go Client GitHub](https://github.com/qdrant/go-client)
- [OpenAI Embeddings Guide](https://platform.openai.com/docs/guides/embeddings)
- [Semantic Search Best Practices](https://qdrant.tech/documentation/tutorials/semantic-search/)
