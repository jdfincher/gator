package main

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jdfincher/gator/internal/database"
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
---> %v
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
		fmt.Printf("....................\n")
		fmt.Printf("%v\n%v\n", RSS.Channel.Item[i].Title, RSS.Channel.Item[i].Link)
	}
	return nil
}
