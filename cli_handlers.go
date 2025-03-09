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

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command expects single argument")
	}
	username := cmd.args[0]

	user, db_err := s.db.GetUser(context.Background(), username)
	if db_err != nil {
		return fmt.Errorf("error getting user from database %v", db_err)
	}

	err := s.config.SetUser(user.Name)
	if err != nil {
		return fmt.Errorf("error while seting user to %s: %w", cmd.args[0], err)
	}

	fmt.Printf("User has been set to %s \n", username)

	return nil

}

func handlerUsers(s *state, cmd command) error {
	users, db_err := s.db.GetUsers(context.Background())
	if db_err != nil {
		return fmt.Errorf("error while listring users from database: %v", db_err)
	}

	loggedUser := s.config.CURRENT_USER_NAME
	for _, user := range users {
		if user.Name == loggedUser {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil

}

func handleRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register command expects single argument of user name")
	}
	username := cmd.args[0]

	createdUser, db_err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      username,
	})
	if db_err != nil {
		return fmt.Errorf("error while adding user %s to database %w", username, db_err)
	}

	err := s.config.SetUser(username)
	if err != nil {
		return fmt.Errorf("error while seting user to %s: %w", cmd.args[0], err)
	}

	fmt.Printf("User was successfuly created \n")
	fmt.Printf("ID: %v \n", createdUser.ID)
	fmt.Printf("Created at: %v \n", createdUser.CreatedAt)
	fmt.Printf("Updated at: %v \n", createdUser.UpdatedAt)
	fmt.Printf("Name: %v \n", createdUser.Name)

	return nil
}

func handleReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
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
