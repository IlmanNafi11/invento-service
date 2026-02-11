package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetEnv tests the getEnv helper function
func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		setupEnv     func()
		expected     string
	}{
		{
			name:         "returns environment value when set",
			key:          "TEST_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("TEST_VAR", "env_value") },
			expected:     "env_value",
		},
		{
			name:         "returns default value when env not set",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default_value",
			setupEnv:     func() { os.Unsetenv("NONEXISTENT_VAR") },
			expected:     "default_value",
		},
		{
			name:         "returns default value when env is empty string",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("EMPTY_VAR", "") },
			expected:     "default",
		},
		{
			name:         "handles special characters in value",
			key:          "SPECIAL_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("SPECIAL_VAR", "value@#$%^&*()") },
			expected:     "value@#$%^&*()",
		},
		{
			name:         "handles numeric string value",
			key:          "NUMERIC_VAR",
			defaultValue: "0",
			setupEnv:     func() { os.Setenv("NUMERIC_VAR", "12345") },
			expected:     "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer os.Unsetenv(tt.key)

			// Execute
			result := getEnv(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetEnvAsInt tests the getEnvAsInt helper function
func TestGetEnvAsInt(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int
		setupEnv     func()
		expected     int
	}{
		{
			name:         "returns valid integer from environment",
			key:          "INT_VAR",
			defaultValue: 0,
			setupEnv:     func() { os.Setenv("INT_VAR", "42") },
			expected:     42,
		},
		{
			name:         "returns default value when env not set",
			key:          "NONEXISTENT_INT",
			defaultValue: 100,
			setupEnv:     func() { os.Unsetenv("NONEXISTENT_INT") },
			expected:     100,
		},
		{
			name:         "returns default value for invalid integer",
			key:          "INVALID_INT",
			defaultValue: 50,
			setupEnv:     func() { os.Setenv("INVALID_INT", "not_a_number") },
			expected:     50,
		},
		{
			name:         "returns default value for empty string",
			key:          "EMPTY_INT",
			defaultValue: 75,
			setupEnv:     func() { os.Setenv("EMPTY_INT", "") },
			expected:     75,
		},
		{
			name:         "handles negative numbers",
			key:          "NEGATIVE_INT",
			defaultValue: 0,
			setupEnv:     func() { os.Setenv("NEGATIVE_INT", "-100") },
			expected:     -100,
		},
		{
			name:         "handles zero value",
			key:          "ZERO_INT",
			defaultValue: 999,
			setupEnv:     func() { os.Setenv("ZERO_INT", "0") },
			expected:     0,
		},
		{
			name:         "handles large positive numbers",
			key:          "LARGE_INT",
			defaultValue: 0,
			setupEnv:     func() { os.Setenv("LARGE_INT", "2147483647") },
			expected:     2147483647,
		},
		{
			name:         "handles decimal string (returns default)",
			key:          "DECIMAL_INT",
			defaultValue: 10,
			setupEnv:     func() { os.Setenv("DECIMAL_INT", "3.14") },
			expected:     10,
		},
		{
			name:         "handles whitespace string (returns default)",
			key:          "WHITESPACE_INT",
			defaultValue: 20,
			setupEnv:     func() { os.Setenv("WHITESPACE_INT", "   ") },
			expected:     20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer os.Unsetenv(tt.key)

			// Execute
			result := getEnvAsInt(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetEnvAsBool tests the getEnvAsBool helper function
func TestGetEnvAsBool(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue bool
		setupEnv     func()
		expected     bool
	}{
		// True values
		{
			name:         "accepts lowercase 'true'",
			key:          "BOOL_VAR",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "true") },
			expected:     true,
		},
		{
			name:         "accepts uppercase 'TRUE'",
			key:          "BOOL_VAR",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "TRUE") },
			expected:     true,
		},
		{
			name:         "accepts mixed case 'True'",
			key:          "BOOL_VAR",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "True") },
			expected:     true,
		},
		{
			name:         "accepts '1' as true",
			key:          "BOOL_VAR",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "1") },
			expected:     true,
		},
		{
			name:         "accepts lowercase 't' as true",
			key:          "BOOL_VAR",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "t") },
			expected:     true,
		},
		{
			name:         "accepts uppercase 'T' as true",
			key:          "BOOL_VAR",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "T") },
			expected:     true,
		},
		// False values
		{
			name:         "accepts lowercase 'false'",
			key:          "BOOL_VAR",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "false") },
			expected:     false,
		},
		{
			name:         "accepts uppercase 'FALSE'",
			key:          "BOOL_VAR",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "FALSE") },
			expected:     false,
		},
		{
			name:         "accepts mixed case 'False'",
			key:          "BOOL_VAR",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "False") },
			expected:     false,
		},
		{
			name:         "accepts '0' as false",
			key:          "BOOL_VAR",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "0") },
			expected:     false,
		},
		{
			name:         "accepts lowercase 'f' as false",
			key:          "BOOL_VAR",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "f") },
			expected:     false,
		},
		{
			name:         "accepts uppercase 'F' as false",
			key:          "BOOL_VAR",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("BOOL_VAR", "F") },
			expected:     false,
		},
		// Default cases
		{
			name:         "returns default value when env not set",
			key:          "NONEXISTENT_BOOL",
			defaultValue: true,
			setupEnv:     func() { os.Unsetenv("NONEXISTENT_BOOL") },
			expected:     true,
		},
		{
			name:         "returns default value for invalid string",
			key:          "INVALID_BOOL",
			defaultValue: true,
			setupEnv:     func() { os.Setenv("INVALID_BOOL", "invalid") },
			expected:     true,
		},
		{
			name:         "returns default value for empty string",
			key:          "EMPTY_BOOL",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("EMPTY_BOOL", "") },
			expected:     false,
		},
		{
			name:         "returns default value for whitespace",
			key:          "WHITESPACE_BOOL",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("WHITESPACE_BOOL", "   ") },
			expected:     false,
		},
		{
			name:         "returns default value for '2' (invalid boolean)",
			key:          "NUMBER_BOOL",
			defaultValue: false,
			setupEnv:     func() { os.Setenv("NUMBER_BOOL", "2") },
			expected:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer os.Unsetenv(tt.key)

			// Execute
			result := getEnvAsBool(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetEnvAsInt64 tests the getEnvAsInt64 helper function
func TestGetEnvAsInt64(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue int64
		setupEnv     func()
		expected     int64
	}{
		{
			name:         "returns valid int64 from environment",
			key:          "INT64_VAR",
			defaultValue: 0,
			setupEnv:     func() { os.Setenv("INT64_VAR", "9223372036854775807") },
			expected:     9223372036854775807,
		},
		{
			name:         "returns default value when env not set",
			key:          "NONEXISTENT_INT64",
			defaultValue: 1000,
			setupEnv:     func() { os.Unsetenv("NONEXISTENT_INT64") },
			expected:     1000,
		},
		{
			name:         "returns default value for invalid integer",
			key:          "INVALID_INT64",
			defaultValue: 500,
			setupEnv:     func() { os.Setenv("INVALID_INT64", "not_a_number") },
			expected:     500,
		},
		{
			name:         "returns default value for empty string",
			key:          "EMPTY_INT64",
			defaultValue: 750,
			setupEnv:     func() { os.Setenv("EMPTY_INT64", "") },
			expected:     750,
		},
		{
			name:         "handles negative numbers",
			key:          "NEGATIVE_INT64",
			defaultValue: 0,
			setupEnv:     func() { os.Setenv("NEGATIVE_INT64", "-9223372036854775808") },
			expected:     -9223372036854775808,
		},
		{
			name:         "handles zero value",
			key:          "ZERO_INT64",
			defaultValue: 99999,
			setupEnv:     func() { os.Setenv("ZERO_INT64", "0") },
			expected:     0,
		},
		{
			name:         "handles positive numbers",
			key:          "POSITIVE_INT64",
			defaultValue: 0,
			setupEnv:     func() { os.Setenv("POSITIVE_INT64", "524288000") },
			expected:     524288000,
		},
		{
			name:         "returns default for overflow (beyond int64 range)",
			key:          "OVERFLOW_INT64",
			defaultValue: 100,
			setupEnv:     func() { os.Setenv("OVERFLOW_INT64", "9223372036854775808") },
			expected:     100,
		},
		{
			name:         "returns default for decimal string",
			key:          "DECIMAL_INT64",
			defaultValue: 200,
			setupEnv:     func() { os.Setenv("DECIMAL_INT64", "3.14") },
			expected:     200,
		},
		{
			name:         "returns default for whitespace string",
			key:          "WHITESPACE_INT64",
			defaultValue: 300,
			setupEnv:     func() { os.Setenv("WHITESPACE_INT64", "   ") },
			expected:     300,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer os.Unsetenv(tt.key)

			// Execute
			result := getEnvAsInt64(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestGetEnvAllowEmpty tests the getEnvAllowEmpty helper function
func TestGetEnvAllowEmpty(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		defaultValue string
		setupEnv     func()
		expected     string
	}{
		{
			name:         "returns environment value when set and not empty",
			key:          "ALLOWED_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("ALLOWED_VAR", "actual_value") },
			expected:     "actual_value",
		},
		{
			name:         "returns empty string when env is explicitly set to empty",
			key:          "EMPTY_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("EMPTY_VAR", "") },
			expected:     "",
		},
		{
			name:         "returns default value when env is not set",
			key:          "NONEXISTENT_VAR",
			defaultValue: "default_value",
			setupEnv:     func() { os.Unsetenv("NONEXISTENT_VAR") },
			expected:     "default_value",
		},
		{
			name:         "returns environment value with whitespace",
			key:          "WHITESPACE_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("WHITESPACE_VAR", "   value   ") },
			expected:     "   value   ",
		},
		{
			name:         "returns empty string when env is set to whitespace",
			key:          "SPACE_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("SPACE_VAR", "   ") },
			expected:     "   ",
		},
		{
			name:         "handles special characters",
			key:          "SPECIAL_VAR",
			defaultValue: "default",
			setupEnv:     func() { os.Setenv("SPECIAL_VAR", "!@#$%^&*()") },
			expected:     "!@#$%^&*()",
		},
		{
			name:         "handles numeric string",
			key:          "NUMERIC_VAR",
			defaultValue: "0",
			setupEnv:     func() { os.Setenv("NUMERIC_VAR", "12345") },
			expected:     "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			tt.setupEnv()
			defer os.Unsetenv(tt.key)

			// Execute
			result := getEnvAllowEmpty(tt.key, tt.defaultValue)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestMaskToken tests the maskToken helper function
func TestMaskToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{
			name:     "masks token shorter than 10 characters",
			token:    "short",
			expected: "***",
		},
		{
			name:     "masks token exactly 10 characters",
			token:    "1234567890",
			expected: "***",
		},
		{
			name:     "masks token longer than 10 characters",
			token:    "re_test_api_key_value",
			expected: "re_test_ap...",
		},
		{
			name:     "handles empty string",
			token:    "",
			expected: "***",
		},
		{
			name:     "handles token with 11 characters",
			token:    "abc123def45",
			expected: "abc123def4...",
		},
		{
			name:     "handles very long token",
			token:    "this_is_a_very_long_api_token_that_should_be_masked_for_security",
			expected: "this_is_a_...",
		},
		{
			name:     "handles token with special characters",
			token:    "sk-1234567890abcdef",
			expected: "sk-1234567...",
		},
		{
			name:     "handles token with spaces",
			token:    "token with spaces inside",
			expected: "token with...",
		},
		{
			name:     "handles token with unicode characters",
			token:    "re_test_🔑api_key",
			expected: "re_test_\xf0\x9f...", // First 10 bytes, not characters
		},
		{
			name:     "handles token exactly 11 characters",
			token:    "12345678901",
			expected: "1234567890...",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Execute
			result := maskToken(tt.token)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}
