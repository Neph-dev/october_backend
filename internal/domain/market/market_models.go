package market

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// StockQuote represents real-time stock market data
type StockQuote struct {
	ID             primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Ticker         string             `bson:"ticker" json:"ticker" validate:"required"`
	CompanyName    string             `bson:"company_name" json:"company_name"`
	CurrentPrice   float64            `bson:"current_price" json:"current_price"`
	OpenPrice      float64            `bson:"open_price" json:"open_price"`
	HighPrice      float64            `bson:"high_price" json:"high_price"`
	LowPrice       float64            `bson:"low_price" json:"low_price"`
	PreviousClose  float64            `bson:"previous_close" json:"previous_close"`
	Change         float64            `bson:"change" json:"change"`
	ChangePercent  float64            `bson:"change_percent" json:"change_percent"`
	Volume         int64              `bson:"volume" json:"volume"`
	MarketCap      int64              `bson:"market_cap,omitempty" json:"market_cap,omitempty"`
	UpdatedAt      time.Time          `bson:"updated_at" json:"updated_at"`
	TradingDay     time.Time          `bson:"trading_day" json:"trading_day"`
	IsMarketOpen   bool               `bson:"is_market_open" json:"is_market_open"`
	Exchange       string             `bson:"exchange" json:"exchange"`
}



// CompanyProfile represents fundamental company information
type CompanyProfile struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Ticker            string             `bson:"ticker" json:"ticker" validate:"required"`
	Name              string             `bson:"name" json:"name"`
	Country           string             `bson:"country" json:"country"`
	Currency          string             `bson:"currency" json:"currency"`
	Exchange          string             `bson:"exchange" json:"exchange"`
	Industry          string             `bson:"industry" json:"industry"`
	Sector            string             `bson:"sector" json:"sector"`
	MarketCap         int64              `bson:"market_cap" json:"market_cap"`
	SharesOutstanding int64              `bson:"shares_outstanding" json:"shares_outstanding"`
	Description       string             `bson:"description" json:"description"`
	Website           string             `bson:"website" json:"website"`
	Logo              string             `bson:"logo" json:"logo"`
	Phone             string             `bson:"phone" json:"phone"`
	IPODate           *time.Time         `bson:"ipo_date,omitempty" json:"ipo_date,omitempty"`
	UpdatedAt         time.Time          `bson:"updated_at" json:"updated_at"`
}

// MarketStatus represents overall market status
type MarketStatus struct {
	Exchange   string    `json:"exchange"`
	Timezone   string    `json:"timezone"`
	IsOpen     bool      `json:"is_open"`
	SessionEnd time.Time `json:"session_end"`
}

// TradingViewData represents data formatted for Trading View widgets
type TradingViewData struct {
	Symbol        string    `json:"symbol"`
	Price         float64   `json:"price"`
	Change        float64   `json:"change"`
	ChangePercent float64   `json:"change_percent"`
	Volume        int64     `json:"volume"`
	MarketCap     int64     `json:"market_cap"`
	High52Week    float64   `json:"high_52_week,omitempty"`
	Low52Week     float64   `json:"low_52_week,omitempty"`
	PERatio       float64   `json:"pe_ratio,omitempty"`
	UpdatedAt     time.Time `json:"updated_at"`
}



// MarketDataCache represents cached market data with TTL
type MarketDataCache struct {
	Key       string    `json:"key"`
	Data      []byte    `json:"data"`
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}



// External API response models

// FinnhubQuoteResponse represents Finnhub real-time quote response
type FinnhubQuoteResponse struct {
	CurrentPrice   float64 `json:"c"`  // Current price
	Change         float64 `json:"d"`  // Change
	ChangePercent  float64 `json:"dp"` // Percent change
	HighPrice      float64 `json:"h"`  // High price of the day
	LowPrice       float64 `json:"l"`  // Low price of the day
	OpenPrice      float64 `json:"o"`  // Open price of the day
	PreviousClose  float64 `json:"pc"` // Previous close price
	Timestamp      int64   `json:"t"`  // UNIX timestamp
}

// FinnhubCompanyProfileResponse represents Finnhub company profile response
type FinnhubCompanyProfileResponse struct {
	Country           string  `json:"country"`
	Currency          string  `json:"currency"`
	EstimateCurrency  string  `json:"estimateCurrency"`
	Exchange          string  `json:"exchange"`
	FinnhubIndustry   string  `json:"finnhubIndustry"`
	IPO               string  `json:"ipo"`
	Logo              string  `json:"logo"`
	MarketCap         float64 `json:"marketCapitalization"`
	Name              string  `json:"name"`
	Phone             string  `json:"phone"`
	SharesOutstanding float64 `json:"shareOutstanding"`
	Ticker            string  `json:"ticker"`
	WebURL            string  `json:"weburl"`
}



// MarketDataFilter represents filtering options for market data
type MarketDataFilter struct {
	Tickers []string `json:"tickers,omitempty"`
	Limit   int      `json:"limit,omitempty"`
	Offset  int      `json:"offset,omitempty"`
}

// Response models for API
type StockQuoteResponse struct {
	*StockQuote
	Profile *CompanyProfile `json:"profile,omitempty"`
}

type MarketOverviewResponse struct {
	Quotes       []StockQuote    `json:"quotes"`
	MarketStatus *MarketStatus   `json:"market_status"`
	LastUpdated  time.Time       `json:"last_updated"`
	Count        int             `json:"count"`
}

// TickerInfo represents ticker and stock exchange information for a company
type TickerInfo struct {
	CompanyName   string `json:"company_name"`
	Ticker        string `json:"ticker"`
	StockExchange string `json:"stock_exchange"`
	Industry      string `json:"industry"`
	Country       string `json:"country"`
}

// AvailableTickersResponse represents the response for available tickers endpoint
type AvailableTickersResponse struct {
	Tickers     []TickerInfo `json:"tickers"`
	Count       int          `json:"count"`
	LastUpdated time.Time    `json:"last_updated"`
}