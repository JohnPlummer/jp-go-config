package config_test

import (
	"os"
	"testing"
	"time"

	config "github.com/JohnPlummer/jp-go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseDatabaseURL(t *testing.T) {
	t.Run("parses valid postgres:// URL", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgres://testuser:testpass@dbhost:5433/testdb?sslmode=require")
		require.NoError(t, err)

		assert.Equal(t, "dbhost", cfg.Host)
		assert.Equal(t, 5433, cfg.Port)
		assert.Equal(t, "testdb", cfg.Database)
		assert.Equal(t, "testuser", cfg.User)
		assert.Equal(t, "testpass", cfg.Password)
		assert.Equal(t, "require", cfg.SSLMode)
	})

	t.Run("parses valid postgresql:// URL", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgresql://pguser:pgpass@pghost:5434/pgdb?sslmode=verify-full")
		require.NoError(t, err)

		assert.Equal(t, "pghost", cfg.Host)
		assert.Equal(t, 5434, cfg.Port)
		assert.Equal(t, "pgdb", cfg.Database)
		assert.Equal(t, "pguser", cfg.User)
		assert.Equal(t, "pgpass", cfg.Password)
		assert.Equal(t, "verify-full", cfg.SSLMode)
	})

	t.Run("handles URL without port (defaults to 5432)", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgres://user:pass@localhost/mydb")
		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, "mydb", cfg.Database)
	})

	t.Run("handles URL without sslmode (defaults to disable)", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgres://user:pass@localhost:5432/mydb")
		require.NoError(t, err)

		assert.Equal(t, "disable", cfg.SSLMode)
	})

	t.Run("handles URL with empty host (defaults to localhost)", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgres://user:pass@/mydb")
		require.NoError(t, err)

		assert.Equal(t, "localhost", cfg.Host)
	})

	t.Run("handles URL without password", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgres://user@localhost:5432/mydb")
		require.NoError(t, err)

		assert.Equal(t, "user", cfg.User)
		assert.Equal(t, "", cfg.Password)
	})

	t.Run("handles URL with special characters in password", func(t *testing.T) {
		cfg, err := config.ParseDatabaseURL("postgres://user:p%40ss%2Fword@localhost:5432/mydb")
		require.NoError(t, err)

		assert.Equal(t, "user", cfg.User)
		assert.Equal(t, "p@ss/word", cfg.Password)
	})

	t.Run("returns error for invalid URL", func(t *testing.T) {
		_, err := config.ParseDatabaseURL("://invalid")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid database URL")
	})

	t.Run("returns error for unsupported scheme (mysql)", func(t *testing.T) {
		_, err := config.ParseDatabaseURL("mysql://user:pass@localhost:3306/mydb")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported database URL scheme")
		assert.Contains(t, err.Error(), "mysql")
	})

	t.Run("returns error for unsupported scheme (http)", func(t *testing.T) {
		_, err := config.ParseDatabaseURL("http://localhost/mydb")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported database URL scheme")
	})

	t.Run("returns error for invalid port", func(t *testing.T) {
		_, err := config.ParseDatabaseURL("postgres://user:pass@localhost:notaport/mydb")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid port")
	})
}

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

	t.Run("aliases no longer work (breaking change v0.4.0)", func(t *testing.T) {
		// Per ADR-003, aliases have been removed in v0.4.0
		// DB_DATABASE, DB_USERNAME, DB_PASS should NOT set the config fields
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

		// Aliases should NOT work - values should be defaults, not the alias values
		assert.Equal(t, "postgres", cfg.Database, "DB_DATABASE alias should not work")
		assert.Equal(t, "postgres", cfg.User, "DB_USERNAME alias should not work")
		assert.Equal(t, "", cfg.Password, "DB_PASS alias should not work")
	})

	t.Run("DATABASE_URL takes precedence over individual vars", func(t *testing.T) {
		// Set both DATABASE_URL and individual vars - DATABASE_URL should win
		os.Setenv("DATABASE_URL", "postgres://urluser:urlpass@urlhost:5555/urldb?sslmode=require")
		os.Setenv("DB_HOST", "ignored-host")
		os.Setenv("DB_PORT", "9999")
		os.Setenv("DB_NAME", "ignored-db")
		os.Setenv("DB_USER", "ignored-user")
		os.Setenv("DB_PASSWORD", "ignored-pass")
		os.Setenv("DB_SSLMODE", "disable")
		defer func() {
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
			os.Unsetenv("DB_SSLMODE")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.DatabaseConfigFromViper(std)

		// Values from DATABASE_URL should be used
		assert.Equal(t, "urlhost", cfg.Host)
		assert.Equal(t, 5555, cfg.Port)
		assert.Equal(t, "urldb", cfg.Database)
		assert.Equal(t, "urluser", cfg.User)
		assert.Equal(t, "urlpass", cfg.Password)
		assert.Equal(t, "require", cfg.SSLMode)
	})

	t.Run("pool settings work when DATABASE_URL is set", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "postgres://user:pass@host:5432/db")
		os.Setenv("DB_MAX_CONNS", "100")
		os.Setenv("DB_MIN_CONNS", "20")
		os.Setenv("DB_RETRY_ATTEMPTS", "5")
		defer func() {
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("DB_MAX_CONNS")
			os.Unsetenv("DB_MIN_CONNS")
			os.Unsetenv("DB_RETRY_ATTEMPTS")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.DatabaseConfigFromViper(std)

		// Connection settings from URL
		assert.Equal(t, "host", cfg.Host)
		assert.Equal(t, 5432, cfg.Port)
		assert.Equal(t, "db", cfg.Database)
		assert.Equal(t, "user", cfg.User)
		assert.Equal(t, "pass", cfg.Password)

		// Pool settings from individual env vars
		assert.Equal(t, 100, cfg.MaxConns)
		assert.Equal(t, 20, cfg.MinConns)
		assert.Equal(t, 5, cfg.RetryAttempts)
	})

	t.Run("falls back to individual vars when DATABASE_URL is empty", func(t *testing.T) {
		os.Setenv("DATABASE_URL", "")
		os.Setenv("DB_HOST", "fallback-host")
		os.Setenv("DB_PORT", "5433")
		os.Setenv("DB_NAME", "fallback-db")
		os.Setenv("DB_USER", "fallback-user")
		os.Setenv("DB_PASSWORD", "fallback-pass")
		defer func() {
			os.Unsetenv("DATABASE_URL")
			os.Unsetenv("DB_HOST")
			os.Unsetenv("DB_PORT")
			os.Unsetenv("DB_NAME")
			os.Unsetenv("DB_USER")
			os.Unsetenv("DB_PASSWORD")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.DatabaseConfigFromViper(std)

		assert.Equal(t, "fallback-host", cfg.Host)
		assert.Equal(t, 5433, cfg.Port)
		assert.Equal(t, "fallback-db", cfg.Database)
		assert.Equal(t, "fallback-user", cfg.User)
		assert.Equal(t, "fallback-pass", cfg.Password)
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
