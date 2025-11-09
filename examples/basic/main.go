package main

import (
	"fmt"
	"log"

	"github.com/JohnPlummer/go-config"
)

func main() {
	// Create a standard config loader
	// By default, it loads .env files and reads environment variables with APP_ prefix
	std, err := config.NewStandard()
	if err != nil {
		log.Fatalf("Failed to create config: %v", err)
	}

	// Load database configuration from environment variables
	dbConfig := config.DatabaseConfigFromViper(std)

	fmt.Println("Database Configuration:")
	fmt.Printf("  Host: %s\n", dbConfig.Host)
	fmt.Printf("  Port: %d\n", dbConfig.Port)
	fmt.Printf("  Database: %s\n", dbConfig.Database)
	fmt.Printf("  User: %s\n", dbConfig.User)
	fmt.Printf("  SSLMode: %s\n", dbConfig.SSLMode)
	fmt.Printf("  MaxConns: %d\n", dbConfig.MaxConns)
	fmt.Printf("  Connection String: %s\n", dbConfig.ConnectionString())

	// Load server configuration
	serverConfig := config.ServerConfigFromViper(std)

	fmt.Println("\nServer Configuration:")
	fmt.Printf("  Host: %s\n", serverConfig.Host)
	fmt.Printf("  Port: %d\n", serverConfig.Port)
	fmt.Printf("  Address: %s\n", serverConfig.Address())
	fmt.Printf("  ReadTimeout: %v\n", serverConfig.ReadTimeout)
	fmt.Printf("  WriteTimeout: %v\n", serverConfig.WriteTimeout)

	// Load OpenAI configuration
	openaiConfig := config.OpenAIConfigFromViper(std)

	fmt.Println("\nOpenAI Configuration:")
	fmt.Printf("  Model: %s\n", openaiConfig.Model)
	fmt.Printf("  Temperature: %.2f\n", openaiConfig.Temperature)
	fmt.Printf("  MaxTokens: %d\n", openaiConfig.MaxTokens)
	fmt.Printf("  Timeout: %v\n", openaiConfig.Timeout)
}
