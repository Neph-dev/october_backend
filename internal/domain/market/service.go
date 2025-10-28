package market

import (
	"context"
	"errors"
)

var (
	ErrTickerNotFound    = errors.New("ticker not found")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrMarketDataService = errors.New("market data service error")
	ErrCacheError        = errors.New("cache error")
)

type Service interface {
	// GetQuote retrieves current market data for a ticker directly from Finnhub
	GetQuote(ctx context.Context, ticker string) (*StockQuoteResponse, error)
	
	// GetQuotes retrieves current market data for multiple tickers directly from Finnhub
	GetQuotes(ctx context.Context, tickers []string) ([]StockQuote, error)
	
	// GetAvailableTickers retrieves available tickers from stored companies
	GetAvailableTickers(ctx context.Context) (*AvailableTickersResponse, error)
	
	// GetMarketStatus retrieves current market status for a given exchange
	GetMarketStatus(ctx context.Context, exchange string) (*MarketStatus, error)
}

type ExternalMarketAPI interface {
	// GetQuote fetches current quote from external API
	GetQuote(ctx context.Context, ticker string) (*StockQuote, error)
	
	// GetCompanyProfile fetches company profile from external API
	GetCompanyProfile(ctx context.Context, ticker string) (*CompanyProfile, error)
	
	// GetMarketStatus fetches market status from external API
	GetMarketStatus(ctx context.Context, exchange string) (*MarketStatus, error)
}