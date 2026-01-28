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

func TestLoad_WithBothEnvAndConfigPaths(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create .env file
	envContent := `COMBINED_TEST_HOST=envhost
COMBINED_TEST_PASSWORD=envpass
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Create config.yaml file
	configContent := `server:
  host: confighost
  port: 6666
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config.yaml file: %v", err)
	}

	// Save and restore env vars
	origHost := os.Getenv("COMBINED_TEST_HOST")
	origPassword := os.Getenv("COMBINED_TEST_PASSWORD")
	t.Cleanup(func() {
		if origHost != "" {
			os.Setenv("COMBINED_TEST_HOST", origHost)
		} else {
			os.Unsetenv("COMBINED_TEST_HOST")
		}
		if origPassword != "" {
			os.Setenv("COMBINED_TEST_PASSWORD", origPassword)
		} else {
			os.Unsetenv("COMBINED_TEST_PASSWORD")
		}
	})
	os.Unsetenv("COMBINED_TEST_HOST")
	os.Unsetenv("COMBINED_TEST_PASSWORD")

	// Load with both options combined
	cfg, err := Load(WithEnvPath(envPath), WithConfigPath(configPath))
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Verify env vars were loaded from .env
	if got := os.Getenv("COMBINED_TEST_HOST"); got != "envhost" {
		t.Errorf("COMBINED_TEST_HOST = %q, want %q", got, "envhost")
	}

	// Verify config values were loaded from config.yaml
	srv, err := cfg.Server()
	if err != nil {
		t.Fatalf("Server() error = %v", err)
	}
	if srv.Host != "confighost" {
		t.Errorf("Server().Host = %q, want %q", srv.Host, "confighost")
	}
	if srv.Port != 6666 {
		t.Errorf("Server().Port = %d, want %d", srv.Port, 6666)
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

func TestValidateAll_Success(t *testing.T) {
	// Set all required config values
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	err = cfg.ValidateAll()
	if err != nil {
		t.Errorf("ValidateAll() error = %v, want nil", err)
	}
}

func TestValidateAll_DatabaseError(t *testing.T) {
	// Set OpenAI key but not DB password
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	os.Unsetenv("DB_PASSWORD")
	defer os.Unsetenv("OPENAI_API_KEY")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	err = cfg.ValidateAll()
	if err == nil {
		t.Fatal("ValidateAll() error = nil, want database validation error")
	}
}

func TestValidateAll_OpenAIError(t *testing.T) {
	// Set DB password but not OpenAI key
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Unsetenv("OPENAI_API_KEY")
	defer os.Unsetenv("DB_PASSWORD")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	err = cfg.ValidateAll()
	if err == nil {
		t.Fatal("ValidateAll() error = nil, want openai validation error")
	}
}

func TestValidateAll_CachesConfigs(t *testing.T) {
	// Set all required config values
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Setenv("OPENAI_API_KEY", "test-api-key")
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("OPENAI_API_KEY")
	}()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	err = cfg.ValidateAll()
	if err != nil {
		t.Fatalf("ValidateAll() error = %v", err)
	}

	// Subsequent accessor calls should return cached configs
	db1, _ := cfg.Database()
	db2, _ := cfg.Database()
	if db1 != db2 {
		t.Error("Database() did not return cached instance after ValidateAll()")
	}
}

func TestDiscovery_EnvFileInCwd(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create temp directory with .env file
	tmpDir := t.TempDir()
	envContent := `DISCOVERY_TEST_VAR=from_cwd
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Save and restore env var
	origVal := os.Getenv("DISCOVERY_TEST_VAR")
	t.Cleanup(func() {
		if origVal != "" {
			os.Setenv("DISCOVERY_TEST_VAR", origVal)
		} else {
			os.Unsetenv("DISCOVERY_TEST_VAR")
		}
	})
	os.Unsetenv("DISCOVERY_TEST_VAR")

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should auto-discover .env in cwd
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was loaded
	if got := os.Getenv("DISCOVERY_TEST_VAR"); got != "from_cwd" {
		t.Errorf("DISCOVERY_TEST_VAR = %q, want %q", got, "from_cwd")
	}
}

func TestDiscovery_ConfigYamlInCwd(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create temp directory with config.yaml file
	tmpDir := t.TempDir()
	configContent := `server:
  host: discovered_host
  port: 7777
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config.yaml file: %v", err)
	}

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should auto-discover config.yaml in cwd
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config was loaded
	srv, err := cfg.Server()
	if err != nil {
		t.Fatalf("Server() error = %v", err)
	}
	if srv.Host != "discovered_host" {
		t.Errorf("Server().Host = %q, want %q", srv.Host, "discovered_host")
	}
	if srv.Port != 7777 {
		t.Errorf("Server().Port = %d, want %d", srv.Port, 7777)
	}
}

func TestDiscovery_MissingFilesDoNotCauseErrors(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create empty temp directory (no .env or config.yaml)
	tmpDir := t.TempDir()

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should succeed even without any config files
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil (missing files should not cause error)", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil Config")
	}
}

func TestDiscovery_EnvFileDoesNotOverrideExisting(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create temp directory with .env file
	tmpDir := t.TempDir()
	envContent := `NO_OVERRIDE_TEST=from_file
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Set env var before Load (should not be overridden)
	os.Setenv("NO_OVERRIDE_TEST", "from_env")
	t.Cleanup(func() {
		os.Unsetenv("NO_OVERRIDE_TEST")
	})

	// Change to temp directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should auto-discover .env but not override existing var
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was NOT overridden
	if got := os.Getenv("NO_OVERRIDE_TEST"); got != "from_env" {
		t.Errorf("NO_OVERRIDE_TEST = %q, want %q (should not be overridden)", got, "from_env")
	}
}

// TestDiscovery_MonorepoScenario tests .env discovery when go.mod is in a subdirectory
// but .env is at the git repository root. This is the key fix for ST-1098.
func TestDiscovery_MonorepoScenario(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create directory structure simulating monorepo:
	// tmpDir/              (git root with .git and .env)
	//   ├── .git/
	//   ├── .env           (config file at repo root)
	//   └── submodule/     (cwd - Go module root)
	//       └── go.mod
	tmpDir := t.TempDir()

	// Create .git directory (simulates repo root)
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o750); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create .env at repo root
	envContent := `MONOREPO_TEST_VAR=from_git_root
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Create submodule directory with go.mod
	submoduleDir := filepath.Join(tmpDir, "submodule")
	if err := os.Mkdir(submoduleDir, 0o750); err != nil {
		t.Fatalf("failed to create submodule dir: %v", err)
	}
	goModContent := `module example.com/submodule

go 1.21
`
	goModPath := filepath.Join(submoduleDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}

	// Save and restore env var
	origVal := os.Getenv("MONOREPO_TEST_VAR")
	t.Cleanup(func() {
		if origVal != "" {
			os.Setenv("MONOREPO_TEST_VAR", origVal)
		} else {
			os.Unsetenv("MONOREPO_TEST_VAR")
		}
	})
	os.Unsetenv("MONOREPO_TEST_VAR")

	// Change to submodule directory (simulates running from Go module)
	if err := os.Chdir(submoduleDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should discover .env at git root even though go.mod is in submodule
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was loaded from git root
	if got := os.Getenv("MONOREPO_TEST_VAR"); got != "from_git_root" {
		t.Errorf("MONOREPO_TEST_VAR = %q, want %q (should find .env at git root)", got, "from_git_root")
	}
}

// TestDiscovery_ConfigYamlInMonorepo tests config.yaml discovery in monorepo scenario.
func TestDiscovery_ConfigYamlInMonorepo(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create monorepo structure
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o750); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create config.yaml at repo root
	configContent := `server:
  host: monorepo_host
  port: 9999
`
	configPath := filepath.Join(tmpDir, "config.yaml")
	if err := os.WriteFile(configPath, []byte(configContent), 0o644); err != nil {
		t.Fatalf("failed to write config.yaml file: %v", err)
	}

	// Create submodule directory with go.mod
	submoduleDir := filepath.Join(tmpDir, "submodule")
	if err := os.Mkdir(submoduleDir, 0o750); err != nil {
		t.Fatalf("failed to create submodule dir: %v", err)
	}
	goModContent := `module example.com/submodule

go 1.21
`
	goModPath := filepath.Join(submoduleDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}

	// Change to submodule directory
	if err := os.Chdir(submoduleDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should discover config.yaml at git root
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify config was loaded from git root
	srv, err := cfg.Server()
	if err != nil {
		t.Fatalf("Server() error = %v", err)
	}
	if srv.Host != "monorepo_host" {
		t.Errorf("Server().Host = %q, want %q", srv.Host, "monorepo_host")
	}
	if srv.Port != 9999 {
		t.Errorf("Server().Port = %d, want %d", srv.Port, 9999)
	}
}

// TestDiscovery_StandardRepo tests discovery when go.mod and .git are in the same directory.
func TestDiscovery_StandardRepo(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create standard repo structure (go.mod and .git in same directory)
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o750); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create go.mod in same directory
	goModContent := `module example.com/standard

go 1.21
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}

	// Create .env
	envContent := `STANDARD_REPO_VAR=standard_value
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Save and restore env var
	origVal := os.Getenv("STANDARD_REPO_VAR")
	t.Cleanup(func() {
		if origVal != "" {
			os.Setenv("STANDARD_REPO_VAR", origVal)
		} else {
			os.Unsetenv("STANDARD_REPO_VAR")
		}
	})
	os.Unsetenv("STANDARD_REPO_VAR")

	// Change to repo directory
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should work normally
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was loaded
	if got := os.Getenv("STANDARD_REPO_VAR"); got != "standard_value" {
		t.Errorf("STANDARD_REPO_VAR = %q, want %q", got, "standard_value")
	}
}

// TestDiscovery_GitWorktree tests discovery with git worktree (.git is a file, not directory).
func TestDiscovery_GitWorktree(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create worktree structure:
	// tmpDir/
	//   └── worktree/       (git worktree root)
	//       ├── .git        (FILE, not directory - points to bare repo)
	//       ├── .env
	//       └── submodule/  (cwd)
	//           └── go.mod
	tmpDir := t.TempDir()

	// Create worktree directory
	worktreeDir := filepath.Join(tmpDir, "worktree")
	if err := os.Mkdir(worktreeDir, 0o750); err != nil {
		t.Fatalf("failed to create worktree dir: %v", err)
	}

	// Create .git as a FILE (how git worktrees work)
	// Content points to the bare repo's git directory
	gitFileContent := "gitdir: /some/bare/repo/.git/worktrees/main\n"
	gitFilePath := filepath.Join(worktreeDir, ".git")
	if err := os.WriteFile(gitFilePath, []byte(gitFileContent), 0o644); err != nil {
		t.Fatalf("failed to write .git file: %v", err)
	}

	// Create .env at worktree root
	envContent := `WORKTREE_TEST_VAR=from_worktree_root
`
	envPath := filepath.Join(worktreeDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Create submodule directory with go.mod
	submoduleDir := filepath.Join(worktreeDir, "submodule")
	if err := os.Mkdir(submoduleDir, 0o750); err != nil {
		t.Fatalf("failed to create submodule dir: %v", err)
	}
	goModContent := `module example.com/submodule

go 1.21
`
	goModPath := filepath.Join(submoduleDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}

	// Save and restore env var
	origVal := os.Getenv("WORKTREE_TEST_VAR")
	t.Cleanup(func() {
		if origVal != "" {
			os.Setenv("WORKTREE_TEST_VAR", origVal)
		} else {
			os.Unsetenv("WORKTREE_TEST_VAR")
		}
	})
	os.Unsetenv("WORKTREE_TEST_VAR")

	// Change to submodule directory
	if err := os.Chdir(submoduleDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should discover .env at worktree root
	// This works because os.Stat() detects .git as a file, not just directories
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was loaded from worktree root
	if got := os.Getenv("WORKTREE_TEST_VAR"); got != "from_worktree_root" {
		t.Errorf("WORKTREE_TEST_VAR = %q, want %q", got, "from_worktree_root")
	}
}

// TestDiscovery_NoGitOnlyGoMod tests discovery when there's no .git (just a go module).
func TestDiscovery_NoGitOnlyGoMod(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create structure with go.mod but no .git
	tmpDir := t.TempDir()

	// Create go.mod
	goModContent := `module example.com/no-git

go 1.21
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}

	// Create .env
	envContent := `NO_GIT_VAR=from_gomod_root
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "cmd")
	if err := os.Mkdir(subDir, 0o750); err != nil {
		t.Fatalf("failed to create cmd dir: %v", err)
	}

	// Save and restore env var
	origVal := os.Getenv("NO_GIT_VAR")
	t.Cleanup(func() {
		if origVal != "" {
			os.Setenv("NO_GIT_VAR", origVal)
		} else {
			os.Unsetenv("NO_GIT_VAR")
		}
	})
	os.Unsetenv("NO_GIT_VAR")

	// Change to subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should find .env at go.mod root
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was loaded
	if got := os.Getenv("NO_GIT_VAR"); got != "from_gomod_root" {
		t.Errorf("NO_GIT_VAR = %q, want %q", got, "from_gomod_root")
	}
}

// TestDiscovery_NoGoModOnlyGit tests discovery when there's no go.mod (just .git).
func TestDiscovery_NoGoModOnlyGit(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create structure with .git but no go.mod
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o750); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create .env
	envContent := `NO_GOMOD_VAR=from_git_root
`
	envPath := filepath.Join(tmpDir, ".env")
	if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
		t.Fatalf("failed to write .env file: %v", err)
	}

	// Create subdirectory
	subDir := filepath.Join(tmpDir, "scripts")
	if err := os.Mkdir(subDir, 0o750); err != nil {
		t.Fatalf("failed to create scripts dir: %v", err)
	}

	// Save and restore env var
	origVal := os.Getenv("NO_GOMOD_VAR")
	t.Cleanup(func() {
		if origVal != "" {
			os.Setenv("NO_GOMOD_VAR", origVal)
		} else {
			os.Unsetenv("NO_GOMOD_VAR")
		}
	})
	os.Unsetenv("NO_GOMOD_VAR")

	// Change to subdirectory
	if err := os.Chdir(subDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// Load should find .env at git root
	_, err = Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify env var was loaded
	if got := os.Getenv("NO_GOMOD_VAR"); got != "from_git_root" {
		t.Errorf("NO_GOMOD_VAR = %q, want %q", got, "from_git_root")
	}
}

// TestCollectSearchPaths_Deduplication verifies search paths are deduplicated.
func TestCollectSearchPaths_Deduplication(t *testing.T) {
	// Save original directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	t.Cleanup(func() {
		os.Chdir(origDir)
	})

	// Create standard repo (go.mod and .git in same directory as cwd)
	tmpDir := t.TempDir()

	// Create .git directory
	gitDir := filepath.Join(tmpDir, ".git")
	if err := os.Mkdir(gitDir, 0o750); err != nil {
		t.Fatalf("failed to create .git dir: %v", err)
	}

	// Create go.mod
	goModContent := `module example.com/dedup

go 1.21
`
	goModPath := filepath.Join(tmpDir, "go.mod")
	if err := os.WriteFile(goModPath, []byte(goModContent), 0o644); err != nil {
		t.Fatalf("failed to write go.mod file: %v", err)
	}

	// Change to repo directory (cwd == go.mod root == .git root)
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Chdir() error = %v", err)
	}

	// collectSearchPaths should return only one path (deduplicated)
	paths := collectSearchPaths()
	if len(paths) != 1 {
		t.Errorf("collectSearchPaths() returned %d paths, want 1 (should deduplicate)", len(paths))
	}
}
