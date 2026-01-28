// Package config provides enterprise-standard configuration management wrapping Viper.
package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Config holds all application configuration loaded via Load().
// Access typed configs via accessor methods like Database(), Server(), etc.
type Config struct {
	standard *Standard
	logger   *slog.Logger

	// Cached typed configs (populated on first access)
	database   *DatabaseConfig
	server     *ServerConfig
	resilience *ResilienceConfig
	openai     *OpenAIConfig
}

// LoadOption configures the Load function using the functional options pattern.
type LoadOption func(*loadOptions)

// loadOptions holds the internal options for Load().
type loadOptions struct {
	envPath    string
	configPath string
	envPrefix  string
	logger     *slog.Logger
}

// WithEnvPath specifies an explicit .env file path, disabling auto-discovery.
func WithEnvPath(path string) LoadOption {
	return func(o *loadOptions) {
		o.envPath = path
	}
}

// WithConfigPath specifies an explicit config file path, disabling auto-discovery.
func WithConfigPath(path string) LoadOption {
	return func(o *loadOptions) {
		o.configPath = path
	}
}

// WithPrefix sets the environment variable prefix (default: APP).
func WithPrefix(prefix string) LoadOption {
	return func(o *loadOptions) {
		o.envPrefix = prefix
	}
}

// WithLogger sets a custom slog logger for discovery events.
func WithLogger(logger *slog.Logger) LoadOption {
	return func(o *loadOptions) {
		o.logger = logger
	}
}

// Load creates a fully configured Config by auto-discovering and loading
// .env and config.yaml files from the current directory or project root.
//
// Discovery order (first found wins):
//  1. Current working directory
//  2. Go module root (directory containing go.mod)
//  3. Git repository root (directory containing .git)
//
// This supports monorepos and git worktrees where config files are at the
// repository root but the Go module is in a subdirectory.
//
// Files loaded:
//   - .env: Environment variables (does not override existing)
//   - config.yaml: Configuration values
//
// Use WithEnvPath/WithConfigPath to override auto-discovery.
//
// Example:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	db := cfg.Database()
func Load(opts ...LoadOption) (*Config, error) {
	// Apply default options
	options := &loadOptions{
		envPrefix: "APP",
		logger:    slog.Default(),
	}
	for _, opt := range opts {
		opt(options)
	}

	// Discover and load .env file
	envPath, err := discoverEnvFile(options)
	if err != nil {
		return nil, fmt.Errorf("env file discovery failed: %w", err)
	}
	if envPath != "" {
		if err := loadEnvFileWithoutOverride(envPath); err != nil {
			return nil, fmt.Errorf("failed to load .env file %s: %w", envPath, err)
		}
		options.logger.Info("loaded env file", "path", envPath)
	}

	// Create Standard config with Viper
	v := viper.New()
	v.SetEnvPrefix(options.envPrefix)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Discover and load config file
	configPath, err := discoverConfigFile(options)
	if err != nil {
		return nil, fmt.Errorf("config file discovery failed: %w", err)
	}
	if configPath != "" {
		v.SetConfigFile(configPath)
		if err := v.ReadInConfig(); err != nil {
			return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
		}
		options.logger.Info("loaded config file", "path", configPath)
	}

	standard := &Standard{viper: v}

	return &Config{
		standard: standard,
		logger:   options.logger,
	}, nil
}

// discoverEnvFile finds the .env file path based on options or auto-discovery.
func discoverEnvFile(options *loadOptions) (string, error) {
	// Use explicit path if provided
	if options.envPath != "" {
		if _, err := os.Stat(options.envPath); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf(".env file not found: %s", options.envPath)
			}
			return "", err
		}
		return options.envPath, nil
	}

	// Auto-discover: check cwd, go.mod root, then .git root
	// This order supports monorepos and git worktrees where .env is at repo root
	// but go.mod is in a subdirectory
	searchPaths := collectSearchPaths()

	// Search for .env file
	for _, basePath := range searchPaths {
		envPath := basePath + "/.env"
		if _, err := os.Stat(envPath); err == nil {
			return envPath, nil
		}
	}

	// No .env file found - this is OK
	return "", nil
}

// discoverConfigFile finds the config.yaml file path based on options or auto-discovery.
func discoverConfigFile(options *loadOptions) (string, error) {
	// Use explicit path if provided
	if options.configPath != "" {
		if _, err := os.Stat(options.configPath); err != nil {
			if os.IsNotExist(err) {
				return "", fmt.Errorf("config file not found: %s", options.configPath)
			}
			return "", err
		}
		return options.configPath, nil
	}

	// Auto-discover: check cwd, go.mod root, then .git root
	// This order supports monorepos and git worktrees where config is at repo root
	// but go.mod is in a subdirectory
	searchPaths := collectSearchPaths()

	// Search for config files (yaml preferred, then yml)
	configNames := []string{"config.yaml", "config.yml"}
	for _, basePath := range searchPaths {
		for _, configName := range configNames {
			configPath := basePath + "/" + configName
			if _, err := os.Stat(configPath); err == nil {
				return configPath, nil
			}
		}
	}

	// No config file found - this is OK
	return "", nil
}

// collectSearchPaths returns deduplicated search paths in priority order:
// 1. Current working directory
// 2. Go module root (directory containing go.mod)
// 3. Git repository root (directory containing .git)
//
// This supports monorepos and git worktrees where config files are at the
// repository root but go.mod is in a subdirectory.
func collectSearchPaths() []string {
	cwd, _ := os.Getwd()

	// Start with cwd
	searchPaths := []string{"."}
	seen := map[string]bool{cwd: true}

	// Add go.mod root if different from cwd
	goModRoot := FindProjectRoot("go.mod")
	if goModRoot != "" && !seen[goModRoot] {
		searchPaths = append(searchPaths, goModRoot)
		seen[goModRoot] = true
	}

	// Add .git root if different from both cwd and go.mod root
	// This is the key fix: in monorepos/worktrees, .git root may be
	// a parent of go.mod root, and that's where .env typically lives
	gitRoot := FindProjectRoot(".git")
	if gitRoot != "" && !seen[gitRoot] {
		searchPaths = append(searchPaths, gitRoot)
	}

	return searchPaths
}

// Database returns the database configuration.
// Validates on first access and caches the result.
func (c *Config) Database() (*DatabaseConfig, error) {
	if c.database != nil {
		return c.database, nil
	}

	cfg := DatabaseConfigFromViper(c.standard)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("database config validation failed: %w", err)
	}

	c.database = &cfg
	return c.database, nil
}

// Server returns the server configuration.
// Validates on first access and caches the result.
func (c *Config) Server() (*ServerConfig, error) {
	if c.server != nil {
		return c.server, nil
	}

	cfg := ServerConfigFromViper(c.standard)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("server config validation failed: %w", err)
	}

	c.server = &cfg
	return c.server, nil
}

// Resilience returns the resilience configuration.
// Validates on first access and caches the result.
func (c *Config) Resilience() (*ResilienceConfig, error) {
	if c.resilience != nil {
		return c.resilience, nil
	}

	cfg := ResilienceConfigFromViper(c.standard)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("resilience config validation failed: %w", err)
	}

	c.resilience = &cfg
	return c.resilience, nil
}

// OpenAI returns the OpenAI configuration.
// Validates on first access and caches the result.
func (c *Config) OpenAI() (*OpenAIConfig, error) {
	if c.openai != nil {
		return c.openai, nil
	}

	cfg := OpenAIConfigFromViper(c.standard)
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("openai config validation failed: %w", err)
	}

	c.openai = &cfg
	return c.openai, nil
}

// Standard returns the underlying Standard config for advanced usage.
func (c *Config) Standard() *Standard {
	return c.standard
}

// ValidateAll validates all configuration types and returns the first error.
// Validation order: database, server, openai, resilience.
// Use this method for fail-fast behavior when all configs are required.
//
// Example:
//
//	cfg, err := config.Load()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	if err := cfg.ValidateAll(); err != nil {
//	    log.Fatal(err)
//	}
func (c *Config) ValidateAll() error {
	if _, err := c.Database(); err != nil {
		return err
	}
	if _, err := c.Server(); err != nil {
		return err
	}
	if _, err := c.OpenAI(); err != nil {
		return err
	}
	if _, err := c.Resilience(); err != nil {
		return err
	}
	return nil
}
