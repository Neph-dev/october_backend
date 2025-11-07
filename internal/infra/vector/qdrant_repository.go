package vector

import (
	"context"
	"fmt"
	"time"

	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/qdrant/go-client/qdrant"
)

const (
	// DefenseArticlesCollection is the name of the Qdrant collection for defense articles
	DefenseArticlesCollection = "defense_articles"
	
	// VectorSize is the dimensionality of our embeddings (text-embedding-3-small)
	VectorSize = 1536
)

// QdrantRepository handles all interactions with Qdrant vector database
type QdrantRepository struct {
	client     *qdrant.Client
	logger     logger.Logger
	collection string
}

// ArticleSearchResult represents a search result from Qdrant
type ArticleSearchResult struct {
	ArticleID string
	Score     float32
	Company   string
	Title     string
}

// NewQdrantRepository creates a new Qdrant repository
func NewQdrantRepository(url, apiKey string, logger logger.Logger) (*QdrantRepository, error) {
	client, err := qdrant.NewClient(&qdrant.Config{
		Host:   url,
		APIKey: apiKey,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create Qdrant client: %w", err)
	}

	repo := &QdrantRepository{
		client:     client,
		logger:     logger,
		collection: DefenseArticlesCollection,
	}

	return repo, nil
}

// EnsureCollection creates the collection if it doesn't exist
func (r *QdrantRepository) EnsureCollection(ctx context.Context) error {
	r.logger.Info("Ensuring Qdrant collection exists", "collection", r.collection)

	// Check if collection exists
	collections, err := r.client.ListCollections(ctx)
	if err != nil {
		return fmt.Errorf("failed to list collections: %w", err)
	}

	// Check if our collection exists
	for _, col := range collections {
		if col == r.collection {
			r.logger.Info("Collection already exists", "collection", r.collection)
			return nil
		}
	}

	// Create the collection
	r.logger.Info("Creating new collection", "collection", r.collection)

	err = r.client.CreateCollection(ctx, &qdrant.CreateCollection{
		CollectionName: r.collection,
		VectorsConfig: qdrant.NewVectorsConfig(&qdrant.VectorParams{
			Size:     VectorSize,
			Distance: qdrant.Distance_Cosine,
		}),
		OptimizersConfig: &qdrant.OptimizersConfigDiff{
			IndexingThreshold: qdrant.PtrOf(uint64(10000)),
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create collection: %w", err)
	}

	r.logger.Info("Collection created successfully", "collection", r.collection)
	return nil
}

// UpsertArticle stores or updates an article's vector embedding
func (r *QdrantRepository) UpsertArticle(ctx context.Context, articleID string, vector []float32, metadata map[string]interface{}) error {
	r.logger.Debug("Upserting article to Qdrant", "article_id", articleID)

	// Convert article ID to Qdrant point ID (use hash or keep as string)
	pointID := qdrant.NewIDNum(hashStringToUint64(articleID))

	// Create the point
	point := &qdrant.PointStruct{
		Id:      pointID,
		Vectors: qdrant.NewVectors(vector...),
		Payload: qdrant.NewValueMap(metadata),
	}

	// Upsert the point
	_, err := r.client.Upsert(ctx, &qdrant.UpsertPoints{
		CollectionName: r.collection,
		Points:         []*qdrant.PointStruct{point},
	})

	if err != nil {
		r.logger.Error("Failed to upsert article", "error", err, "article_id", articleID)
		return fmt.Errorf("failed to upsert article: %w", err)
	}

	r.logger.Debug("Article upserted successfully", "article_id", articleID)
	return nil
}

// UpsertBatch stores multiple articles in a single batch operation
func (r *QdrantRepository) UpsertBatch(ctx context.Context, articles []ArticleVector) error {
	if len(articles) == 0 {
		return nil
	}

	r.logger.Info("Batch upserting articles to Qdrant", "count", len(articles))

	points := make([]*qdrant.PointStruct, len(articles))
	for i, article := range articles {
		points[i] = &qdrant.PointStruct{
			Id:      qdrant.NewIDNum(hashStringToUint64(article.ID)),
			Vectors: qdrant.NewVectors(article.Vector...),
			Payload: qdrant.NewValueMap(article.Metadata),
		}
	}

	// Upsert in batches of 100 to avoid overwhelming the server
	const batchSize = 100
	for i := 0; i < len(points); i += batchSize {
		end := i + batchSize
		if end > len(points) {
			end = len(points)
		}

		batch := points[i:end]
		_, err := r.client.Upsert(ctx, &qdrant.UpsertPoints{
			CollectionName: r.collection,
			Points:         batch,
		})

		if err != nil {
			return fmt.Errorf("failed to upsert batch %d-%d: %w", i, end, err)
		}

		r.logger.Debug("Batch upserted", "from", i, "to", end)
	}

	r.logger.Info("Batch upsert completed successfully", "total", len(articles))
	return nil
}

// Search performs a semantic search for similar articles
func (r *QdrantRepository) Search(ctx context.Context, queryVector []float32, limit uint64, filter *SearchFilter) ([]ArticleSearchResult, error) {
	r.logger.Debug("Searching Qdrant", "limit", limit, "has_filter", filter != nil)

	// Build filter if provided
	var qdrantFilter *qdrant.Filter
	if filter != nil {
		qdrantFilter = buildQdrantFilter(filter)
	}

	// Perform the search
	searchResult, err := r.client.Query(ctx, &qdrant.QueryPoints{
		CollectionName: r.collection,
		Query:          qdrant.NewQuery(queryVector...),
		Limit:          qdrant.PtrOf(limit),
		Filter:         qdrantFilter,
		WithPayload:    qdrant.NewWithPayload(true),
	})

	if err != nil {
		r.logger.Error("Search failed", "error", err)
		return nil, fmt.Errorf("search failed: %w", err)
	}

	// Convert results
	results := make([]ArticleSearchResult, len(searchResult))
	for i, point := range searchResult {
		results[i] = ArticleSearchResult{
			ArticleID: getStringFromPayload(point.Payload, "article_id"),
			Score:     point.Score,
			Company:   getStringFromPayload(point.Payload, "company"),
			Title:     getStringFromPayload(point.Payload, "title"),
		}
	}

	r.logger.Debug("Search completed", "results_found", len(results))
	return results, nil
}

// DeleteArticle removes an article from the vector database
func (r *QdrantRepository) DeleteArticle(ctx context.Context, articleID string) error {
	r.logger.Debug("Deleting article from Qdrant", "article_id", articleID)

	pointID := qdrant.NewIDNum(hashStringToUint64(articleID))

	_, err := r.client.Delete(ctx, &qdrant.DeletePoints{
		CollectionName: r.collection,
		Points: &qdrant.PointsSelector{
			PointsSelectorOneOf: &qdrant.PointsSelector_Points{
				Points: &qdrant.PointsIdsList{
					Ids: []*qdrant.PointId{pointID},
				},
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	r.logger.Debug("Article deleted successfully", "article_id", articleID)
	return nil
}

// GetCollectionInfo returns information about the collection  
func (r *QdrantRepository) GetCollectionInfo(ctx context.Context) (map[string]interface{}, error) {
	// For now, return basic info - can be expanded later
	return map[string]interface{}{
		"collection": r.collection,
		"status":     "active",
	}, nil
}

// ArticleVector represents an article with its vector embedding
type ArticleVector struct {
	ID       string
	Vector   []float32
	Metadata map[string]interface{}
}

// SearchFilter represents filter criteria for searches
type SearchFilter struct {
	Companies      []string
	StartDate      *time.Time
	EndDate        *time.Time
	MinRelevance   *float64
}

// Helper functions

func buildQdrantFilter(filter *SearchFilter) *qdrant.Filter {
	var conditions []*qdrant.Condition

	// Company filter
	if len(filter.Companies) > 0 {
		conditions = append(conditions, &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: "company",
					Match: &qdrant.Match{
						MatchValue: &qdrant.Match_Keywords{
							Keywords: &qdrant.RepeatedStrings{
								Strings: filter.Companies,
							},
						},
					},
				},
			},
		})
	}

	// Date range filter
	if filter.StartDate != nil || filter.EndDate != nil {
		rangeCondition := &qdrant.Range{}
		
		if filter.StartDate != nil {
			rangeCondition.Gte = qdrant.PtrOf(float64(filter.StartDate.Unix()))
		}
		
		if filter.EndDate != nil {
			rangeCondition.Lte = qdrant.PtrOf(float64(filter.EndDate.Unix()))
		}

		conditions = append(conditions, &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key:   "published_date",
					Range: rangeCondition,
				},
			},
		})
	}

	// Relevance filter
	if filter.MinRelevance != nil {
		conditions = append(conditions, &qdrant.Condition{
			ConditionOneOf: &qdrant.Condition_Field{
				Field: &qdrant.FieldCondition{
					Key: "relevance_score",
					Range: &qdrant.Range{
						Gte: qdrant.PtrOf(*filter.MinRelevance),
					},
				},
			},
		})
	}

	if len(conditions) == 0 {
		return nil
	}

	return &qdrant.Filter{
		Must: conditions,
	}
}

func getStringFromPayload(payload map[string]*qdrant.Value, key string) string {
	if val, ok := payload[key]; ok {
		if strVal := val.GetStringValue(); strVal != "" {
			return strVal
		}
	}
	return ""
}

// hashStringToUint64 creates a deterministic uint64 from a string
func hashStringToUint64(s string) uint64 {
	var hash uint64 = 5381
	for _, c := range s {
		hash = ((hash << 5) + hash) + uint64(c)
	}
	return hash
}
