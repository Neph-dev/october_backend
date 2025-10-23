package mongodb

import (
	"context"

	"github.com/Neph-dev/october_backend/internal/domain/news"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const newsCollection = "news"

// NewsRepository implements news.Repository for MongoDB
type NewsRepository struct {
	collection *mongo.Collection
}

// NewNewsRepository creates a new MongoDB news repository
func NewNewsRepository(db *mongo.Database) *NewsRepository {
	return &NewsRepository{
		collection: db.Collection(newsCollection),
	}
}

// Create saves a new article to MongoDB
func (r *NewsRepository) Create(ctx context.Context, article *news.Article) error {
	_, err := r.collection.InsertOne(ctx, article)
	return err
}

// GetByID retrieves an article by its ID
func (r *NewsRepository) GetByID(ctx context.Context, id string) (*news.Article, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, news.ErrArticleNotFound
	}

	var article news.Article
	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, news.ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}

// GetByGUID retrieves an article by its GUID
func (r *NewsRepository) GetByGUID(ctx context.Context, guid string) (*news.Article, error) {
	var article news.Article
	err := r.collection.FindOne(ctx, bson.M{"guid": guid}).Decode(&article)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, news.ErrArticleNotFound
		}
		return nil, err
	}

	return &article, nil
}
// GetByCompany retrieves articles by company name
func (r *NewsRepository) GetByCompany(ctx context.Context, companyName string) ([]*news.Article, error) {
	filter := bson.M{
		"companies": bson.M{
			"$in": []string{companyName},
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*news.Article
	for cursor.Next(ctx) {
		var article news.Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// List retrieves articles with optional filtering
func (r *NewsRepository) List(ctx context.Context, filter *news.NewsFilter) ([]*news.Article, error) {
	mongoFilter := r.buildFilter(filter)
	opts := r.buildOptions(filter)

	cursor, err := r.collection.Find(ctx, mongoFilter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var articles []*news.Article
	for cursor.Next(ctx) {
		var article news.Article
		if err := cursor.Decode(&article); err != nil {
			return nil, err
		}
		articles = append(articles, &article)
	}

	if err := cursor.Err(); err != nil {
		return nil, err
	}

	return articles, nil
}

// Count returns the total number of articles matching the filter
func (r *NewsRepository) Count(ctx context.Context, filter *news.NewsFilter) (int64, error) {
	mongoFilter := r.buildFilter(filter)
	return r.collection.CountDocuments(ctx, mongoFilter)
}

// Update updates an existing article
func (r *NewsRepository) Update(ctx context.Context, article *news.Article) error {
	filter := bson.M{"_id": article.ID}
	update := bson.M{"$set": article}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return news.ErrArticleNotFound
	}
	
	return nil
}

// Delete removes an article by ID
func (r *NewsRepository) Delete(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return news.ErrArticleNotFound
	}

	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": objectID})
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return news.ErrArticleNotFound
	}

	return nil
}

// ExistsByGUID checks if an article with the given GUID exists
func (r *NewsRepository) ExistsByGUID(ctx context.Context, guid string) (bool, error) {
	count, err := r.collection.CountDocuments(ctx, bson.M{"guid": guid})
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// CreateIndexes creates necessary indexes for the news collection
func (r *NewsRepository) CreateIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.M{"guid": 1},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.M{"companies": 1},
		},
		{
			Keys: bson.M{"published_date": -1},
		},
		{
			Keys: bson.M{"relevance_score": -1},
		},
		{
			Keys: bson.M{"feed_source": 1},
		},
		{
			Keys: bson.M{
				"companies":      1,
				"published_date": -1,
			},
		},
	}

	_, err := r.collection.Indexes().CreateMany(ctx, indexes)
	return err
}

// buildFilter constructs MongoDB filter from NewsFilter
func (r *NewsRepository) buildFilter(filter *news.NewsFilter) bson.M {
	mongoFilter := bson.M{}

	if filter == nil {
		return mongoFilter
	}

	if filter.Company != "" {
		mongoFilter["companies"] = bson.M{"$in": []string{filter.Company}}
	}

	if filter.StartDate != nil || filter.EndDate != nil {
		dateFilter := bson.M{}
		if filter.StartDate != nil {
			dateFilter["$gte"] = *filter.StartDate
		}
		if filter.EndDate != nil {
			dateFilter["$lte"] = *filter.EndDate
		}
		mongoFilter["published_date"] = dateFilter
	}

	if filter.MinRelevance != nil {
		mongoFilter["relevance_score"] = bson.M{"$gte": *filter.MinRelevance}
	}

	return mongoFilter
}

// buildOptions constructs MongoDB options from NewsFilter
func (r *NewsRepository) buildOptions(filter *news.NewsFilter) *options.FindOptions {
	opts := options.Find()

	if filter == nil {
		opts.SetSort(bson.M{"published_date": -1})
		opts.SetLimit(50)
		return opts
	}

	// Sort by published date (newest first)
	opts.SetSort(bson.M{"published_date": -1})

	if filter.Limit > 0 {
		opts.SetLimit(int64(filter.Limit))
	}

	if filter.Offset > 0 {
		opts.SetSkip(int64(filter.Offset))
	}

	return opts
}