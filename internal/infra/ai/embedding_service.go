package ai

import (
	"context"
	"fmt"

	"github.com/Neph-dev/october_backend/pkg/logger"
	"github.com/sashabaranov/go-openai"
)

// EmbeddingService handles text to vector embedding conversion
type EmbeddingService struct {
	client *openai.Client
	model  openai.EmbeddingModel
	logger logger.Logger
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(client *openai.Client, logger logger.Logger) *EmbeddingService {
	return &EmbeddingService{
		client: client,
		model:  openai.SmallEmbedding3, // text-embedding-3-small (1536 dimensions)
		logger: logger,
	}
}

// EmbedText converts text to a vector embedding
func (s *EmbeddingService) EmbedText(ctx context.Context, text string) ([]float32, error) {
	if text == "" {
		return nil, fmt.Errorf("cannot embed empty text")
	}

	s.logger.Debug("Generating embedding", "text_length", len(text))

	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: []string{text},
		Model: s.model,
	})

	if err != nil {
		s.logger.Error("Failed to generate embedding", "error", err)
		return nil, fmt.Errorf("failed to generate embedding: %w", err)
	}

	if len(resp.Data) == 0 {
		return nil, fmt.Errorf("no embedding returned from API")
	}

	s.logger.Debug("Embedding generated successfully", "dimensions", len(resp.Data[0].Embedding))

	return resp.Data[0].Embedding, nil
}

// EmbedArticle creates an embedding optimized for article content
// Combines title and summary with appropriate weighting
func (s *EmbeddingService) EmbedArticle(ctx context.Context, title, summary, content string) ([]float32, error) {
	// Build composite text with weighted components
	// Title is most important, then summary, then selected content
	var compositeText string
	
	// Weight title more heavily (repeat 2x)
	compositeText = title + ". " + title + ". "
	
	if summary != "" {
		compositeText += summary + ". "
	}
	
	// Include beginning of content if available (first 500 chars)
	if content != "" {
		contentSnippet := content
		if len(content) > 500 {
			contentSnippet = content[:500]
		}
		compositeText += contentSnippet
	}

	s.logger.Debug("Creating article embedding", 
		"title", title,
		"has_summary", summary != "",
		"has_content", content != "",
		"composite_length", len(compositeText))

	return s.EmbedText(ctx, compositeText)
}

// EmbedQuery creates an embedding optimized for search queries
func (s *EmbeddingService) EmbedQuery(ctx context.Context, query string) ([]float32, error) {
	// For queries, we can enhance with context
	enhancedQuery := query + " defense aerospace industry news"
	
	s.logger.Debug("Creating query embedding", "query", query)
	
	return s.EmbedText(ctx, enhancedQuery)
}

// EmbedBatch creates embeddings for multiple texts in a single API call
func (s *EmbeddingService) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("cannot embed empty batch")
	}

	// OpenAI has a limit on batch size, chunk if necessary
	const maxBatchSize = 100
	if len(texts) > maxBatchSize {
		return s.embedLargeBatch(ctx, texts, maxBatchSize)
	}

	s.logger.Debug("Generating batch embeddings", "count", len(texts))

	resp, err := s.client.CreateEmbeddings(ctx, openai.EmbeddingRequest{
		Input: texts,
		Model: s.model,
	})

	if err != nil {
		s.logger.Error("Failed to generate batch embeddings", "error", err)
		return nil, fmt.Errorf("failed to generate batch embeddings: %w", err)
	}

	if len(resp.Data) != len(texts) {
		return nil, fmt.Errorf("expected %d embeddings, got %d", len(texts), len(resp.Data))
	}

	embeddings := make([][]float32, len(resp.Data))
	for i, data := range resp.Data {
		embeddings[i] = data.Embedding
	}

	s.logger.Debug("Batch embeddings generated successfully", "count", len(embeddings))

	return embeddings, nil
}

// embedLargeBatch handles batches larger than the API limit
func (s *EmbeddingService) embedLargeBatch(ctx context.Context, texts []string, chunkSize int) ([][]float32, error) {
	var allEmbeddings [][]float32

	for i := 0; i < len(texts); i += chunkSize {
		end := i + chunkSize
		if end > len(texts) {
			end = len(texts)
		}

		chunk := texts[i:end]
		embeddings, err := s.EmbedBatch(ctx, chunk)
		if err != nil {
			return nil, fmt.Errorf("failed to embed chunk %d: %w", i/chunkSize, err)
		}

		allEmbeddings = append(allEmbeddings, embeddings...)
	}

	return allEmbeddings, nil
}

// GetEmbeddingDimensions returns the dimensionality of embeddings from this service
func (s *EmbeddingService) GetEmbeddingDimensions() int {
	// text-embedding-3-small produces 1536-dimensional vectors
	return 1536
}
