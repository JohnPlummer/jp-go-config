package config

import (
	"fmt"
	"time"
)

// ValidateRequired validates that a string field is not empty
func ValidateRequired(field, value string) error {
	if value == "" {
		return fmt.Errorf("%s is required", field)
	}
	return nil
}

// ValidatePort validates that a port number is in the valid range (1-65535)
func ValidatePort(field string, port int) error {
	if port < 1 || port > 65535 {
		return fmt.Errorf("%s must be between 1 and 65535, got %d", field, port)
	}
	return nil
}

// ValidateDuration validates that a duration is positive
func ValidateDuration(field string, duration time.Duration) error {
	if duration < 0 {
		return fmt.Errorf("%s must be positive, got %v", field, duration)
	}
	return nil
}

// ValidatePositive validates that an integer is positive (> 0)
func ValidatePositive(field string, value int) error {
	if value <= 0 {
		return fmt.Errorf("%s must be positive, got %d", field, value)
	}
	return nil
}

// ValidateRange validates that a value is within a range (inclusive)
func ValidateRange[T int | float64](field string, value, min, max T) error {
	if value < min || value > max {
		return fmt.Errorf("%s must be between %v and %v, got %v", field, min, max, value)
	}
	return nil
}
