package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jdfincher/gator/internal/database"
	"github.com/lib/pq"
)

func scrapeFeeds(s *state) error {
	nextfeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error: could not retrieve next feed to fetch -> %w", err)
	}
	fmt.Printf(`

░█▀▀░█▀▀░▀█▀░█▀▀░█░█░▀█▀░█▀█░█▀▀░░░█▀▀░█▀▀░█▀▀░█▀▄
░█▀▀░█▀▀░░█░░█░░░█▀█░░█░░█░█░█░█░░░█▀▀░█▀▀░█▀▀░█░█
░▀░░░▀▀▀░░▀░░▀▀▀░▀░▀░▀▀▀░▀░▀░▀▀▀░░░▀░░░▀▀▀░▀▀▀░▀▀░
»»»» %v
`+"\n", nextfeed.Url)
	RSS, err := fetchFeed(context.Background(), nextfeed.Url)
	if err != nil {
		return err
	}
	fetched := database.MarkFeedFetchedParams{
		UpdatedAt: time.Now(),
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		ID: nextfeed.ID,
	}
	if err := s.db.MarkFeedFetched(context.Background(), fetched); err != nil {
		return fmt.Errorf("error: could not mark feed as fetched -> %w", err)
	}

	for i := range RSS.Channel.Item {
		pubDate, err := parsePubDate(RSS.Channel.Item[i].PubDate)
		if err != nil {
		}
		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       RSS.Channel.Item[i].Title,
			Url:         RSS.Channel.Item[i].Link,
			Description: RSS.Channel.Item[i].Description,
			PublishedAt: sql.NullTime{
				Time:  pubDate,
				Valid: true,
			},
			FeedID: nextfeed.ID,
		}
		_, err = s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			handleInsertErr(err)
		} else {
			fmt.Printf("Adding record for -> %v\n", RSS.Channel.Item[i].Title)
		}
	}
	fmt.Printf(`
░█▀▀░█░░░█▀▀░█▀▀░█▀█░▀█▀░█▀█░█▀▀░░░░░░░░░
░▀▀█░█░░░█▀▀░█▀▀░█▀▀░░█░░█░█░█░█░░░░░░░░░
░▀▀▀░▀▀▀░▀▀▀░▀▀▀░▀░░░▀▀▀░▀░▀░▀▀▀░▀░░▀░░▀░` + "\n")
	fmt.Printf("»»»» Awaiting next fetch round...\n")
	return nil
}

func parsePubDate(pubdate string) (time.Time, error) {
	s := strings.TrimSpace(pubdate)
	layouts := []string{
		time.RFC1123,     // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC1123Z,    // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC3339,     // "2006-01-02T15:04:05Z07:00"
		time.RFC3339Nano, // "2006-01-02T15:04:05.999999999Z07:00"
		time.RFC822,      // "02 Jan 06 15:04 MST"
		time.RFC822Z,     // "02 Jan 06 15:04 -0700"
		time.RFC850,      // "Monday, 02-Jan-06 15:04:05 MST"
		time.DateTime,    // "2006-01-02 15:04:05"
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Now(), fmt.Errorf("error: pubDate not in a recognizable format or is null\n PublishedAt set to current time")
}

func handleInsertErr(err error) {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) {
		if pqErr.Code == "23505" && pqErr.Constraint == "posts_url_key" {
			fmt.Printf("skipping: post already recorded in previous fetch ✓\n")
			return
		}
		fmt.Printf("db error (%s/%s): %s\n", pqErr.Code, pqErr.Constraint, pqErr.Message)
		return
	}
	fmt.Printf("insert error: %v\n", err)
}
