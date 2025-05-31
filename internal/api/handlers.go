package api

import (
	"net/http"
	"strings"
	"time"

	"goodreads-scraper/internal/cache"
	"goodreads-scraper/internal/middleware"
	"goodreads-scraper/internal/scraper"
	"goodreads-scraper/pkg/config"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies for API handlers
type Handler struct {
	scraper scraper.Interface
	cache   *cache.MemoryCache
}

// NewHandler creates a new API handler
func NewHandler(s scraper.Interface, c *cache.MemoryCache) *Handler {
	return &Handler{
		scraper: s,
		cache:   c,
	}
}

// SetupRoutes configures the API routes
func (h *Handler) SetupRoutes(cfg *config.Config) *gin.Engine {
	r := gin.Default()

	// Configure trusted proxies for security
	// Parse trusted proxies from config (comma-separated)
	trustedProxies := strings.Split(cfg.TrustedProxies, ",")
	for i, proxy := range trustedProxies {
		trustedProxies[i] = strings.TrimSpace(proxy)
	}
	r.SetTrustedProxies(trustedProxies)

	// Add CORS headers for frontend consumption
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	r.GET("/health", h.healthCheck)

	// Debug endpoints
	r.GET("/debug/:username", h.debugHTML)
	r.GET("/debug/:username/shelf/:shelf", h.debugShelf)

	// General rate limiting for all API endpoints
	v1 := r.Group("/api/v1", middleware.RateLimitMiddleware(cfg.RateLimitPerMinute, cfg.RateLimitPerMinute))

	// Apply stricter rate limiting to scraping endpoints
	scrapeGroup := v1.Group("/")
	scrapeGroup.Use(middleware.ScrapeRateLimitMiddleware(cfg.ScrapeRateLimit, cfg.ScrapeRateLimit))
	{
		scrapeGroup.GET("/reading-stats/:username", h.getReadingStats)
		scrapeGroup.GET("/reading-stats/:username/favorites", h.getFavorites)
		scrapeGroup.GET("/reading-stats/:username/study", h.getStudyBooks)
		scrapeGroup.GET("/portfolio/:username", h.getPortfolioData)
	}

	return r
}

// healthCheck returns service health status
func (h *Handler) healthCheck(c *gin.Context) {
	cacheStats := h.cache.Stats()

	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now(),
		"cache":     cacheStats,
	})
}

// getReadingStats returns complete reading statistics
func (h *Handler) getReadingStats(c *gin.Context) {
	username := c.Param("username")

	// Check cache first
	cacheKey := "stats:" + username
	if cached, found := h.cache.Get(cacheKey); found {
		if stats, ok := cached.(*scraper.ReadingStats); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, stats)
			return
		}
	}

	// Scrape fresh data
	stats, err := h.scraper.GetReadingStats(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, scraper.ErrorResponse{
			Error:   "scraping_failed",
			Message: "Failed to scrape reading statistics: " + err.Error(),
		})
		return
	}

	// Cache the result
	h.cache.Set(cacheKey, stats)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, stats)
}

// getFavorites returns only favorite books
func (h *Handler) getFavorites(c *gin.Context) {
	username := c.Param("username")

	// Try to get from cache first
	cacheKey := "favorites:" + username
	if cached, found := h.cache.Get(cacheKey); found {
		if books, ok := cached.([]scraper.Book); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{
				"username":  username,
				"favorites": books,
				"count":     len(books),
			})
			return
		}
	}

	// Get from full stats (this will use cache if available)
	stats, err := h.scraper.GetReadingStats(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, scraper.ErrorResponse{
			Error:   "scraping_failed",
			Message: "Failed to get favorites: " + err.Error(),
		})
		return
	}

	// Cache just the favorites
	h.cache.Set(cacheKey, stats.Favorites)
	c.Header("X-Cache", "MISS")

	c.JSON(http.StatusOK, gin.H{
		"username":  username,
		"favorites": stats.Favorites,
		"count":     len(stats.Favorites),
	})
}

// getStudyBooks returns only study shelf books
func (h *Handler) getStudyBooks(c *gin.Context) {
	username := c.Param("username")

	// Try to get from cache first
	cacheKey := "study:" + username
	if cached, found := h.cache.Get(cacheKey); found {
		if books, ok := cached.([]scraper.Book); ok {
			c.Header("X-Cache", "HIT")
			c.JSON(http.StatusOK, gin.H{
				"username":    username,
				"study_books": books,
				"count":       len(books),
			})
			return
		}
	}

	// Get from full stats (this will use cache if available)
	stats, err := h.scraper.GetReadingStats(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, scraper.ErrorResponse{
			Error:   "scraping_failed",
			Message: "Failed to get study books: " + err.Error(),
		})
		return
	}

	// Cache just the study books
	h.cache.Set(cacheKey, stats.StudyBooks)
	c.Header("X-Cache", "MISS")

	c.JSON(http.StatusOK, gin.H{
		"username":    username,
		"study_books": stats.StudyBooks,
		"count":       len(stats.StudyBooks),
	})
}

// debugHTML returns HTML structure debug information
func (h *Handler) debugHTML(c *gin.Context) {
	username := c.Param("username")

	err := h.scraper.DebugHTML(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "debug_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Debug output written to console logs",
		"username": username,
	})
}

// debugShelf returns HTML structure debug information for a shelf
func (h *Handler) debugShelf(c *gin.Context) {
	username := c.Param("username")
	shelf := c.Param("shelf")

	// Get user ID (this is hardcoded for now)
	userID := "101839711-kaine" // TODO: make this dynamic

	err := h.scraper.DebugShelf(userID, shelf)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "debug_failed",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "Shelf debug output written to console logs",
		"username": username,
		"shelf":    shelf,
	})
}

// getPortfolioData returns optimized data for portfolio websites
func (h *Handler) getPortfolioData(c *gin.Context) {
	username := c.Param("username")

	// Check cache first
	cacheKey := "portfolio:" + username
	if cached, found := h.cache.Get(cacheKey); found {
		c.Header("X-Cache", "HIT")
		c.JSON(http.StatusOK, cached)
		return
	}

	// Get full stats
	stats, err := h.scraper.GetReadingStats(username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "scraping_failed",
			"message": "Failed to get portfolio data: " + err.Error(),
		})
		return
	}

	// Create portfolio-optimized response
	portfolioData := gin.H{
		"username": username,
		"stats": gin.H{
			"total_ratings":  stats.TotalRatings,
			"total_reviews":  stats.TotalReviews,
			"average_rating": stats.AverageRating,
		},
		"favorite_books": stats.Favorites,
		"book_count": gin.H{
			"favorites": len(stats.Favorites),
			"study":     len(stats.StudyBooks),
		},
		"last_updated": stats.LastUpdated,
	}

	// Cache the result
	h.cache.Set(cacheKey, portfolioData)
	c.Header("X-Cache", "MISS")
	c.JSON(http.StatusOK, portfolioData)
}
