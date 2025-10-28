package market

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/market"
	"github.com/Neph-dev/october_backend/pkg/logger"
)

const (
	finnhubBaseURL = "https://finnhub.io/api/v1"
	
	// API endpoints
	finnhubQuoteEndpoint        = "/quote"
	finnhubProfileEndpoint      = "/stock/profile2"
	finnhubMarketStatusEndpoint = "/stock/market-status"
)

// FinnhubService implements ExternalMarketAPI for Finnhub integration
type FinnhubService struct {
	apiKey     string
	httpClient *http.Client
	logger     logger.Logger
}

// NewFinnhubService creates a new Finnhub service instance
func NewFinnhubService(apiKey string, logger logger.Logger) *FinnhubService {
	return &FinnhubService{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// GetQuote fetches current quote from Finnhub API
func (f *FinnhubService) GetQuote(ctx context.Context, ticker string) (*market.StockQuote, error) {
	f.logger.Info("Fetching quote from Finnhub", "ticker", ticker)
	
	u, err := f.buildURL(finnhubQuoteEndpoint, map[string]string{
		"symbol": ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var finnhubResponse market.FinnhubQuoteResponse
	if err := json.NewDecoder(resp.Body).Decode(&finnhubResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check if we got a valid response (Finnhub returns 0 values for invalid tickers)
	if finnhubResponse.CurrentPrice == 0 && finnhubResponse.Timestamp == 0 {
		return nil, fmt.Errorf("%w: %s", market.ErrTickerNotFound, ticker)
	}

	fmt.Println("finnhubResponse...", finnhubResponse)
	// Convert Finnhub response to our standard format
	quote := f.convertFinnhubQuote(ticker, finnhubResponse)

	f.logger.Info("Quote fetched successfully", "ticker", ticker, "price", quote.CurrentPrice)
	return quote, nil
}

// GetCompanyProfile fetches company profile from Finnhub API
func (f *FinnhubService) GetCompanyProfile(ctx context.Context, ticker string) (*market.CompanyProfile, error) {
	f.logger.Info("Fetching company profile from Finnhub", "ticker", ticker)
	
	u, err := f.buildURL(finnhubProfileEndpoint, map[string]string{
		"symbol": ticker,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var finnhubProfile market.FinnhubCompanyProfileResponse
	if err := json.NewDecoder(resp.Body).Decode(&finnhubProfile); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Validate the response
	if finnhubProfile.Name == "" || finnhubProfile.Ticker == "" {
		return nil, fmt.Errorf("%w: %s", market.ErrTickerNotFound, ticker)
	}

	// Convert Finnhub profile to our standard format
	profile := f.convertFinnhubProfile(ticker, finnhubProfile)

	f.logger.Info("Company profile fetched successfully", "ticker", ticker, "name", profile.Name)
	return profile, nil
}

// GetMarketStatus fetches market status from Finnhub API
func (f *FinnhubService) GetMarketStatus(ctx context.Context, exchange string) (*market.MarketStatus, error) {
	f.logger.Info("Fetching market status from Finnhub", "exchange", exchange)
	
	u, err := f.buildURL(finnhubMarketStatusEndpoint, map[string]string{
		"exchange": exchange,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", u, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	var finnhubResponse market.FinnhubMarketStatusResponse
	if err := json.NewDecoder(resp.Body).Decode(&finnhubResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert Finnhub response to our standard format
	status := f.convertFinnhubMarketStatus(finnhubResponse)

	f.logger.Info("Market status fetched successfully", "exchange", exchange, "is_open", status.IsOpen, "session", status.Session)
	return status, nil
}

// buildURL builds a complete URL with query parameters
func (f *FinnhubService) buildURL(endpoint string, params map[string]string) (string, error) {
	u, err := url.Parse(finnhubBaseURL + endpoint)
	if err != nil {
		return "", err
	}

	q := u.Query()
	q.Set("token", f.apiKey)
	
	for key, value := range params {
		q.Set(key, value)
	}
	
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// convertFinnhubQuote converts Finnhub quote to standard format
func (f *FinnhubService) convertFinnhubQuote(ticker string, finnhubQuote market.FinnhubQuoteResponse) *market.StockQuote {
	return &market.StockQuote{
		Ticker:        ticker,
		CurrentPrice:  finnhubQuote.CurrentPrice,
		Change:        finnhubQuote.Change,
		ChangePercent: finnhubQuote.ChangePercent,
		HighPrice:     finnhubQuote.HighPrice,
		LowPrice:      finnhubQuote.LowPrice,
		OpenPrice:     finnhubQuote.OpenPrice,
		PreviousClose: finnhubQuote.PreviousClose,
		UpdatedAt:     time.Now(),
		TradingDay:    time.Unix(finnhubQuote.Timestamp, 0),
		Exchange:      "NYSE", // Default exchange
	}
}

// convertFinnhubProfile converts Finnhub profile to standard format
func (f *FinnhubService) convertFinnhubProfile(ticker string, finnhubProfile market.FinnhubCompanyProfileResponse) *market.CompanyProfile {
	// Parse IPO date
	var ipoDate *time.Time
	if finnhubProfile.IPO != "" {
		if parsed, err := time.Parse("2006-01-02", finnhubProfile.IPO); err == nil {
			ipoDate = &parsed
		}
	}

	// Convert market cap from millions to actual value
	marketCap := int64(finnhubProfile.MarketCap * 1000000)
	sharesOutstanding := int64(finnhubProfile.SharesOutstanding * 1000000)

	return &market.CompanyProfile{
		Ticker:            ticker,
		Name:              finnhubProfile.Name,
		Country:           finnhubProfile.Country,
		Currency:          finnhubProfile.Currency,
		Exchange:          finnhubProfile.Exchange,
		Industry:          finnhubProfile.FinnhubIndustry,
		Sector:            "", // Finnhub doesn't provide sector in this endpoint
		MarketCap:         marketCap,
		SharesOutstanding: sharesOutstanding,
		Description:       "", // Finnhub doesn't provide description in this endpoint
		Website:           finnhubProfile.WebURL,
		Logo:              finnhubProfile.Logo,
		Phone:             finnhubProfile.Phone,
		IPODate:           ipoDate,
		UpdatedAt:         time.Now(),
	}
}

// convertFinnhubMarketStatus converts Finnhub market status to standard format
func (f *FinnhubService) convertFinnhubMarketStatus(finnhubStatus market.FinnhubMarketStatusResponse) *market.MarketStatus {
	// Calculate session end based on market session
	var sessionEnd time.Time
	now := time.Unix(finnhubStatus.T, 0)
	
	// Parse timezone and convert to location
	loc, err := time.LoadLocation(finnhubStatus.Timezone)
	if err != nil {
		loc = time.UTC // fallback to UTC if timezone parsing fails
	}
	
	nowInTz := now.In(loc)
	
	if finnhubStatus.IsOpen {
		// If market is open, assume it closes at 4 PM in the market's timezone
		sessionEnd = time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(), 16, 0, 0, 0, loc)
		// If current time is already past 4 PM, set to next business day
		if nowInTz.Hour() >= 16 {
			sessionEnd = sessionEnd.Add(24 * time.Hour)
			for sessionEnd.Weekday() == time.Saturday || sessionEnd.Weekday() == time.Sunday {
				sessionEnd = sessionEnd.Add(24 * time.Hour)
			}
		}
	} else {
		// If market is closed, find next opening (9:30 AM next business day)
		sessionEnd = time.Date(nowInTz.Year(), nowInTz.Month(), nowInTz.Day(), 9, 30, 0, 0, loc)
		if nowInTz.Hour() >= 16 || nowInTz.Weekday() == time.Saturday || nowInTz.Weekday() == time.Sunday {
			sessionEnd = sessionEnd.Add(24 * time.Hour)
		}
		for sessionEnd.Weekday() == time.Saturday || sessionEnd.Weekday() == time.Sunday {
			sessionEnd = sessionEnd.Add(24 * time.Hour)
		}
	}

	return &market.MarketStatus{
		Exchange:   finnhubStatus.Exchange,
		Timezone:   finnhubStatus.Timezone,
		IsOpen:     finnhubStatus.IsOpen,
		Session:    finnhubStatus.Session,
		Holiday:    finnhubStatus.Holiday,
		SessionEnd: sessionEnd,
		UpdatedAt:  time.Now(),
	}
}
