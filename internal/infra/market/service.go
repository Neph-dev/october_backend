package market

import (
	"context"
	"fmt"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/market"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

// MarketService implements the market data service
type MarketService struct {
	finnhub     market.ExternalMarketAPI
	companyRepo company.Repository
	logger      logger.Logger
}

// NewMarketService creates a new market service instance
func NewMarketService(
	finnhub market.ExternalMarketAPI,
	companyRepo company.Repository,
	logger logger.Logger,
) *MarketService {
	return &MarketService{
		finnhub:     finnhub,
		companyRepo: companyRepo,
		logger:      logger,
	}
}

// GetQuote retrieves current market data for a ticker directly from Finnhub
func (s *MarketService) GetQuote(ctx context.Context, ticker string) (*market.StockQuoteResponse, error) {
	s.logger.Info("Getting quote from Finnhub", "ticker", ticker)
	
	// Fetch data directly from Finnhub API
	quote, err := s.finnhub.GetQuote(ctx, ticker)
	if err != nil {
		s.logger.Error("Failed to fetch quote from Finnhub", "ticker", ticker, "error", err)
		return nil, fmt.Errorf("%w: failed to fetch quote from Finnhub", market.ErrMarketDataService)
	}
	
	s.logger.Info("Quote fetched successfully from Finnhub", "ticker", ticker, "price", quote.CurrentPrice)
	
	return &market.StockQuoteResponse{
		StockQuote: quote,
		Profile:    nil, // No company profile needed since we're not storing in DB
	}, nil
}

// GetQuotes retrieves current market data for multiple tickers directly from Finnhub
func (s *MarketService) GetQuotes(ctx context.Context, tickers []string) ([]market.StockQuote, error) {
	s.logger.Info("Getting multiple quotes from Finnhub", "tickers", tickers, "count", len(tickers))
	
	var quotes []market.StockQuote
	
	for _, ticker := range tickers {
		// Fetch data directly from Finnhub API
		quote, err := s.finnhub.GetQuote(ctx, ticker)
		if err != nil {
			s.logger.Warn("Failed to fetch quote from Finnhub", "ticker", ticker, "error", err)
			continue // Skip this ticker entirely
		}
		
		s.logger.Info("Quote fetched successfully from Finnhub", "ticker", ticker, "price", quote.CurrentPrice)
		quotes = append(quotes, *quote)
		
		// Add small delay to respect rate limits (Finnhub allows up to 60 requests per minute)
		time.Sleep(250 * time.Millisecond)
	}
	
	s.logger.Info("Multiple quotes retrieval completed", "requested", len(tickers), "retrieved", len(quotes))
	return quotes, nil
}










// GetAvailableTickers retrieves available tickers from stored companies
func (s *MarketService) GetAvailableTickers(ctx context.Context) (*market.AvailableTickersResponse, error) {
	s.logger.Info("Getting available tickers from companies")
	
	// Get all companies from the database
	companies, err := s.companyRepo.List(ctx, 100, 0) // Get up to 100 companies
	if err != nil {
		s.logger.Error("Failed to get companies list", "error", err)
		return nil, fmt.Errorf("%w: failed to get companies list", market.ErrMarketDataService)
	}
	
	var tickers []market.TickerInfo
	for _, comp := range companies {
		// Only include companies that have a ticker (non-government entities)
		if comp.Ticker != "" && comp.StockExchange != "" {
			tickerInfo := market.TickerInfo{
				CompanyName:   comp.Name,
				Ticker:        comp.Ticker,
				StockExchange: comp.StockExchange,
				Industry:      string(comp.Industry),
				Country:       comp.Country,
			}
			tickers = append(tickers, tickerInfo)
		}
	}
	
	response := &market.AvailableTickersResponse{
		Tickers:     tickers,
		Count:       len(tickers),
		LastUpdated: time.Now(),
	}
	
	s.logger.Info("Available tickers retrieved successfully", "count", len(tickers))
	return response, nil
}





