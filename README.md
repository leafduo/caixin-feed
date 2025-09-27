# Caixin Feed

A Go-based RSS/JSON feed service that fetches articles from Caixin (财新) and provides them as a JSON feed. The service uses background fetching with SQLite storage for optimal performance and reliability.

## Features

- **Background Fetching**: Articles are fetched from Caixin API every 15 minutes in the background
- **Database Storage**: SQLite database for persistent article storage
- **Fast API**: Sub-second response times with cached articles
- **Reliability**: Service continues to work even if Caixin API is temporarily down
- **Docker Support**: Easy deployment with Docker and Docker Compose
- **JSON Feed**: Modern JSON feed format for better compatibility

## Live Feed

Visit the live feed at: [https://leafduo.com/caixin-feed/feed.json](https://leafduo.com/caixin-feed/feed.json)

## Quick Start

### Using Docker Compose (Recommended)

```bash
git clone https://github.com/leafduo/caixin-feed.git
cd caixin-feed
docker-compose up -d
```

The service will be available at `http://localhost:1323/feed.json`

### Manual Setup

1. **Prerequisites**
   - Go 1.24+ 
   - SQLite3

2. **Installation**
   ```bash
   git clone https://github.com/leafduo/caixin-feed.git
   cd caixin-feed
   go mod download
   go build -o caixin-feed main.go
   ```

3. **Run**
   ```bash
   ./caixin-feed
   ```

## Configuration

The service can be configured by modifying constants in `main.go`:

- `FetchInterval`: How often to fetch new articles (default: 15 minutes)
- `MaxArticles`: Maximum number of articles to return (default: 100)

## API Documentation

### Endpoints

#### GET `/feed.json`

Returns a JSON feed containing the latest articles from Caixin.

**Response Format:**
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

## Architecture

The application uses a background fetching pattern with database storage:

1. **Background Fetcher**: Runs every 15 minutes to fetch new articles from Caixin API
2. **Database Storage**: SQLite database stores articles with metadata
3. **API Endpoint**: Serves cached articles from database for fast responses

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

## Deployment

### Docker

The application includes a Dockerfile for containerized deployment:

```bash
docker build -t caixin-feed .
docker run -p 1323:1323 caixin-feed
```

### Docker Compose

Use the included `docker-compose.yml` for easy deployment:

```bash
docker-compose up -d
```

### GitHub Actions

The repository includes GitHub Actions workflow for automated Docker image building and publishing to GitHub Container Registry.

## Development

### Project Structure

```
caixin-feed/
├── main.go              # Main application code
├── go.mod              # Go module dependencies
├── Dockerfile          # Docker configuration
├── docker-compose.yml  # Docker Compose configuration
├── ARCHITECTURE.md     # Detailed architecture documentation
└── .github/workflows/  # GitHub Actions workflows
```

### Dependencies

- [Echo v4](https://echo.labstack.com/) - Web framework
- [Gorilla Feeds](https://github.com/gorilla/feeds) - RSS/JSON feed generation
- [SQLite3](https://github.com/mattn/go-sqlite3) - Database driver

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

## Author

Created by [leafduo](https://github.com/leafduo)
