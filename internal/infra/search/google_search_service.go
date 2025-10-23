package search

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/pkg/logger"
)

// GoogleSearchService handles Google Custom Search API requests
type GoogleSearchService struct {
	apiKey     string
	searchEngineID string
	httpClient *http.Client
	logger     logger.Logger
}

// GoogleSearchResult represents a search result from Google Custom Search API
type GoogleSearchResult struct {
	Title   string `json:"title"`
	Link    string `json:"link"`
	Snippet string `json:"snippet"`
}

// GoogleSearchResponse represents the full response from Google Custom Search API
type GoogleSearchResponse struct {
	Items []GoogleSearchResult `json:"items"`
}

// NewGoogleSearchService creates a new Google Custom Search service
func NewGoogleSearchService(apiKey, searchEngineID string, logger logger.Logger) *GoogleSearchService {
	return &GoogleSearchService{
		apiKey:         apiKey,
		searchEngineID: searchEngineID,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger,
	}
}

// SearchDefenseAndAerospace performs a Google search focused on defense and aerospace topics
func (g *GoogleSearchService) SearchDefenseAndAerospace(ctx context.Context, query string) ([]GoogleSearchResult, error) {
	// Enhance query with defense/aerospace context
	enhancedQuery := g.enhanceQueryForDefense(query)
	
	g.logger.Info("Performing Google Custom Search", "original_query", query, "enhanced_query", enhancedQuery)

	// Build the search URL
	searchURL := g.buildSearchURL(enhancedQuery)

	// Create request with context
	req, err := http.NewRequestWithContext(ctx, "GET", searchURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create search request: %w", err)
	}

	// Perform the search
	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to perform search request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("search API returned status %d", resp.StatusCode)
	}

	// Parse the response
	var searchResponse GoogleSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResponse); err != nil {
		return nil, fmt.Errorf("failed to decode search response: %w", err)
	}

	// Filter and limit results
	filteredResults := g.filterDefenseResults(searchResponse.Items)
	
	g.logger.Info("Google search completed", 
		"total_results", len(searchResponse.Items),
		"filtered_results", len(filteredResults))

	return filteredResults, nil
}

// enhanceQueryForDefense adds defense/aerospace context to the search query
func (g *GoogleSearchService) enhanceQueryForDefense(query string) string {
	lowerQuery := strings.ToLower(query)
	
	// If query already contains defense/aerospace terms, don't modify
	defenseTerms := []string{
		"defense", "aerospace", "military", "rtx", "raytheon", "lockheed",
		"boeing", "northrop", "grumman", "war department", "pentagon",
	}
	
	for _, term := range defenseTerms {
		if strings.Contains(lowerQuery, term) {
			return query // Query already has defense context
		}
	}
	
	// Add defense context to improve relevance
	return fmt.Sprintf("%s defense aerospace industry", query)
}

// buildSearchURL constructs the Google Custom Search API URL
func (g *GoogleSearchService) buildSearchURL(query string) string {
	baseURL := "https://www.googleapis.com/customsearch/v1"
	params := url.Values{}
	params.Set("key", g.apiKey)
	params.Set("cx", g.searchEngineID)
	params.Set("q", query)
	params.Set("num", "10") // Maximum 10 results
	params.Set("dateRestrict", "y1") // Last year for recent information
	
	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}

// filterDefenseResults filters search results to focus on defense/aerospace content
func (g *GoogleSearchService) filterDefenseResults(results []GoogleSearchResult) []GoogleSearchResult {
	var filtered []GoogleSearchResult
	
	defenseKeywords := []string{
		"defense", "aerospace", "military", "rtx", "raytheon", "lockheed",
		"boeing", "northrop", "grumman", "war department", "pentagon",
		"contract", "missile", "aircraft", "satellite", "cybersecurity",
	}
	
	for _, result := range results {
		if g.isDefenseRelated(result, defenseKeywords) {
			filtered = append(filtered, result)
		}
		
		// Limit to top 5 most relevant results
		if len(filtered) >= 5 {
			break
		}
	}
	
	return filtered
}

// isDefenseRelated checks if a search result is related to defense/aerospace
func (g *GoogleSearchService) isDefenseRelated(result GoogleSearchResult, keywords []string) bool {
	content := strings.ToLower(result.Title + " " + result.Snippet)
	
	for _, keyword := range keywords {
		if strings.Contains(content, keyword) {
			return true
		}
	}
	
	return false
}