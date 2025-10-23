package ai

import (
	"context"
	"errors"
)

var (
	ErrInvalidQuery = errors.New("invalid query")
	ErrNoResults    = errors.New("no relevant articles found")
	ErrAIService    = errors.New("AI service error")
)

// Service defines the AI service interface
type Service interface {
	// ProcessQuery processes a natural language query and returns an AI-generated response
	ProcessQuery(ctx context.Context, req *QueryRequest) (*QueryResponse, error)
	
	// AnalyzeQuery analyzes the query to extract intent, companies, and search terms
	AnalyzeQuery(ctx context.Context, question string) (*QueryAnalysisResult, error)
	
	// SearchWeb searches the internet for defense/aeronautics information when DB context is insufficient
	SearchWeb(ctx context.Context, query string, companies []string) ([]WebSearchSource, error)
	
	// SummarizeArticle generates a concise summary of an article using AI
	SummarizeArticle(ctx context.Context, articleID string) (*ArticleSummaryResponse, error)
	
	// GetCacheStats returns statistics about the summary cache
	GetCacheStats() map[string]interface{}
}

// Repository defines the interface for AI-related data operations
type Repository interface {
	// SearchArticles performs semantic search on articles based on the analyzed query
	SearchArticles(ctx context.Context, analysis *QueryAnalysisResult, limit int) ([]*SourceReference, error)
}