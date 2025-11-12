package config_test

import (
	"os"
	"testing"
	"time"

	config "github.com/JohnPlummer/jp-go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenAIConfigFromViper(t *testing.T) {
	t.Run("uses defaults when no config provided", func(t *testing.T) {
		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.OpenAIConfigFromViper(std)

		assert.Equal(t, "gpt-3.5-turbo", cfg.Model)
		assert.Equal(t, 0.7, cfg.Temperature)
		assert.Equal(t, 2000, cfg.MaxTokens)
		assert.Equal(t, 30*time.Second, cfg.Timeout)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("OPENAI_API_KEY", "sk-test123")
		os.Setenv("OPENAI_MODEL", "gpt-4")
		os.Setenv("OPENAI_TEMPERATURE", "0.5")
		os.Setenv("OPENAI_MAX_TOKENS", "4000")
		os.Setenv("OPENAI_TIMEOUT", "60s")
		defer func() {
			os.Unsetenv("OPENAI_API_KEY")
			os.Unsetenv("OPENAI_MODEL")
			os.Unsetenv("OPENAI_TEMPERATURE")
			os.Unsetenv("OPENAI_MAX_TOKENS")
			os.Unsetenv("OPENAI_TIMEOUT")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.OpenAIConfigFromViper(std)

		assert.Equal(t, "sk-test123", cfg.APIKey)
		assert.Equal(t, "gpt-4", cfg.Model)
		assert.Equal(t, 0.5, cfg.Temperature)
		assert.Equal(t, 4000, cfg.MaxTokens)
		assert.Equal(t, 60*time.Second, cfg.Timeout)
	})
}

func TestOpenAIConfig_Validate(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		cfg := config.OpenAIConfig{
			APIKey:      "sk-test123",
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   2000,
			Timeout:     30 * time.Second,
		}

		require.NoError(t, cfg.Validate())
	})

	t.Run("missing API key fails", func(t *testing.T) {
		cfg := config.OpenAIConfig{
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   2000,
			Timeout:     30 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai.api_key is required")
	})

	t.Run("missing model fails", func(t *testing.T) {
		cfg := config.OpenAIConfig{
			APIKey:      "sk-test123",
			Temperature: 0.7,
			MaxTokens:   2000,
			Timeout:     30 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai.model is required")
	})

	t.Run("temperature out of range fails", func(t *testing.T) {
		cfg := config.OpenAIConfig{
			APIKey:      "sk-test123",
			Model:       "gpt-3.5-turbo",
			Temperature: 3.0,
			MaxTokens:   2000,
			Timeout:     30 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai.temperature must be between")
	})

	t.Run("negative max tokens fails", func(t *testing.T) {
		cfg := config.OpenAIConfig{
			APIKey:      "sk-test123",
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   0,
			Timeout:     30 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai.max_tokens must be positive")
	})

	t.Run("negative timeout fails", func(t *testing.T) {
		cfg := config.OpenAIConfig{
			APIKey:      "sk-test123",
			Model:       "gpt-3.5-turbo",
			Temperature: 0.7,
			MaxTokens:   2000,
			Timeout:     -1 * time.Second,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "openai.timeout must be positive")
	})
}
