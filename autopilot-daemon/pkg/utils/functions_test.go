package utils

import (
	"testing"
	"time"
)

// TestParseDuration tests the ParseDuration function for various valid and invalid inputs.
func TestParseInterval(t *testing.T) {
	// Test valid durations
	validDurations := []string{"1h30m", "45m", "2s", "1h", "30m", "2h15m30s", "0"}
	validResults := []time.Duration{time.Hour + 30*time.Minute, 45 * time.Minute, 2 * time.Second, time.Hour, 30 * time.Minute, 2*time.Hour + 15*time.Minute + 30*time.Second, 0}
	for i, duration := range validDurations {
		result, err := ParseInterval(duration)
		if err != nil {
			t.Errorf("Expected no error for %q, got %v", duration, err)
		}
		if result != validResults[i] {
			t.Errorf("Expected %v for %q, got %v", validResults[i], duration, result)
		}
	}

	// Test invalid durations
	invalidDurations := []string{"1h2x", "abc", "123"}
	for _, duration := range invalidDurations {
		_, err := ParseInterval(duration)
		if err == nil {
			t.Errorf("Expected error for %q, got nil", duration)
		}
	}
}
