package main

import (
	"context"
	"fmt"
	"time"

	"github.com/MichalGul/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

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

func handleAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("addfeed command expects two arguments of feed name and url")
	}

	feedName := cmd.args[0]
	feedUrl := cmd.args[1]

	createdFeed, create_error := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedUrl,
		UserID:    user.ID,
	})
	if create_error != nil {
		return fmt.Errorf("error adding feed: %s to database: %v", feedName, create_error)
	}

	_, errorFeedFollow := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    createdFeed.ID,
	})
	if errorFeedFollow != nil {
		return fmt.Errorf("error adding feed follow for user %s to database %v", user.Name, errorFeedFollow)
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