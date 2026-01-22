# jp-go-config

Configuration management for Go applications with auto-discovery, typed configuration structs, and comprehensive validation.

## Installation

```bash
go get github.com/JohnPlummer/jp-go-config
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/JohnPlummer/jp-go-config"
)

func main() {
    // Load configuration with auto-discovery
    cfg, err := config.Load()
    if err != nil {
        log.Fatal(err)
    }

    // Get validated database configuration
    db, err := cfg.Database()
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Connecting to %s", db.ConnectionString())
}
```

## Configuration Precedence

Configuration values are resolved in this order (highest to lowest priority):

1. Environment variables
2. `.env` file values
3. `config.yaml` file values
4. Default values

Environment variables always win. The `.env` file does not override existing environment variables.

## Auto-Discovery

`Load()` automatically discovers and loads configuration files from:

1. Current working directory
2. Project root (directory containing `go.mod` or `.git`)

Files discovered:

- `.env` - Environment variables (loaded first, does not override existing env vars)
- `config.yaml` or `config.yml` - Configuration values

No configuration files are required. Missing files are silently skipped.

## Environment Variables

### Database Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_URL` | (none) | PostgreSQL connection URL (takes precedence over individual vars) |
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_NAME` | `postgres` | Database name |
| `DB_USER` | `postgres` | Database user |
| `DB_PASSWORD` | (none) | Database password (required) |
| `DB_SSLMODE` | `disable` | SSL mode (disable, require, verify-ca, verify-full) |
| `DB_MAX_CONNS` | `25` | Maximum connections in pool |
| `DB_MIN_CONNS` | `5` | Minimum connections in pool |
| `DB_CONN_MAX_LIFETIME` | `1h` | Maximum connection lifetime |
| `DB_CONN_MAX_IDLE_TIME` | `10m` | Maximum connection idle time |
| `DB_RETRY_ATTEMPTS` | `3` | Number of retry attempts |
| `DB_RETRY_DELAY` | `2s` | Delay between retries |
| `DB_HEALTH_CHECK_PERIOD` | `30s` | Health check interval |

### Server Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `localhost` | Server host |
| `SERVER_PORT` | `8080` | Server port |
| `SERVER_READ_TIMEOUT` | `15s` | Read timeout |
| `SERVER_WRITE_TIMEOUT` | `15s` | Write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | Idle timeout |

### OpenAI Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | (none) | OpenAI API key (required) |
| `OPENAI_MODEL` | `gpt-3.5-turbo` | Model to use |
| `OPENAI_TEMPERATURE` | `0.7` | Temperature (0.0 - 2.0) |
| `OPENAI_MAX_TOKENS` | `2000` | Maximum tokens in response |
| `OPENAI_TIMEOUT` | `30s` | Request timeout |

### Resilience Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `RESILIENCE_MAX_RETRIES` | `3` | Maximum retry attempts |
| `RESILIENCE_INITIAL_DELAY` | `1s` | Initial retry delay |
| `RESILIENCE_MAX_DELAY` | `30s` | Maximum retry delay |
| `RESILIENCE_MULTIPLIER` | `2.0` | Backoff multiplier |
| `RESILIENCE_MAX_REQUESTS` | `10` | Circuit breaker max requests |
| `RESILIENCE_INTERVAL` | `10s` | Circuit breaker interval |
| `RESILIENCE_TIMEOUT` | `60s` | Circuit breaker timeout |
| `RESILIENCE_FAILURE_THRESHOLD` | `0.6` | Failure threshold (0.0 - 1.0) |

## DB_URL Support

Use a PostgreSQL connection URL instead of individual environment variables:

```bash
DB_URL=postgres://user:password@localhost:5432/mydb?sslmode=require
```

When `DB_URL` is set, it takes precedence over `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`, and `DB_SSLMODE`.

Pool settings (`DB_MAX_CONNS`, etc.) are always read from individual environment variables and applied regardless of whether `DB_URL` is used.

Supported URL schemes: `postgres://` and `postgresql://`

## Explicit Paths

Override auto-discovery with explicit file paths:

```go
cfg, err := config.Load(
    config.WithEnvPath("/path/to/.env"),
    config.WithConfigPath("/path/to/config.yaml"),
)
```

When explicit paths are provided:

- The file must exist (returns error if not found)
- Auto-discovery is disabled for that file type

## Load Options

| Option | Description |
|--------|-------------|
| `WithEnvPath(path)` | Explicit `.env` file path |
| `WithConfigPath(path)` | Explicit config file path |
| `WithPrefix(prefix)` | Environment variable prefix (default: `APP`) |
| `WithLogger(logger)` | Custom `slog.Logger` for discovery events |

## Accessing Configuration

The `Config` struct provides typed accessor methods that validate and cache on first access:

```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

// Each accessor validates and caches the config
db, err := cfg.Database()
server, err := cfg.Server()
openai, err := cfg.OpenAI()
resilience, err := cfg.Resilience()
```

## Fail-Fast Validation

Use `ValidateAll()` to validate all configuration types at startup:

```go
cfg, err := config.Load()
if err != nil {
    log.Fatal(err)
}

if err := cfg.ValidateAll(); err != nil {
    log.Fatal(err)
}

// All configs are now validated and cached
db, _ := cfg.Database()
server, _ := cfg.Server()
```

Validation order: database, server, openai, resilience.

## Advanced Usage

For lower-level access or custom configurations, use `NewStandard()` directly:

```go
// Create standard Viper wrapper
std, err := config.NewStandard(
    config.WithEnvPrefix("MYAPP"),
    config.WithConfigFile("config.yaml"),
)
if err != nil {
    log.Fatal(err)
}

// Load typed configuration manually
dbConfig := config.DatabaseConfigFromViper(std)
if err := dbConfig.Validate(); err != nil {
    log.Fatal(err)
}
```

### Custom Configuration Structs

```go
type MyConfig struct {
    APIKey  string `mapstructure:"api_key"`
    Timeout int    `mapstructure:"timeout"`
}

std, _ := config.NewStandard()
std.BindEnv("myservice.api_key", "MYSERVICE_API_KEY")
std.BindEnv("myservice.timeout", "MYSERVICE_TIMEOUT")

var myConfig MyConfig
if err := std.Unmarshal(&myConfig); err != nil {
    log.Fatal(err)
}
```

### Access Underlying Standard

```go
cfg, _ := config.Load()
std := cfg.Standard()

// Use Standard methods
value := std.GetString("custom.key")
```

## Validation Helpers

Use these functions for custom configuration validation:

```go
// Required string field
if err := config.ValidateRequired("field.name", value); err != nil {
    return err
}

// Port number (1-65535)
if err := config.ValidatePort("server.port", port); err != nil {
    return err
}

// Positive duration
if err := config.ValidateDuration("timeout", duration); err != nil {
    return err
}

// Positive integer
if err := config.ValidatePositive("count", count); err != nil {
    return err
}

// Value in range
if err := config.ValidateRange("temperature", temp, 0.0, 2.0); err != nil {
    return err
}
```

## Development

### Requirements

- Go 1.21 or higher

### Testing

```bash
go test -v ./...
go test -v -race -cover ./...
```

### Linting

```bash
golangci-lint run
```

## License

MIT License - see [LICENSE](LICENSE) for details.
