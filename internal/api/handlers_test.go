package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"goodreads-scraper/internal/cache"
	"goodreads-scraper/internal/scraper"
)

// MockScraper implements the scraper interface for testing
type MockScraper struct {
	mock.Mock
}

func (m *MockScraper) GetReadingStats(username string) (*scraper.ReadingStats, error) {
	args := m.Called(username)
	return args.Get(0).(*scraper.ReadingStats), args.Error(1)
}

func (m *MockScraper) DebugHTML(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockScraper) DebugShelf(userID, shelf string) error {
	args := m.Called(userID, shelf)
	return args.Error(0)
}

func setupTestRouter(mockScraper *MockScraper) *gin.Engine {
	gin.SetMode(gin.TestMode)

	memCache := cache.NewMemoryCache(1 * time.Hour)

	// Create handler with mock scraper
	handler := NewHandler(mockScraper, memCache)

	r := gin.New()

	// Add CORS
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")
		c.Next()
	})

	// Health check
	r.GET("/health", handler.healthCheck)

	// API routes
	v1 := r.Group("/api/v1")
	v1.GET("/portfolio/:username", handler.getPortfolioData)
	v1.GET("/reading-stats/:username", handler.getReadingStats)
	v1.GET("/reading-stats/:username/favorites", handler.getFavorites)
	v1.GET("/reading-stats/:username/study", handler.getStudyBooks)

	return r
}

func TestHealthHandler(t *testing.T) {
	mockScraper := &MockScraper{}
	router := setupTestRouter(mockScraper)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
	assert.NotEmpty(t, response["timestamp"])
}

func TestPortfolioHandler_Success(t *testing.T) {
	mockScraper := &MockScraper{}
	router := setupTestRouter(mockScraper)

	// Mock scraper response
	stats := &scraper.ReadingStats{
		Username:      "testuser",
		TotalRatings:  61,
		TotalReviews:  9,
		AverageRating: 4.18,
		Favorites: []scraper.Book{
			{
				Title:        "Test Book 1",
				Author:       "Test Author 1",
				CoverURL:     "https://example.com/cover1.jpg",
				GoodreadsURL: "https://goodreads.com/book/1",
			},
		},
		StudyBooks: []scraper.Book{
			{
				Title:        "Study Book 1",
				Author:       "Study Author 1",
				CoverURL:     "https://example.com/study1.jpg",
				GoodreadsURL: "https://goodreads.com/book/3",
			},
		},
		LastUpdated: time.Now(),
	}

	mockScraper.On("GetReadingStats", "testuser").Return(stats, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/portfolio/testuser", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "testuser", response["username"])

	// Check stats
	statsMap := response["stats"].(map[string]interface{})
	assert.Equal(t, float64(61), statsMap["total_ratings"])
	assert.Equal(t, float64(9), statsMap["total_reviews"])
	assert.Equal(t, 4.18, statsMap["average_rating"])

	// Check favorite books
	favBooks := response["favorite_books"].([]interface{})
	assert.Len(t, favBooks, 1)

	// Check book counts
	bookCount := response["book_count"].(map[string]interface{})
	assert.Equal(t, float64(1), bookCount["favorites"])
	assert.Equal(t, float64(1), bookCount["study"])

	// Check timestamp
	assert.NotEmpty(t, response["last_updated"])

	mockScraper.AssertExpectations(t)
}

func TestReadingStatsHandler_Success(t *testing.T) {
	mockScraper := &MockScraper{}
	router := setupTestRouter(mockScraper)

	stats := &scraper.ReadingStats{
		Username:      "testuser",
		TotalRatings:  61,
		TotalReviews:  9,
		AverageRating: 4.18,
		Favorites: []scraper.Book{
			{
				Title:        "Test Book",
				Author:       "Test Author",
				CoverURL:     "https://example.com/cover.jpg",
				GoodreadsURL: "https://goodreads.com/book/1",
			},
		},
		LastUpdated: time.Now(),
	}

	mockScraper.On("GetReadingStats", "testuser").Return(stats, nil)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/reading-stats/testuser", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "testuser", response["username"])
	assert.Equal(t, float64(61), response["total_ratings"])
	assert.Equal(t, float64(9), response["total_reviews"])
	assert.Equal(t, 4.18, response["average_rating"])

	mockScraper.AssertExpectations(t)
}

func TestCaching(t *testing.T) {
	mockScraper := &MockScraper{}
	router := setupTestRouter(mockScraper)

	stats := &scraper.ReadingStats{
		Username:      "testuser",
		TotalRatings:  61,
		TotalReviews:  9,
		AverageRating: 4.18,
		LastUpdated:   time.Now(),
	}

	// Mock should only be called once due to caching
	mockScraper.On("GetReadingStats", "testuser").Return(stats, nil).Once()

	// First request
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/api/v1/reading-stats/testuser", nil)
	router.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request should hit cache
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/v1/reading-stats/testuser", nil)
	router.ServeHTTP(w2, req2)
	assert.Equal(t, 200, w2.Code)

	// Both responses should be identical
	assert.Equal(t, w1.Body.String(), w2.Body.String())

	mockScraper.AssertExpectations(t)
}

func TestCORSHeaders(t *testing.T) {
	mockScraper := &MockScraper{}
	router := setupTestRouter(mockScraper)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	// Should have CORS headers
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
}
