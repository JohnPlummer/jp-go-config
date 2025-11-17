package config

import (
	"fmt"
	"time"
)

// ResilienceConfig holds retry and circuit breaker configuration.
// This provides standardized resilience settings that can be used across
// all packages that implement retry and circuit breaker patterns.
type ResilienceConfig struct {
	// Retry settings
	MaxRetries   int           `mapstructure:"max_retries"`
	InitialDelay time.Duration `mapstructure:"initial_delay"`
	MaxDelay     time.Duration `mapstructure:"max_delay"`
	Multiplier   float64       `mapstructure:"multiplier"`

	// Circuit breaker settings
	MaxRequests      uint32        `mapstructure:"max_requests"`
	Interval         time.Duration `mapstructure:"interval"`
	Timeout          time.Duration `mapstructure:"timeout"`
	FailureThreshold float64       `mapstructure:"failure_threshold"`
}

// ResilienceConfigFromViper creates a ResilienceConfig from a Standard config loader.
//
// Environment variable mappings:
//   - RESILIENCE_MAX_RETRIES -> max_retries (default: 3)
//   - RESILIENCE_INITIAL_DELAY -> initial_delay (default: 1s)
//   - RESILIENCE_MAX_DELAY -> max_delay (default: 30s)
//   - RESILIENCE_MULTIPLIER -> multiplier (default: 2.0)
//   - RESILIENCE_MAX_REQUESTS -> max_requests (default: 10)
//   - RESILIENCE_INTERVAL -> interval (default: 10s)
//   - RESILIENCE_TIMEOUT -> timeout (default: 60s)
//   - RESILIENCE_FAILURE_THRESHOLD -> failure_threshold (default: 0.6)
func ResilienceConfigFromViper(s *Standard) ResilienceConfig {
	// Bind environment variables
	_ = s.BindEnv("resilience.max_retries", "RESILIENCE_MAX_RETRIES")
	_ = s.BindEnv("resilience.initial_delay", "RESILIENCE_INITIAL_DELAY")
	_ = s.BindEnv("resilience.max_delay", "RESILIENCE_MAX_DELAY")
	_ = s.BindEnv("resilience.multiplier", "RESILIENCE_MULTIPLIER")
	_ = s.BindEnv("resilience.max_requests", "RESILIENCE_MAX_REQUESTS")
	_ = s.BindEnv("resilience.interval", "RESILIENCE_INTERVAL")
	_ = s.BindEnv("resilience.timeout", "RESILIENCE_TIMEOUT")
	_ = s.BindEnv("resilience.failure_threshold", "RESILIENCE_FAILURE_THRESHOLD")

	config := ResilienceConfig{
		MaxRetries:       s.GetInt("resilience.max_retries"),
		InitialDelay:     s.viper.GetDuration("resilience.initial_delay"),
		MaxDelay:         s.viper.GetDuration("resilience.max_delay"),
		Multiplier:       s.viper.GetFloat64("resilience.multiplier"),
		MaxRequests:      s.viper.GetUint32("resilience.max_requests"),
		Interval:         s.viper.GetDuration("resilience.interval"),
		Timeout:          s.viper.GetDuration("resilience.timeout"),
		FailureThreshold: s.viper.GetFloat64("resilience.failure_threshold"),
	}

	// Apply defaults
	config.setDefaults()

	return config
}

// setDefaults sets default values for optional fields
func (c *ResilienceConfig) setDefaults() {
	// Retry defaults
	if c.MaxRetries == 0 {
		c.MaxRetries = 3
	}
	if c.InitialDelay == 0 {
		c.InitialDelay = 1 * time.Second
	}
	if c.MaxDelay == 0 {
		c.MaxDelay = 30 * time.Second
	}
	if c.Multiplier == 0 {
		c.Multiplier = 2.0
	}

	// Circuit breaker defaults
	if c.MaxRequests == 0 {
		c.MaxRequests = 10
	}
	if c.Interval == 0 {
		c.Interval = 10 * time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = 60 * time.Second
	}
	if c.FailureThreshold == 0 {
		c.FailureThreshold = 0.6
	}
}

// Validate validates the resilience configuration
func (c *ResilienceConfig) Validate() error {
	// Validate retry settings
	if err := ValidateRange("resilience.max_retries", c.MaxRetries, 0, 10); err != nil {
		return err
	}
	if err := ValidateDuration("resilience.initial_delay", c.InitialDelay); err != nil {
		return err
	}
	if err := ValidateDuration("resilience.max_delay", c.MaxDelay); err != nil {
		return err
	}
	if c.MaxDelay < c.InitialDelay {
		return fmt.Errorf("resilience.max_delay (%v) must be greater than or equal to initial_delay (%v)",
			c.MaxDelay, c.InitialDelay)
	}
	if err := ValidateRange("resilience.multiplier", c.Multiplier, 1.0, 10.0); err != nil {
		return err
	}

	// Validate circuit breaker settings
	if err := ValidatePositive("resilience.max_requests", int(c.MaxRequests)); err != nil {
		return err
	}
	if err := ValidateDuration("resilience.interval", c.Interval); err != nil {
		return err
	}
	if err := ValidateDuration("resilience.timeout", c.Timeout); err != nil {
		return err
	}
	if err := ValidateRange("resilience.failure_threshold", c.FailureThreshold, 0.0, 1.0); err != nil {
		return err
	}

	return nil
}
