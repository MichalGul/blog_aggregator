package main

import (
	"fmt"
	"os"

	"github.com/MichalGul/blog_aggregator/internal/config"
)

func main() {

	configData, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config data %w", err)
	}
	fmt.Printf("Config %v \n", configData)

	appState := state{
		config: &configData,
	}

	cliCommands := commands{
		handableCommands: map[string]func(*state, command) error{},
	}

	cliCommands.register("login", handlerLogin)

	providedCommands := os.Args
	if len(providedCommands) < 2 {
		fmt.Println("Missing arguments")
		os.Exit(1)
	}

	command := command{
		name: providedCommands[1],
		args: providedCommands[2:],
	}

	cmdErr := cliCommands.run(&appState, command)
	if cmdErr != nil {
		fmt.Println(cmdErr)
		os.Exit(1)
	}

}
