package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Routes
	e.GET("/feed.json", generateFeed)

	// Start server
	e.Logger.Fatal(e.Start("127.0.0.1:1323"))
}

type jsonStruct struct {
	Data struct {
		List []struct {
			Title       string `json:"title"`
			Summary     string `json:"summary"`
			ChannelName string `json:"channel_name"`
			Time        string `json:"time"`
			SourceID    string `json:"source_id"`
		} `json:"list"`
	} `json:"data"`
}

// Handler
func generateFeed(c echo.Context) error {

	feed := &feeds.Feed{
		Title:       "财新首页文档",
		Link:        &feeds.Link{Href: "https://leafduo.com/caixin-feed/feed.json"},
		Description: "财新首页文档",
		Author:      &feeds.Author{Name: "leafduo", Email: "leafduo@gmail.com"},
		Created:     time.Now(),
	}

	err := addPage(1, feed)
	if err != nil {
		return echo.NewHTTPError(500, err.Error)
	}

	err = addPage(2, feed)
	if err != nil {
		return echo.NewHTTPError(500, err.Error)
	}

	err = addPage(3, feed)
	if err != nil {
		return echo.NewHTTPError(500, err.Error)
	}

	jsonFeed, err := feed.ToJSON()
	if err != nil {
		return echo.NewHTTPError(500, err.Error)
	}

	return c.JSONBlob(200, []byte(jsonFeed))
}

func addPage(pageNumber int, feed *feeds.Feed) error {
	client := http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(fmt.Sprintf("https://mappsv5.caixin.com/index_page_v5/index_page_%d.json", pageNumber))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	jsonData := jsonStruct{}
	err = json.NewDecoder(resp.Body).Decode(&jsonData)
	if err != nil {
		return err
	}

	_, _ = ioutil.ReadAll(resp.Body)

	for _, item := range jsonData.Data.List {
		timestamp, err := strconv.ParseInt(item.Time, 10, 64)
		if err != nil {
			return echo.NewHTTPError(500, err.Error())
		}
		feed.Items = append(feed.Items, &feeds.Item{
			Title:       item.Title,
			Description: item.Summary,
			Created:     time.Unix(timestamp, 0),
			Link:        &feeds.Link{Href: getLink(item.ChannelName, timestamp, item.SourceID)},
			Id:          item.SourceID,
		})
	}

	return nil
}

var channelMap = map[string]string{
	"金融":   "finance",
	"财新周刊": "weekly",
	"公司":   "companies",
	"世界":   "international",
	"政经":   "china",
	"图片":   "photos",
	"观点":   "opinion",
}

func getLink(channel string, timestamp int64, sourceID string) string {
	domain := channelMap[channel]
	if len(domain) == 0 {
		domain = "www"
	}
	date := time.Unix(timestamp, 0).Format("2006-01-02")

	return fmt.Sprintf("https://%s.caixin.com/%s/%s.html", domain, date, sourceID)
}
