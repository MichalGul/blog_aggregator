package main

import (
	"context"
	"fmt"
	"time"
	"github.com/MichalGul/blog_aggregator/internal/database"
	"github.com/google/uuid"
)

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