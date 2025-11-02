# Alpha Vantage Service Implementation

## Overview

The Alpha Vantage service implementation provides market data integration using the Alpha Vantage API. This service replaces the previous Finnhub integration to provide better access to historical market data on the free tier.

## File Structure

```
internal/infra/market/
├── alphavantage.go         # Alpha Vantage service implementation
├── market_service.go       # Main market service (uses Alpha Vantage)
└── README.md              # This documentation file

internal/domain/market/
├── market_models.go        # Data models including Alpha Vantage response models
└── market_interfaces.go   # Interface definitions
```

## Key Components

### AlphaVantageService

**File:** `internal/infra/market/alphavantage.go`

**Purpose:** Implements the `ExternalMarketAPI` interface using Alpha Vantage's REST API.

**Key Methods:**
- `GetQuote(ctx, ticker)` - Fetches real-time quote data using `GLOBAL_QUOTE` function
- `GetCompanyProfile(ctx, ticker)` - Fetches company overview using `OVERVIEW` function  
- `GetHistoricalData(ctx, ticker, from, to)` - Fetches historical data using `TIME_SERIES_DAILY` function

**Configuration:**
- Requires `ALPHA_VANTAGE_API_KEY` environment variable
- Uses 30-second HTTP timeout
- Implements rate limiting awareness (5 requests/minute)

### Data Models

**File:** `internal/domain/market/market_models.go`

**Alpha Vantage Response Models:**
- `AlphaVantageQuoteResponse` - Maps to Alpha Vantage's Global Quote response
- `AlphaVantageCompanyOverviewResponse` - Maps to Alpha Vantage's Overview response
- `AlphaVantageTimeSeriesResponse` - Maps to Alpha Vantage's Time Series Daily response

**Internal Models:**
- `Quote` - Standardized quote data model
- `CompanyProfile` - Standardized company information model
- `HistoricalData` - Standardized historical price data model

## API Functions Used

### 1. GLOBAL_QUOTE
- **Purpose:** Real-time stock quotes
- **Endpoint:** `https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol={ticker}&apikey={key}`
- **Rate Limit Impact:** 1 request per call
- **Data Returned:** Current price, open, high, low, volume, change, etc.

### 2. OVERVIEW  
- **Purpose:** Company profile and fundamental data
- **Endpoint:** `https://www.alphavantage.co/query?function=OVERVIEW&symbol={ticker}&apikey={key}`
- **Rate Limit Impact:** 1 request per call
- **Data Returned:** Company description, financials, sector, industry, etc.

### 3. TIME_SERIES_DAILY
- **Purpose:** Historical daily price data
- **Endpoint:** `https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol={ticker}&apikey={key}`
- **Rate Limit Impact:** 1 request per call
- **Data Returned:** Daily OHLCV data for up to 100 days

## Response Conversion

The service includes conversion methods to transform Alpha Vantage's string-based responses into strongly-typed internal models:

### Quote Conversion
```go
// Alpha Vantage returns quote data as map[string]string
// Example: {"05. price": "485.41", "09. change": "-2.64"}
// Converts to: Quote{CurrentPrice: 485.41, Change: -2.64}
```

### Company Profile Conversion
```go
// Alpha Vantage returns all fields as strings
// Example: {"MarketCapitalization": "123456789000", "PERatio": "15.5"}
// Converts to: CompanyProfile{MarketCap: 123456789000, PERatio: 15.5}
```

### Historical Data Conversion
```go
// Alpha Vantage returns time series as nested maps
// Example: {"2025-10-24": {"1. open": "490.28", "4. close": "485.41"}}
// Converts to: []HistoricalData{{Date: time.Time, Open: 490.28, Close: 485.41}}
```

## Error Handling

The service handles various error scenarios:

### 1. Rate Limit Exceeded
- **Detection:** HTTP 429 status or specific error messages
- **Handling:** Returns appropriate error with retry suggestions
- **Mitigation:** Built-in delays between requests

### 2. Invalid API Key
- **Detection:** HTTP 401 status or authentication errors
- **Handling:** Returns configuration error
- **Resolution:** Check `ALPHA_VANTAGE_API_KEY` environment variable

### 3. Invalid Ticker Symbol
- **Detection:** Empty response data or "Error Message" in response
- **Handling:** Returns "ticker not found" error
- **Examples:** Non-existent symbols, delisted stocks

### 4. Network Errors
- **Detection:** HTTP client errors, timeouts
- **Handling:** Returns network-related errors
- **Mitigation:** 30-second timeout, proper connection management

## Rate Limiting

### Alpha Vantage Free Tier Limits
- **Requests per minute:** 5
- **Requests per day:** 500
- **Premium upgrade:** Available for higher limits

### Implementation Strategy
1. **Request Spacing:** Automatic delays between requests
2. **Caching:** Aggressive caching in market service layer
3. **Error Handling:** Graceful degradation when limits exceeded
4. **Monitoring:** Logging of API usage patterns

## Testing

### Unit Tests
Test the Alpha Vantage service with:
```bash
go test ./internal/infra/market/... -v
```

### Integration Tests
Test with real API (requires valid API key):
```bash
ALPHA_VANTAGE_API_KEY="your-key" go test ./internal/infra/market/... -tags=integration -v
```

### Manual Testing
Use curl to test the endpoints:
```bash
# Start the application
SERVER_PORT=8082 ./bin/api

# Test quote endpoint
curl "http://localhost:8082/market/quote/LMT"

# Test historical data
curl "http://localhost:8082/market/historical/LMT"
```

## Configuration

### Environment Variables
```env
# Required: Alpha Vantage API key
ALPHA_VANTAGE_API_KEY="your_api_key_here"

# Optional: Override default timeout
ALPHA_VANTAGE_TIMEOUT="30s"
```

### Service Registration
The service is registered in `cmd/api/main.go`:
```go
// Initialize market service with Alpha Vantage integration
marketRepo := mongodb.NewMarketRepository(app.dbClient.Database())
alphaVantageService := marketInfra.NewAlphaVantageService(app.config.Market.AlphaVantageAPIKey, app.logger)
marketCache := cache.NewGenericCache()
app.marketService = marketInfra.NewMarketService(
    marketRepo,
    alphaVantageService,
    marketCache,
    companyRepo,
    app.logger,
)
```

## Migration from Finnhub

### What Changed
1. **Service Implementation:** Replaced `FinnhubService` with `AlphaVantageService`
2. **Configuration:** Added `AlphaVantageAPIKey` to config, updated validation
3. **API Mapping:** 
   - Quotes: `/quote` → `GLOBAL_QUOTE`
   - Profiles: `/stock/profile2` → `OVERVIEW`
   - Historical: `/stock/candle` → `TIME_SERIES_DAILY`

### Benefits
- ✅ **Free Historical Data:** Available on Alpha Vantage free tier
- ✅ **No 403 Errors:** Resolves Finnhub historical data limitations
- ✅ **Comprehensive Data:** More detailed company information
- ✅ **Better Documentation:** Extensive API documentation available

### Considerations
- ⚠️ **Lower Rate Limits:** 5/min vs Finnhub's higher limits
- ⚠️ **Different Format:** String-based responses require conversion
- ⚠️ **Premium Features:** Advanced features require paid plans

## Troubleshooting

### Common Issues

#### 1. "Failed to get historical data"
**Cause:** Usually rate limiting or invalid ticker
**Solution:** 
- Check API key configuration
- Verify ticker symbol is valid
- Wait for rate limit reset (1 minute)

#### 2. Empty quote responses
**Cause:** Invalid ticker or API key issues
**Solution:**
- Validate ticker symbol format
- Check API key in Alpha Vantage dashboard
- Review application logs for detailed errors

#### 3. Rate limit exceeded errors
**Cause:** Too many requests in short timeframe
**Solution:**
- Implement proper caching
- Space out requests appropriately
- Consider upgrading to premium tier

### Debug Logging
Enable debug logging for troubleshooting:
```bash
LOG_LEVEL=debug ./bin/api
```

Look for log entries like:
```
{"level":"INFO","msg":"Getting quote","ticker":"LMT"}
{"level":"INFO","msg":"Successfully retrieved quote","ticker":"LMT","price":485.41}
```

## Performance Considerations

### Caching Strategy
- **Quote Data:** Cache for 1-5 minutes during market hours
- **Historical Data:** Cache for several hours (data doesn't change)
- **Company Profiles:** Cache for days (fundamental data changes slowly)

### Request Optimization
- **Batch Operations:** Group requests when possible
- **Smart Scheduling:** Refresh data during off-peak hours
- **Circuit Breaker:** Implement fallback mechanisms for failures

## Future Enhancements

1. **WebSocket Support:** Real-time data streaming (premium feature)
2. **Multiple Providers:** Fallback to other APIs when Alpha Vantage is unavailable
3. **Advanced Caching:** Redis-based distributed caching
4. **Analytics:** Technical indicators and market analysis
5. **Monitoring:** Comprehensive API usage and performance metrics

## References

- [Alpha Vantage API Documentation](https://www.alphavantage.co/documentation/)
- [Alpha Vantage Support](https://www.alphavantage.co/support/)
- [Get Free API Key](https://www.alphavantage.co/support/#api-key)
- [Premium Plans](https://www.alphavantage.co/premium/)

## Support

For issues with the Alpha Vantage integration:

1. Check application logs for detailed error messages
2. Verify API key configuration and validity
3. Review Alpha Vantage API status and limits
4. Test with simple curl requests to isolate issues
5. Check network connectivity and firewall settings

## License

This implementation is part of the October Backend project and follows the same license terms.