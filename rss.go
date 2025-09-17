package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedurl string) (*RSSFeed, error) {
	feed := new(RSSFeed)
	req, err := http.NewRequestWithContext(ctx, "GET", feedurl, nil)
	if err != nil {
		return feed, fmt.Errorf("error: request -> %w", err)
	}
	req.Header.Set("User-Agent", "gator")
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return feed, fmt.Errorf("error: response -> %w", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return feed, fmt.Errorf("error: Reading response -> %w", err)
	}
	if err := xml.Unmarshal(data, feed); err != nil {
		return feed, fmt.Errorf("error: Unmarshal -> %w", err)
	}
	feed.unescapeHTML()
	return feed, nil
}

func (r *RSSFeed) unescapeHTML() {
	r.Channel.Title = html.UnescapeString(r.Channel.Title)
	r.Channel.Description = html.UnescapeString(r.Channel.Description)
	for i := range r.Channel.Item {
		r.Channel.Item[i].Title = html.UnescapeString(r.Channel.Item[i].Title)
		r.Channel.Item[i].Description = html.UnescapeString(r.Channel.Item[i].Description)
	}
}
