package ai

import (
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/news"
)

// QueryRequest represents a user question about companies/news
type QueryRequest struct {
	Question string `json:"question" validate:"required,min=1,max=1000"`
	CompanyContext []string `json:"company_context,omitempty"` // Optional: focus on specific companies
}

// QueryResponse represents the AI-generated response
type QueryResponse struct {
	Answer string `json:"answer"`
	Sources []SourceReference `json:"sources"`
	WebSources []WebSearchSource `json:"web_sources,omitempty"`
	UsedWebSearch bool `json:"used_web_search"`
	Confidence float64 `json:"confidence"`
	ProcessingTime time.Duration `json:"processing_time"`
	CompaniesReferenced []string `json:"companies_referenced"`
}

// SourceReference represents a news article used as context
type SourceReference struct {
	ArticleID string `json:"article_id"`
	Title string `json:"title"`
	Summary string `json:"summary"`
	CompanyName string `json:"company_name"`
	PublishedDate time.Time `json:"published_date"`
	SourceURL string `json:"source_url"`
	RelevanceScore float64 `json:"relevance_score"`
}

// WebSearchSource represents a web search result used as context
type WebSearchSource struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Snippet     string    `json:"snippet"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Relevance   float64   `json:"relevance"`
}

// QueryContext represents processed context for the AI query
type QueryContext struct {
	Question string
	RelevantArticles []*news.Article
	CompanyNames []string
	QueryType QueryType
	TimeWindow *TimeWindow
}

// QueryType categorizes the type of question
type QueryType string

const (
	QueryTypeFinancial   QueryType = "financial"
	QueryTypeContracts   QueryType = "contracts"
	QueryTypeGeneral     QueryType = "general"
	QueryTypeComparison  QueryType = "comparison"
	QueryTypeNews        QueryType = "news"
)

// TimeWindow represents a time-based filter for queries
type TimeWindow struct {
	StartDate *time.Time
	EndDate   *time.Time
	Period    string // "this quarter", "this year", "recent", etc.
}

// QueryAnalysisResult represents the analysis of a user query
type QueryAnalysisResult struct {
	QueryType QueryType
	CompanyNames []string
	Keywords []string
	TimeWindow *TimeWindow
	SearchTerms []string
}