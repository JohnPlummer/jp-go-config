# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.4.1] - 2025-01-22

### Breaking Changes

- **Renamed `DB_URL` to `DATABASE_URL`** - Aligns with industry-standard naming used by Heroku, Render, Railway, and most PostgreSQL hosting providers. No mapping required when deploying to these platforms.

## [0.4.0] - 2025-01-22

### Breaking Changes

- **Removed alias environment variables:**
  - `DB_DATABASE` removed (use `DB_NAME`)
  - `DB_USERNAME` removed (use `DB_USER`)
  - `DB_PASS` removed (use `DB_PASSWORD`)

### Added

- **`Load()` API with auto-discovery** - New primary entry point that automatically discovers and loads `.env` and `config.yaml` files from the current directory or project root (directory containing `go.mod` or `.git`).

- **`DB_URL` support** - Parse PostgreSQL connection URLs (`postgres://` or `postgresql://`) with the new `DB_URL` environment variable. When set, takes precedence over individual connection variables (`DB_HOST`, `DB_PORT`, etc.). Pool settings are always applied from individual env vars.

- **Load options:**
  - `WithEnvPath(path)` - Specify explicit `.env` file path, disabling auto-discovery
  - `WithConfigPath(path)` - Specify explicit config file path, disabling auto-discovery
  - `WithPrefix(prefix)` - Set environment variable prefix (default: `APP`)
  - `WithLogger(logger)` - Set custom `slog.Logger` for discovery events

- **`Config` struct with cached, validated accessors:**
  - `Database()` - Returns validated `*DatabaseConfig`
  - `Server()` - Returns validated `*ServerConfig`
  - `OpenAI()` - Returns validated `*OpenAIConfig`
  - `Resilience()` - Returns validated `*ResilienceConfig`
  - `Standard()` - Returns underlying `*Standard` for advanced usage

- **`ValidateAll()` method** - Validates all configuration types for fail-fast behavior at startup. Validation order: database, server, openai, resilience.

- **`ParseDatabaseURL()` function** - Standalone function to parse PostgreSQL connection URLs into `DatabaseConfig` structs.

- **`FindProjectRoot()` function** - Searches parent directories for project root markers like `go.mod` or `.git`.

## [0.3.0] - Previous Release

See git history for earlier changes.
