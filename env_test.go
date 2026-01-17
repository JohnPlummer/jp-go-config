package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindProjectRoot(t *testing.T) {
	// Create a temporary directory structure for testing
	tmpDir, err := os.MkdirTemp("", "findroot-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create nested directory structure: tmpDir/project/sub1/sub2
	projectDir := filepath.Join(tmpDir, "project")
	sub1Dir := filepath.Join(projectDir, "sub1")
	sub2Dir := filepath.Join(sub1Dir, "sub2")

	if err := os.MkdirAll(sub2Dir, 0o755); err != nil {
		t.Fatal(err)
	}

	// Create marker files in project dir
	if err := os.WriteFile(filepath.Join(projectDir, "CLAUDE.md"), []byte("# Test"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(projectDir, "config.yaml"), []byte("test: true"), 0o644); err != nil {
		t.Fatal(err)
	}

	t.Run("no markers specified", func(t *testing.T) {
		result := FindProjectRoot()
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("finds root with single marker", func(t *testing.T) {
		// Save current dir and change to sub2
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		if err := os.Chdir(sub2Dir); err != nil {
			t.Fatal(err)
		}

		result := FindProjectRoot("CLAUDE.md")
		// Normalize paths to handle macOS /var -> /private/var symlinks
		expectedNorm, _ := filepath.EvalSymlinks(projectDir)
		resultNorm, _ := filepath.EvalSymlinks(result)
		if resultNorm != expectedNorm {
			t.Errorf("expected %s, got %s", expectedNorm, resultNorm)
		}
	})

	t.Run("finds root with multiple markers", func(t *testing.T) {
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		if err := os.Chdir(sub2Dir); err != nil {
			t.Fatal(err)
		}

		result := FindProjectRoot("CLAUDE.md", "config.yaml")
		// Normalize paths to handle macOS /var -> /private/var symlinks
		expectedNorm, _ := filepath.EvalSymlinks(projectDir)
		resultNorm, _ := filepath.EvalSymlinks(result)
		if resultNorm != expectedNorm {
			t.Errorf("expected %s, got %s", expectedNorm, resultNorm)
		}
	})

	t.Run("returns empty when marker not found", func(t *testing.T) {
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		if err := os.Chdir(sub2Dir); err != nil {
			t.Fatal(err)
		}

		result := FindProjectRoot("nonexistent.marker")
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})

	t.Run("returns empty when only some markers found", func(t *testing.T) {
		origDir, _ := os.Getwd()
		defer os.Chdir(origDir)

		if err := os.Chdir(sub2Dir); err != nil {
			t.Fatal(err)
		}

		// CLAUDE.md exists but nonexistent.marker doesn't
		result := FindProjectRoot("CLAUDE.md", "nonexistent.marker")
		if result != "" {
			t.Errorf("expected empty string, got %s", result)
		}
	})
}

func TestExpandEnvStrict(t *testing.T) {
	t.Run("expands single variable", func(t *testing.T) {
		os.Setenv("TEST_HOST", "localhost")
		defer os.Unsetenv("TEST_HOST")

		result, err := ExpandEnvStrict("host=${TEST_HOST}")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "host=localhost" {
			t.Errorf("expected 'host=localhost', got '%s'", result)
		}
	})

	t.Run("expands multiple variables", func(t *testing.T) {
		os.Setenv("TEST_HOST", "localhost")
		os.Setenv("TEST_PORT", "5432")
		defer os.Unsetenv("TEST_HOST")
		defer os.Unsetenv("TEST_PORT")

		result, err := ExpandEnvStrict("${TEST_HOST}:${TEST_PORT}")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "localhost:5432" {
			t.Errorf("expected 'localhost:5432', got '%s'", result)
		}
	})

	t.Run("returns error for unset variable", func(t *testing.T) {
		os.Unsetenv("UNSET_VAR")

		_, err := ExpandEnvStrict("value=${UNSET_VAR}")
		if err == nil {
			t.Error("expected error for unset variable")
		}
		if err.Error() != "environment variable UNSET_VAR not set" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("returns unchanged value without variables", func(t *testing.T) {
		result, err := ExpandEnvStrict("plain value")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "plain value" {
			t.Errorf("expected 'plain value', got '%s'", result)
		}
	})

	t.Run("ignores default syntax but still requires var to be set", func(t *testing.T) {
		os.Setenv("TEST_VAR", "actual_value")
		defer os.Unsetenv("TEST_VAR")

		// The :- default syntax is parsed but ignored; we use the actual env value
		result, err := ExpandEnvStrict("${TEST_VAR:-default}")
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if result != "actual_value" {
			t.Errorf("expected 'actual_value', got '%s'", result)
		}
	})

	t.Run("fails for variable with default syntax when var not set", func(t *testing.T) {
		os.Unsetenv("MISSING_VAR")

		_, err := ExpandEnvStrict("${MISSING_VAR:-default}")
		if err == nil {
			t.Error("expected error for unset variable even with default syntax")
		}
	})

	t.Run("handles empty variable value as unset", func(t *testing.T) {
		os.Setenv("EMPTY_VAR", "")
		defer os.Unsetenv("EMPTY_VAR")

		_, err := ExpandEnvStrict("value=${EMPTY_VAR}")
		if err == nil {
			t.Error("expected error for empty variable")
		}
	})
}

func TestLoadEnvFiles(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "loadenv-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Store original env values to restore later
	originalEnv := make(map[string]string)
	envVars := []string{"BASE_VAR", "OVERRIDE_VAR", "TEST_ONLY_VAR"}
	for _, v := range envVars {
		originalEnv[v] = os.Getenv(v)
	}
	defer func() {
		for k, v := range originalEnv {
			if v == "" {
				os.Unsetenv(k)
			} else {
				os.Setenv(k, v)
			}
		}
	}()

	t.Run("loads base env file", func(t *testing.T) {
		// Clear env vars
		for _, v := range envVars {
			os.Unsetenv(v)
		}

		// Create .env file
		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("BASE_VAR=base_value\nOVERRIDE_VAR=base_override"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := LoadEnvFiles("", tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if os.Getenv("BASE_VAR") != "base_value" {
			t.Errorf("expected BASE_VAR=base_value, got %s", os.Getenv("BASE_VAR"))
		}
	})

	t.Run("loads environment-specific file with override", func(t *testing.T) {
		// Clear env vars
		for _, v := range envVars {
			os.Unsetenv(v)
		}

		// Create .env file
		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("BASE_VAR=base_value\nOVERRIDE_VAR=base_override"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Create .env.test file
		envTestFile := filepath.Join(tmpDir, ".env.test")
		if err := os.WriteFile(envTestFile, []byte("OVERRIDE_VAR=test_override\nTEST_ONLY_VAR=test_only"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := LoadEnvFiles("test", tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// BASE_VAR should be from base .env
		if os.Getenv("BASE_VAR") != "base_value" {
			t.Errorf("expected BASE_VAR=base_value, got %s", os.Getenv("BASE_VAR"))
		}

		// OVERRIDE_VAR should be overridden by .env.test
		if os.Getenv("OVERRIDE_VAR") != "test_override" {
			t.Errorf("expected OVERRIDE_VAR=test_override, got %s", os.Getenv("OVERRIDE_VAR"))
		}

		// TEST_ONLY_VAR should be from .env.test
		if os.Getenv("TEST_ONLY_VAR") != "test_only" {
			t.Errorf("expected TEST_ONLY_VAR=test_only, got %s", os.Getenv("TEST_ONLY_VAR"))
		}
	})

	t.Run("does not fail when files are missing", func(t *testing.T) {
		// Clear env vars
		for _, v := range envVars {
			os.Unsetenv(v)
		}

		// Create a directory without any .env files
		emptyDir := filepath.Join(tmpDir, "empty")
		if err := os.MkdirAll(emptyDir, 0o755); err != nil {
			t.Fatal(err)
		}

		err := LoadEnvFiles("production", emptyDir)
		if err != nil {
			t.Errorf("unexpected error for missing files: %v", err)
		}
	})

	t.Run("base env does not override existing vars", func(t *testing.T) {
		// Clear env vars
		for _, v := range envVars {
			os.Unsetenv(v)
		}

		// Set a var before loading
		os.Setenv("BASE_VAR", "already_set")

		// Create .env file
		envFile := filepath.Join(tmpDir, ".env")
		if err := os.WriteFile(envFile, []byte("BASE_VAR=should_not_override"), 0o644); err != nil {
			t.Fatal(err)
		}

		err := LoadEnvFiles("", tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		// Should keep original value
		if os.Getenv("BASE_VAR") != "already_set" {
			t.Errorf("expected BASE_VAR=already_set, got %s", os.Getenv("BASE_VAR"))
		}
	})

	t.Run("handles quoted values", func(t *testing.T) {
		// Clear env vars
		for _, v := range envVars {
			os.Unsetenv(v)
		}
		os.Unsetenv("DOUBLE_QUOTED")
		os.Unsetenv("SINGLE_QUOTED")

		// Create .env file with quoted values
		envFile := filepath.Join(tmpDir, ".env")
		content := `DOUBLE_QUOTED="value with spaces"
SINGLE_QUOTED='another value'`
		if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := LoadEnvFiles("", tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if os.Getenv("DOUBLE_QUOTED") != "value with spaces" {
			t.Errorf("expected DOUBLE_QUOTED='value with spaces', got '%s'", os.Getenv("DOUBLE_QUOTED"))
		}
		if os.Getenv("SINGLE_QUOTED") != "another value" {
			t.Errorf("expected SINGLE_QUOTED='another value', got '%s'", os.Getenv("SINGLE_QUOTED"))
		}
	})

	t.Run("ignores comments and empty lines", func(t *testing.T) {
		// Clear env vars
		for _, v := range envVars {
			os.Unsetenv(v)
		}
		os.Unsetenv("ACTUAL_VAR")

		// Create .env file with comments
		envFile := filepath.Join(tmpDir, ".env")
		content := `# This is a comment
ACTUAL_VAR=value

# Another comment
`
		if err := os.WriteFile(envFile, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}

		err := LoadEnvFiles("", tmpDir)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}

		if os.Getenv("ACTUAL_VAR") != "value" {
			t.Errorf("expected ACTUAL_VAR=value, got '%s'", os.Getenv("ACTUAL_VAR"))
		}
	})
}

func TestTrimQuotes(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"quoted"`, "quoted"},
		{`'single'`, "single"},
		{`no quotes`, "no quotes"},
		{`"partial`, `"partial`},
		{`partial"`, `partial"`},
		{`""`, ""},
		{`''`, ""},
		{`"a"`, "a"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := trimQuotes(tt.input)
			if result != tt.expected {
				t.Errorf("trimQuotes(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestParentDir(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/a/b/c", "/a/b"},
		{"/a/b", "/a"},
		{"/a", ""},
		{"/", ""},
		{"", ""},
		{"no/leading/slash", "no/leading"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parentDir(tt.input)
			if result != tt.expected {
				t.Errorf("parentDir(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}
