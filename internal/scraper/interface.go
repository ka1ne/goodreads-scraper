package scraper

// Interface defines the contract for Goodreads scraping operations
type Interface interface {
	GetReadingStats(username string) (*ReadingStats, error)
	DebugHTML(username string) error
	DebugShelf(userID, shelf string) error
}
