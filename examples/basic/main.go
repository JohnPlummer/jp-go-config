package main

import (
	"fmt"
	"log"

	"github.com/JohnPlummer/jp-go-config"
)

func main() {
	// Load configuration with auto-discovery
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Get validated database configuration
	dbConfig, err := cfg.Database()
	if err != nil {
		log.Printf("Database config not available: %v", err)
	} else {
		fmt.Println("Database Configuration:")
		fmt.Printf("  Host: %s\n", dbConfig.Host)
		fmt.Printf("  Port: %d\n", dbConfig.Port)
		fmt.Printf("  Database: %s\n", dbConfig.Database)
		fmt.Printf("  User: %s\n", dbConfig.User)
		fmt.Printf("  SSLMode: %s\n", dbConfig.SSLMode)
		fmt.Printf("  Connection String: %s\n", dbConfig.ConnectionString())
	}

	// Get validated server configuration
	serverConfig, err := cfg.Server()
	if err != nil {
		log.Printf("Server config not available: %v", err)
	} else {
		fmt.Println("\nServer Configuration:")
		fmt.Printf("  Address: %s\n", serverConfig.Address())
		fmt.Printf("  ReadTimeout: %v\n", serverConfig.ReadTimeout)
	}
}
