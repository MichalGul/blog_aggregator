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

func handleFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("follow command expects one argument of url")
	}

	feedUrl := cmd.args[0]

	feed, feedErr := s.db.GetFeedByUrl(context.Background(), feedUrl)
	if feedErr != nil {
		return fmt.Errorf("error getting feed from db by url %s: %v", feedUrl, feedErr)
	}

	createdFeedFollow, createError := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if createError != nil {
		return fmt.Errorf("error adding feed follow: %s to database: %v", feedUrl, createError)
	}

	fmt.Printf("Feed Follow was successfuly created \n")
	fmt.Printf("FeedFollow Name: %v \n", createdFeedFollow.FeedName)
	fmt.Printf("FeedFollow User id: %v \n", createdFeedFollow.UserName)

	return nil

}

func handleUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("unfollow command expects one argument of url")
	}
	feedUrl := cmd.args[0]

	delErr := s.db.DeleteFeedsFollow(context.Background(), database.DeleteFeedsFollowParams{
		Url:    feedUrl,
		UserID: user.ID,
	})

	if delErr != nil {
		return fmt.Errorf("error while handling unfollow by url %s: %v", feedUrl, delErr)
	}

	return nil

}

func handleFollowing(s *state, cmd command, user database.User) error {

	feedFollowsForUser, feedFollowsErr := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if feedFollowsErr != nil {
		return fmt.Errorf("error getting feed follow for user %s: %v", user.Name, feedFollowsErr)
	}

	fmt.Printf("Feeds that user %s follows:\n", user.Name)
	fmt.Printf("-----------------------------------\n")
	for i := range feedFollowsForUser {
		fmt.Printf("Name: %v \n", feedFollowsForUser[i].FeedName)
	}
	fmt.Printf("-----------------------------------\n")
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
