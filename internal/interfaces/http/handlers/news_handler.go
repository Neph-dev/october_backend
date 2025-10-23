package handlers

import (
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/Neph-dev/october_backend/internal/interfaces/dto"
	"github.com/gorilla/mux"
)

// NewsHandler handles HTTP requests for news operations
type NewsHandler struct {
	newsService *news.Service
	logger      *slog.Logger
}

// NewNewsHandler creates a new news handler
func NewNewsHandler(newsService *news.Service, logger *slog.Logger) *NewsHandler {
	return &NewsHandler{
		newsService: newsService,
		logger:      logger,
	}
}

// GetNews handles GET /news requests
func (h *NewsHandler) GetNews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	filter, err := h.parseNewsFilter(r)
	if err != nil {
		h.logger.Error("Invalid filter parameters", "error", err)
		dto.WriteErrorResponse(w, http.StatusBadRequest, "Invalid filter parameters: "+err.Error())
		return
	}

	// Get articles
	articles, total, err := h.newsService.ListArticles(ctx, filter)
	if err != nil {
		h.logger.Error("Failed to list articles", "error", err)
		dto.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve articles")
		return
	}

	// Convert to DTOs
	articleDTOs := make([]*dto.ArticleResponse, len(articles))
	for i, article := range articles {
		articleDTOs[i] = dto.ToArticleResponse(article)
	}

	response := dto.NewsListResponse{
		Articles: articleDTOs,
		Total:    total,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	}

	h.logger.Info("Successfully retrieved articles", 
		"count", len(articles), 
		"total", total,
		"company", filter.Company)

	dto.WriteJSONResponse(w, http.StatusOK, response)
}

// GetNewsById handles GET /news/{id} requests
func (h *NewsHandler) GetNewsById(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["id"]

	if id == "" {
		dto.WriteErrorResponse(w, http.StatusBadRequest, "Article ID is required")
		return
	}

	article, err := h.newsService.GetArticleByID(ctx, id)
	if err != nil {
		if err == news.ErrArticleNotFound {
			dto.WriteErrorResponse(w, http.StatusNotFound, "Article not found")
			return
		}
		h.logger.Error("Failed to get article by ID", "error", err, "id", id)
		dto.WriteErrorResponse(w, http.StatusInternalServerError, "Failed to retrieve article")
		return
	}

	response := dto.ToArticleResponse(article)
	dto.WriteJSONResponse(w, http.StatusOK, response)
}

// parseNewsFilter parses query parameters into a NewsFilter
func (h *NewsHandler) parseNewsFilter(r *http.Request) (*news.NewsFilter, error) {
	filter := &news.NewsFilter{}

	// Company filter
	if company := r.URL.Query().Get("company"); company != "" {
		filter.Company = company
	}

	// Date filters
	if startDateStr := r.URL.Query().Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			return nil, err
		}
		filter.StartDate = &startDate
	}

	if endDateStr := r.URL.Query().Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			return nil, err
		}
		// Set to end of day
		endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		filter.EndDate = &endDate
	}

	// Relevance filter
	if relevanceStr := r.URL.Query().Get("min_relevance"); relevanceStr != "" {
		relevance, err := strconv.ParseFloat(relevanceStr, 64)
		if err != nil {
			return nil, err
		}
		filter.MinRelevance = &relevance
	}

	// Pagination
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, err
		}
		filter.Limit = limit
	} else {
		filter.Limit = 50 // Default limit
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, err
		}
		filter.Offset = offset
	}

	return filter, nil
}