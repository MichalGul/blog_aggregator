package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"
	"github.com/MichalGul/blog_aggregator/internal/config"
	"github.com/MichalGul/blog_aggregator/internal/database"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handableCommands map[string]func(*state, command) error
}

// Checks for logged in user in system. DRY code
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {

	return func(s *state, cmd command) error {
		currentUser, userErr := s.db.GetUser(context.Background(), s.config.CURRENT_USER_NAME)
		if userErr != nil {
			return fmt.Errorf("error getting current user: %v", userErr)
		}

		return handler(s, cmd, currentUser)
	}
}

// Register new handler function for command name
func (c *commands) register(name string, f func(*state, command) error) {
	c.handableCommands[name] = f

}

// Runs given command with provided state if exists
func (c *commands) run(s *state, cmd command) error {
	avaliableCommand, commandExists := c.handableCommands[cmd.name]
	if !commandExists {
		return fmt.Errorf("command %s not avaliable", cmd.name)
	}
	return avaliableCommand(s, cmd)
}

func handleReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())

	if err != nil {
		return fmt.Errorf("error reseting database state %v", err)
	}

	ferr := s.db.DeleteFeeds(context.Background())
	if ferr != nil {
		return fmt.Errorf("error reseting database state %v", err)
	}

	fmt.Println("Database was cleaned from data")

	return err
}

func handleAgg(s *state, cmd command) error {
	// time_between_reqs interval feed 1s 1m 1h
	// var feedStr string = "https://www.wagslane.dev/index.xml"
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()

	if (len(cmd.args) < 1 || len(cmd.args) >2 ) {
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
		fmt.Printf("Name: %s \n", rssFeed.Channel.Item[i].Title)
	}

	return nil
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
