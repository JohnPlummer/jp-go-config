// Package config provides enterprise-standard configuration management wrapping Viper.
package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
)

// envVarPattern matches ${VAR} or ${VAR:-default} patterns.
// Only captures the variable name, ignoring any default value syntax.
var envVarPattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?::-[^}]*)?\}`)

// FindProjectRoot searches parent directories for project root markers.
// Looks for directories containing ALL specified markerFiles.
// Returns empty string if no root found or if no markers specified.
//
// Example:
//
//	// Find directory containing both CLAUDE.md and config.yaml
//	root := FindProjectRoot("CLAUDE.md", "config.yaml")
//	if root != "" {
//	    configPath := root + "/config.yaml"
//	}
func FindProjectRoot(markerFiles ...string) string {
	if len(markerFiles) == 0 {
		return ""
	}

	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	for {
		// Check if all marker files exist in this directory
		allFound := true
		for _, marker := range markerFiles {
			markerPath := dir + "/" + marker
			if _, err := os.Stat(markerPath); err != nil {
				allFound = false
				break
			}
		}

		if allFound {
			return dir
		}

		// Move to parent directory
		parent := parentDir(dir)
		if parent == "" || parent == dir {
			break
		}
		dir = parent
	}

	return ""
}

// parentDir returns the parent directory of the given path.
// Returns empty string if at root or path is invalid.
func parentDir(dir string) string {
	lastSlash := strings.LastIndex(dir, "/")
	if lastSlash <= 0 {
		return ""
	}
	return dir[:lastSlash]
}

// ExpandEnvStrict expands ${VAR} patterns in value, returning error if VAR is not set.
// Unlike os.ExpandEnv, this function:
//   - Fails if any referenced environment variable is not set
//   - Does NOT support ${VAR:-default} syntax (ignores default values)
//   - Forces explicit configuration to prevent production surprises
//
// IMPORTANT: Call AFTER LoadEnvFile() so .env values are available.
//
// Example:
//
//	// Given DATABASE_URL is set in environment
//	value := "postgres://${DB_HOST}:${DB_PORT}/mydb"
//	expanded, err := ExpandEnvStrict(value)
//	if err != nil {
//	    // Error: environment variable DB_HOST not set
//	}
func ExpandEnvStrict(value string) (string, error) {
	// Find all ${VAR} references
	matches := envVarPattern.FindAllStringSubmatch(value, -1)
	if len(matches) == 0 {
		return value, nil
	}

	// Verify all referenced variables are set and build replacement map
	result := value
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		fullMatch := match[0] // e.g., ${VAR} or ${VAR:-default}
		varName := match[1]   // e.g., VAR
		envValue := os.Getenv(varName)
		if envValue == "" {
			return "", fmt.Errorf("environment variable %s not set", varName)
		}
		// Replace the full match (including any :-default) with the actual value
		result = strings.Replace(result, fullMatch, envValue, 1)
	}

	return result, nil
}

// LoadEnvFiles loads .env and optional .env.{env} in correct order.
// Environment-specific file overrides base .env values.
// Returns nil if files are missing (this is expected in production).
//
// Search paths are checked in order; first existing file in each category is used.
// If env is empty, only loads base .env file.
//
// Example:
//
//	// Load .env then .env.test for test environment
//	if err := LoadEnvFiles("test", ".", ".."); err != nil {
//	    log.Printf("Warning: failed to load env files: %v", err)
//	}
//
//	// Load only base .env
//	_ = LoadEnvFiles("", ".", "..")
func LoadEnvFiles(env string, searchPaths ...string) error {
	if len(searchPaths) == 0 {
		searchPaths = []string{"."}
	}

	// Load base .env file first
	for _, path := range searchPaths {
		envFile := path + "/.env"
		if _, err := os.Stat(envFile); err == nil {
			if err := loadEnvFileWithoutOverride(envFile); err != nil {
				return fmt.Errorf("failed to load %s: %w", envFile, err)
			}
			break
		}
	}

	// Load environment-specific file if env is specified
	if env != "" {
		envSpecificFile := ".env." + env
		for _, path := range searchPaths {
			fullPath := path + "/" + envSpecificFile
			if _, err := os.Stat(fullPath); err == nil {
				if err := loadEnvFileWithOverride(fullPath); err != nil {
					return fmt.Errorf("failed to load %s: %w", fullPath, err)
				}
				break
			}
		}
	}

	return nil
}

// loadEnvFileWithoutOverride loads a .env file without overriding existing env vars.
// This is the standard behavior for base .env files.
func loadEnvFileWithoutOverride(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = trimQuotes(value)

		// Only set if not already defined
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("failed to set %s: %w", key, err)
			}
		}
	}

	return nil
}

// loadEnvFileWithOverride loads a .env file and overrides existing env vars.
// This is used for environment-specific files that should take precedence.
func loadEnvFileWithOverride(path string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove surrounding quotes if present
		value = trimQuotes(value)

		// Override existing values
		if err := os.Setenv(key, value); err != nil {
			return fmt.Errorf("failed to set %s: %w", key, err)
		}
	}

	return nil
}

// trimQuotes removes surrounding single or double quotes from a string.
func trimQuotes(s string) string {
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
