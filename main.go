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
			ID         string `json:"id"`
			Title      string `json:"title"`
			Summary    string `json:"summary"`
			PictureURL string `json:"pics"`
			Time       string `json:"time"`
			URL        string `json:"web_url"`
		} `json:"list"`
	} `json:"data"`
}

// Handler
func generateFeed(c echo.Context) error {

	feed := &feeds.Feed{
		Title:       "财新首页文章",
		Link:        &feeds.Link{Href: "https://leafduo.com/caixin-feed/feed.json"},
		Description: "财新首页文章",
		Author:      &feeds.Author{Name: "leafduo", Email: "leafduo@gmail.com"},
		Created:     time.Now(),
	}

	err := addPage(1, feed)
	if err != nil {
		return echo.NewHTTPError(500, err.Error)
	}

	err = addPage(2, feed)
	if err != nil {
		return echo.NewHTTPError(500, err.Error())
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
	resp, err := client.Get(fmt.Sprintf("http://mapiv5.caixin.com//m/api/getWapIndexListByPage?page=%d", pageNumber))
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
			Content:     fmt.Sprintf("<p>%s</p><img src=\"%s\">", item.Summary, item.PictureURL),
			Created:     time.Unix(timestamp, 0),
			Link:        &feeds.Link{Href: item.URL},
			Id:          item.ID,
		})
	}

	return nil
}
