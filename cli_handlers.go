package main

import (
	"fmt"
	"github.com/MichalGul/blog_aggregator/internal/config"
)

type state struct {
	config  *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	handableCommands map[string]func(*state, command) error 

}

// Register new handler function for command name
func (c *commands) register(name string, f func(*state, command) error){
	c.handableCommands[name]=f

}

// Runs given command with provided state if exists
func (c *commands) run (s *state, cmd command) error {
	avaliableCommand, commandExists := c.handableCommands[cmd.name]
	if !commandExists {
		return fmt.Errorf("command %s not avaliable", cmd.name)
	}
	return avaliableCommand(s,cmd)	
}


func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0{
		return fmt.Errorf("login command expects single argument")
	}
	username := cmd.args[0]

	err := s.config.SetUser(username)
	if err != nil {
		return fmt.Errorf("error while seting user to %s: %w", cmd.args[0], err)
	}
	
	fmt.Printf("User has been set to %s \n", username)

	return nil

}