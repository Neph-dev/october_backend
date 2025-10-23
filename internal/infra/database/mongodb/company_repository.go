package mongodb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

type CompanyRepository struct {
	collection *mongo.Collection
	logger     logger.Logger
}

func NewCompanyRepository(db *mongo.Database, logger logger.Logger) company.Repository {
	collection := db.Collection("companies")
	
	// Create indexes for better performance
	// Following NASA's rule of being explicit about assumptions
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		
		// Create unique index on company name
		nameIndex := mongo.IndexModel{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		}
		
		// Create unique index on ticker (partial index, only for non-empty tickers)
		tickerIndex := mongo.IndexModel{
			Keys:    bson.D{{Key: "ticker", Value: 1}},
			Options: options.Index().SetUnique(true).SetPartialFilterExpression(bson.D{
				{Key: "ticker", Value: bson.D{{Key: "$ne", Value: ""}}},
			}),
		}
		
		// Create compound index on country and industry
		compoundIndex := mongo.IndexModel{
			Keys: bson.D{
				{Key: "country", Value: 1},
				{Key: "industry", Value: 1},
			},
		}
		
		indexes := []mongo.IndexModel{nameIndex, tickerIndex, compoundIndex}
		_, err := collection.Indexes().CreateMany(ctx, indexes)
		if err != nil {
			logger.Error("Failed to create indexes", "error", err)
		} else {
			logger.Info("Database indexes created successfully")
		}
	}()
	
	return &CompanyRepository{
		collection: collection,
		logger:     logger,
	}
}

// Create inserts a new company into the database
func (r *CompanyRepository) Create(ctx context.Context, comp *company.Company) error {
	if comp == nil {
		return company.ErrInvalidCompanyData
	}
	
	if !comp.Industry.IsValid() {
		return fmt.Errorf("%w: invalid industry value", company.ErrInvalidCompanyData)
	}
	
	now := time.Now()
	comp.CreatedAt = now
	comp.UpdatedAt = now
	
	result, err := r.collection.InsertOne(ctx, comp)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return company.ErrCompanyExists
		}
		r.logger.Error("Failed to create company", "error", err, "company", comp.Name)
		return fmt.Errorf("failed to create company: %w", err)
	}
	
	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		comp.ID = oid
	}
	
	r.logger.Info("Company created successfully", "id", comp.ID.Hex(), "name", comp.Name)
	return nil
}

func (r *CompanyRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*company.Company, error) {
	if id.IsZero() {
		return nil, company.ErrInvalidCompanyData
	}
	
	var comp company.Company
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&comp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, company.ErrCompanyNotFound
		}
		r.logger.Error("Failed to get company by ID", "error", err, "id", id.Hex())
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	
	return &comp, nil
}

func (r *CompanyRepository) GetByName(ctx context.Context, name string) (*company.Company, error) {
	if name == "" {
		return nil, company.ErrInvalidCompanyData
	}
	
	filter := bson.M{
		"name": bson.M{
			"$regex":   fmt.Sprintf("^%s$", name),
			"$options": "i", // case-insensitive
		},
	}
	
	var comp company.Company
	err := r.collection.FindOne(ctx, filter).Decode(&comp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, company.ErrCompanyNotFound
		}
		r.logger.Error("Failed to get company by name", "error", err, "name", name)
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	
	return &comp, nil
}

func (r *CompanyRepository) GetByTicker(ctx context.Context, ticker string) (*company.Company, error) {
	if ticker == "" {
		return nil, company.ErrInvalidCompanyData
	}
	
	filter := bson.M{"ticker": ticker}
	
	var comp company.Company
	err := r.collection.FindOne(ctx, filter).Decode(&comp)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, company.ErrCompanyNotFound
		}
		r.logger.Error("Failed to get company by ticker", "error", err, "ticker", ticker)
		return nil, fmt.Errorf("failed to get company: %w", err)
	}
	
	return &comp, nil
}

func (r *CompanyRepository) Update(ctx context.Context, comp *company.Company) error {
	if comp == nil || comp.ID.IsZero() {
		return company.ErrInvalidCompanyData
	}
	
	if !comp.Industry.IsValid() {
		return fmt.Errorf("%w: invalid industry value", company.ErrInvalidCompanyData)
	}
	
	comp.UpdatedAt = time.Now()
	
	filter := bson.M{"_id": comp.ID}
	update := bson.M{"$set": comp}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		r.logger.Error("Failed to update company", "error", err, "id", comp.ID.Hex())
		return fmt.Errorf("failed to update company: %w", err)
	}
	
	if result.MatchedCount == 0 {
		return company.ErrCompanyNotFound
	}
	
	r.logger.Info("Company updated successfully", "id", comp.ID.Hex(), "name", comp.Name)
	return nil
}

func (r *CompanyRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	if id.IsZero() {
		return company.ErrInvalidCompanyData
	}
	
	filter := bson.M{"_id": id}
	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		r.logger.Error("Failed to delete company", "error", err, "id", id.Hex())
		return fmt.Errorf("failed to delete company: %w", err)
	}
	
	if result.DeletedCount == 0 {
		return company.ErrCompanyNotFound
	}
	
	r.logger.Info("Company deleted successfully", "id", id.Hex())
	return nil
}

func (r *CompanyRepository) List(ctx context.Context, limit, offset int) ([]*company.Company, error) {
	if limit <= 0 || limit > 100 {
		limit = 20 // Default limit
	}
	if offset < 0 {
		offset = 0
	}
	
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSkip(int64(offset)).
		SetSort(bson.D{{Key: "name", Value: 1}}) // Sort by name
	
	cursor, err := r.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		r.logger.Error("Failed to list companies", "error", err)
		return nil, fmt.Errorf("failed to list companies: %w", err)
	}
	defer cursor.Close(ctx)
	
	var companies []*company.Company
	for cursor.Next(ctx) {
		var comp company.Company
		if err := cursor.Decode(&comp); err != nil {
			r.logger.Error("Failed to decode company", "error", err)
			continue // Skip malformed documents
		}
		companies = append(companies, &comp)
	}
	
	if err := cursor.Err(); err != nil {
		r.logger.Error("Cursor error while listing companies", "error", err)
		return nil, fmt.Errorf("cursor error: %w", err)
	}
	
	return companies, nil
}