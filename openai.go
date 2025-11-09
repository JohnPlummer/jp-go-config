package config

import (
	"time"
)

// OpenAIConfig holds OpenAI API configuration
type OpenAIConfig struct {
	APIKey      string        `mapstructure:"api_key"`
	Model       string        `mapstructure:"model"`
	Temperature float64       `mapstructure:"temperature"`
	MaxTokens   int           `mapstructure:"max_tokens"`
	Timeout     time.Duration `mapstructure:"timeout"`
}

// OpenAIConfigFromViper creates an OpenAIConfig from a Standard config loader.
//
// Environment variable mappings:
//   - OPENAI_API_KEY -> api_key (required)
//   - OPENAI_MODEL -> model (default: gpt-3.5-turbo)
//   - OPENAI_TEMPERATURE -> temperature (default: 0.7)
//   - OPENAI_MAX_TOKENS -> max_tokens (default: 2000)
//   - OPENAI_TIMEOUT -> timeout (default: 30s)
func OpenAIConfigFromViper(s *Standard) OpenAIConfig {
	// Bind environment variables
	_ = s.BindEnv("openai.api_key", "OPENAI_API_KEY")
	_ = s.BindEnv("openai.model", "OPENAI_MODEL")
	_ = s.BindEnv("openai.temperature", "OPENAI_TEMPERATURE")
	_ = s.BindEnv("openai.max_tokens", "OPENAI_MAX_TOKENS")
	_ = s.BindEnv("openai.timeout", "OPENAI_TIMEOUT")

	config := OpenAIConfig{
		APIKey:      s.GetString("openai.api_key"),
		Model:       s.GetString("openai.model"),
		Temperature: s.viper.GetFloat64("openai.temperature"),
		MaxTokens:   s.GetInt("openai.max_tokens"),
		Timeout:     s.viper.GetDuration("openai.timeout"),
	}

	// Apply defaults
	config.setDefaults()

	return config
}

// setDefaults sets default values for optional fields
func (c *OpenAIConfig) setDefaults() {
	if c.Model == "" {
		c.Model = "gpt-3.5-turbo"
	}
	if c.Temperature == 0 {
		c.Temperature = 0.7
	}
	if c.MaxTokens == 0 {
		c.MaxTokens = 2000
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
}

// Validate validates the OpenAI configuration
func (c *OpenAIConfig) Validate() error {
	if err := ValidateRequired("openai.api_key", c.APIKey); err != nil {
		return err
	}
	if err := ValidateRequired("openai.model", c.Model); err != nil {
		return err
	}
	if err := ValidateRange("openai.temperature", c.Temperature, 0.0, 2.0); err != nil {
		return err
	}
	if err := ValidatePositive("openai.max_tokens", c.MaxTokens); err != nil {
		return err
	}
	if err := ValidateDuration("openai.timeout", c.Timeout); err != nil {
		return err
	}

	return nil
}
