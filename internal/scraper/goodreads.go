package scraper

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/go-resty/resty/v2"
)

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Scraper handles Goodreads web scraping
type Scraper struct {
	client    *resty.Client
	userAgent string
	timeout   time.Duration
}

// NewScraper creates a new Goodreads scraper
func NewScraper(userAgent string, timeout time.Duration) *Scraper {
	client := resty.New().
		SetTimeout(timeout).
		SetRetryCount(3).
		SetRetryWaitTime(2*time.Second).
		SetRetryMaxWaitTime(10*time.Second).
		SetHeader("User-Agent", userAgent).
		SetHeader("Accept-Language", "en-US,en;q=0.9").
		SetHeader("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8").
		SetHeader("Accept-Encoding", "gzip, deflate, br").
		SetHeader("DNT", "1").
		SetHeader("Connection", "keep-alive").
		SetHeader("Upgrade-Insecure-Requests", "1")

	return &Scraper{
		client:    client,
		userAgent: userAgent,
		timeout:   timeout,
	}
}

// GetReadingStats scrapes reading statistics for a user
func (s *Scraper) GetReadingStats(username string) (*ReadingStats, error) {
	// Extract user ID from profile URL or use username directly
	userID, err := s.getUserID(username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user ID: %w", err)
	}

	// Build profile URL
	profileURL := fmt.Sprintf("https://www.goodreads.com/user/show/%s", userID)

	log.Printf("Scraping profile: %s", profileURL)

	// Fetch profile page
	resp, err := s.client.R().Get(profileURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch profile: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	// Parse HTML
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body())))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Extract basic stats
	stats := &ReadingStats{
		UserID:      userID,
		Username:    username,
		LastUpdated: time.Now(),
	}

	// Parse profile statistics
	if err := s.parseProfileStats(doc, stats); err != nil {
		log.Printf("Warning: failed to parse profile stats: %v", err)
	}

	// Get books from various shelves
	favorites, err := s.getShelfBooks(userID, "favorites")
	if err != nil {
		log.Printf("Warning: failed to get favorites: %v", err)
	} else {
		stats.Favorites = favorites
	}

	studyBooks, err := s.getShelfBooks(userID, "study")
	if err != nil {
		log.Printf("Warning: failed to get study books: %v", err)
	} else {
		stats.StudyBooks = studyBooks
	}

	// Also try to get some books from main shelves for debugging
	if len(stats.Favorites) == 0 && len(stats.StudyBooks) == 0 {
		log.Printf("No books in favorites/study, trying main shelves...")

		// Try 'read' shelf as favorites fallback
		readBooks, err := s.getShelfBooks(userID, "read")
		if err != nil {
			log.Printf("Warning: failed to get read books: %v", err)
		} else {
			log.Printf("Found %d books in 'read' shelf", len(readBooks))
			if len(readBooks) > 0 {
				stats.Favorites = readBooks[:min(10, len(readBooks))] // Take first 10 as sample
			}
		}
	}

	return stats, nil
}

// getUserID extracts user ID from username or profile URL
func (s *Scraper) getUserID(username string) (string, error) {
	// If already looks like a user ID, return as-is
	if strings.Contains(username, "-") {
		return username, nil
	}

	// For now, assume the username format is "101839711-kaine"
	// In a real implementation, you might need to search for the user
	return "101839711-kaine", nil
}

// getShelfBooks scrapes books from a specific shelf
func (s *Scraper) getShelfBooks(userID, shelf string) ([]Book, error) {
	shelfURL := fmt.Sprintf("https://www.goodreads.com/review/list/%s?shelf=%s", userID, shelf)

	log.Printf("Scraping shelf: %s", shelfURL)

	resp, err := s.client.R().Get(shelfURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch shelf: %w", err)
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body())))
	if err != nil {
		return nil, fmt.Errorf("failed to parse shelf HTML: %w", err)
	}

	return s.parseShelfBooks(doc), nil
}

// DebugShelf outputs HTML structure debug information for a shelf
func (s *Scraper) DebugShelf(userID, shelf string) error {
	shelfURL := fmt.Sprintf("https://www.goodreads.com/review/list/%s?shelf=%s", userID, shelf)

	fmt.Printf("Fetching shelf: %s\n", shelfURL)

	resp, err := s.client.R().Get(shelfURL)
	if err != nil {
		return fmt.Errorf("failed to fetch shelf: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body())))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	fmt.Printf("=== SHELF HTML STRUCTURE DEBUG ===\n")
	fmt.Printf("Page title: %s\n", doc.Find("title").Text())

	// Look for book-related elements
	fmt.Printf("\n=== POTENTIAL BOOK ELEMENTS ===\n")
	doc.Find("tr[id*='review_']").Each(func(i int, sel *goquery.Selection) {
		fmt.Printf("Book row %d:\n", i+1)

		// Title
		titleCell := sel.Find("td.field.title")
		if titleCell.Length() > 0 {
			titleLink := titleCell.Find("a")
			fmt.Printf("  Title: %s\n", strings.TrimSpace(titleLink.Text()))
		}

		// Author
		authorCell := sel.Find("td.field.author")
		if authorCell.Length() > 0 {
			fmt.Printf("  Author: %s\n", strings.TrimSpace(authorCell.Find("a").Text()))
		}
	})

	// Alternative book formats
	fmt.Printf("\n=== ALTERNATIVE BOOK FORMATS ===\n")
	doc.Find(".bookalike").Each(func(i int, sel *goquery.Selection) {
		title := sel.Find(".title a")
		if title.Length() > 0 {
			fmt.Printf("Book %d: %s\n", i+1, strings.TrimSpace(title.Text()))
		}
	})

	// Count total potential book elements
	reviewRows := doc.Find("tr[id*='review_']").Length()
	bookalikeElements := doc.Find(".bookalike").Length()
	fmt.Printf("\n=== SUMMARY ===\n")
	fmt.Printf("Review rows found: %d\n", reviewRows)
	fmt.Printf("Bookalike elements found: %d\n", bookalikeElements)

	return nil
}
