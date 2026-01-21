package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NoFiles(t *testing.T) {
	// Load without any config files should succeed
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil Config")
	}
	if cfg.Standard() == nil {
		t.Fatal("Config.Standard() returned nil")
	}
}

func TestLoad_WithEnvFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create .env file
	envContent := `DB_HOST=testhost
DB_PORT=5433
DB_PASSWORD=testpass
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Save original env vars and restore after test
	origHost := os.Getenv("DB_HOST")
	origPort := os.Getenv("DB_PORT")
	origPassword := os.Getenv("DB_PASSWORD")
	t.Cleanup(func() {
		if origHost != "" {
			os.Setenv("DB_HOST", origHost)
		} else {
			os.Unsetenv("DB_HOST")
		}
		if origPort != "" {
			os.Setenv("DB_PORT", origPort)
		} else {
			os.Unsetenv("DB_PORT")
		}
		if origPassword != "" {
			os.Setenv("DB_PASSWORD", origPassword)
		} else {
			os.Unsetenv("DB_PASSWORD")
		}
	})

	// Clear any existing env vars
	os.Unsetenv("DB_HOST")
	os.Unsetenv("DB_PORT")
	os.Unsetenv("DB_PASSWORD")

	cfg, err := Load(WithEnvPath(envPath))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Check that env vars were loaded
	if got := os.Getenv("DB_HOST"); got != "testhost" {
		t.Errorf("DB_HOST = %q, want %q", got, "testhost")
	}
	if got := os.Getenv("DB_PORT"); got != "5433" {
		t.Errorf("DB_PORT = %q, want %q", got, "5433")
	}

	// Verify we can get database config
	db, err := cfg.Database()
	if err != nil {
		t.Fatalf("Database() error = %v", err)
	}
	if db.Host != "testhost" {
		t.Errorf("Database().Host = %q, want %q", db.Host, "testhost")
	}
	if db.Port != 5433 {
		t.Errorf("Database().Port = %d, want %d", db.Port, 5433)
	}
}

func TestLoad_WithConfigFile(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create config.yaml file
	configContent := `server:
  host: confighost
  port: 9090
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config.yaml file: %v", err)
	}

	cfg, err := Load(WithConfigPath(configPath))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Verify we can get server config
	srv, err := cfg.Server()
	if err != nil {
		t.Fatalf("Server() error = %v", err)
	}
	if srv.Host != "confighost" {
		t.Errorf("Server().Host = %q, want %q", srv.Host, "confighost")
	}
	if srv.Port != 9090 {
		t.Errorf("Server().Port = %d, want %d", srv.Port, 9090)
	}
}

func TestLoad_WithEnvPath_NotFound(t *testing.T) {
	_, err := Load(WithEnvPath("/nonexistent/.env"))
	if err == nil {
		t.Fatal("Load() error = nil, want error for nonexistent .env")
	}
}

func TestLoad_WithConfigPath_NotFound(t *testing.T) {
	_, err := Load(WithConfigPath("/nonexistent/config.yaml"))
	if err == nil {
		t.Fatal("Load() error = nil, want error for nonexistent config")
	}
}

func TestLoad_WithPrefix(t *testing.T) {
	// Set env var with custom prefix
	os.Setenv("MYAPP_SERVER_HOST", "prefixhost")
	defer os.Unsetenv("MYAPP_SERVER_HOST")

	cfg, err := Load(WithPrefix("MYAPP"))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	srv, err := cfg.Server()
	if err != nil {
		t.Fatalf("Server() error = %v", err)
	}
	if srv.Host != "prefixhost" {
		t.Errorf("Server().Host = %q, want %q", srv.Host, "prefixhost")
	}
}

func TestConfig_AccessorCaching(t *testing.T) {
	// Set required password to avoid validation error
	os.Setenv("DB_PASSWORD", "testpassword")
	defer os.Unsetenv("DB_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Call Database twice, should return same pointer
	db1, err := cfg.Database()
	if err != nil {
		t.Fatalf("Database() error = %v", err)
	}

	db2, err := cfg.Database()
	if err != nil {
		t.Fatalf("Database() error = %v", err)
	}

	if db1 != db2 {
		t.Error("Database() did not return cached instance")
	}
}

func TestConfig_Database_ValidationError(t *testing.T) {
	// Clear password to trigger validation error
	os.Unsetenv("DB_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = cfg.Database()
	if err == nil {
		t.Fatal("Database() error = nil, want validation error")
	}
}

func TestConfig_Server_Success(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	srv, err := cfg.Server()
	if err != nil {
		t.Fatalf("Server() error = %v", err)
	}

	// Should have defaults
	if srv.Port == 0 {
		t.Error("Server().Port = 0, want default value")
	}
}

func TestConfig_Resilience_Success(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	res, err := cfg.Resilience()
	if err != nil {
		t.Fatalf("Resilience() error = %v", err)
	}

	// Should have defaults
	if res.MaxRetries == 0 {
		t.Error("Resilience().MaxRetries = 0, want default value")
	}
}

func TestConfig_OpenAI_ValidationError(t *testing.T) {
	// Clear API key to trigger validation error
	os.Unsetenv("OPENAI_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	_, err = cfg.OpenAI()
	if err == nil {
		t.Fatal("OpenAI() error = nil, want validation error for missing API key")
	}
}

func TestConfig_OpenAI_Success(t *testing.T) {
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	oai, err := cfg.OpenAI()
	if err != nil {
		t.Fatalf("OpenAI() error = %v", err)
	}

	if oai.APIKey != "test-api-key" {
		t.Errorf("OpenAI().APIKey = %q, want %q", oai.APIKey, "test-api-key")
	}
}
