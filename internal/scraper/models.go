package scraper

import "time"

// ReadingStats represents the complete reading statistics for a user
type ReadingStats struct {
	UserID           string    `json:"user_id"`
	Username         string    `json:"username"`
	TotalBooks       int       `json:"total_books"`
	BooksThisYear    int       `json:"books_this_year"`
	CurrentlyReading int       `json:"currently_reading"`
	AverageRating    float64   `json:"average_rating"`
	TotalRatings     int       `json:"total_ratings"`
	TotalReviews     int       `json:"total_reviews"`
	LastUpdated      time.Time `json:"last_updated"`
	RecentReads      []Book    `json:"recent_reads"`
	Favorites        []Book    `json:"favorites"`
	StudyBooks       []Book    `json:"study_books"`
}

// Book represents a book with its metadata
type Book struct {
	Title        string `json:"title"`
	Author       string `json:"author"`
	Rating       int    `json:"rating,omitempty"`
	DateRead     string `json:"date_read,omitempty"`
	CoverURL     string `json:"cover_url,omitempty"`
	GoodreadsURL string `json:"goodreads_url,omitempty"`
}

// ErrorResponse represents API error responses
type ErrorResponse struct {
	Error       string    `json:"error"`
	Message     string    `json:"message"`
	CachedData  bool      `json:"cached_data,omitempty"`
	LastUpdated time.Time `json:"last_updated,omitempty"`
}
