package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/JohnPlummer/go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseConfigFromViper(t *testing.T) {
	t.Run("uses defaults when no config provided", func(t *testing.T) {
		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.DatabaseConfigFromViper(std)

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, "postgres", cfg.Database)
		assert.Equal(t, "postgres", cfg.User)
		assert.Equal(t, "disable", cfg.SSLMode)
		assert.Equal(t, 25, cfg.MaxConns)
		assert.Equal(t, 5, cfg.MinConns)
		assert.Equal(t, 1*time.Hour, cfg.ConnMaxLifetime)
		assert.Equal(t, 10*time.Minute, cfg.ConnMaxIdleTime)
		assert.Equal(t, 3, cfg.RetryAttempts)
		assert.Equal(t, 2*time.Second, cfg.RetryDelay)
		assert.Equal(t, 30*time.Second, cfg.HealthCheckPeriod)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("DB_HOST", "dbhost")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_NAME", "testdb")
		os.Setenv("DB_USER", "testuser")
		os.Setenv("DB_PASSWORD", "testpass")
		os.Setenv("DB_SSLMODE", "require")
		os.Setenv("DB_MAX_CONNS", "50")
		os.Setenv("DB_MIN_CONNS", "10")
		defer func() {
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_SSLMODE")
			os.Unsetenv("DB_MAX_CONNS")
			os.Unsetenv("DB_MIN_CONNS")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.DatabaseConfigFromViper(std)

		assert.Equal(t, "dbhost", cfg.Host)
		assert.Equal(t, 5433, cfg.Port)
		assert.Equal(t, "testdb", cfg.Database)
		assert.Equal(t, "testuser", cfg.User)
		assert.Equal(t, "testpass", cfg.Password)
		assert.Equal(t, "require", cfg.SSLMode)
		assert.Equal(t, 50, cfg.MaxConns)
		assert.Equal(t, 10, cfg.MinConns)
	})

	t.Run("supports alternative env var names", func(t *testing.T) {
		os.Setenv("DB_DATABASE", "altdb")
		os.Setenv("DB_USERNAME", "altuser")
		os.Setenv("DB_PASS", "altpass")
		defer func() {
			os.Unsetenv("DB_DATABASE")
			os.Unsetenv("DB_USERNAME")
			os.Unsetenv("DB_PASS")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.DatabaseConfigFromViper(std)

		assert.Equal(t, "altdb", cfg.Database)
		assert.Equal(t, "altuser", cfg.User)
		assert.Equal(t, "altpass", cfg.Password)
	})
}

func TestDatabaseConfig_Validate(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:              "localhost",
			Port:              5432,
			Database:          "testdb",
			User:              "testuser",
			Password:          "testpass",
			SSLMode:           "disable",
			MaxConns:          25,
			MinConns:          5,
			ConnMaxLifetime:   1 * time.Hour,
			ConnMaxIdleTime:   10 * time.Minute,
			RetryAttempts:     3,
			RetryDelay:        2 * time.Second,
			HealthCheckPeriod: 30 * time.Second,
		}

		require.NoError(t, cfg.Validate())
	})

	t.Run("missing host fails", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database.host is required")
	})

	t.Run("invalid port fails", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     99999,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "disable",
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database.port must be between")
	})

	t.Run("invalid SSL mode fails", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "invalid",
			MaxConns: 25,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database.ssl_mode must be one of")
	})

	t.Run("min conns greater than max conns fails", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			Database: "testdb",
			User:     "testuser",
			Password: "testpass",
			SSLMode:  "disable",
			MaxConns: 10,
			MinConns: 20,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database.min_conns")
	})

	t.Run("negative retry attempts fails", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:          "localhost",
			Port:          5432,
			Database:      "testdb",
			User:          "testuser",
			Password:      "testpass",
			SSLMode:       "disable",
			MaxConns:      25,
			RetryAttempts: -1,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "database.retry_attempts")
	})
}

func TestDatabaseConfig_ConnectionString(t *testing.T) {
	cfg := config.DatabaseConfig{
		Host:     "localhost",
		Port:     5432,
		Database: "testdb",
		User:     "testuser",
		Password: "testpass",
		SSLMode:  "disable",
	}

	expected := "postgres://testuser:testpass@localhost:5432/testdb?sslmode=disable"
	assert.Equal(t, expected, cfg.ConnectionString())
}
