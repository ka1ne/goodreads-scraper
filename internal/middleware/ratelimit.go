package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// IPRateLimiter holds rate limiters for each IP
type IPRateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	limit    rate.Limit
	burst    int
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(rps int, burst int) *IPRateLimiter {
	return &IPRateLimiter{
		limiters: make(map[string]*rate.Limiter),
		limit:    rate.Limit(rps),
		burst:    burst,
	}
}

// GetLimiter returns the rate limiter for the given IP
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	limiter, exists := i.limiters[ip]
	if !exists {
		limiter = rate.NewLimiter(i.limit, i.burst)
		i.limiters[ip] = limiter
	}

	return limiter
}

// CleanupOldEntries removes inactive limiters (call periodically)
func (i *IPRateLimiter) CleanupOldEntries() {
	i.mu.Lock()
	defer i.mu.Unlock()

	for ip, limiter := range i.limiters {
		// Remove limiters that haven't been used recently
		if limiter.Tokens() == float64(i.burst) {
			delete(i.limiters, ip)
		}
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(rps int, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rps, burst)

	// Cleanup old entries every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.CleanupOldEntries()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		ipLimiter := limiter.GetLimiter(ip)

		if !ipLimiter.Allow() {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rps))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "rate_limit_exceeded",
				"message":     "Too many requests. Please try again later.",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := int(ipLimiter.Tokens())
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rps))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}

// ScrapeRateLimitMiddleware creates stricter rate limiting for scraping endpoints
func ScrapeRateLimitMiddleware(rps int, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(rps, burst)

	// Cleanup old entries every 5 minutes
	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			limiter.CleanupOldEntries()
		}
	}()

	return func(c *gin.Context) {
		ip := c.ClientIP()
		ipLimiter := limiter.GetLimiter(ip)

		if !ipLimiter.Allow() {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rps))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "60")

			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":       "scrape_rate_limit_exceeded",
				"message":     "Scraping rate limit exceeded. Please wait before making more requests.",
				"retry_after": 60,
			})
			c.Abort()
			return
		}

		// Add rate limit headers
		remaining := int(ipLimiter.Tokens())
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rps))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))

		c.Next()
	}
}
