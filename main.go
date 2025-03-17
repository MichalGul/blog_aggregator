package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/MichalGul/blog_aggregator/internal/config"
	"github.com/MichalGul/blog_aggregator/internal/database"

	_ "github.com/lib/pq"
)

func main() {

	configData, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading config data %v", err)
		os.Exit(1)
	}
	// fmt.Printf("Config %v \n", configData)

	db, db_err := sql.Open("postgres", configData.DB_URL)
	if db_err != nil {
		fmt.Printf("Error connecting to database using %s , error: %v", configData.DB_URL, db_err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	defer db.Close()

	appState := state{
		db:     dbQueries,
		config: &configData,
	}

	cliCommands := commands{
		handableCommands: map[string]func(*state, command) error{},
	}

	cliCommands.register("login", handlerLogin)
	cliCommands.register("register", handleRegister)
	cliCommands.register("reset", handleReset)
	cliCommands.register("users", handlerUsers)
	cliCommands.register("agg", handleAgg)
	cliCommands.register("addfeed", handleAddFeed)
	cliCommands.register("feeds", handleFeeds)
	cliCommands.register("follow", handleFollow)
	cliCommands.register("following", handleFollowing)

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
