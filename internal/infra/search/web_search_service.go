package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/company"
)

// WebSearchService handles internet searches for company-related topics
type WebSearchService struct {
	client         *http.Client
	logger         *slog.Logger
	companyService company.Service
	searchEngines  []SearchEngine
}

// SearchEngine interface for different search providers
type SearchEngine interface {
	Search(ctx context.Context, query string, companies []string) ([]WebSearchResult, error)
	GetName() string
}

// WebSearchResult represents a web search result
type WebSearchResult struct {
	Title       string    `json:"title"`
	URL         string    `json:"url"`
	Snippet     string    `json:"snippet"`
	Source      string    `json:"source"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	Relevance   float64   `json:"relevance"`
}

// DuckDuckGoEngine implements search using DuckDuckGo's instant answer API
type DuckDuckGoEngine struct {
	client *http.Client
	logger *slog.Logger
}

// NewWebSearchService creates a new web search service
func NewWebSearchService(companyService company.Service, logger *slog.Logger) *WebSearchService {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	engines := []SearchEngine{
		&DuckDuckGoEngine{
			client: client,
			logger: logger,
		},
	}

	return &WebSearchService{
		client:         client,
		logger:         logger,
		companyService: companyService,
		searchEngines:  engines,
	}
}

// SearchDefenseAndAeronautics searches for company-related information
func (s *WebSearchService) SearchDefenseAndAeronautics(ctx context.Context, query string, companies []string) ([]ai.WebSearchSource, error) {
	// Check if the query is about companies that exist in our database
	if !s.isAboutDatabaseCompanies(ctx, query, companies) {
		return nil, fmt.Errorf("query is not about companies in our database")
	}

	// Enhance query with company context
	enhancedQuery := s.enhanceQueryForCompanies(query, companies)

	var allResults []WebSearchResult

	// Search using all available engines
	for _, engine := range s.searchEngines {
		results, err := engine.Search(ctx, enhancedQuery, companies)
		if err != nil {
			s.logger.Warn("Search engine failed", "engine", engine.GetName(), "error", err)
			continue
		}
		allResults = append(allResults, results...)
	}

	fmt.Println("Total search results found:", len(allResults))
	// Filter and rank results
	filteredResults := s.filterCompanyResults(allResults, companies)
	rankedResults := s.rankSearchResults(filteredResults, query, companies)

	// Convert to domain models
	sources := make([]ai.WebSearchSource, 0, len(rankedResults))
	for _, result := range rankedResults {
		sources = append(sources, ai.WebSearchSource{
			Title:       result.Title,
			URL:         result.URL,
			Snippet:     result.Snippet,
			Source:      result.Source,
			PublishedAt: result.PublishedAt,
			Relevance:   result.Relevance,
		})
	}

	// Limit to top 5 results
	if len(sources) > 5 {
		sources = sources[:5]
	}

	return sources, nil
}

// isAboutDatabaseCompanies checks if the query is about companies that exist in our database
func (s *WebSearchService) isAboutDatabaseCompanies(ctx context.Context, query string, companies []string) bool {
	lowerQuery := strings.ToLower(query)
	
	// First check explicitly provided companies
	for _, companyName := range companies {
		if s.isCompanyInDatabase(ctx, companyName) {
			return true
		}
	}
	
	// Extract potential company names from the query
	extractedCompanies := s.extractCompanyNamesFromQuery(lowerQuery)
	
	// Check if any extracted company exists in our database
	for _, companyName := range extractedCompanies {
		if s.isCompanyInDatabase(ctx, companyName) {
			return true
		}
	}
	
	return false
}

// isCompanyInDatabase checks if a company exists in our database
func (s *WebSearchService) isCompanyInDatabase(ctx context.Context, companyName string) bool {
	// Try to get the company from the database
	_, err := s.companyService.GetCompanyByName(ctx, companyName)
	return err == nil
}

// extractCompanyNamesFromQuery extracts potential company names from the query
func (s *WebSearchService) extractCompanyNamesFromQuery(lowerQuery string) []string {
	var companies []string
	
	// Known company variations to check for
	companyPatterns := map[string][]string{
		"Raytheon Technologies": {"rtx", "raytheon", "raytheon technologies"},
		"US War Department": {"war department", "us war department", "war dept", "department of war"},
		"Lockheed Martin": {"lockheed", "lockheed martin", "lmt"},
		// Add more as needed
	}
	
	for dbCompanyName, patterns := range companyPatterns {
		for _, pattern := range patterns {
			if strings.Contains(lowerQuery, pattern) {
				companies = append(companies, dbCompanyName)
				break // Only add once per company
			}
		}
	}
	
	return companies
}

// enhanceQueryForCompanies adds company context to the search query
func (s *WebSearchService) enhanceQueryForCompanies(query string, companies []string) string {
	enhanced := query
	
	// Add company context
	if len(companies) > 0 {
		enhanced = fmt.Sprintf("%s %s", enhanced, strings.Join(companies, " "))
	}
	
	// Add company-related terms to improve search results
	enhanced = fmt.Sprintf("%s company corporation business", enhanced)
	
	return enhanced
}

// filterCompanyResults filters results to company-related content
func (s *WebSearchService) filterCompanyResults(results []WebSearchResult, companies []string) []WebSearchResult {
	var filtered []WebSearchResult

	for _, result := range results {
		// For company queries, be more permissive - accept any result that mentions the companies
		if s.isCompanyRelated(result, companies) {
			filtered = append(filtered, result)
		}
	}

	return filtered
}

// isCompanyRelated checks if content is related to the specified companies
func (s *WebSearchService) isCompanyRelated(result WebSearchResult, companies []string) bool {
	content := strings.ToLower(result.Title + " " + result.Snippet)

	// Check for company mentions (case-insensitive)
	for _, company := range companies {
		if strings.Contains(content, strings.ToLower(company)) {
			return true
		}
	}

	// Check for known company variations
	companyVariations := []string{
		"rtx", "raytheon", "war department", "lockheed", "martin",
		"corporation", "company", "technologies", "defense", "aerospace",
	}

	for _, variation := range companyVariations {
		if strings.Contains(content, variation) {
			return true
		}
	}

	// If we can't find company mentions, accept it anyway to let OpenAI filter
	// This is more permissive approach as requested
	return true
}

// rankSearchResults ranks search results by relevance
func (s *WebSearchService) rankSearchResults(results []WebSearchResult, query string, companies []string) []WebSearchResult {
	for i := range results {
		results[i].Relevance = s.calculateRelevance(results[i], query, companies)
	}

	// Sort by relevance (bubble sort for simplicity)
	for i := 0; i < len(results)-1; i++ {
		for j := i + 1; j < len(results); j++ {
			if results[i].Relevance < results[j].Relevance {
				results[i], results[j] = results[j], results[i]
			}
		}
	}

	return results
}

// calculateRelevance calculates relevance score for a search result
func (s *WebSearchService) calculateRelevance(result WebSearchResult, query string, companies []string) float64 {
	score := 0.0
	content := strings.ToLower(result.Title + " " + result.Snippet)
	lowerQuery := strings.ToLower(query)

	// Title relevance (higher weight)
	if strings.Contains(strings.ToLower(result.Title), lowerQuery) {
		score += 0.4
	}

	// Snippet relevance
	if strings.Contains(strings.ToLower(result.Snippet), lowerQuery) {
		score += 0.2
	}

	// Company mention bonus
	for _, company := range companies {
		if strings.Contains(content, strings.ToLower(company)) {
			score += 0.3
		}
	}

	// Source credibility bonus
	if s.isCredibleSource(result.URL) {
		score += 0.2
	}

	// Recent content bonus (if we have date)
	if !result.PublishedAt.IsZero() && time.Since(result.PublishedAt) < 30*24*time.Hour {
		score += 0.1
	}

	return score
}

// isCredibleSource checks if the source is from a credible news/defense outlet
func (s *WebSearchService) isCredibleSource(url string) bool {
	credibleDomains := []string{
		"defensenews.com", "janes.com", "aviationweek.com", "flightglobal.com",
		"reuters.com", "bloomberg.com", "wsj.com", "ft.com", "cnn.com",
		"bbc.com", "npr.org", "politico.com", "thehill.com", "defense.gov",
		"navy.mil", "af.mil", "army.mil", "marines.mil",
	}

	lowerURL := strings.ToLower(url)
	for _, domain := range credibleDomains {
		if strings.Contains(lowerURL, domain) {
			return true
		}
	}

	return false
}

// DuckDuckGo Engine Implementation

func (d *DuckDuckGoEngine) GetName() string {
	return "DuckDuckGo"
}

func (d *DuckDuckGoEngine) Search(ctx context.Context, query string, companies []string) ([]WebSearchResult, error) {
	// Use DuckDuckGo's instant answer API (limited but no API key required)
	encodedQuery := url.QueryEscape(query)
	searchURL := fmt.Sprintf("https://api.duckduckgo.com/?q=%s&format=json&no_html=1&skip_disambig=1", encodedQuery)

	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "October-Backend/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var duckDuckGoResp DuckDuckGoResponse
	if err := json.Unmarshal(body, &duckDuckGoResp); err != nil {
		return nil, err
	}

	var results []WebSearchResult

	// Process related topics
	for _, topic := range duckDuckGoResp.RelatedTopics {
		if topic.Text != "" {
			result := WebSearchResult{
				Title:     topic.Text[:min(len(topic.Text), 100)],
				URL:       topic.FirstURL,
				Snippet:   topic.Text,
				Source:    "DuckDuckGo",
				Relevance: 0.5, // Base relevance
			}
			results = append(results, result)
		}
	}

	// Fallback: create synthetic results based on abstract
	if len(results) == 0 && duckDuckGoResp.Abstract != "" {
		result := WebSearchResult{
			Title:     duckDuckGoResp.Heading,
			URL:       duckDuckGoResp.AbstractURL,
			Snippet:   duckDuckGoResp.Abstract,
			Source:    duckDuckGoResp.AbstractSource,
			Relevance: 0.7,
		}
		results = append(results, result)
	}

	return results, nil
}

// DuckDuckGo API response structures
type DuckDuckGoResponse struct {
	Abstract       string                  `json:"Abstract"`
	AbstractSource string                  `json:"AbstractSource"`
	AbstractURL    string                  `json:"AbstractURL"`
	Heading        string                  `json:"Heading"`
	RelatedTopics  []DuckDuckGoRelatedTopic `json:"RelatedTopics"`
}

type DuckDuckGoRelatedTopic struct {
	Text     string `json:"Text"`
	FirstURL string `json:"FirstURL"`
}

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}