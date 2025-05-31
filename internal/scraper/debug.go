package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// DebugHTML inspects and logs the HTML structure for debugging
func (s *Scraper) DebugHTML(username string) error {
	userID, err := s.getUserID(username)
	if err != nil {
		return fmt.Errorf("failed to get user ID: %w", err)
	}

	profileURL := fmt.Sprintf("https://www.goodreads.com/user/show/%s", userID)
	fmt.Printf("Fetching: %s\n", profileURL)

	resp, err := s.client.R().Get(profileURL)
	if err != nil {
		return fmt.Errorf("failed to fetch profile: %w", err)
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode())
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(resp.Body())))
	if err != nil {
		return fmt.Errorf("failed to parse HTML: %w", err)
	}

	fmt.Printf("=== HTML STRUCTURE DEBUG ===\n")
	fmt.Printf("Page title: %s\n", doc.Find("title").Text())

	// Look for any text containing numbers
	fmt.Printf("\n=== POTENTIAL STATS ELEMENTS ===\n")
	doc.Find("*").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if len(text) > 0 && len(text) < 100 {
			// Look for ratings, reviews, or avg patterns
			lowerText := strings.ToLower(text)
			if strings.Contains(lowerText, "rating") ||
				strings.Contains(lowerText, "review") ||
				strings.Contains(lowerText, "avg") ||
				strings.Contains(lowerText, "book") {
				fmt.Printf("Element: %s | Text: %s\n", goquery.NodeName(sel), text)
				if class, exists := sel.Attr("class"); exists {
					fmt.Printf("  Class: %s\n", class)
				}
				if id, exists := sel.Attr("id"); exists {
					fmt.Printf("  ID: %s\n", id)
				}
			}
		}
	})

	// Look for links that might contain stats
	fmt.Printf("\n=== LINKS WITH POTENTIAL STATS ===\n")
	doc.Find("a").Each(func(i int, sel *goquery.Selection) {
		href, exists := sel.Attr("href")
		if exists && strings.Contains(href, "review") {
			text := strings.TrimSpace(sel.Text())
			fmt.Printf("Link: %s | Text: %s\n", href, text)
		}
	})

	return nil
}
