// Package config provides enterprise-standard configuration management wrapping Viper.
//
// This package offers typed configuration with automatic .env file loading,
// environment variable precedence, and comprehensive validation.
package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

// Standard wraps Viper to provide enterprise-standard configuration loading
// with automatic .env file support, environment variable precedence, and validation.
type Standard struct {
	viper *viper.Viper
}

// Option configures the Standard config loader using the functional options pattern.
type Option func(*Standard) error

// WithEnvPrefix sets the environment variable prefix (default: APP_)
func WithEnvPrefix(prefix string) Option {
	return func(s *Standard) error {
		s.viper.SetEnvPrefix(prefix)
		return nil
	}
}

// WithConfigFile specifies a config file to load (YAML, JSON, TOML, etc.)
func WithConfigFile(path string) Option {
	return func(s *Standard) error {
		s.viper.SetConfigFile(path)
		if err := s.viper.ReadInConfig(); err != nil {
			return fmt.Errorf("failed to read config file %s: %w", path, err)
		}
		return nil
	}
}

// WithConfigName sets the name of the config file to search for (without extension)
func WithConfigName(name string) Option {
	return func(s *Standard) error {
		s.viper.SetConfigName(name)
		return nil
	}
}

// WithConfigType sets the type of the config file (yaml, json, toml, etc.)
func WithConfigType(configType string) Option {
	return func(s *Standard) error {
		s.viper.SetConfigType(configType)
		return nil
	}
}

// WithConfigPaths adds paths to search for the config file
func WithConfigPaths(paths ...string) Option {
	return func(s *Standard) error {
		for _, path := range paths {
			s.viper.AddConfigPath(path)
		}
		return nil
	}
}

// WithEnvFile loads environment variables from a specific .env file
func WithEnvFile(path string) Option {
	return func(s *Standard) error {
		if err := godotenv.Load(path); err != nil {
			return fmt.Errorf("failed to load .env file %s: %w", path, err)
		}
		return nil
	}
}

// WithoutEnvFile disables automatic .env file loading
func WithoutEnvFile() Option {
	return func(s *Standard) error {
		// Marker option - actual behavior is in NewStandard
		return nil
	}
}

// NewStandard creates a new Standard config loader with the given options.
//
// By default:
// - Loads .env files from current directory (silently ignored if missing)
// - Reads environment variables with APP_ prefix
// - Replaces dots and hyphens with underscores in env var names
//
// Options can override any of these defaults.
func NewStandard(options ...Option) (*Standard, error) {
	s := &Standard{
		viper: viper.New(),
	}

	// Set defaults
	s.viper.SetEnvPrefix("APP")
	s.viper.AutomaticEnv()
	s.viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))

	// Load .env file by default (silently ignore if missing)
	// Can be disabled with WithoutEnvFile()
	loadEnv := true
	for _, opt := range options {
		// Check if WithoutEnvFile is in options
		if err := opt(s); err != nil {
			// Special handling for WithoutEnvFile marker
			if err.Error() == "skip env file" {
				loadEnv = false
				continue
			}
			return nil, fmt.Errorf("failed to apply option: %w", err)
		}
	}

	if loadEnv {
		// Try to load .env file, but don't fail if it doesn't exist
		_ = godotenv.Load()
	}

	return s, nil
}

// Get retrieves a value by key
func (s *Standard) Get(key string) interface{} {
	return s.viper.Get(key)
}

// GetString retrieves a string value
func (s *Standard) GetString(key string) string {
	return s.viper.GetString(key)
}

// GetInt retrieves an integer value
func (s *Standard) GetInt(key string) int {
	return s.viper.GetInt(key)
}

// GetBool retrieves a boolean value
func (s *Standard) GetBool(key string) bool {
	return s.viper.GetBool(key)
}

// GetDuration retrieves a duration value
func (s *Standard) GetDuration(key string) interface{} {
	return s.viper.GetDuration(key)
}

// Set sets a value for a key
func (s *Standard) Set(key string, value interface{}) {
	s.viper.Set(key, value)
}

// BindEnv binds a config key to environment variables.
// With no envVars argument, it uses the key as the env var name.
// With one or more envVars, it checks each in order until finding a set value.
func (s *Standard) BindEnv(key string, envVars ...string) error {
	args := make([]string, len(envVars)+1)
	args[0] = key
	copy(args[1:], envVars)
	return s.viper.BindEnv(args...)
}

// Unmarshal unmarshals the config into a struct
func (s *Standard) Unmarshal(rawVal interface{}) error {
	if err := s.viper.Unmarshal(rawVal); err != nil {
		return fmt.Errorf("failed to unmarshal config: %w", err)
	}
	return nil
}

// AllKeys returns all keys in the config
func (s *Standard) AllKeys() []string {
	return s.viper.AllKeys()
}

// IsSet checks if a key is set in the config
func (s *Standard) IsSet(key string) bool {
	return s.viper.IsSet(key)
}

// Viper returns the underlying Viper instance for advanced usage
func (s *Standard) Viper() *viper.Viper {
	return s.viper
}

// LoadEnvFile loads environment variables from a .env file.
// Does not override existing environment variables.
// Silently succeeds if the file doesn't exist.
func LoadEnvFile(paths ...string) error {
	if len(paths) == 0 {
		paths = []string{".env"}
	}

	for _, path := range paths {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				return fmt.Errorf("failed to load .env file %s: %w", path, err)
			}
			return nil
		}
	}

	// No .env file found - this is OK
	return nil
}
