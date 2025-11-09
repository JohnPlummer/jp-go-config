package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JohnPlummer/go-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewStandard(t *testing.T) {
	t.Run("creates with defaults", func(t *testing.T) {
		std, err := config.NewStandard()
		require.NoError(t, err)
		require.NotNil(t, std)
	})

	t.Run("loads .env file by default", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer func() {
			_ = os.Chdir(oldWd)
		}()
		require.NoError(t, os.Chdir(tmpDir))

		envContent := "TEST_VAR=from_env_file"
		require.NoError(t, os.WriteFile(".env", []byte(envContent), 0o644))

		_, err := config.NewStandard()
		require.NoError(t, err)

		// Verify env var was loaded
		assert.Equal(t, "from_env_file", os.Getenv("TEST_VAR"))
		t.Cleanup(func() {
			os.Unsetenv("TEST_VAR")
		})
	})

	t.Run("works without .env file", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer func() {
			_ = os.Chdir(oldWd)
		}()
		require.NoError(t, os.Chdir(tmpDir))

		std, err := config.NewStandard()
		require.NoError(t, err)
		require.NotNil(t, std)
	})
}

func TestStandard_WithEnvPrefix(t *testing.T) {
	os.Setenv("CUSTOM_KEY", "value")
	defer os.Unsetenv("CUSTOM_KEY")

	std, err := config.NewStandard(config.WithEnvPrefix("CUSTOM"))
	require.NoError(t, err)

	// Need to bind the env var for it to be picked up
	require.NoError(t, std.BindEnv("key", "CUSTOM_KEY"))
	assert.Equal(t, "value", std.GetString("key"))
}

func TestStandard_WithConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "config.yaml")
	configContent := `
test:
  value: from_file
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0o644))

	std, err := config.NewStandard(config.WithConfigFile(configFile))
	require.NoError(t, err)

	assert.Equal(t, "from_file", std.GetString("test.value"))
}

func TestStandard_WithConfigFile_Error(t *testing.T) {
	_, err := config.NewStandard(config.WithConfigFile("/nonexistent/config.yaml"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read config file")
}

func TestStandard_WithConfigName(t *testing.T) {
	tmpDir := t.TempDir()
	configFile := filepath.Join(tmpDir, "myconfig.yaml")
	configContent := `
test:
  value: from_named_file
`
	require.NoError(t, os.WriteFile(configFile, []byte(configContent), 0o644))

	std, err := config.NewStandard(
		config.WithConfigName("myconfig"),
		config.WithConfigType("yaml"),
		config.WithConfigPaths(tmpDir),
	)
	require.NoError(t, err)

	// Read the config
	require.NoError(t, std.Viper().ReadInConfig())
	assert.Equal(t, "from_named_file", std.GetString("test.value"))
}

func TestStandard_WithEnvFile(t *testing.T) {
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "custom.env")
	envContent := "CUSTOM_ENV=from_custom_env"
	require.NoError(t, os.WriteFile(envFile, []byte(envContent), 0o644))
	defer os.Unsetenv("CUSTOM_ENV")

	std, err := config.NewStandard(config.WithEnvFile(envFile))
	require.NoError(t, err)
	require.NotNil(t, std)

	assert.Equal(t, "from_custom_env", os.Getenv("CUSTOM_ENV"))
}

func TestStandard_WithEnvFile_Error(t *testing.T) {
	_, err := config.NewStandard(config.WithEnvFile("/nonexistent/.env"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load .env file")
}

func TestStandard_GetMethods(t *testing.T) {
	std, err := config.NewStandard()
	require.NoError(t, err)

	std.Set("string_key", "value")
	std.Set("int_key", 42)
	std.Set("bool_key", true)
	std.Set("duration_key", 5*time.Second)

	assert.Equal(t, "value", std.GetString("string_key"))
	assert.Equal(t, 42, std.GetInt("int_key"))
	assert.Equal(t, true, std.GetBool("bool_key"))
	assert.Equal(t, 5*time.Second, std.GetDuration("duration_key"))
}

func TestStandard_BindEnv(t *testing.T) {
	os.Setenv("TEST_BIND_VAR", "bound_value")
	defer os.Unsetenv("TEST_BIND_VAR")

	std, err := config.NewStandard()
	require.NoError(t, err)

	require.NoError(t, std.BindEnv("test.key", "TEST_BIND_VAR"))
	assert.Equal(t, "bound_value", std.GetString("test.key"))
}

func TestStandard_Unmarshal(t *testing.T) {
	std, err := config.NewStandard()
	require.NoError(t, err)

	std.Set("name", "test")
	std.Set("count", 5)

	type TestConfig struct {
		Name  string `mapstructure:"name"`
		Count int    `mapstructure:"count"`
	}

	var cfg TestConfig
	require.NoError(t, std.Unmarshal(&cfg))

	assert.Equal(t, "test", cfg.Name)
	assert.Equal(t, 5, cfg.Count)
}

func TestStandard_AllKeys(t *testing.T) {
	std, err := config.NewStandard()
	require.NoError(t, err)

	std.Set("key1", "value1")
	std.Set("key2", "value2")

	keys := std.AllKeys()
	assert.Contains(t, keys, "key1")
	assert.Contains(t, keys, "key2")
}

func TestStandard_IsSet(t *testing.T) {
	std, err := config.NewStandard()
	require.NoError(t, err)

	std.Set("existing_key", "value")

	assert.True(t, std.IsSet("existing_key"))
	assert.False(t, std.IsSet("nonexistent_key"))
}

func TestStandard_Viper(t *testing.T) {
	std, err := config.NewStandard()
	require.NoError(t, err)

	viper := std.Viper()
	require.NotNil(t, viper)

	viper.Set("test", "value")
	assert.Equal(t, "value", std.GetString("test"))
}

func TestLoadEnvFile(t *testing.T) {
	t.Run("loads existing file", func(t *testing.T) {
		tmpDir := t.TempDir()
		envFile := filepath.Join(tmpDir, "test.env")
		envContent := "LOAD_TEST_VAR=loaded"
		require.NoError(t, os.WriteFile(envFile, []byte(envContent), 0o644))
		defer os.Unsetenv("LOAD_TEST_VAR")

		require.NoError(t, config.LoadEnvFile(envFile))
		assert.Equal(t, "loaded", os.Getenv("LOAD_TEST_VAR"))
	})

	t.Run("succeeds with nonexistent file", func(t *testing.T) {
		require.NoError(t, config.LoadEnvFile("/nonexistent/.env"))
	})

	t.Run("uses default .env path", func(t *testing.T) {
		tmpDir := t.TempDir()
		oldWd, _ := os.Getwd()
		defer func() {
			_ = os.Chdir(oldWd)
		}()
		require.NoError(t, os.Chdir(tmpDir))

		envContent := "DEFAULT_ENV_VAR=default"
		require.NoError(t, os.WriteFile(".env", []byte(envContent), 0o644))
		defer os.Unsetenv("DEFAULT_ENV_VAR")

		require.NoError(t, config.LoadEnvFile())
		assert.Equal(t, "default", os.Getenv("DEFAULT_ENV_VAR"))
	})
}

func TestStandard_EnvironmentPrecedence(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, _ := os.Getwd()
	defer func() {
		_ = os.Chdir(oldWd)
	}()
	require.NoError(t, os.Chdir(tmpDir))

	// Create .env file
	envContent := "TEST_PRECEDENCE=from_file"
	require.NoError(t, os.WriteFile(".env", []byte(envContent), 0o644))

	// Set environment variable (should take precedence)
	os.Setenv("APP_TEST_PRECEDENCE", "from_env")
	defer os.Unsetenv("APP_TEST_PRECEDENCE")
	defer os.Unsetenv("TEST_PRECEDENCE")

	std, err := config.NewStandard()
	require.NoError(t, err)

	require.NoError(t, std.BindEnv("test_precedence", "TEST_PRECEDENCE"))

	// Environment variable should take precedence over .env file
	assert.Equal(t, "from_env", std.GetString("test_precedence"))
}
