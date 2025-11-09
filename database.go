package config

import (
	"fmt"
	"time"
)

// DatabaseConfig holds PostgreSQL database configuration with connection pooling settings.
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	SSLMode  string `mapstructure:"ssl_mode"`

	// Connection pool settings
	MaxConns        int           `mapstructure:"max_conns"`
	MinConns        int           `mapstructure:"min_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`

	// Retry settings
	RetryAttempts int           `mapstructure:"retry_attempts"`
	RetryDelay    time.Duration `mapstructure:"retry_delay"`

	// Health check
	HealthCheckPeriod time.Duration `mapstructure:"health_check_period"`
}

// DatabaseConfigFromViper creates a DatabaseConfig from a Standard config loader.
//
// Environment variable mappings:
//   - DB_HOST -> host (default: localhost)
//   - DB_PORT -> port (default: 5432)
//   - DB_NAME or DB_DATABASE -> database (default: postgres)
//   - DB_USER or DB_USERNAME -> user (default: postgres)
//   - DB_PASSWORD or DB_PASS -> password
//   - DB_SSLMODE -> ssl_mode (default: disable)
//   - DB_MAX_CONNS -> max_conns (default: 25)
//   - DB_MIN_CONNS -> min_conns (default: 5)
//   - DB_CONN_MAX_LIFETIME -> conn_max_lifetime (default: 1h)
//   - DB_CONN_MAX_IDLE_TIME -> conn_max_idle_time (default: 10m)
//   - DB_RETRY_ATTEMPTS -> retry_attempts (default: 3)
//   - DB_RETRY_DELAY -> retry_delay (default: 2s)
//   - DB_HEALTH_CHECK_PERIOD -> health_check_period (default: 30s)
func DatabaseConfigFromViper(s *Standard) DatabaseConfig {
	// Bind environment variables
	_ = s.BindEnv("database.host", "DB_HOST")
	_ = s.BindEnv("database.port", "DB_PORT")
	_ = s.BindEnv("database.database", "DB_NAME", "DB_DATABASE")
	_ = s.BindEnv("database.user", "DB_USER", "DB_USERNAME")
	_ = s.BindEnv("database.password", "DB_PASSWORD", "DB_PASS")
	_ = s.BindEnv("database.ssl_mode", "DB_SSLMODE")
	_ = s.BindEnv("database.max_conns", "DB_MAX_CONNS")
	_ = s.BindEnv("database.min_conns", "DB_MIN_CONNS")
	_ = s.BindEnv("database.conn_max_lifetime", "DB_CONN_MAX_LIFETIME")
	_ = s.BindEnv("database.conn_max_idle_time", "DB_CONN_MAX_IDLE_TIME")
	_ = s.BindEnv("database.retry_attempts", "DB_RETRY_ATTEMPTS")
	_ = s.BindEnv("database.retry_delay", "DB_RETRY_DELAY")
	_ = s.BindEnv("database.health_check_period", "DB_HEALTH_CHECK_PERIOD")

	config := DatabaseConfig{
		Host:              s.GetString("database.host"),
		Port:              s.GetInt("database.port"),
		Database:          s.GetString("database.database"),
		User:              s.GetString("database.user"),
		Password:          s.GetString("database.password"),
		SSLMode:           s.GetString("database.ssl_mode"),
		MaxConns:          s.GetInt("database.max_conns"),
		MinConns:          s.GetInt("database.min_conns"),
		ConnMaxLifetime:   s.viper.GetDuration("database.conn_max_lifetime"),
		ConnMaxIdleTime:   s.viper.GetDuration("database.conn_max_idle_time"),
		RetryAttempts:     s.GetInt("database.retry_attempts"),
		RetryDelay:        s.viper.GetDuration("database.retry_delay"),
		HealthCheckPeriod: s.viper.GetDuration("database.health_check_period"),
	}

	// Apply defaults
	config.setDefaults()

	return config
}

// setDefaults sets default values for optional fields
func (c *DatabaseConfig) setDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.Database == "" {
		c.Database = "postgres"
	}
	if c.User == "" {
		c.User = "postgres"
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	if c.MaxConns == 0 {
		c.MaxConns = 25
	}
	if c.MinConns == 0 {
		c.MinConns = 5
	}
	if c.ConnMaxLifetime == 0 {
		c.ConnMaxLifetime = 1 * time.Hour
	}
	if c.ConnMaxIdleTime == 0 {
		c.ConnMaxIdleTime = 10 * time.Minute
	}
	if c.RetryAttempts == 0 {
		c.RetryAttempts = 3
	}
	if c.RetryDelay == 0 {
		c.RetryDelay = 2 * time.Second
	}
	if c.HealthCheckPeriod == 0 {
		c.HealthCheckPeriod = 30 * time.Second
	}
}

// Validate validates the database configuration
func (c *DatabaseConfig) Validate() error {
	if err := ValidateRequired("database.host", c.Host); err != nil {
		return err
	}
	if err := ValidatePort("database.port", c.Port); err != nil {
		return err
	}
	if err := ValidateRequired("database.database", c.Database); err != nil {
		return err
	}
	if err := ValidateRequired("database.user", c.User); err != nil {
		return err
	}
	if err := ValidateRequired("database.password", c.Password); err != nil {
		return err
	}

	// Validate SSL mode
	validSSLModes := []string{"disable", "require", "verify-ca", "verify-full"}
	valid := false
	for _, mode := range validSSLModes {
		if c.SSLMode == mode {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("database.ssl_mode must be one of: %v", validSSLModes)
	}

	// Validate connection pool settings
	if err := ValidatePositive("database.max_conns", c.MaxConns); err != nil {
		return err
	}
	if err := ValidateRange("database.min_conns", c.MinConns, 0, c.MaxConns); err != nil {
		return err
	}
	if err := ValidateDuration("database.conn_max_lifetime", c.ConnMaxLifetime); err != nil {
		return err
	}
	if err := ValidateDuration("database.conn_max_idle_time", c.ConnMaxIdleTime); err != nil {
		return err
	}

	// Validate retry settings
	if err := ValidateRange("database.retry_attempts", c.RetryAttempts, 0, 10); err != nil {
		return err
	}
	if err := ValidateDuration("database.retry_delay", c.RetryDelay); err != nil {
		return err
	}
	if err := ValidateDuration("database.health_check_period", c.HealthCheckPeriod); err != nil {
		return err
	}

	return nil
}

// ConnectionString returns a PostgreSQL connection string
func (c *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User,
		c.Password,
		c.Host,
		c.Port,
		c.Database,
		c.SSLMode,
	)
}
