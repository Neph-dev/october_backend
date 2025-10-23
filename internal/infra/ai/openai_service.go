package ai

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/company"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client         *openai.Client
	newsService    *news.Service
	companyService company.Service
	logger         *slog.Logger
	model          string
}

func NewOpenAIService(
	apiKey string,
	newsService *news.Service,
	companyService company.Service,
	logger *slog.Logger,
) *OpenAIService {
	return &OpenAIService{
		client:         openai.NewClient(apiKey),
		newsService:    newsService,
		companyService: companyService,
		logger:         logger,
		model:          openai.GPT4oMini,
	}
}

func (s *OpenAIService) ProcessQuery(ctx context.Context, req *ai.QueryRequest) (*ai.QueryResponse, error) {
	startTime := time.Now()
	
	if strings.TrimSpace(req.Question) == "" {
		return nil, fmt.Errorf("%w: question cannot be empty", ai.ErrInvalidQuery)
	}

	s.logger.Info("Processing AI query", "question", req.Question)

	// Step 1: Analyze the query to understand intent
	analysis, err := s.AnalyzeQuery(ctx, req.Question)
	if err != nil {
		s.logger.Error("Failed to analyze query", "error", err)
		return nil, fmt.Errorf("%w: failed to analyze query", ai.ErrAIService)
	}

	// Step 2: Retrieve relevant articles
	sources, err := s.retrieveRelevantArticles(ctx, analysis, req.CompanyContext)
	if err != nil {
		s.logger.Error("Failed to retrieve articles", "error", err)
		return nil, fmt.Errorf("%w: failed to retrieve articles", ai.ErrAIService)
	}

	if len(sources) == 0 {
		return &ai.QueryResponse{
			Answer:              "I couldn't find any relevant information to answer your question. Please try asking about recent news or developments for RTX or US War Department.",
			Sources:             []ai.SourceReference{},
			Confidence:          0.0,
			ProcessingTime:      time.Since(startTime),
			CompaniesReferenced: analysis.CompanyNames,
		}, nil
	}

	// Step 3: Generate AI response using retrieved context
	response, err := s.generateResponse(ctx, req.Question, sources, analysis)
	if err != nil {
		s.logger.Error("Failed to generate AI response", "error", err)
		return nil, fmt.Errorf("%w: failed to generate response", ai.ErrAIService)
	}

	// Step 4: Build final response
	result := &ai.QueryResponse{
		Answer:              response,
		Sources:             sources,
		Confidence:          s.calculateConfidence(sources, analysis),
		ProcessingTime:      time.Since(startTime),
		CompaniesReferenced: analysis.CompanyNames,
	}

	s.logger.Info("AI query processed successfully", 
		"processing_time", result.ProcessingTime,
		"sources_used", len(sources),
		"confidence", result.Confidence)

	return result, nil
}

// AnalyzeQuery analyzes the user's question to extract intent and entities
func (s *OpenAIService) AnalyzeQuery(ctx context.Context, question string) (*ai.QueryAnalysisResult, error) {
	systemPrompt := `You are a query analyzer for a defense industry news system. 
Analyze the user's question and extract:
1. Query type (financial, contracts, general, comparison, news)
2. Company names mentioned (RTX, Raytheon Technologies, US War Department, Lockheed Martin, etc.)
3. Key search terms and keywords
4. Time window if mentioned (this quarter, recent, this year, etc.)

Respond in this exact JSON format:
{
	"query_type": "financial|contracts|general|comparison|news",
	"company_names": ["Company1", "Company2"],
	"keywords": ["keyword1", "keyword2"],
	"time_window": "recent|this_quarter|this_year|null",
	"search_terms": ["term1", "term2"]
}`

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		MaxTokens:   200,
		Temperature: 0.1, // Low temperature for consistent analysis
	})

	if err != nil {
		return nil, err
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	// Parse the JSON response (simplified - in production, use proper JSON parsing)
	analysis := s.parseAnalysisResponse(resp.Choices[0].Message.Content, question)
	
	return analysis, nil
}

// retrieveRelevantArticles finds articles relevant to the query
func (s *OpenAIService) retrieveRelevantArticles(ctx context.Context, analysis *ai.QueryAnalysisResult, companyContext []string) ([]ai.SourceReference, error) {
	filter := &news.NewsFilter{
		Limit:  20,
		Offset: 0,
	}

	// Add company filter
	companies := analysis.CompanyNames
	if len(companyContext) > 0 {
		companies = append(companies, companyContext...)
	}

	// Add time window if specified
	if analysis.TimeWindow != nil {
		filter.StartDate = analysis.TimeWindow.StartDate
		filter.EndDate = analysis.TimeWindow.EndDate
	} else {
		// Default to recent articles (last 90 days)
		recent := time.Now().AddDate(0, 0, -90)
		filter.StartDate = &recent
	}

	var allArticles []*news.Article

	// If specific companies mentioned, get articles for each
	if len(companies) > 0 {
		for _, companyName := range companies {
			filter.Company = companyName
			articles, _, err := s.newsService.ListArticles(ctx, filter)
			if err != nil {
				s.logger.Warn("Failed to get articles for company", "company", companyName, "error", err)
				continue
			}
			allArticles = append(allArticles, articles...)
		}
	} else {
		// Get general articles
		articles, _, err := s.newsService.ListArticles(ctx, filter)
		if err != nil {
			return nil, err
		}
		allArticles = articles
	}

	// Convert to source references and rank by relevance
	sources := s.rankArticlesByRelevance(allArticles, analysis)
	
	// Return top 10 most relevant
	if len(sources) > 10 {
		sources = sources[:10]
	}

	return sources, nil
}

// generateResponse creates an AI response using the retrieved context
func (s *OpenAIService) generateResponse(ctx context.Context, question string, sources []ai.SourceReference, analysis *ai.QueryAnalysisResult) (string, error) {
	contextText := s.buildContextFromSources(sources)

	systemPrompt := `You are an expert analyst for defense industry news and information. 
Answer the user's question based ONLY on the provided context from recent news articles.

Guidelines:
- Be factual and cite specific information from the articles
- If the context doesn't contain enough information, say so
- Focus on the companies mentioned: RTX (Raytheon Technologies) and US War Department
- Provide specific details like dates, numbers, and contract values when available
- Keep responses concise but informative (2-3 paragraphs max)
- Do not make up information not present in the context

Context from recent articles:
` + contextText

	resp, err := s.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: s.model,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: question,
			},
		},
		MaxTokens:   500,
		Temperature: 0.3, // Slightly higher for more natural responses
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}

// Helper methods

func (s *OpenAIService) parseAnalysisResponse(content, originalQuestion string) *ai.QueryAnalysisResult {
	// Simplified parsing - in production, use proper JSON unmarshaling
	analysis := &ai.QueryAnalysisResult{
		QueryType:    ai.QueryTypeGeneral, // Default
		CompanyNames: []string{},
		Keywords:     []string{},
		SearchTerms:  []string{},
	}

	// Basic parsing logic - extract company names
	lowerContent := strings.ToLower(originalQuestion)
	if strings.Contains(lowerContent, "rtx") || strings.Contains(lowerContent, "raytheon") {
		analysis.CompanyNames = append(analysis.CompanyNames, "Raytheon Technologies")
	}
	if strings.Contains(lowerContent, "war department") || strings.Contains(lowerContent, "defense") || strings.Contains(lowerContent, "military") {
		analysis.CompanyNames = append(analysis.CompanyNames, "US War Department")
	}

	// Determine query type
	if strings.Contains(lowerContent, "quarter") || strings.Contains(lowerContent, "earnings") || strings.Contains(lowerContent, "revenue") || strings.Contains(lowerContent, "financial") {
		analysis.QueryType = ai.QueryTypeFinancial
	} else if strings.Contains(lowerContent, "contract") || strings.Contains(lowerContent, "award") || strings.Contains(lowerContent, "deal") {
		analysis.QueryType = ai.QueryTypeContracts
	}

	// Set time window for recent queries
	if strings.Contains(lowerContent, "recent") || strings.Contains(lowerContent, "latest") || strings.Contains(lowerContent, "this quarter") {
		analysis.TimeWindow = &ai.TimeWindow{
			Period: "recent",
		}
	}

	return analysis
}

func (s *OpenAIService) rankArticlesByRelevance(articles []*news.Article, analysis *ai.QueryAnalysisResult) []ai.SourceReference {
	sources := make([]ai.SourceReference, 0, len(articles))

	for _, article := range articles {
		// Calculate relevance score based on query analysis
		relevanceScore := article.RelevanceScore

		// Boost score if article matches query type
		if analysis.QueryType == ai.QueryTypeFinancial && 
		   (strings.Contains(strings.ToLower(article.Title), "earnings") || 
			strings.Contains(strings.ToLower(article.Title), "quarter") ||
			strings.Contains(strings.ToLower(article.Title), "revenue")) {
			relevanceScore += 0.2
		}

		if analysis.QueryType == ai.QueryTypeContracts && 
		   (strings.Contains(strings.ToLower(article.Title), "contract") || 
			strings.Contains(strings.ToLower(article.Title), "award")) {
			relevanceScore += 0.2
		}

		// Determine primary company for this article
		companyName := "Unknown"
		if len(article.Companies) > 0 {
			companyName = article.Companies[0]
		}

		source := ai.SourceReference{
			ArticleID:      article.ID.Hex(),
			Title:          article.Title,
			Summary:        article.Summary,
			CompanyName:    companyName,
			PublishedDate:  article.PublishedDate,
			SourceURL:      article.SourceURL,
			RelevanceScore: relevanceScore,
		}

		sources = append(sources, source)
	}

	// Sort by relevance score (descending)
	for i := 0; i < len(sources)-1; i++ {
		for j := i + 1; j < len(sources); j++ {
			if sources[i].RelevanceScore < sources[j].RelevanceScore {
				sources[i], sources[j] = sources[j], sources[i]
			}
		}
	}

	return sources
}

func (s *OpenAIService) buildContextFromSources(sources []ai.SourceReference) string {
	var contextBuilder strings.Builder
	
	for i, source := range sources {
		contextBuilder.WriteString(fmt.Sprintf("\n--- Article %d ---\n", i+1))
		contextBuilder.WriteString(fmt.Sprintf("Company: %s\n", source.CompanyName))
		contextBuilder.WriteString(fmt.Sprintf("Title: %s\n", source.Title))
		contextBuilder.WriteString(fmt.Sprintf("Date: %s\n", source.PublishedDate.Format("2006-01-02")))
		contextBuilder.WriteString(fmt.Sprintf("Summary: %s\n", source.Summary))
		contextBuilder.WriteString(fmt.Sprintf("URL: %s\n", source.SourceURL))
	}
	
	return contextBuilder.String()
}

func (s *OpenAIService) calculateConfidence(sources []ai.SourceReference, analysis *ai.QueryAnalysisResult) float64 {
	if len(sources) == 0 {
		return 0.0
	}

	// Base confidence on number and quality of sources
	confidence := 0.5 // Base confidence

	// Boost confidence based on number of sources
	if len(sources) >= 5 {
		confidence += 0.2
	} else if len(sources) >= 3 {
		confidence += 0.1
	}

	// Boost confidence based on average relevance score
	totalRelevance := 0.0
	for _, source := range sources {
		totalRelevance += source.RelevanceScore
	}
	avgRelevance := totalRelevance / float64(len(sources))
	confidence += avgRelevance * 0.3

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}