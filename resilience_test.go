package config_test

import (
	"os"
	"testing"
	"time"

	config "github.com/JohnPlummer/jp-go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResilienceConfigFromViper(t *testing.T) {
	t.Run("uses defaults when no config provided", func(t *testing.T) {
		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.ResilienceConfigFromViper(std)

		// Retry defaults
		assert.Equal(t, 3, cfg.MaxRetries)
		assert.Equal(t, 1*time.Second, cfg.InitialDelay)
		assert.Equal(t, 30*time.Second, cfg.MaxDelay)
		assert.Equal(t, 2.0, cfg.Multiplier)

		// Circuit breaker defaults
		assert.Equal(t, uint32(10), cfg.MaxRequests)
		assert.Equal(t, 10*time.Second, cfg.Interval)
		assert.Equal(t, 60*time.Second, cfg.Timeout)
		assert.Equal(t, 0.6, cfg.FailureThreshold)
	})

	t.Run("loads from environment variables", func(t *testing.T) {
		os.Setenv("RESILIENCE_MAX_RETRIES", "5")
		os.Setenv("RESILIENCE_INITIAL_DELAY", "2s")
		os.Setenv("RESILIENCE_MAX_DELAY", "60s")
		os.Setenv("RESILIENCE_MULTIPLIER", "3.0")
		os.Setenv("RESILIENCE_MAX_REQUESTS", "20")
		os.Setenv("RESILIENCE_INTERVAL", "30s")
		os.Setenv("RESILIENCE_TIMEOUT", "120s")
		os.Setenv("RESILIENCE_FAILURE_THRESHOLD", "0.7")
		defer func() {
			os.Unsetenv("RESILIENCE_MAX_RETRIES")
			os.Unsetenv("RESILIENCE_INITIAL_DELAY")
			os.Unsetenv("RESILIENCE_MAX_DELAY")
			os.Unsetenv("RESILIENCE_MULTIPLIER")
			os.Unsetenv("RESILIENCE_MAX_REQUESTS")
			os.Unsetenv("RESILIENCE_INTERVAL")
			os.Unsetenv("RESILIENCE_TIMEOUT")
			os.Unsetenv("RESILIENCE_FAILURE_THRESHOLD")
		}()

		std, err := config.NewStandard()
		require.NoError(t, err)

		cfg := config.ResilienceConfigFromViper(std)

		// Retry settings
		assert.Equal(t, 5, cfg.MaxRetries)
		assert.Equal(t, 2*time.Second, cfg.InitialDelay)
		assert.Equal(t, 60*time.Second, cfg.MaxDelay)
		assert.Equal(t, 3.0, cfg.Multiplier)

		// Circuit breaker settings
		assert.Equal(t, uint32(20), cfg.MaxRequests)
		assert.Equal(t, 30*time.Second, cfg.Interval)
		assert.Equal(t, 120*time.Second, cfg.Timeout)
		assert.Equal(t, 0.7, cfg.FailureThreshold)
	})
}

func TestResilienceConfig_Validate(t *testing.T) {
	t.Run("valid config passes", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		require.NoError(t, cfg.Validate())
	})

	t.Run("max retries out of range fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       15,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.max_retries must be between")
	})

	t.Run("negative initial delay fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     -1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.initial_delay must be positive")
	})

	t.Run("max delay less than initial delay fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     30 * time.Second,
			MaxDelay:         1 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.max_delay")
		assert.Contains(t, err.Error(), "must be greater than or equal to initial_delay")
	})

	t.Run("multiplier out of range fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       15.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.multiplier must be between")
	})

	t.Run("zero max requests fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      0,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.max_requests must be positive")
	})

	t.Run("negative interval fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         -10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.interval must be positive")
	})

	t.Run("negative timeout fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          -60 * time.Second,
			FailureThreshold: 0.6,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.timeout must be positive")
	})

	t.Run("failure threshold out of range fails", func(t *testing.T) {
		cfg := config.ResilienceConfig{
			MaxRetries:       3,
			InitialDelay:     1 * time.Second,
			MaxDelay:         30 * time.Second,
			Multiplier:       2.0,
			MaxRequests:      10,
			Interval:         10 * time.Second,
			Timeout:          60 * time.Second,
			FailureThreshold: 1.5,
		}

		err := cfg.Validate()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "resilience.failure_threshold must be between")
	})
}
