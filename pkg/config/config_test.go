package config

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoad_DefaultValues(t *testing.T) {
	// Clear any existing env vars
	clearTestEnvVars()

	config := Load()

	assert.Equal(t, "8080", config.Port)
	assert.Equal(t, 6*time.Hour, config.CacheTTL)
	assert.Equal(t, 30*time.Second, config.ScrapeTimeout)
	assert.Equal(t, "info", config.LogLevel)
	assert.Equal(t, 60, config.RateLimitPerMinute)
	assert.Equal(t, 10, config.ScrapeRateLimit)
	assert.Equal(t, "127.0.0.1,::1", config.TrustedProxies)
	assert.Contains(t, config.UserAgent, "Mozilla")
}

func TestLoad_EnvironmentVariables(t *testing.T) {
	// Clear any existing env vars
	clearTestEnvVars()

	// Set test environment variables
	os.Setenv("PORT", "9090")
	os.Setenv("CACHE_TTL", "2h")
	os.Setenv("SCRAPE_TIMEOUT", "45s")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("RATE_LIMIT_PER_MINUTE", "120")
	os.Setenv("SCRAPE_RATE_LIMIT", "20")
	os.Setenv("TRUSTED_PROXIES", "10.0.0.0/8,172.16.0.0/12")
	os.Setenv("USER_AGENT", "TestBot/1.0")

	defer clearTestEnvVars()

	config := Load()

	assert.Equal(t, "9090", config.Port)
	assert.Equal(t, 2*time.Hour, config.CacheTTL)
	assert.Equal(t, 45*time.Second, config.ScrapeTimeout)
	assert.Equal(t, "debug", config.LogLevel)
	assert.Equal(t, 120, config.RateLimitPerMinute)
	assert.Equal(t, 20, config.ScrapeRateLimit)
	assert.Equal(t, "10.0.0.0/8,172.16.0.0/12", config.TrustedProxies)
	assert.Equal(t, "TestBot/1.0", config.UserAgent)
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		envValue     string
		expected     string
	}{
		{
			"returns default when env not set",
			"TEST_KEY", "default", "", "default",
		},
		{
			"returns env value when set",
			"TEST_KEY", "default", "env_value", "env_value",
		},
		{
			"returns env value when empty string",
			"TEST_KEY", "default", "", "default",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetDurationEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue time.Duration
		envValue     string
		expected     time.Duration
	}{
		{
			"returns default when env not set",
			"TEST_DURATION", 5 * time.Minute, "", 5 * time.Minute,
		},
		{
			"returns parsed duration when valid",
			"TEST_DURATION", 5 * time.Minute, "10m", 10 * time.Minute,
		},
		{
			"returns default when invalid duration",
			"TEST_DURATION", 5 * time.Minute, "invalid", 5 * time.Minute,
		},
		{
			"handles complex durations",
			"TEST_DURATION", 5 * time.Minute, "1h30m", 90 * time.Minute,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getDurationEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIntEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		envValue     string
		expected     int
	}{
		{
			"returns default when env not set",
			"TEST_INT", 42, "", 42,
		},
		{
			"returns parsed int when valid",
			"TEST_INT", 42, "100", 100,
		},
		{
			"returns default when invalid int",
			"TEST_INT", 42, "invalid", 42,
		},
		{
			"handles zero value",
			"TEST_INT", 42, "0", 0,
		},
		{
			"handles negative value",
			"TEST_INT", 42, "-10", -10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			result := getIntEnv(tt.key, tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func clearTestEnvVars() {
	envVars := []string{
		"PORT", "CACHE_TTL", "SCRAPE_TIMEOUT", "LOG_LEVEL",
		"RATE_LIMIT_PER_MINUTE", "SCRAPE_RATE_LIMIT",
		"TRUSTED_PROXIES", "USER_AGENT",
	}

	for _, env := range envVars {
		os.Unsetenv(env)
	}
}
