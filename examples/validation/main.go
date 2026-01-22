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

	fmt.Println("\n=== Example 4: ValidateAll Fail-Fast ===")
	demonstrateValidateAll()
}

func demonstrateValidConfig() {
	os.Setenv("DB_PASSWORD", "secret")
	defer os.Unsetenv("DB_PASSWORD")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Load error: %v", err)
		return
	}

	db, err := cfg.Database()
	if err != nil {
		log.Printf("Validation error: %v", err)
		return
	}

	fmt.Println("Configuration is valid!")
	fmt.Printf("Database: %s@%s:%d/%s\n", db.User, db.Host, db.Port, db.Database)
}

func demonstrateMissingField() {
	os.Unsetenv("DB_PASSWORD")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Load error: %v", err)
		return
	}

	_, err = cfg.Database()
	if err != nil {
		fmt.Printf("Expected error caught: %v\n", err)
		return
	}

	fmt.Println("Unexpected: Validation passed")
}

func demonstrateInvalidPort() {
	os.Setenv("SERVER_PORT", "99999")
	defer os.Unsetenv("SERVER_PORT")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Load error: %v", err)
		return
	}

	_, err = cfg.Server()
	if err != nil {
		fmt.Printf("Expected error caught: %v\n", err)
		return
	}

	fmt.Println("Unexpected: Validation passed")
}

func demonstrateValidateAll() {
	os.Setenv("DB_PASSWORD", "secret")
	os.Setenv("OPENAI_API_KEY", "sk-test")
	defer os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, err := config.Load()
	if err != nil {
		log.Printf("Load error: %v", err)
		return
	}

	if err := cfg.ValidateAll(); err != nil {
		fmt.Printf("ValidateAll error: %v\n", err)
		return
	}

	fmt.Println("All configurations validated successfully!")
}
