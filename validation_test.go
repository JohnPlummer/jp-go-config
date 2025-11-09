package config_test

import (
	"testing"
	"time"

	"github.com/JohnPlummer/go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateRequired(t *testing.T) {
	t.Run("empty string fails", func(t *testing.T) {
		err := config.ValidateRequired("test.field", "")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.field is required")
	})

	t.Run("non-empty string passes", func(t *testing.T) {
		err := config.ValidateRequired("test.field", "value")
		require.NoError(t, err)
	})
}

func TestValidatePort(t *testing.T) {
	t.Run("valid port passes", func(t *testing.T) {
		err := config.ValidatePort("test.port", 8080)
		require.NoError(t, err)
	})

	t.Run("port 1 passes", func(t *testing.T) {
		err := config.ValidatePort("test.port", 1)
		require.NoError(t, err)
	})

	t.Run("port 65535 passes", func(t *testing.T) {
		err := config.ValidatePort("test.port", 65535)
		require.NoError(t, err)
	})

	t.Run("port 0 fails", func(t *testing.T) {
		err := config.ValidatePort("test.port", 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.port must be between 1 and 65535")
	})

	t.Run("port >65535 fails", func(t *testing.T) {
		err := config.ValidatePort("test.port", 70000)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.port must be between 1 and 65535")
	})

	t.Run("negative port fails", func(t *testing.T) {
		err := config.ValidatePort("test.port", -1)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.port must be between 1 and 65535")
	})
}

func TestValidateDuration(t *testing.T) {
	t.Run("positive duration passes", func(t *testing.T) {
		err := config.ValidateDuration("test.duration", 5*time.Second)
		require.NoError(t, err)
	})

	t.Run("zero duration passes", func(t *testing.T) {
		err := config.ValidateDuration("test.duration", 0)
		require.NoError(t, err)
	})

	t.Run("negative duration fails", func(t *testing.T) {
		err := config.ValidateDuration("test.duration", -1*time.Second)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.duration must be positive")
	})
}

func TestValidatePositive(t *testing.T) {
	t.Run("positive integer passes", func(t *testing.T) {
		err := config.ValidatePositive("test.value", 10)
		require.NoError(t, err)
	})

	t.Run("1 passes", func(t *testing.T) {
		err := config.ValidatePositive("test.value", 1)
		require.NoError(t, err)
	})

	t.Run("zero fails", func(t *testing.T) {
		err := config.ValidatePositive("test.value", 0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.value must be positive")
	})

	t.Run("negative fails", func(t *testing.T) {
		err := config.ValidatePositive("test.value", -5)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.value must be positive")
	})
}

func TestValidateRange(t *testing.T) {
	t.Run("value in range passes (int)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 5, 0, 10)
		require.NoError(t, err)
	})

	t.Run("value at min passes (int)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 0, 0, 10)
		require.NoError(t, err)
	})

	t.Run("value at max passes (int)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 10, 0, 10)
		require.NoError(t, err)
	})

	t.Run("value below min fails (int)", func(t *testing.T) {
		err := config.ValidateRange("test.value", -1, 0, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.value must be between 0 and 10")
	})

	t.Run("value above max fails (int)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 11, 0, 10)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.value must be between 0 and 10")
	})

	t.Run("value in range passes (float64)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 0.5, 0.0, 1.0)
		require.NoError(t, err)
	})

	t.Run("value at min passes (float64)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 0.0, 0.0, 1.0)
		require.NoError(t, err)
	})

	t.Run("value at max passes (float64)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 1.0, 0.0, 1.0)
		require.NoError(t, err)
	})

	t.Run("value below min fails (float64)", func(t *testing.T) {
		err := config.ValidateRange("test.value", -0.1, 0.0, 1.0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.value must be between 0 and 1")
	})

	t.Run("value above max fails (float64)", func(t *testing.T) {
		err := config.ValidateRange("test.value", 1.1, 0.0, 1.0)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "test.value must be between 0 and 1")
	})
}
