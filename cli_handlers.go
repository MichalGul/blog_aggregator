package main

import (
	"context"
	"fmt"
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
