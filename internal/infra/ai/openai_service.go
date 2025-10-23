package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/ai"
	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

type OpenAIService struct {
	client      *openai.Client
	newsService *news.Service
	model       string
	logger      logger.Logger
}

// NewOpenAIService creates a new OpenAI service instance
func NewOpenAIService(client *openai.Client, newsService *news.Service, logger logger.Logger) *OpenAIService {
	return &OpenAIService{
		client:      client,
		newsService: newsService,
		model:       "gpt-4o-mini", // Default model
		logger:      logger,
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

	// Step 3: If insufficient database context, use OpenAI's knowledge directly
	if len(sources) < 3 || s.hasLowConfidenceContext(sources) {
		s.logger.Info("Insufficient database context, using OpenAI's direct knowledge", "db_sources", len(sources))
		
		// Check if the question is about defense/aeronautics companies or topics
		if s.isDefenseAeronauticsQuestion(req.Question, analysis.CompanyNames) {
			directResponse, err := s.generateDirectResponse(ctx, req.Question, analysis)
			if err != nil {
				s.logger.Error("Failed to generate direct response", "error", err)
				return nil, fmt.Errorf("%w: failed to generate direct response", ai.ErrAIService)
			}

			// Return only the answer when using OpenAI's direct knowledge
			result := &ai.QueryResponse{
				Answer:              directResponse,
				Sources:             []ai.SourceReference{},
				WebSources:          []ai.WebSearchSource{},
				UsedWebSearch:       false,
				Confidence:          0.7, // Medium confidence for direct OpenAI responses
				ProcessingTime:      time.Since(startTime),
				CompaniesReferenced: analysis.CompanyNames,
			}

			s.logger.Info("Direct OpenAI response provided", 
				"processing_time", result.ProcessingTime,
				"confidence", result.Confidence)

			return result, nil
		} else {
			// Not a defense/aeronautics question, return no results
			return &ai.QueryResponse{
				Answer:              "I can only provide information about defense and aeronautics companies and topics. Please ask about RTX, US War Department, or related defense/aerospace subjects.",
				Sources:             []ai.SourceReference{},
				WebSources:          []ai.WebSearchSource{},
				UsedWebSearch:       false,
				Confidence:          0.0,
				ProcessingTime:      time.Since(startTime),
				CompaniesReferenced: analysis.CompanyNames,
			}, nil
		}
	}

	// Step 4: Generate AI response using retrieved context from database
	response, err := s.generateResponse(ctx, req.Question, sources, []ai.WebSearchSource{}, analysis)
	if err != nil {
		s.logger.Error("Failed to generate AI response", "error", err)
		return nil, fmt.Errorf("%w: failed to generate response", ai.ErrAIService)
	}

	// Step 5: Build final response
	result := &ai.QueryResponse{
		Answer:              response,
		Sources:             sources,
		WebSources:          []ai.WebSearchSource{},
		UsedWebSearch:       false,
		Confidence:          s.calculateConfidence(sources, []ai.WebSearchSource{}, analysis),
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

// generateResponse creates an AI response using the retrieved context (both DB and web)
func (s *OpenAIService) generateResponse(ctx context.Context, question string, sources []ai.SourceReference, webSources []ai.WebSearchSource, analysis *ai.QueryAnalysisResult) (string, error) {
	contextText := s.buildContextFromSources(sources, webSources)

	systemPrompt := `You are an expert analyst for defense industry news and information. 
Answer the user's question based ONLY on the provided context from recent news articles and web sources.

Guidelines:
- Be factual and cite specific information from the articles
- If the context doesn't contain enough information, say so
- Focus on the companies mentioned: RTX (Raytheon Technologies) and US War Department
- Provide specific details like dates, numbers, and contract values when available
- Keep responses concise but informative (2-3 paragraphs max)
- Do not make up information not present in the context
- If using web sources, mention that additional information was found from recent web searches
- Clearly distinguish between database sources and web sources when referencing information

Context from recent articles and web sources:
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

func (s *OpenAIService) buildContextFromSources(sources []ai.SourceReference, webSources []ai.WebSearchSource) string {
	var contextBuilder strings.Builder
	
	// Add database sources
	if len(sources) > 0 {
		contextBuilder.WriteString("\n=== DATABASE SOURCES ===\n")
		for i, source := range sources {
			contextBuilder.WriteString(fmt.Sprintf("\n--- Article %d ---\n", i+1))
			contextBuilder.WriteString(fmt.Sprintf("Company: %s\n", source.CompanyName))
			contextBuilder.WriteString(fmt.Sprintf("Title: %s\n", source.Title))
			contextBuilder.WriteString(fmt.Sprintf("Date: %s\n", source.PublishedDate.Format("2006-01-02")))
			contextBuilder.WriteString(fmt.Sprintf("Summary: %s\n", source.Summary))
			contextBuilder.WriteString(fmt.Sprintf("URL: %s\n", source.SourceURL))
		}
	}
	
	// Add web sources
	if len(webSources) > 0 {
		contextBuilder.WriteString("\n=== WEB SOURCES ===\n")
		for i, source := range webSources {
			contextBuilder.WriteString(fmt.Sprintf("\n--- Web Result %d ---\n", i+1))
			contextBuilder.WriteString(fmt.Sprintf("Source: %s\n", source.Source))
			contextBuilder.WriteString(fmt.Sprintf("Title: %s\n", source.Title))
			if !source.PublishedAt.IsZero() {
				contextBuilder.WriteString(fmt.Sprintf("Date: %s\n", source.PublishedAt.Format("2006-01-02")))
			}
			contextBuilder.WriteString(fmt.Sprintf("Content: %s\n", source.Snippet))
			contextBuilder.WriteString(fmt.Sprintf("URL: %s\n", source.URL))
		}
	}
	
	return contextBuilder.String()
}

func (s *OpenAIService) calculateConfidence(sources []ai.SourceReference, webSources []ai.WebSearchSource, analysis *ai.QueryAnalysisResult) float64 {
	if len(sources) == 0 && len(webSources) == 0 {
		return 0.0
	}

	// Base confidence on number and quality of sources
	confidence := 0.5 // Base confidence

	// Boost confidence based on number of database sources (higher weight)
	if len(sources) >= 5 {
		confidence += 0.3
	} else if len(sources) >= 3 {
		confidence += 0.2
	} else if len(sources) >= 1 {
		confidence += 0.1
	}

	// Add confidence from web sources (lower weight than DB sources)
	if len(webSources) >= 3 {
		confidence += 0.15
	} else if len(webSources) >= 1 {
		confidence += 0.1
	}

	// Boost confidence based on average relevance score of database sources
	if len(sources) > 0 {
		totalRelevance := 0.0
		for _, source := range sources {
			totalRelevance += source.RelevanceScore
		}
		avgRelevance := totalRelevance / float64(len(sources))
		confidence += avgRelevance * 0.2
	}

	// Add confidence from web source relevance (lower weight)
	if len(webSources) > 0 {
		totalRelevance := 0.0
		for _, source := range webSources {
			totalRelevance += source.Relevance
		}
		avgRelevance := totalRelevance / float64(len(webSources))
		confidence += avgRelevance * 0.1
	}

	// Ensure confidence is within bounds
	if confidence > 1.0 {
		confidence = 1.0
	}

	return confidence
}

// hasLowConfidenceContext checks if the database context has low confidence
func (s *OpenAIService) hasLowConfidenceContext(sources []ai.SourceReference) bool {
	if len(sources) == 0 {
		return true
	}

	// Calculate average relevance score
	totalRelevance := 0.0
	for _, source := range sources {
		totalRelevance += source.RelevanceScore
	}
	avgRelevance := totalRelevance / float64(len(sources))

	// Consider low confidence if average relevance is below 0.6
	return avgRelevance < 0.6
}

// SearchWeb implements the Service interface for web searching (now deprecated)
func (s *OpenAIService) SearchWeb(ctx context.Context, query string, companies []string) ([]ai.WebSearchSource, error) {
	// Web search functionality has been removed in favor of direct OpenAI responses
	return []ai.WebSearchSource{}, fmt.Errorf("web search functionality has been replaced with direct OpenAI responses")
}

// isDefenseAeronauticsQuestion checks if the question is about defense/aeronautics companies or topics
func (s *OpenAIService) isDefenseAeronauticsQuestion(question string, companyNames []string) bool {
	lowerQuestion := strings.ToLower(question)
	
	// Check for defense/aeronautics keywords
	defenseKeywords := []string{
		"defense", "defence", "military", "aerospace", "aeronautics", "aviation",
		"aircraft", "fighter", "missile", "radar", "satellite", "contract",
		"pentagon", "air force", "navy", "army", "marines", "weapons",
		"jet", "helicopter", "drone", "uav", "ceo", "executive", "leadership",
		"financial", "earnings", "revenue", "stock", "market", "performance",
		"founder", "history", "established", "founded", "company", "corporation",
	}

	// Check if question contains defense/aeronautics keywords
	for _, keyword := range defenseKeywords {
		if strings.Contains(lowerQuestion, keyword) {
			return true
		}
	}

	// Check if question mentions known defense companies
	defenseCompanies := []string{
		"rtx", "raytheon", "lockheed", "boeing", "northrop", "grumman",
		"war department", "defense department", "pentagon",
	}

	for _, company := range defenseCompanies {
		if strings.Contains(lowerQuestion, company) {
			return true
		}
	}

	// Check if any of the identified company names are defense-related
	for _, companyName := range companyNames {
		lowerCompany := strings.ToLower(companyName)
		if strings.Contains(lowerCompany, "raytheon") ||
		   strings.Contains(lowerCompany, "rtx") ||
		   strings.Contains(lowerCompany, "war department") ||
		   strings.Contains(lowerCompany, "lockheed") ||
		   strings.Contains(lowerCompany, "defense") {
			return true
		}
	}

	return false
}

// generateDirectResponse generates a response using OpenAI's knowledge without database context
func (s *OpenAIService) generateDirectResponse(ctx context.Context, question string, analysis *ai.QueryAnalysisResult) (string, error) {
	systemPrompt := `You are a concise defense and aerospace industry analyst. Answer questions directly and briefly.

Guidelines:
- Only answer questions about defense and aerospace companies or topics
- Focus on companies like RTX (Raytheon Technologies), Lockheed Martin, Boeing, Northrop Grumman, etc.
- Give short, direct answers (1-2 sentences maximum)
- Provide only the most essential information requested
- No lengthy explanations or background context
- If you don't have current information, briefly mention your knowledge may be outdated
- If the question is not about defense/aerospace, politely decline to answer

Important: Keep responses short, direct, and to the point. Only provide information about defense and aerospace companies and related topics. Give longer answers only if specifically requested. such as "explain in detail" or "provide a comprehensive overview".`

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
		Temperature: 0.3,
	})

	if err != nil {
		return "", err
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return resp.Choices[0].Message.Content, nil
}