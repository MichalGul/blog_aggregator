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
	"github.com/google/uuid"
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
	var feedStr string = "https://www.wagslane.dev/index.xml"
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rssFeed, err := fetchFeed(ctx, feedStr)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed at %s, err: %v", feedStr, err)
	}

	fmt.Printf("%+v\n", rssFeed)

	return nil

}

func handleFeeds(s *state, cmd command) error {

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all feeds from database %v", err)
	}

	fmt.Printf("-----------------------------------\n")
	for i := range feeds {
		printFeedDB(feeds[i], s)
	}

	return nil

}

func printFeedDB(feed database.GetFeedsRow, s *state) {

	creatorName, err := s.db.GetUsernameById(context.Background(), feed.UserID)
	if err != nil {
		fmt.Printf("Error while parsing user id to name for feed: %s, %w", feed.Name, err)
	}

	fmt.Printf("Feed Name: %v \n", feed.Name)
	fmt.Printf("Feed Url: %v \n", feed.Url)
	fmt.Printf("Feed Creator: %v \n", creatorName)

	fmt.Printf("-----------------------------------\n")

}

func handleAddFeed(s *state, cmd command) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed command expects two arguments of feed name and url")
	}
	current_user, user_err := s.db.GetUser(context.Background(), s.config.CURRENT_USER_NAME)
	if user_err != nil {
		return fmt.Errorf("error getting user from db to add feed with name %s: %v", cmd.args[0], user_err)
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	createdFeed, create_error := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    current_user.ID,
	})
	if create_error != nil {
		return fmt.Errorf("error adding feed: %s to database: %v", feedName, create_error)
	}

	fmt.Printf("Feed was successfuly created \n")
	fmt.Printf("ID: %v \n", createdFeed.ID)
	fmt.Printf("Created at: %v \n", createdFeed.CreatedAt)
	fmt.Printf("Updated at: %v \n", createdFeed.UpdatedAt)
	fmt.Printf("Feed Name: %v \n", createdFeed.Name)
	fmt.Printf("Feed Url: %v \n", createdFeed.Url)
	fmt.Printf("Feed User id: %v \n", createdFeed.UserID)

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
