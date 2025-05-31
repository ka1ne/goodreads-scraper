package scraper

import (
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/stretchr/testify/assert"
)

func TestExtractNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"simple number", "123", 123},
		{"number with text", "123 ratings", 123},
		{"number with comma", "1,234 reviews", 1234},
		{"multiple numbers", "123 ratings and 456 reviews", 123},
		{"no number", "no numbers here", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractNumber(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestExtractRating(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"simple rating", "4.18", 4.18},
		{"rating with text", "avg rating: 4.18", 4.18},
		{"rating with more text", "average rating is 3.50 stars", 3.50},
		{"no rating", "no rating here", 0},
		{"empty string", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractRating(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseRatingsReviews(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedRatings int
		expectedReviews int
	}{
		{
			"pipe separator",
			"61 ratings | 9 reviews",
			61, 9,
		},
		{
			"and separator",
			"61 ratings and 9 reviews",
			61, 9,
		},
		{
			"with extra text",
			"Total: 61 ratings | 9 reviews avg rating: 4.18",
			61, 9,
		},
		{
			"no match",
			"some other text",
			0, 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ratings, reviews := parseRatingsReviews(tt.input)
			assert.Equal(t, tt.expectedRatings, ratings)
			assert.Equal(t, tt.expectedReviews, reviews)
		})
	}
}

func TestParseShelfBooks(t *testing.T) {
	// Mock HTML for a book shelf
	htmlContent := `
	<html>
		<body>
			<table>
				<tr id="review_123">
					<td class="field title">
						<a href="/book/show/123">Test Book Title</a>
					</td>
					<td class="field author">
						<a href="/author/456">Test Author</a>
					</td>
					<td class="field rating">
						<span>4</span>
					</td>
					<td>
						<img src="https://example.com/cover_SX50_.jpg" />
					</td>
				</tr>
				<tr id="review_456">
					<td class="field title">
						<a href="/book/show/456">Another Book
						(Series #1)</a>
					</td>
					<td class="field author">
						<a href="/author/789">Another Author</a>
					</td>
				</tr>
			</table>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	assert.NoError(t, err)

	scraper := &Scraper{}
	books := scraper.parseShelfBooks(doc)

	assert.Len(t, books, 2)

	// Test first book
	assert.Equal(t, "Test Book Title", books[0].Title)
	assert.Equal(t, "Test Author", books[0].Author)
	assert.Equal(t, "https://www.goodreads.com/book/show/123", books[0].GoodreadsURL)
	assert.Equal(t, "https://example.com/cover_SX150_.jpg", books[0].CoverURL) // Should be upgraded
	assert.Equal(t, 4, books[0].Rating)

	// Test second book with normalized title
	assert.Equal(t, "Another Book (Series #1)", books[1].Title) // Should be normalized
	assert.Equal(t, "Another Author", books[1].Author)
}

func TestParseProfileStats(t *testing.T) {
	// Mock HTML for profile stats
	htmlContent := `
	<html>
		<body>
			<div class="smallText">
				61 ratings |
				9 reviews
				
				avg rating: 4.18
			</div>
		</body>
	</html>`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	assert.NoError(t, err)

	scraper := &Scraper{}
	stats := &ReadingStats{}

	err = scraper.parseProfileStats(doc, stats)
	assert.NoError(t, err)

	assert.Equal(t, 61, stats.TotalRatings)
	assert.Equal(t, 9, stats.TotalReviews)
	assert.Equal(t, 4.18, stats.AverageRating)
}
