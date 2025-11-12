package main

import (
	"fmt"
	"log"
	"os"

	"github.com/JohnPlummer/jp-go-config"
)

func main() {
	fmt.Println("=== Example 1: Valid Configuration ===")
	demonstrateValidConfig()

	fmt.Println("\n=== Example 2: Missing Required Field ===")
	demonstrateMissingField()

	fmt.Println("\n=== Example 3: Invalid Port ===")
	demonstrateInvalidPort()

	fmt.Println("\n=== Example 4: Out of Range Temperature ===")
	demonstrateInvalidTemperature()
}

func demonstrateValidConfig() {
	// Set valid environment variables
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PASSWORD", "secret")

	std, err := config.NewStandard()
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return
	}

	dbConfig := config.DatabaseConfigFromViper(std)

	if err := dbConfig.Validate(); err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Println("Configuration is valid!")
	fmt.Printf("Database: %s@%s:%d/%s\n", dbConfig.User, dbConfig.Host, dbConfig.Port, dbConfig.Database)
}

func demonstrateMissingField() {
	// Clear required environment variable
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PASSWORD")

	std, err := config.NewStandard()
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return
	}

	dbConfig := config.DatabaseConfigFromViper(std)
	dbConfig.Host = ""
	dbConfig.Password = ""

	if err := dbConfig.Validate(); err != nil {
		fmt.Printf("Expected error caught: %v\n", err)
		return
	}

	fmt.Println("Unexpected: Validation passed")
}

func demonstrateInvalidPort() {
	os.Setenv("SERVER_PORT", "99999")
	defer os.Unsetenv("SERVER_PORT")

	std, err := config.NewStandard()
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return
	}

	serverConfig := config.ServerConfigFromViper(std)

	if err := serverConfig.Validate(); err != nil {
		fmt.Printf("Expected error caught: %v\n", err)
		return
	}

	fmt.Println("Unexpected: Validation passed")
}

func demonstrateInvalidTemperature() {
	os.Setenv("OPENAI_API_KEY", "sk-test")
	os.Setenv("OPENAI_TEMPERATURE", "3.0")
	defer os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("OPENAI_TEMPERATURE")

	std, err := config.NewStandard()
	if err != nil {
		log.Printf("Error creating config: %v", err)
		return
	}

	openaiConfig := config.OpenAIConfigFromViper(std)

	if err := openaiConfig.Validate(); err != nil {
		fmt.Printf("Expected error caught: %v\n", err)
		return
	}

	fmt.Println("Unexpected: Validation passed")
}
