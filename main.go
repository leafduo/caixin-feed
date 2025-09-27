package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// Configuration
const (
	FetchInterval = 15 * time.Minute // Fetch every 10 minutes
	MaxArticles   = 100              // Maximum articles to return
)

func main() {
	// Initialize database
	var err error
	db, err = initDB()
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Start background fetcher
	go startBackgroundFetcher()

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/feed.json", generateFeed)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}

// ArticleItem represents a single article item returned from the API
type ArticleItem struct {
	ID         string `json:"id"`
	Title      string `json:"title"`
	Summary    string `json:"summary"`
	PictureURL string `json:"pics"`
	Time       string `json:"time"`
	URL        string `json:"web_url"`
}

type jsonStruct struct {
	Data struct {
		Header []ArticleItem `json:"header"`
		List   []ArticleItem `json:"list"`
	} `json:"data"`
}

type Article struct {
	ID            string
	Title         string
	Summary       string
	PictureURL    string
	PublishedAt   time.Time
	URL           string
	LastSeenAt    time.Time
	LastSeenIndex int
}

func initDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "articles.db")
	if err != nil {
		return nil, err
	}

	// Create articles table if it doesn't exist
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS articles (
		id TEXT PRIMARY KEY,
		title TEXT NOT NULL,
		summary TEXT,
		picture_url TEXT,
		published_at DATETIME,
		url TEXT,
		last_seen_at DATETIME,
		last_seen_index INTEGER
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func startBackgroundFetcher() {
	// Fetch immediately on startup
	fetchAndStoreArticles()

	// Then fetch at configured interval
	ticker := time.NewTicker(FetchInterval)
	defer ticker.Stop()

	for range ticker.C {
		fetchAndStoreArticles()
	}
}

func fetchAndStoreArticles() {
	log.Println("Starting background fetch...")

	// Use the same LastSeenAt time for all articles in this fetch batch
	fetchTime := time.Now()
	articleIndex := 0

	// Fetch articles from pages 1, 2, and 3
	for page := 1; page <= 3; page++ {
		articles, err := fetchPage(page, fetchTime, articleIndex)
		if err != nil {
			log.Printf("Error fetching page %d: %v", page, err)
			continue
		}

		// Store articles in database
		for _, article := range articles {
			err := storeArticle(article)
			if err != nil {
				log.Printf("Error storing article %s: %v", article.ID, err)
			}
		}

		// Update article index for next page
		articleIndex += len(articles)
	}

	log.Println("Background fetch completed")
}

func fetchPage(pageNumber int, fetchTime time.Time, startIndex int) ([]Article, error) {
	client := http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://mapiv5.caixin.com//m/api/getWapIndexListByPage?page=%d", pageNumber))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	jsonData := jsonStruct{}
	err = json.NewDecoder(resp.Body).Decode(&jsonData)
	if err != nil {
		return nil, err
	}

	_, _ = io.ReadAll(resp.Body)

	var articles []Article
	articleIndex := startIndex

	// First, process header articles (banner articles) - these should come first
	headerArticles := parseArticleItems(jsonData.Data.Header, fetchTime, articleIndex)
	articles = append(articles, headerArticles...)
	articleIndex += len(headerArticles)

	// Then, process list articles
	listArticles := parseArticleItems(jsonData.Data.List, fetchTime, articleIndex)
	articles = append(articles, listArticles...)

	return articles, nil
}

// parseArticleItems parses a list of article items and returns Article slice
func parseArticleItems(items []ArticleItem, fetchTime time.Time, startIndex int) []Article {
	var articles []Article
	articleIndex := startIndex

	for _, item := range items {
		// Use URL as ID if ID is empty or "0"
		articleID := item.ID
		if articleID == "" || articleID == "0" {
			articleID = item.URL
		}

		// Parse the timestamp from string to time.Time
		var publishedAt time.Time
		if timestamp, err := strconv.ParseInt(item.Time, 10, 64); err == nil {
			publishedAt = time.Unix(timestamp, 0)
		} else {
			// If parsing fails, use current time
			publishedAt = time.Now()
		}

		articles = append(articles, Article{
			ID:            articleID,
			Title:         item.Title,
			Summary:       item.Summary,
			PictureURL:    item.PictureURL,
			PublishedAt:   publishedAt,
			URL:           item.URL,
			LastSeenAt:    fetchTime,
			LastSeenIndex: articleIndex,
		})
		articleIndex++
	}

	return articles
}

func storeArticle(article Article) error {
	// Use INSERT OR REPLACE to handle duplicates
	stmt, err := db.Prepare(`
		INSERT OR REPLACE INTO articles (id, title, summary, picture_url, published_at, url, last_seen_at, last_seen_index)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(article.ID, article.Title, article.Summary, article.PictureURL, article.PublishedAt, article.URL, article.LastSeenAt, article.LastSeenIndex)
	return err
}

func getArticlesFromDB() ([]Article, error) {
	rows, err := db.Query(`
		SELECT id, title, summary, picture_url, published_at, url, last_seen_at, last_seen_index
		FROM articles
		ORDER BY last_seen_at DESC, last_seen_index ASC
		LIMIT ?
	`, MaxArticles)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var article Article
		err := rows.Scan(&article.ID, &article.Title, &article.Summary, &article.PictureURL, &article.PublishedAt, &article.URL, &article.LastSeenAt, &article.LastSeenIndex)
		if err != nil {
			return nil, err
		}
		articles = append(articles, article)
	}

	return articles, nil
}

// Handler
func generateFeed(c echo.Context) error {
	// Get articles from database
	articles, err := getArticlesFromDB()
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
	}

	feed := &feeds.Feed{
		Title:       "财新首页文章",
		Link:        &feeds.Link{Href: "https://leafduo.com/caixin-feed/feed.json"},
		Description: "财新首页文章",
		Author:      &feeds.Author{Name: "leafduo", Email: "leafduo@gmail.com"},
		Created:     time.Now(),
	}

	// Convert articles to feed items
	for _, article := range articles {
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       article.Title,
			Description: article.Summary,
			Content:     fmt.Sprintf("<p>%s</p><img src=\"%s\">", article.Summary, article.PictureURL),
			Created:     article.PublishedAt,
			Link:        &feeds.Link{Href: article.URL},
			Id:          article.ID,
		})
	}

	jsonFeed, err := feed.ToJSON()
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
	}

	return c.JSONBlob(200, []byte(jsonFeed))
}
