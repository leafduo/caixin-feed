# Architecture Documentation

## Overview
The Caixin Feed service uses a background fetching pattern with SQLite database storage to provide fast, reliable access to Caixin articles via a JSON feed API.

## System Architecture

### Components

1. **Background Fetcher**: Goroutine that periodically fetches articles from Caixin API
2. **Database Layer**: SQLite database for persistent article storage
3. **Web Server**: Echo-based HTTP server serving the JSON feed
4. **API Endpoint**: Single endpoint `/feed.json` that returns cached articles

### Data Flow

```
Caixin API → Background Fetcher → SQLite Database → Web Server → JSON Feed
```

## Key Features

### 1. Background Fetching
- Articles are fetched from Caixin API every 15 minutes
- Fetches happen immediately on startup, then at regular intervals
- Fetches from multiple pages (1, 2, 3) to get comprehensive coverage
- Fetch interval is configurable via `FetchInterval` constant (default: 15 minutes)

### 2. Database Storage
- SQLite database (`articles.db`) stores all fetched articles
- Articles table with comprehensive metadata including:
  - `id`: Unique article identifier
  - `title`: Article title
  - `summary`: Article summary/description
  - `picture_url`: Article image URL
  - `published_at`: Original publication timestamp
  - `url`: Article URL
  - `last_seen_at`: When article was last fetched
  - `last_seen_index`: Order of article in fetch batch
- Duplicate articles are handled with `INSERT OR REPLACE`
- Articles are ordered by `last_seen_at DESC, last_seen_index ASC`

### 3. API Response
- `/feed.json` endpoint serves articles from the database
- Sub-second response times since no external API calls are made
- Returns up to 100 most recent articles (configurable via `MaxArticles`)
- JSON Feed format (https://jsonfeed.org/version/1.1)

## Implementation Details

### Background Fetcher Process

1. **Startup**: Immediate fetch on application start
2. **Scheduling**: Ticker-based periodic fetching every 15 minutes
3. **Multi-page Fetching**: Fetches pages 1, 2, and 3 from Caixin API
4. **Article Processing**: 
   - Header articles (banner articles) processed first
   - List articles processed second
   - Maintains article order with `last_seen_index`
5. **Database Storage**: Each article stored with `INSERT OR REPLACE`

### Article Processing

- **ID Generation**: Uses article ID from API, falls back to URL if ID is empty or "0"
- **Timestamp Parsing**: Converts Unix timestamp to Go `time.Time`
- **Order Preservation**: Maintains original order from API response
- **Error Handling**: Continues processing even if individual articles fail

### Database Schema

```sql
CREATE TABLE articles (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    summary TEXT,
    picture_url TEXT,
    published_at DATETIME,
    url TEXT,
    last_seen_at DATETIME,
    last_seen_index INTEGER
);
```

### API Response Format

The service returns a JSON Feed with the following structure:

```json
{
  "version": "https://jsonfeed.org/version/1.1",
  "title": "财新首页文章",
  "description": "财新首页文章",
  "home_page_url": "https://leafduo.com/caixin-feed/feed.json",
  "feed_url": "https://leafduo.com/caixin-feed/feed.json",
  "author": {
    "name": "leafduo",
    "email": "leafduo@gmail.com"
  },
  "items": [
    {
      "id": "article_id",
      "title": "Article Title",
      "content_html": "<p>Article summary</p><img src=\"image_url\">",
      "url": "https://article-url",
      "date_published": "2024-01-01T00:00:00Z"
    }
  ]
}
```

## Configuration

### Constants

- `FetchInterval`: How often to fetch new articles (default: 15 minutes)
- `MaxArticles`: Maximum number of articles to return (default: 100)

### Environment

- **Database**: SQLite file `articles.db` in application directory
- **Port**: HTTP server runs on port 1323
- **Timeout**: HTTP client timeout of 20 seconds for API calls

## Benefits

1. **Performance**: API responses are sub-second with cached data
2. **Reliability**: Service continues to work even if Caixin API is temporarily down
3. **Consistency**: Articles are available even during API maintenance
4. **Scalability**: Can handle high concurrent request loads
5. **Freshness**: Regular background updates ensure content stays current
6. **Ordering**: Maintains proper article ordering from original source

## Error Handling

- **API Failures**: Individual page fetch failures don't stop the entire process
- **Database Errors**: Article storage failures are logged but don't crash the service
- **HTTP Timeouts**: 20-second timeout prevents hanging requests
- **Graceful Degradation**: Service continues to serve cached articles even during API outages

## Monitoring

- **Logging**: Comprehensive logging of fetch operations and errors
- **Database Growth**: SQLite database grows with article history
- **Performance**: Background fetching doesn't impact API response times
