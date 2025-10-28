package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"

	"github.com/Neph-dev/october_backend/internal/domain/market"
	"github.com/gorilla/mux"
)

// MarketHandler handles market data HTTP requests
type MarketHandler struct {
	marketService market.Service
	logger        *slog.Logger
}

// NewMarketHandler creates a new market handler
func NewMarketHandler(marketService market.Service, logger *slog.Logger) *MarketHandler {
	return &MarketHandler{
		marketService: marketService,
		logger:        logger,
	}
}

// GetQuoteHandler handles GET /market/quote/{ticker} requests
func (h *MarketHandler) GetQuoteHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ticker := strings.ToUpper(strings.TrimSpace(vars["ticker"]))

	if ticker == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "ticker is required")
		return
	}

	h.logger.Info("Getting quote", "ticker", ticker)

	quote, err := h.marketService.GetQuote(r.Context(), ticker)
	if err != nil {
		h.logger.Error("Failed to get quote", "error", err, "ticker", ticker)
		
		if err == market.ErrTickerNotFound {
			h.writeErrorResponse(w, http.StatusNotFound, "ticker not found")
			return
		}
		
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get quote")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, quote)
}

// GetQuotesHandler handles GET /market/quotes requests with ticker query params
func (h *MarketHandler) GetQuotesHandler(w http.ResponseWriter, r *http.Request) {
	tickersParam := r.URL.Query().Get("tickers")
	if tickersParam == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "tickers parameter is required")
		return
	}

	tickers := strings.Split(tickersParam, ",")
	for i, ticker := range tickers {
		tickers[i] = strings.ToUpper(strings.TrimSpace(ticker))
	}

	h.logger.Info("Getting multiple quotes", "tickers", tickers, "count", len(tickers))

	quotes, err := h.marketService.GetQuotes(r.Context(), tickers)
	if err != nil {
		h.logger.Error("Failed to get quotes", "error", err, "tickers", tickers)
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get quotes")
		return
	}

	response := map[string]interface{}{
		"quotes": quotes,
		"count":  len(quotes),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// GetAvailableTickersHandler handles GET /market/tickers requests
func (h *MarketHandler) GetAvailableTickersHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting available tickers")

	tickers, err := h.marketService.GetAvailableTickers(r.Context())
	if err != nil {
		h.logger.Error("Failed to get available tickers", "error", err)
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get available tickers")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, tickers)
}

// GetMarketStatusHandler handles GET /market/status/{exchange} requests
func (h *MarketHandler) GetMarketStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	exchange := strings.ToUpper(strings.TrimSpace(vars["exchange"]))

	if exchange == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "exchange is required")
		return
	}

	h.logger.Info("Getting market status", "exchange", exchange)

	status, err := h.marketService.GetMarketStatus(r.Context(), exchange)
	if err != nil {
		h.logger.Error("Failed to get market status", "error", err, "exchange", exchange)
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to get market status")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, status)
}

// writeJSONResponse writes a JSON response
func (h *MarketHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// writeErrorResponse writes an error response
func (h *MarketHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}