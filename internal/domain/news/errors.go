package news

import "errors"

// Domain errors for news
var (
	ErrInvalidTitle          = errors.New("title cannot be empty")
	ErrInvalidSourceURL      = errors.New("source URL cannot be empty")
	ErrInvalidGUID           = errors.New("GUID cannot be empty")
	ErrInvalidRelevanceScore = errors.New("relevance score must be between 0 and 1")
	ErrArticleNotFound       = errors.New("article not found")
	ErrDuplicateArticle      = errors.New("article already exists")
	ErrInvalidFilter         = errors.New("invalid filter parameters")
)