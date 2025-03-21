package main

import (
	"context"
	"fmt"
	"time"

	"github.com/MichalGul/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

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
