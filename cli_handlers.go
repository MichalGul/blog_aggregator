package main

import (
	"context"
	"fmt"
	"strconv"

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

func handleBrowse(s *state, cmd command, user database.User) error {
	var limitRaw string
	if len(cmd.args) == 0 {
		limitRaw = "2"
	} else {
		limitRaw = cmd.args[0]
	}

	limit, parsErr := strconv.Atoi(limitRaw)
	if parsErr != nil {
		return fmt.Errorf("parse error:", parsErr)
	}

	posts, browseErr := s.db.GetPostForUser(context.Background(), database.GetPostForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})

	if browseErr != nil {
		return fmt.Errorf("error while browsing posts for user %s: %v", user.Name, browseErr)
	}

	fmt.Printf("Saved posts for user: %s \n", user.Name)
	fmt.Printf("==================== \n")
	for i := range posts {
		fmt.Printf("Title: %s \n", posts[i].Title)
		fmt.Printf("Published: %s \n", posts[i].PublishedAt.Time)
		fmt.Printf("Description: %s \n", posts[i].Description.String)
		feed, _ := s.db.GetFeedById(context.Background(), posts[i].FeedID)
		fmt.Printf("Feed source: %s \n", feed.Name)

		fmt.Printf("==================== \n")
	}

	return nil

}
