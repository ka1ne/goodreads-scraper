package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRateLimitMiddleware_WithinLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware(10, 10)) // 10 requests per minute, burst 10
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Make requests within limit
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "10", w.Header().Get("X-RateLimit-Limit"))
		// Remaining should decrease
		assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	}
}

func TestRateLimitMiddleware_ExceedsLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware(2, 2)) // 2 requests per minute, burst 2
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Make requests up to limit
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		r.ServeHTTP(w, req)
		assert.Equal(t, 200, w.Code)
	}

	// Next request should be rate limited
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, 429, w.Code)
	assert.Equal(t, "2", w.Header().Get("X-RateLimit-Limit"))
	assert.Equal(t, "0", w.Header().Get("X-RateLimit-Remaining"))
	assert.Equal(t, "60", w.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_DifferentIPs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(RateLimitMiddleware(1, 1)) // 1 request per minute, burst 1
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// First IP - should work
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	req1.RemoteAddr = "192.168.1.1:12345"
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Same IP - should be rate limited
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	req2.RemoteAddr = "192.168.1.1:12346"
	r.ServeHTTP(w2, req2)
	assert.Equal(t, 429, w2.Code)

	// Different IP - should work
	w3 := httptest.NewRecorder()
	req3, _ := http.NewRequest("GET", "/test", nil)
	req3.RemoteAddr = "192.168.1.2:12347"
	r.ServeHTTP(w3, req3)
	assert.Equal(t, 200, w3.Code)
}

func TestScrapeRateLimitMiddleware_ErrorMessage(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(ScrapeRateLimitMiddleware(1, 1)) // 1 request per minute, burst 1
	r.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// First request - should work
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w1, req1)
	assert.Equal(t, 200, w1.Code)

	// Second request - should be rate limited with scrape-specific message
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w2, req2)

	assert.Equal(t, 429, w2.Code)
	assert.Contains(t, w2.Body.String(), "scrape_rate_limit_exceeded")
	assert.Contains(t, w2.Body.String(), "Scraping rate limit exceeded")
}

func TestIPRateLimiter_CleanupOldEntries(t *testing.T) {
	limiter := NewIPRateLimiter(10, 10)

	// Add some limiters and use them
	l1 := limiter.GetLimiter("192.168.1.1")
	l2 := limiter.GetLimiter("192.168.1.2")

	// Use the limiters so they're not at full tokens
	l1.Allow()
	l2.Allow()

	// Store initial count by checking if we can get the same limiters
	same1 := limiter.GetLimiter("192.168.1.1")
	same2 := limiter.GetLimiter("192.168.1.2")

	// Should be the same instances
	assert.Same(t, l1, same1)
	assert.Same(t, l2, same2)

	// Cleanup should not remove active limiters that were recently used
	limiter.CleanupOldEntries()

	// Should still get the same limiters
	after1 := limiter.GetLimiter("192.168.1.1")
	after2 := limiter.GetLimiter("192.168.1.2")
	assert.Same(t, l1, after1)
	assert.Same(t, l2, after2)
}

func TestIPRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewIPRateLimiter(100, 100)

	done := make(chan bool, 2)

	// Concurrent access from different goroutines
	go func() {
		for i := 0; i < 50; i++ {
			limiter.GetLimiter("192.168.1.1").Allow()
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 50; i++ {
			limiter.GetLimiter("192.168.1.2").Allow()
		}
		done <- true
	}()

	// Wait for completion - should not panic
	<-done
	<-done

	// Verify limiters exist by getting them again
	l1 := limiter.GetLimiter("192.168.1.1")
	l2 := limiter.GetLimiter("192.168.1.2")
	assert.NotNil(t, l1)
	assert.NotNil(t, l2)
}
