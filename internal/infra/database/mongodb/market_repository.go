package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/market"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	quotesCollection         = "stock_quotes"
	companyProfileCollection = "company_profiles"
)

// MarketRepository implements market.Repository using MongoDB
type MarketRepository struct {
	db *mongo.Database
}

// NewMarketRepository creates a new MongoDB market repository
func NewMarketRepository(db *mongo.Database) *MarketRepository {
	return &MarketRepository{
		db: db,
	}
}

// SaveQuote saves a stock quote to the database
func (r *MarketRepository) SaveQuote(ctx context.Context, quote *market.StockQuote) error {
	collection := r.db.Collection(quotesCollection)
	
	// Use upsert to replace existing quote for the same ticker and trading day
	filter := bson.M{
		"ticker":      quote.Ticker,
		"trading_day": quote.TradingDay,
	}
	
	quote.UpdatedAt = time.Now()
	
	opts := options.Replace().SetUpsert(true)
	_, err := collection.ReplaceOne(ctx, filter, quote, opts)
	if err != nil {
		return fmt.Errorf("failed to save quote: %w", err)
	}
	
	return nil
}

// GetQuote retrieves the latest quote for a ticker
func (r *MarketRepository) GetQuote(ctx context.Context, ticker string) (*market.StockQuote, error) {
	collection := r.db.Collection(quotesCollection)

	filter := bson.M{"ticker": ticker}
	opts := options.FindOne().SetSort(bson.M{"updated_at": -1})
	
	var quote market.StockQuote
	err := collection.FindOne(ctx, filter, opts).Decode(&quote)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, market.ErrTickerNotFound
		}
		return nil, fmt.Errorf("failed to get quote: %w", err)
	}
	
	return &quote, nil
}

// GetQuotes retrieves the latest quotes for multiple tickers
func (r *MarketRepository) GetQuotes(ctx context.Context, tickers []string) ([]market.StockQuote, error) {
	collection := r.db.Collection(quotesCollection)
	
	// Build aggregation pipeline to get latest quote for each ticker
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"ticker": bson.M{"$in": tickers},
			},
		},
		{
			"$sort": bson.M{"updated_at": -1},
		},
		{
			"$group": bson.M{
				"_id": "$ticker",
				"doc": bson.M{"$first": "$$ROOT"},
			},
		},
		{
			"$replaceRoot": bson.M{"newRoot": "$doc"},
		},
	}
	
	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("failed to get quotes: %w", err)
	}
	defer cursor.Close(ctx)
	
	var quotes []market.StockQuote
	if err := cursor.All(ctx, &quotes); err != nil {
		return nil, fmt.Errorf("failed to decode quotes: %w", err)
	}
	
	return quotes, nil
}

// GetLatestQuotes retrieves the most recent quotes with a limit
func (r *MarketRepository) GetLatestQuotes(ctx context.Context, limit int) ([]market.StockQuote, error) {
	collection := r.db.Collection(quotesCollection)
	
	opts := options.Find().
		SetSort(bson.M{"updated_at": -1}).
		SetLimit(int64(limit))
	
	cursor, err := collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest quotes: %w", err)
	}
	defer cursor.Close(ctx)
	
	var quotes []market.StockQuote
	if err := cursor.All(ctx, &quotes); err != nil {
		return nil, fmt.Errorf("failed to decode quotes: %w", err)
	}
	
	return quotes, nil
}







// SaveCompanyProfile saves company profile information
func (r *MarketRepository) SaveCompanyProfile(ctx context.Context, profile *market.CompanyProfile) error {
	collection := r.db.Collection(companyProfileCollection)
	
	filter := bson.M{"ticker": profile.Ticker}
	profile.UpdatedAt = time.Now()
	
	opts := options.Replace().SetUpsert(true)
	_, err := collection.ReplaceOne(ctx, filter, profile, opts)
	if err != nil {
		return fmt.Errorf("failed to save company profile: %w", err)
	}
	
	return nil
}

// GetCompanyProfile retrieves company profile information
func (r *MarketRepository) GetCompanyProfile(ctx context.Context, ticker string) (*market.CompanyProfile, error) {
	collection := r.db.Collection(companyProfileCollection)
	
	filter := bson.M{"ticker": ticker}
	
	var profile market.CompanyProfile
	err := collection.FindOne(ctx, filter).Decode(&profile)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, market.ErrTickerNotFound
		}
		return nil, fmt.Errorf("failed to get company profile: %w", err)
	}
	
	return &profile, nil
}

// GetAllTickers retrieves all unique tickers from stored quotes
func (r *MarketRepository) GetAllTickers(ctx context.Context) ([]string, error) {
	collection := r.db.Collection(quotesCollection)
	
	tickers, err := collection.Distinct(ctx, "ticker", bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to get tickers: %w", err)
	}
	
	// Convert to string slice
	stringTickers := make([]string, len(tickers))
	for i, ticker := range tickers {
		stringTickers[i] = ticker.(string)
	}
	
	return stringTickers, nil
}

// CreateIndexes creates necessary indexes for market data collections
func (r *MarketRepository) CreateIndexes(ctx context.Context) error {
	// Stock quotes indexes
	quotesCol := r.db.Collection(quotesCollection)
	quotesIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "ticker", Value: 1}, {Key: "trading_day", Value: -1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "ticker", Value: 1}, {Key: "updated_at", Value: -1}},
		},
		{
			Keys: bson.D{{Key: "updated_at", Value: -1}},
		},
	}
	
	_, err := quotesCol.Indexes().CreateMany(ctx, quotesIndexes)
	if err != nil {
		return fmt.Errorf("failed to create quotes indexes: %w", err)
	}
	
	// Company profiles indexes
	profileCol := r.db.Collection(companyProfileCollection)
	profileIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "ticker", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
	}
	
	_, err = profileCol.Indexes().CreateMany(ctx, profileIndexes)
	if err != nil {
		return fmt.Errorf("failed to create profile indexes: %w", err)
	}
	
	return nil
}