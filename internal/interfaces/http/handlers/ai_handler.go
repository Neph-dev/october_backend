package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/gorilla/mux"
)

// AIHandler handles AI/RAG-related HTTP requests
type AIHandler struct {
	aiService ai.Service
	logger    *slog.Logger
}

// NewAIHandler creates a new AI handler
func NewAIHandler(aiService ai.Service, logger *slog.Logger) *AIHandler {
	return &AIHandler{
		aiService: aiService,
		logger:    logger,
	}
}

// QueryHandler handles POST /ai/query requests
func (h *AIHandler) QueryHandler(w http.ResponseWriter, r *http.Request) {
	var req ai.QueryRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Invalid JSON in AI query request", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	if strings.TrimSpace(req.Question) == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "question is required")
		return
	}

	if len(req.Question) > 1000 {
		h.writeErrorResponse(w, http.StatusBadRequest, "question too long (max 1000 characters)")
		return
	}

	h.logger.Info("Processing AI query", "question", req.Question, "company_context", req.CompanyContext)

	response, err := h.aiService.ProcessQuery(r.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to process AI query", "error", err, "question", req.Question)
		
		if err == ai.ErrInvalidQuery {
			h.writeErrorResponse(w, http.StatusBadRequest, "invalid query: "+err.Error())
			return
		}
		if err == ai.ErrNoResults {
			h.writeErrorResponse(w, http.StatusNotFound, "no relevant information found")
			return
		}
		
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to process query")
		return
	}

	h.logger.Info("AI query processed successfully", 
		"processing_time", response.ProcessingTime,
		"confidence", response.Confidence,
		"sources_count", len(response.Sources))

	h.writeJSONResponse(w, http.StatusOK, response)
}

// AnalyzeQueryHandler handles POST /ai/analyze requests for query analysis only
func (h *AIHandler) AnalyzeQueryHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Question string `json:"question"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warn("Invalid JSON in analyze query request", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	if strings.TrimSpace(request.Question) == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "question is required")
		return
	}

	analysis, err := h.aiService.AnalyzeQuery(r.Context(), request.Question)
	if err != nil {
		h.logger.Error("Failed to analyze query", "error", err, "question", request.Question)
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to analyze query")
		return
	}

	h.writeJSONResponse(w, http.StatusOK, analysis)
}

// WebSearchHandler handles POST /ai/web-search requests for testing web search functionality
func (h *AIHandler) WebSearchHandler(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Question string   `json:"question"`
		Companies []string `json:"companies,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.logger.Warn("Invalid JSON in web search request", "error", err)
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid JSON format")
		return
	}

	if strings.TrimSpace(request.Question) == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "question is required")
		return
	}

	h.logger.Info("Processing web search", "question", request.Question, "companies", request.Companies)

	results, err := h.aiService.SearchWeb(r.Context(), request.Question, request.Companies)
	if err != nil {
		h.logger.Error("Failed to perform web search", "error", err, "question", request.Question)
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to perform web search: "+err.Error())
		return
	}

	response := map[string]interface{}{
		"question": request.Question,
		"companies": request.Companies,
		"results": results,
		"count": len(results),
	}

	h.writeJSONResponse(w, http.StatusOK, response)
}

// writeJSONResponse writes a JSON response
func (h *AIHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("Failed to encode JSON response", "error", err)
	}
}

// writeErrorResponse writes an error response
func (h *AIHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	errorResponse := map[string]string{
		"error":   http.StatusText(statusCode),
		"message": message,
	}

	h.writeJSONResponse(w, statusCode, errorResponse)
}

// SummarizeArticleHandler handles GET /ai/summarise/{articleId} requests
func (h *AIHandler) SummarizeArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleID := vars["articleId"]

	if strings.TrimSpace(articleID) == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "article ID is required")
		return
	}

	h.logger.Info("Processing article summarization request", "article_id", articleID)

	response, err := h.aiService.SummarizeArticle(r.Context(), articleID)
	if err != nil {
		h.logger.Error("Failed to summarize article", "error", err, "article_id", articleID)
		
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "failed to retrieve article") {
			h.writeErrorResponse(w, http.StatusNotFound, "article not found")
			return
		}
		
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to summarize article")
		return
	}

	h.logger.Info("Article summarization completed successfully", 
		"article_id", articleID,
		"processing_time", response.ProcessingTime)

	h.writeJSONResponse(w, http.StatusOK, response)
}

// CacheStatsHandler handles GET /ai/cache/stats requests
func (h *AIHandler) CacheStatsHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Getting cache statistics")

	stats := h.aiService.GetCacheStats()

	h.writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"cache_stats": stats,
		"timestamp":   time.Now().UTC(),
	})
}