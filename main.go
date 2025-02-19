package main

import (
	"fmt"

	"github.com/MichalGul/blog_aggregator/internal/config"
)

func main() {

	configData, err := config.Read()

	if err != nil {
		fmt.Printf("Error reading config data %w", err)
	}

	configData.SetUser("testuser")
	fmt.Printf("Config %v \n", configData)
}
