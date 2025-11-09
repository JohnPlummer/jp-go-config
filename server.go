package config

import (
	"fmt"
	"time"
)

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
}

// ServerConfigFromViper creates a ServerConfig from a Standard config loader.
//
// Environment variable mappings:
//   - SERVER_HOST -> host (default: localhost)
//   - SERVER_PORT -> port (default: 8080)
//   - SERVER_READ_TIMEOUT -> read_timeout (default: 15s)
//   - SERVER_WRITE_TIMEOUT -> write_timeout (default: 15s)
//   - SERVER_IDLE_TIMEOUT -> idle_timeout (default: 60s)
func ServerConfigFromViper(s *Standard) ServerConfig {
	// Bind environment variables
	_ = s.BindEnv("server.host", "SERVER_HOST")
	_ = s.BindEnv("server.port", "SERVER_PORT")
	_ = s.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	_ = s.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	_ = s.BindEnv("server.idle_timeout", "SERVER_IDLE_TIMEOUT")

	config := ServerConfig{
		Host:         s.GetString("server.host"),
		Port:         s.GetInt("server.port"),
		ReadTimeout:  s.viper.GetDuration("server.read_timeout"),
		WriteTimeout: s.viper.GetDuration("server.write_timeout"),
		IdleTimeout:  s.viper.GetDuration("server.idle_timeout"),
	}

	// Apply defaults
	config.setDefaults()

	return config
}

// setDefaults sets default values for optional fields
func (c *ServerConfig) setDefaults() {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 8080
	}
	if c.ReadTimeout == 0 {
		c.ReadTimeout = 15 * time.Second
	}
	if c.WriteTimeout == 0 {
		c.WriteTimeout = 15 * time.Second
	}
	if c.IdleTimeout == 0 {
		c.IdleTimeout = 60 * time.Second
	}
}

// Validate validates the server configuration
func (c *ServerConfig) Validate() error {
	if err := ValidateRequired("server.host", c.Host); err != nil {
		return err
	}
	if err := ValidatePort("server.port", c.Port); err != nil {
		return err
	}
	if err := ValidateDuration("server.read_timeout", c.ReadTimeout); err != nil {
		return err
	}
	if err := ValidateDuration("server.write_timeout", c.WriteTimeout); err != nil {
		return err
	}
	if err := ValidateDuration("server.idle_timeout", c.IdleTimeout); err != nil {
		return err
	}

	return nil
}

// Address returns the server address in host:port format
func (c *ServerConfig) Address() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
