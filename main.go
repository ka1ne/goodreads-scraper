package main

import (
	"log"

	"goodreads-scraper/internal/api"
	"goodreads-scraper/internal/cache"
	"goodreads-scraper/internal/scraper"
	"goodreads-scraper/pkg/config"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting Goodreads Scraper on port %s", cfg.Port)
	log.Printf("Cache TTL: %s, Scrape timeout: %s", cfg.CacheTTL, cfg.ScrapeTimeout)

	// Initialize dependencies
	memCache := cache.NewMemoryCache(cfg.CacheTTL)
	goodreadsScraper := scraper.NewScraper(cfg.UserAgent, cfg.ScrapeTimeout)
	apiHandler := api.NewHandler(goodreadsScraper, memCache)

	// Setup routes
	router := apiHandler.SetupRoutes(cfg)

	// Start server
	log.Printf("Server starting on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
