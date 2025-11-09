package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/JohnPlummer/go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServerConfigFromViper(t *testing.T) {
	t.Run("uses defaults when no config provided", func(t *testing.T) {
		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.ServerConfigFromViper(std)

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, 8080, cfg.Port)
		assert.Equal(t, 15*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 15*time.Second, cfg.WriteTimeout)
		assert.Equal(t, 60*time.Second, cfg.IdleTimeout)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("SERVER_HOST", "0.0.0.0")
		os.Setenv("SERVER_PORT", "9000")
		os.Setenv("SERVER_READ_TIMEOUT", "30s")
		os.Setenv("SERVER_WRITE_TIMEOUT", "30s")
		os.Setenv("SERVER_IDLE_TIMEOUT", "120s")
		defer func() {
			os.Unsetenv("SERVER_HOST")
			os.Unsetenv("SERVER_PORT")
			os.Unsetenv("SERVER_READ_TIMEOUT")
			os.Unsetenv("SERVER_WRITE_TIMEOUT")
			os.Unsetenv("SERVER_IDLE_TIMEOUT")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.ServerConfigFromViper(std)

		assert.Equal(t, "0.0.0.0", cfg.Host)
		assert.Equal(t, 9000, cfg.Port)
		assert.Equal(t, 30*time.Second, cfg.ReadTimeout)
		assert.Equal(t, 30*time.Second, cfg.WriteTimeout)
		assert.Equal(t, 120*time.Second, cfg.IdleTimeout)
	})
}

func TestServerConfig_Validate(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		cfg := config.ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		require.NoError(t, cfg.Validate())
	})

	t.Run("missing host fails", func(t *testing.T) {
		cfg := config.ServerConfig{
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server.host is required")
	})

	t.Run("invalid port fails", func(t *testing.T) {
		cfg := config.ServerConfig{
			Host:         "localhost",
			Port:         0,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server.port must be between")
	})

	t.Run("negative timeout fails", func(t *testing.T) {
		cfg := config.ServerConfig{
			Host:         "localhost",
			Port:         8080,
			ReadTimeout:  -1 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "server.read_timeout must be positive")
	})
}

func TestServerConfig_Address(t *testing.T) {
	cfg := config.ServerConfig{
		Host: "localhost",
		Port: 8080,
	}

	assert.Equal(t, "localhost:8080", cfg.Address())

	cfg.Host = "0.0.0.0"
	cfg.Port = 9000
	assert.Equal(t, "0.0.0.0:9000", cfg.Address())
}
