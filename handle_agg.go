package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MichalGul/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

func handleAgg(s *state, cmd command) error {
	// time_between_reqs interval feed 1s 1m 1h
	// var feedStr string = "https://www.wagslane.dev/index.xml"
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	if len(cmd.args) < 1 || len(cmd.args) > 2 {
		return fmt.Errorf("agg command expects one argument of time aggregation interval: %v <timme_between_reqs>", cmd.name)
	}

	time_between_reqs := cmd.args[0]
	parsedTime, _ := time.ParseDuration(time_between_reqs)

	fmt.Printf("Collecting feeds every: %s \n", parsedTime)

	ticker := time.NewTicker(parsedTime)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

}

func scrapeFeeds(s *state) error {

	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error getting next feed to fetch: %v", err)
	}

	nextFeed, markErr := s.db.MarkFeedFetched(context.Background(), nextFeed.ID)
	if markErr != nil {
		return fmt.Errorf("error marking feed as fetched %s: %v", nextFeed.Name, markErr)
	}

	rssFeed, feedErr := fetchFeed(context.Background(), nextFeed.Url)
	if feedErr != nil {
		return fmt.Errorf("error fetching feed %s: %v", nextFeed.Url, feedErr)
	}

	fmt.Printf("Aggregated items in Feed:")
	fmt.Printf("------------------------------------------------------\n")

	for i := range rssFeed.Channel.Item {
		fmt.Printf("Adding post: %s \n", rssFeed.Channel.Item[i].Title)
		post, createErr := s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       rssFeed.Channel.Item[i].Title,
			Url:         rssFeed.Channel.Item[i].Link,
			Description: parseToNullString(rssFeed.Channel.Item[i].Description),
			PublishedAt: parseStringToNullTime(rssFeed.Channel.Item[i].PubDate),
			FeedID:      nextFeed.ID,
		})
		if createErr != nil {
			// Check if it's a duplicate URL error
			if strings.Contains(createErr.Error(), "duplicate key") && strings.Contains(createErr.Error(), "url") {
				// This is a duplicate URL - just ignore it as per requirements
				fmt.Printf("Skipping duplicate post: %s\n", rssFeed.Channel.Item[i].Link)
				continue // Continue to the next post
			}

			// It's a different kind of error - log it
			fmt.Printf("Error creating post %s: %v\n", rssFeed.Channel.Item[i].Title, err)
			// You might want to continue anyway to process other posts
			continue
		}

		fmt.Printf("Successfuly added post: %s \n", post.Title)
	}

	return nil
}

func parseToNullString(input string) sql.NullString {
	if input != "" {
		return sql.NullString{
			String: input,
			Valid:  true,
		}
	} else {
		return sql.NullString{
			Valid: false,
		}
	}
}

func parseStringToNullTime(input string) sql.NullTime {
	if input != "" {
		// RSS feeds can use different date formats
		// Try the common formats, you might need to adjust these
		parsedTime, err := time.Parse(time.RFC1123Z, input)
		if err != nil {
			// Try alternative format
			parsedTime, err = time.Parse(time.RFC822, input)
			if err != nil {
				// Try another common format
				parsedTime, err = time.Parse("2006-01-02T15:04:05Z", input)
			}
		}

		if err == nil {
			return sql.NullTime{
				Time:  parsedTime,
				Valid: true,
			}
		} else {
			// Log the error but continue
			fmt.Printf("Failed to parse date %s: %v\n", input, err)
			return sql.NullTime{Valid: false}
		}

	} else {
		return sql.NullTime{Valid: false}
	}
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return &RSSFeed{}, fmt.Errorf("error creating request %v", err)
	}
	req.Header.Set("User-Agent", "gator")

	httpClient := &http.Client{}

	resp, resp_err := httpClient.Do(req)
	if resp_err != nil {
		return &RSSFeed{}, fmt.Errorf("error getting response %v", resp_err)
	}
	defer resp.Body.Close()

	byteArray, read_err := io.ReadAll(resp.Body)
	if read_err != nil {
		return &RSSFeed{}, fmt.Errorf("error reading response %v", read_err)
	}

	rssFeed := RSSFeed{}
	unmarshallErr := xml.Unmarshal(byteArray, &rssFeed)
	if unmarshallErr != nil {
		return &RSSFeed{}, fmt.Errorf("error unmarshalling response %v", unmarshallErr)

	}

	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)

	for i := range rssFeed.Channel.Item {
		rssFeed.Channel.Item[i].Title = html.UnescapeString(rssFeed.Channel.Item[i].Title)
		rssFeed.Channel.Item[i].Description = html.UnescapeString(rssFeed.Channel.Item[i].Description)
	}

	return &rssFeed, nil
}
