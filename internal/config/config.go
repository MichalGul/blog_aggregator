package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const configFileName = ".gatorconfig.json"

type Config struct {
	DB_URL            string `json:"db_url"`
	CURRENT_USER_NAME string `json:"current_user_name"`
}

func Read() (Config, error) {
	confiFilePath, pathError := getConfigFilePath()
	if pathError != nil {
		return Config{}, fmt.Errorf("error getting config path: %w", pathError)
	}

	jsonFile, err := os.Open(confiFilePath)
	if err != nil {
		fmt.Println(err)
		return Config{}, err
	}
	defer jsonFile.Close()

	// Read the file's content
	data, readErr := io.ReadAll(jsonFile)
	if err != nil {
		return Config{}, fmt.Errorf("error reading file: %w", readErr)
	}

	// Parse json data
	var config Config
	marshallErr := json.Unmarshal(data, &config)
	if marshallErr != nil {
		return Config{}, fmt.Errorf("error parsing data")
	}

	return config, nil

}

func (c *Config) SetUser(username string) error{
	c.CURRENT_USER_NAME = username
	return write(*c)

}

func getConfigFilePath() (string, error) {
	basePath, err := os.UserHomeDir()
	if err != nil {
		fmt.Println("Error reading config file:", err)
		return "", err
	}

	fullPath := filepath.Join(basePath, configFileName)
	return fullPath, nil
}

func write(cfg Config) error {

	confiFilePath, pathError := getConfigFilePath()
	if pathError != nil {
		return fmt.Errorf("error getting config path: %w", pathError)
	}

	// Open a file for writing
	file, err := os.Create(confiFilePath)
	if err != nil {
		return fmt.Errorf("Error creating file:", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", " ")
	encodeErr := encoder.Encode(cfg)
	if encodeErr != nil {
		return fmt.Errorf("error encoding Json: %w", err)
	}

	return nil

}
