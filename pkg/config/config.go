package config

import (
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the application
type Config struct {
	Port          string        `env:"PORT"`
	CacheTTL      time.Duration `env:"CACHE_TTL"`
	ScrapeTimeout time.Duration `env:"SCRAPE_TIMEOUT"`
	UserAgent     string        `env:"USER_AGENT"`
	LogLevel      string        `env:"LOG_LEVEL"`

	// Rate limiting
	RateLimitPerMinute int `env:"RATE_LIMIT_PER_MINUTE"`
	ScrapeRateLimit    int `env:"SCRAPE_RATE_LIMIT"`

	// Security
	TrustedProxies string `env:"TRUSTED_PROXIES"`
}

// Load creates a new Config with values from environment variables or defaults
func Load() *Config {
	return &Config{
		Port:          getEnv("PORT", "8080"),
		CacheTTL:      getDurationEnv("CACHE_TTL", 6*time.Hour),
		ScrapeTimeout: getDurationEnv("SCRAPE_TIMEOUT", 30*time.Second),
		UserAgent:     getEnv("USER_AGENT", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),

		// Rate limiting defaults
		RateLimitPerMinute: getIntEnv("RATE_LIMIT_PER_MINUTE", 60), // 60 requests per minute general
		ScrapeRateLimit:    getIntEnv("SCRAPE_RATE_LIMIT", 10),     // 10 scrape requests per minute

		// Security defaults
		TrustedProxies: getEnv("TRUSTED_PROXIES", "127.0.0.1,::1"), // localhost only by default
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getDurationEnv gets a duration from environment variable or returns default
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// getIntEnv gets an integer from environment variable or returns default
func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
