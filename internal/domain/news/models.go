package news

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Article represents a news article entry
type Article struct {
	ID             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Title          string             `json:"title" bson:"title"`
	Summary        string             `json:"summary" bson:"summary"`
	SourceURL      string             `json:"source_url" bson:"source_url"`
	Companies      []string           `json:"companies" bson:"companies"`
	PublishedDate  time.Time          `json:"published_date" bson:"published_date"`
	RelevanceScore float64            `json:"relevance_score" bson:"relevance_score"`
	ProcessedDate  time.Time          `json:"processed_date" bson:"processed_date"`
	FeedSource     string             `json:"feed_source" bson:"feed_source"`
	Content        string             `json:"content,omitempty" bson:"content,omitempty"`
	GUID           string             `json:"guid" bson:"guid"`
}

// Validate validates the Article fields
func (a *Article) Validate() error {
	if a.Title == "" {
		return ErrInvalidTitle
	}
	if a.SourceURL == "" {
		return ErrInvalidSourceURL
	}
	if a.GUID == "" {
		return ErrInvalidGUID
	}
	if a.RelevanceScore < 0 || a.RelevanceScore > 1 {
		return ErrInvalidRelevanceScore
	}
	return nil
}

// NewsFilter represents filters for news queries
type NewsFilter struct {
	Company      string     `json:"company,omitempty"`
	StartDate    *time.Time `json:"start_date,omitempty"`
	EndDate      *time.Time `json:"end_date,omitempty"`
	MinRelevance *float64   `json:"min_relevance,omitempty"`
	Limit        int        `json:"limit,omitempty"`
	Offset       int        `json:"offset,omitempty"`
}

// RSSFeedItem represents a parsed RSS feed item
type RSSFeedItem struct {
	Title       string
	Summary     string
	Link        string
	PublishDate time.Time
	GUID        string
	Content     string
}