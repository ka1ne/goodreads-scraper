package scraper

import (
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// parseProfileStats extracts reading statistics from the profile page
func (s *Scraper) parseProfileStats(doc *goquery.Document, stats *ReadingStats) error {
	// Look for ratings count
	doc.Find("a[href*='/review/list/']").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if strings.Contains(text, "rating") {
			count := extractNumber(text)
			if count > 0 {
				stats.TotalRatings = count
			}
		}
		if strings.Contains(text, "review") {
			count := extractNumber(text)
			if count > 0 {
				stats.TotalReviews = count
			}
		}
	})

	// Look for average rating
	doc.Find(".userStats").Each(func(i int, sel *goquery.Selection) {
		text := strings.TrimSpace(sel.Text())
		if strings.Contains(text, "avg rating") {
			rating := extractRating(text)
			if rating > 0 {
				stats.AverageRating = rating
			}
		}
	})

	// Alternative selectors based on dev.to article patterns
	if stats.TotalRatings == 0 {
		// Try alternative selectors
		doc.Find("*").Each(func(i int, sel *goquery.Selection) {
			text := strings.TrimSpace(sel.Text())
			if strings.Contains(text, "ratings") && strings.Contains(text, "reviews") {
				stats.TotalRatings, stats.TotalReviews = parseRatingsReviews(text)
				// Also try to extract average rating from the same element
				if stats.AverageRating == 0 && strings.Contains(text, "avg rating") {
					stats.AverageRating = extractRating(text)
				}
			}
		})
	}

	log.Printf("Parsed stats - Ratings: %d, Reviews: %d, Avg: %.2f",
		stats.TotalRatings, stats.TotalReviews, stats.AverageRating)

	return nil
}

// parseShelfBooks extracts books from a shelf page
func (s *Scraper) parseShelfBooks(doc *goquery.Document) []Book {
	var books []Book

	// Look for book entries in various possible formats
	doc.Find("tr[id*='review_']").Each(func(i int, sel *goquery.Selection) {
		book := Book{}

		// Extract title and author
		titleCell := sel.Find("td.field.title")
		if titleCell.Length() > 0 {
			titleLink := titleCell.Find("a")
			// Clean up title - remove extra whitespace and newlines
			title := strings.TrimSpace(titleLink.Text())
			title = strings.ReplaceAll(title, "\n", " ")
			title = strings.Join(strings.Fields(title), " ") // Normalize whitespace
			book.Title = title

			if href, exists := titleLink.Attr("href"); exists {
				book.GoodreadsURL = "https://www.goodreads.com" + href
			}
		}

		authorCell := sel.Find("td.field.author")
		if authorCell.Length() > 0 {
			book.Author = strings.TrimSpace(authorCell.Find("a").Text())
		}

		// Extract rating
		ratingCell := sel.Find("td.field.rating")
		if ratingCell.Length() > 0 {
			ratingText := strings.TrimSpace(ratingCell.Text())
			if rating := extractNumber(ratingText); rating > 0 {
				book.Rating = rating
			}
		}

		// Extract cover URL
		coverImg := sel.Find("img")
		if coverImg.Length() > 0 {
			if src, exists := coverImg.Attr("src"); exists {
				// Upgrade cover image size for better portfolio display
				// Replace small sizes with medium/large for better quality
				src = strings.ReplaceAll(src, "_SX50_", "_SX150_")
				src = strings.ReplaceAll(src, "_SY75_", "_SY225_")
				src = strings.ReplaceAll(src, "_SX50.", "_SX150.")
				book.CoverURL = src
			}
		}

		// Only add if we have at least title
		if book.Title != "" {
			books = append(books, book)
		}
	})

	// Fallback: try alternative selectors
	if len(books) == 0 {
		doc.Find(".bookalike").Each(func(i int, sel *goquery.Selection) {
			book := Book{}

			title := sel.Find(".title a")
			if title.Length() > 0 {
				book.Title = strings.TrimSpace(title.Text())
				if href, exists := title.Attr("href"); exists {
					book.GoodreadsURL = "https://www.goodreads.com" + href
				}
			}

			author := sel.Find(".author a")
			if author.Length() > 0 {
				book.Author = strings.TrimSpace(author.Text())
			}

			if book.Title != "" {
				books = append(books, book)
			}
		})
	}

	log.Printf("Parsed %d books from shelf", len(books))
	return books
}

// extractNumber extracts the first number from a string
func extractNumber(text string) int {
	re := regexp.MustCompile(`\d+`)
	match := re.FindString(strings.ReplaceAll(text, ",", ""))
	if match == "" {
		return 0
	}

	num, err := strconv.Atoi(match)
	if err != nil {
		return 0
	}

	return num
}

// extractRating extracts a rating (decimal number) from text
func extractRating(text string) float64 {
	re := regexp.MustCompile(`\d+\.\d+`)
	match := re.FindString(text)
	if match == "" {
		return 0
	}

	rating, err := strconv.ParseFloat(match, 64)
	if err != nil {
		return 0
	}

	return rating
}

// parseRatingsReviews extracts ratings and reviews count from combined text
func parseRatingsReviews(text string) (int, int) {
	// Look for pattern like "61 ratings | 9 reviews" or "61 ratings and 9 reviews"
	re := regexp.MustCompile(`(\d+)\s+ratings?\s*[|&and]*\s*(\d+)\s+reviews?`)
	matches := re.FindStringSubmatch(text)

	if len(matches) >= 3 {
		ratings := extractNumber(matches[1])
		reviews := extractNumber(matches[2])
		return ratings, reviews
	}

	return 0, 0
}
