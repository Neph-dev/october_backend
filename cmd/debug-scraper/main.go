package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

func main() {
	urls := []string{
		"https://boeing.mediaroom.com",
		"https://news.northropgrumman.com/News-Stream",
		"https://www.gd.com/news",
	}

	for _, url := range urls {
		fmt.Printf("\n=== Analyzing %s ===\n", url)
		analyzeWebsite(url)
	}
}

func analyzeWebsite(pageURL string) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), "GET", pageURL, nil)
	if err != nil {
		fmt.Printf("Failed to create request: %v\n", err)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; OctoberBot/1.0; +https://october.ai)")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Failed to fetch page: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("HTTP error: %d %s\n", resp.StatusCode, resp.Status)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		fmt.Printf("Failed to parse HTML: %v\n", err)
		return
	}

	fmt.Printf("Page title: %s\n", doc.Find("title").Text())

	// Look for common news article patterns
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
		"h1", "h2", "h3", // Headlines
		"a[href*='news']",
		"a[href*='press']",
	}

	for _, selector := range selectors {
		elements := doc.Find(selector)
		if elements.Length() > 0 {
			fmt.Printf("Found %d elements with selector '%s'\n", elements.Length(), selector)
			
			// Show first few elements
			elements.EachWithBreak(func(i int, s *goquery.Selection) bool {
				if i >= 3 { // Show only first 3
					return false
				}
				text := strings.TrimSpace(s.Text())
				if len(text) > 100 {
					text = text[:100] + "..."
				}
				if text != "" {
					fmt.Printf("  [%d]: %s\n", i+1, text)
				}
				return true
			})
		}
	}

	// Look for class names that might contain news items
	fmt.Println("Common class names:")
	classNames := make(map[string]int)
	doc.Find("*[class]").Each(func(i int, s *goquery.Selection) {
		if class, exists := s.Attr("class"); exists {
			classes := strings.Fields(class)
			for _, className := range classes {
				if strings.Contains(strings.ToLower(className), "news") ||
				   strings.Contains(strings.ToLower(className), "article") ||
				   strings.Contains(strings.ToLower(className), "press") ||
				   strings.Contains(strings.ToLower(className), "release") {
					classNames[className]++
				}
			}
		}
	})

	for className, count := range classNames {
		if count > 1 { // Only show classes that appear multiple times
			fmt.Printf("  .%s (%d occurrences)\n", className, count)
		}
	}
}