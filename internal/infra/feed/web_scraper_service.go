package feed

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/Neph-dev/october_backend/internal/domain/news"
	"github.com/PuerkitoBio/goquery"
)

// WebScraperService handles scraping news articles from HTML pages
type WebScraperService struct {
	client *http.Client
	logger *slog.Logger
}

// NewWebScraperService creates a new web scraper service
func NewWebScraperService(logger *slog.Logger) *WebScraperService {
	return &WebScraperService{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// ScrapePage scrapes news articles from an HTML news page
func (s *WebScraperService) ScrapePage(ctx context.Context, pageURL, companyName string) ([]*news.RSSFeedItem, error) {
	s.logger.Info("Scraping news page", "url", pageURL, "company", companyName)

	// First, try to find RSS feed links on the page
	if rssURL := s.findRSSFeedURL(ctx, pageURL); rssURL != "" {
		s.logger.Info("Found RSS feed URL, trying RSS first", "rss_url", rssURL, "company", companyName)
		// TODO: We could try RSS parsing here if we had access to RSS service
	}

	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent to avoid being blocked
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OctoberBot/1.0; +https://october.ai)")

	// Make the request
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP error: %d %s", resp.StatusCode, resp.Status)
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Determine scraping strategy based on company
	return s.scrapeByCompany(doc, pageURL, companyName)
}

// findRSSFeedURL attempts to find RSS feed URLs on the page (simplified for now)
func (s *WebScraperService) findRSSFeedURL(ctx context.Context, pageURL string) string {
	// TODO: Implement RSS feed discovery
	// For now, return empty to always use HTML scraping
	return ""
}

// scrapeByCompany applies company-specific scraping rules
func (s *WebScraperService) scrapeByCompany(doc *goquery.Document, pageURL, companyName string) ([]*news.RSSFeedItem, error) {
	switch strings.ToLower(companyName) {
	case "northrop grumman":
		return s.scrapeNorthropGrumman(doc, pageURL)
	case "boeing":
		return s.scrapeBoeing(doc, pageURL)
	case "general dynamics":
		return s.scrapeGeneralDynamics(doc, pageURL)
	case "l3harris technologies":
		return s.scrapeL3Harris(doc, pageURL)
	case "huntington ingalls industries":
		return s.scrapeHuntingtonIngalls(doc, pageURL)
	case "textron inc":
		return s.scrapeTextron(doc, pageURL)
	default:
		// Generic scraping approach
		return s.scrapeGeneric(doc, pageURL)
	}
}

// scrapeNorthropGrumman scrapes https://news.northropgrumman.com/News-Stream
func (s *WebScraperService) scrapeNorthropGrumman(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem

	// Look for news articles - typical patterns for news sites
	selectors := []string{
		"article",
		".news-item",
		".press-release",
		".news-article",
		"[class*='news']",
		"[class*='article']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "Northrop Grumman")
			if item != nil {
				items = append(items, item)
			}
		})
		
		// If we found items with this selector, use them
		if len(items) > 0 {
			break
		}
	}

	s.logger.Info("Scraped Northrop Grumman news", "items", len(items))
	return items, nil
}

// scrapeBoeing scrapes https://boeing.mediaroom.com
func (s *WebScraperService) scrapeBoeing(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem

	// Boeing mediaroom specific selectors - try broader patterns
	selectors := []string{
		"div[class*='news']",
		"div[class*='release']",
		"div[class*='press']",
		"li[class*='news']",
		"li[class*='release']", 
		".news-release",
		".press-release",
		"article",
		".newsroom-item",
		"[class*='release']",
		"a[href*='/news/']",
		"a[href*='/press/']",
		"a[href*='release']",
		"h2 a", "h3 a", "h4 a", // Headlines with links
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "Boeing")
			if item != nil {
				items = append(items, item)
			}
		})
		
		if len(items) > 0 {
			s.logger.Info("Boeing: found items with selector", "selector", selector, "count", len(items))
			break
		}
	}

	s.logger.Info("Scraped Boeing news", "items", len(items))
	return items, nil
}

// scrapeGeneralDynamics scrapes https://www.gd.com/news
func (s *WebScraperService) scrapeGeneralDynamics(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem

	selectors := []string{
		".news-item",
		"article",
		".press-release",
		"[class*='news']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "General Dynamics")
			if item != nil {
				items = append(items, item)
			}
		})
		
		if len(items) > 0 {
			break
		}
	}

	s.logger.Info("Scraped General Dynamics news", "items", len(items))
	return items, nil
}

// scrapeL3Harris scrapes https://www.l3harris.com/newsroom/search (search results page)
func (s *WebScraperService) scrapeL3Harris(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem
	
	// L3Harris newsroom page selectors - target their main newsroom content
	selectors := []string{
		".node--storytelling-page .paragraph--type--featured-news .featured-content-item",
		".featured-content-item", 
		".card-wrapper",
		".view-content .views-row",
		"article",
		".news-item",
		".content-item",
	}
	
	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "L3Harris Technologies")
			if item != nil {
				items = append(items, item)
			}
		})
		
		if len(items) > 0 {
			break // Found items with this selector
		}
	}
	
	s.logger.Info("Scraped L3Harris news", "items", len(items))
	return items, nil
}

// scrapeHuntingtonIngalls scrapes https://hii.com/newsroom
func (s *WebScraperService) scrapeHuntingtonIngalls(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem

	selectors := []string{
		".newsroom-item",
		".news-item",
		"article",
		"[class*='news']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "Huntington Ingalls Industries")
			if item != nil {
				items = append(items, item)
			}
		})
		
		if len(items) > 0 {
			break
		}
	}

	s.logger.Info("Scraped Huntington Ingalls news", "items", len(items))
	return items, nil
}

// scrapeTextron scrapes https://investor.textron.com/news-releases/default.aspx
func (s *WebScraperService) scrapeTextron(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem

	selectors := []string{
		".news-release",
		".press-release",
		"article",
		"[class*='release']",
		"[class*='news']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "Textron Inc")
			if item != nil {
				items = append(items, item)
			}
		})
		
		if len(items) > 0 {
			break
		}
	}

	s.logger.Info("Scraped Textron news", "items", len(items))
	return items, nil
}

// scrapeGeneric provides a fallback scraping approach
func (s *WebScraperService) scrapeGeneric(doc *goquery.Document, baseURL string) ([]*news.RSSFeedItem, error) {
	var items []*news.RSSFeedItem

	// Generic selectors for common news page structures
	selectors := []string{
		"article",
		".news-item",
		".news-article",
		".press-release",
		".news-release",
		"[class*='news']",
		"[class*='article']",
		"[class*='press']",
		"[class*='release']",
	}

	for _, selector := range selectors {
		doc.Find(selector).Each(func(i int, article *goquery.Selection) {
			item := s.extractArticleInfo(article, baseURL, "Unknown")
			if item != nil {
				items = append(items, item)
			}
		})
		
		if len(items) > 0 {
			break
		}
	}

	s.logger.Info("Scraped generic news", "items", len(items))
	return items, nil
}

// extractArticleInfo extracts article information from a DOM element
func (s *WebScraperService) extractArticleInfo(article *goquery.Selection, baseURL, companyName string) *news.RSSFeedItem {
	// Extract title
	title := s.extractTitle(article)
	if title == "" {
		return nil
	}

	// Extract link
	link := s.extractLink(article, baseURL)
	if link == "" {
		link = baseURL // fallback to base URL
	}

	// Extract summary/description
	summary := s.extractSummary(article)

	// Extract date
	publishDate := s.extractDate(article)

	// Generate GUID from link or title
	guid := link
	if guid == "" {
		guid = fmt.Sprintf("%s-%s", companyName, title)
	}

	return &news.RSSFeedItem{
		Title:       title,
		Summary:     summary,
		Link:        link,
		PublishDate: publishDate,
		GUID:        guid,
		Content:     summary, // Use summary as content for now
	}
}

// extractTitle finds the article title from various selectors
func (s *WebScraperService) extractTitle(article *goquery.Selection) string {
	selectors := []string{
		"h1", "h2", "h3", "h4", ".title", ".headline", 
		"[class*='title']", "[class*='headline']",
		"a[href]", // Sometimes the link text is the title
		"strong", "b", // Bold text might be titles
	}

	for _, selector := range selectors {
		if title := strings.TrimSpace(article.Find(selector).First().Text()); title != "" && len(title) > 10 {
			return title
		}
	}

	// If we found an element but no good title, try the whole text content
	if text := strings.TrimSpace(article.Text()); text != "" && len(text) > 10 && len(text) < 200 {
		return text
	}

	return ""
}

// extractLink finds the article link
func (s *WebScraperService) extractLink(article *goquery.Selection, baseURL string) string {
	selectors := []string{
		"a[href]",
		"[data-url]",
		"[data-link]",
	}

	for _, selector := range selectors {
		if href, exists := article.Find(selector).First().Attr("href"); exists && href != "" {
			return s.resolveURL(href, baseURL)
		}
		if dataURL, exists := article.Find(selector).First().Attr("data-url"); exists && dataURL != "" {
			return s.resolveURL(dataURL, baseURL)
		}
	}

	return ""
}

// extractSummary finds the article summary/description
func (s *WebScraperService) extractSummary(article *goquery.Selection) string {
	selectors := []string{
		".summary", ".description", ".excerpt", "p",
		"[class*='summary']", "[class*='description']", "[class*='excerpt']",
	}

	for _, selector := range selectors {
		if summary := strings.TrimSpace(article.Find(selector).First().Text()); summary != "" {
			// Limit summary length
			if len(summary) > 500 {
				summary = summary[:500] + "..."
			}
			return summary
		}
	}

	return ""
}

// extractDate attempts to find and parse the publication date
func (s *WebScraperService) extractDate(article *goquery.Selection) time.Time {
	selectors := []string{
		"time", ".date", ".published", "[datetime]",
		"[class*='date']", "[class*='time']", "[class*='published']",
	}

	for _, selector := range selectors {
		element := article.Find(selector).First()
		
		// Try datetime attribute first
		if datetime, exists := element.Attr("datetime"); exists {
			if date, err := time.Parse(time.RFC3339, datetime); err == nil {
				return date
			}
		}
		
		// Try parsing text content
		if dateText := strings.TrimSpace(element.Text()); dateText != "" {
			if date := s.parseDate(dateText); !date.IsZero() {
				return date
			}
		}
	}

	// Default to current time if no date found
	return time.Now()
}

// parseDate attempts to parse various date formats
func (s *WebScraperService) parseDate(dateStr string) time.Time {
	formats := []string{
		"January 2, 2006",
		"Jan 2, 2006",
		"2006-01-02",
		"01/02/2006",
		"2006/01/02",
		"2 January 2006",
		"2 Jan 2006",
		time.RFC3339,
		time.RFC822,
	}

	// Clean up the date string
	dateStr = strings.TrimSpace(dateStr)
	dateStr = regexp.MustCompile(`\s+`).ReplaceAllString(dateStr, " ")

	for _, format := range formats {
		if date, err := time.Parse(format, dateStr); err == nil {
			return date
		}
	}

	return time.Time{}
}

// resolveURL converts relative URLs to absolute URLs
func (s *WebScraperService) resolveURL(href, baseURL string) string {
	if href == "" {
		return ""
	}

	// If already absolute URL, return as is
	if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
		return href
	}

	// Parse base URL
	base, err := url.Parse(baseURL)
	if err != nil {
		return href
	}

	// Resolve relative URL
	relative, err := url.Parse(href)
	if err != nil {
		return href
	}

	return base.ResolveReference(relative).String()
}