# go-config

Enterprise-standard configuration management for Go applications, wrapping Viper with typed configuration, automatic .env file loading, and comprehensive validation.

## Features

- **Viper wrapper** with sensible defaults
- **Automatic .env file loading** with environment variable precedence
- **Typed configuration structs** for database, server, and OpenAI
- **Functional options pattern** for flexible initialization
- **Comprehensive validation** with helpful error messages
- **Zero configuration required** - works with defaults out of the box

## Installation

```bash
go get github.com/JohnPlummer/go-config
```

## Quick Start

```go
package main

import (
    "log"
    "github.com/JohnPlummer/go-config"
)

func main() {
    // Create standard config loader (loads .env files automatically)
    std, err := config.NewStandard()
    if err != nil {
        log.Fatal(err)
    }

    // Load typed database configuration
    dbConfig := config.DatabaseConfigFromViper(std)

    // Validate configuration
    if err := dbConfig.Validate(); err != nil {
        log.Fatal(err)
    }

    // Use the configuration
    log.Printf("Connecting to %s", dbConfig.ConnectionString())
}
```

## Configuration Loading

The Standard loader follows this precedence (highest to lowest):

1. Environment variables with configured prefix (default: `APP_`)
2. .env file values
3. Config file values (if provided)
4. Default values

### Creating a Standard Config Loader

```go
// With defaults (loads .env, uses APP_ prefix)
std, err := config.NewStandard()

// With custom prefix
std, err := config.NewStandard(
    config.WithEnvPrefix("MYAPP"),
)

// With config file
std, err := config.NewStandard(
    config.WithConfigFile("config.yaml"),
)

// With config name and search paths
std, err := config.NewStandard(
    config.WithConfigName("config"),
    config.WithConfigType("yaml"),
    config.WithConfigPaths(".", "/etc/myapp"),
)

// With custom .env file
std, err := config.NewStandard(
    config.WithEnvFile("/path/to/custom.env"),
)
```

## Database Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `DB_HOST` | `localhost` | Database host |
| `DB_PORT` | `5432` | Database port |
| `DB_NAME` or `DB_DATABASE` | `postgres` | Database name |
| `DB_USER` or `DB_USERNAME` | `postgres` | Database user |
| `DB_PASSWORD` or `DB_PASS` | (none) | Database password (required) |
| `DB_SSLMODE` | `disable` | SSL mode (disable, require, verify-ca, verify-full) |
| `DB_MAX_CONNS` | `25` | Maximum connections in pool |
| `DB_MIN_CONNS` | `5` | Minimum connections in pool |
| `DB_CONN_MAX_LIFETIME` | `1h` | Maximum connection lifetime |
| `DB_CONN_MAX_IDLE_TIME` | `10m` | Maximum connection idle time |
| `DB_RETRY_ATTEMPTS` | `3` | Number of retry attempts |
| `DB_RETRY_DELAY` | `2s` | Delay between retries |
| `DB_HEALTH_CHECK_PERIOD` | `30s` | Health check interval |

### Usage

```go
std, _ := config.NewStandard()
dbConfig := config.DatabaseConfigFromViper(std)

if err := dbConfig.Validate(); err != nil {
    log.Fatal(err)
}

// Get connection string
connStr := dbConfig.ConnectionString()
// postgres://user:pass@localhost:5432/mydb?sslmode=disable
```

## Server Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `localhost` | Server host |
| `SERVER_PORT` | `8080` | Server port |
| `SERVER_READ_TIMEOUT` | `15s` | Read timeout |
| `SERVER_WRITE_TIMEOUT` | `15s` | Write timeout |
| `SERVER_IDLE_TIMEOUT` | `60s` | Idle timeout |

### Usage

```go
std, _ := config.NewStandard()
serverConfig := config.ServerConfigFromViper(std)

if err := serverConfig.Validate(); err != nil {
    log.Fatal(err)
}

// Get address for net/http
addr := serverConfig.Address() // "localhost:8080"
```

## OpenAI Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `OPENAI_API_KEY` | (none) | OpenAI API key (required) |
| `OPENAI_MODEL` | `gpt-3.5-turbo` | Model to use |
| `OPENAI_TEMPERATURE` | `0.7` | Temperature (0.0 - 2.0) |
| `OPENAI_MAX_TOKENS` | `2000` | Maximum tokens in response |
| `OPENAI_TIMEOUT` | `30s` | Request timeout |

### Usage

```go
std, _ := config.NewStandard()
openaiConfig := config.OpenAIConfigFromViper(std)

if err := openaiConfig.Validate(); err != nil {
    log.Fatal(err)
}

client := openai.NewClient(openaiConfig.APIKey)
```

## Validation

All configuration structs provide a `Validate()` method that checks:

- Required fields are present
- Values are within acceptable ranges
- Cross-field constraints are satisfied

```go
dbConfig := config.DatabaseConfigFromViper(std)

if err := dbConfig.Validate(); err != nil {
    // Error messages are clear and actionable:
    // "database.port must be between 1 and 65535, got 99999"
    // "database.password is required"
    log.Fatal(err)
}
```

### Validation Helpers

The package provides validation helper functions you can use for custom configurations:

```go
// Validate required string field
if err := config.ValidateRequired("field.name", value); err != nil {
    return err
}

// Validate port number (1-65535)
if err := config.ValidatePort("server.port", port); err != nil {
    return err
}

// Validate positive duration
if err := config.ValidateDuration("timeout", duration); err != nil {
    return err
}

// Validate positive integer
if err := config.ValidatePositive("count", count); err != nil {
    return err
}

// Validate value in range
if err := config.ValidateRange("temperature", temp, 0.0, 2.0); err != nil {
    return err
}
```

## Migration from Monorepo

If migrating from the Some Things To Do monorepo:

### Before

```go
// pipeline/pkg/config
config, err := config.LoadConfig("config.yaml")
dbConfig := config.Database
```

### After

```go
// github.com/JohnPlummer/go-config
std, err := config.NewStandard(config.WithConfigFile("config.yaml"))
dbConfig := config.DatabaseConfigFromViper(std)
```

The database configuration struct is compatible - you may only need to update import paths.

## Usage Patterns

### Environment-Only Configuration

For containerized deployments without config files:

```go
// Just loads from environment variables and .env files
std, err := config.NewStandard()
dbConfig := config.DatabaseConfigFromViper(std)
```

### Config File + Environment Overrides

For local development with environment-specific overrides:

```go
std, err := config.NewStandard(
    config.WithConfigFile("config.yaml"),
)
// Environment variables override config file values
dbConfig := config.DatabaseConfigFromViper(std)
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

## Examples

See the [examples](./examples) directory for complete, runnable examples:

- [Basic usage](./examples/basic/main.go) - Loading and using typed configurations
- [Validation](./examples/validation/main.go) - Error handling and validation examples

## Development

### Requirements

- Go 1.21 or higher

### Testing

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -v -race -cover ./...

# View coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Linting

```bash
golangci-lint run
```

## License

MIT License - see [LICENSE](LICENSE) for details.

## Contributing

This package is extracted from the Some Things To Do monorepo and follows enterprise Go standards:

- Comprehensive test coverage (>80%)
- Clear, actionable error messages
- Backward-compatible API changes
- Documentation for all exported types and functions
