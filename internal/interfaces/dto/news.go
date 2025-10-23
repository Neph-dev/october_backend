package dto

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/news"
)

// ArticleResponse represents the API response for an article
type ArticleResponse struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Summary        string    `json:"summary"`
	SourceURL      string    `json:"source_url"`
	Companies      []string  `json:"companies"`
	PublishedDate  time.Time `json:"published_date"`
	RelevanceScore float64   `json:"relevance_score"`
	ProcessedDate  time.Time `json:"processed_date"`
	FeedSource     string    `json:"feed_source"`
}

// NewsListResponse represents the API response for news list
type NewsListResponse struct {
	Articles []*ArticleResponse `json:"articles"`
	Total    int64              `json:"total"`
	Limit    int                `json:"limit"`
	Offset   int                `json:"offset"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ToArticleResponse converts a domain Article to ArticleResponse
func ToArticleResponse(article *news.Article) *ArticleResponse {
	return &ArticleResponse{
		ID:             article.ID.Hex(),
		Title:          article.Title,
		Summary:        article.Summary,
		SourceURL:      article.SourceURL,
		Companies:      article.Companies,
		PublishedDate:  article.PublishedDate,
		RelevanceScore: article.RelevanceScore,
		ProcessedDate:  article.ProcessedDate,
		FeedSource:     article.FeedSource,
	}
}

// WriteJSONResponse writes a JSON response
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	response := ErrorResponse{
		Error:   http.StatusText(statusCode),
		Message: message,
	}
	WriteJSONResponse(w, statusCode, response)
}