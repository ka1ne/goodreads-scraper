# Goodreads Scraper API

A Go-based REST API for scraping Goodreads user profiles and book data. Built for portfolio websites and data analysis.

## Features

- **Profile scraping** - User stats (ratings, reviews, average rating)
- **Shelf scraping** - Books from favorites, study, and other shelves  
- **Portfolio-optimized** - Clean JSON endpoints for frontend consumption
- **Rate limiting** - Tiered protection (60/min general, 10/min scraping)
- **Smart caching** - 6-hour TTL to avoid rate limiting
- **Enhanced images** - Higher resolution book covers (150px vs 50px)
- **CORS enabled** - Ready for browser requests
- **Security** - Configurable trusted proxies

## Quick Start

```bash
# Clone and start
git clone <repo-url>
cd goodreads-scraper
docker compose up -d

# Test the API
curl "http://localhost:8080/api/v1/portfolio/your-goodreads-username" | jq .
```

## API Endpoints

### Portfolio Data (Recommended)
```
GET /api/v1/portfolio/:username
```
Returns optimized data for portfolio websites with stats + favorite books.

### Individual Endpoints
```
GET /api/v1/reading-stats/:username           # Complete user data
GET /api/v1/reading-stats/:username/favorites # Favorite books only
GET /api/v1/reading-stats/:username/study     # Study shelf only
```

### Health & Debug
```
GET /health                                   # Service health
GET /debug/:username                         # HTML structure debug
GET /debug/:username/shelf/:shelf            # Shelf debug
```

## Response Format

```json
{
  "username": "example-user",
  "stats": {
    "total_ratings": 61,
    "total_reviews": 9,
    "average_rating": 4.18
  },
  "favorite_books": [
    {
      "title": "Book Title",
      "author": "Author Name",
      "cover_url": "https://...",
      "goodreads_url": "https://..."
    }
  ],
  "book_count": {
    "favorites": 4,
    "study": 2
  },
  "last_updated": "2025-06-01T19:35:08Z"
}
```

## Configuration

Environment variables:

```bash
# Basic
PORT=8080
CACHE_TTL=6h
SCRAPE_TIMEOUT=30s
LOG_LEVEL=info

# Rate Limiting
RATE_LIMIT_PER_MINUTE=60    # General API requests
SCRAPE_RATE_LIMIT=10        # Scraping endpoints

# Security
TRUSTED_PROXIES="127.0.0.1,::1"    # Comma-separated IPs/CIDRs

# User Agent
USER_AGENT="Mozilla/5.0 ..."
```

## Deployment

### Docker Compose (Development)
```bash
docker compose up -d
```

### Production Considerations
- Set `GIN_MODE=release` 
- Configure `TRUSTED_PROXIES` for your infrastructure
- Use persistent storage for enhanced caching
- Monitor rate limits and adjust as needed
- Consider adding authentication for private profiles

## Rate Limiting

Built-in protection with HTTP headers:
- `X-RateLimit-Limit`: Requests allowed per minute
- `X-RateLimit-Remaining`: Requests left
- `Retry-After`: Seconds to wait when blocked

**429 responses** when limits exceeded.

## Frontend Integration

### Next.js Example
```typescript
// Client-side
const { data } = useSWR('/api/goodreads', 
  () => fetch('http://localhost:8080/api/v1/portfolio/username').then(r => r.json())
);

// Build-time
export async function getStaticProps() {
  const data = await fetch('http://localhost:8080/api/v1/portfolio/username').then(r => r.json());
  return { props: { goodreadsData: data }, revalidate: 3600 };
}
```

## Requirements

- **Public Goodreads profile** - Private profiles require authentication
- **Docker & Docker Compose** - For easy deployment
- **Go 1.23+** - For local development

## Tech Stack

- **Go 1.23** with Gin framework
- **goquery** for HTML parsing  
- **resty** for HTTP requests
- **golang.org/x/time/rate** for rate limiting
- **Docker** for containerization

## License

MIT License
