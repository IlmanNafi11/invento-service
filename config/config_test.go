package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidate_BothPresent(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Supabase: SupabaseConfig{
			URL:        "https://example.supabase.co",
			ServiceKey: "service-role-key-123",
		},
	}
	err := cfg.Validate()
	assert.NoError(t, err)
}

func TestValidate_MissingURL(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Supabase: SupabaseConfig{
			URL:        "",
			ServiceKey: "service-role-key-123",
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SUPABASE_URL")
	assert.NotContains(t, err.Error(), "SUPABASE_SERVICE_ROLE_KEY")
}

func TestValidate_MissingServiceKey(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Supabase: SupabaseConfig{
			URL:        "https://example.supabase.co",
			ServiceKey: "",
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SUPABASE_SERVICE_ROLE_KEY")
	assert.NotContains(t, err.Error(), "SUPABASE_URL")
}

func TestValidate_BothMissing(t *testing.T) {
	t.Parallel()
	cfg := &Config{
		Supabase: SupabaseConfig{
			URL:        "",
			ServiceKey: "",
		},
	}
	err := cfg.Validate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "SUPABASE_URL")
	assert.Contains(t, err.Error(), "SUPABASE_SERVICE_ROLE_KEY")
}

func TestParseMemLimit_MiB(t *testing.T) {
	t.Parallel()
	val, err := ParseMemLimit("350MiB")
	assert.NoError(t, err)
	assert.Equal(t, int64(350*1024*1024), val)
}

func TestParseMemLimit_GiB(t *testing.T) {
	t.Parallel()
	val, err := ParseMemLimit("1GiB")
	assert.NoError(t, err)
	assert.Equal(t, int64(1*1024*1024*1024), val)
}

func TestParseMemLimit_RawBytes(t *testing.T) {
	t.Parallel()
	val, err := ParseMemLimit("1048576")
	assert.NoError(t, err)
	assert.Equal(t, int64(1048576), val)
}

func TestParseMemLimit_WithWhitespace(t *testing.T) {
	t.Parallel()
	val, err := ParseMemLimit("  512MiB  ")
	assert.NoError(t, err)
	assert.Equal(t, int64(512*1024*1024), val)
}

func TestParseMemLimit_InvalidInput(t *testing.T) {
	t.Parallel()
	_, err := ParseMemLimit("notanumber")
	assert.Error(t, err)
}

func TestParseMemLimit_EmptyString(t *testing.T) {
	t.Parallel()
	_, err := ParseMemLimit("")
	assert.Error(t, err)
}

func TestParseMemLimit_InvalidMiB(t *testing.T) {
	t.Parallel()
	_, err := ParseMemLimit("abcMiB")
	assert.Error(t, err)
}

func TestParseMemLimit_InvalidGiB(t *testing.T) {
	t.Parallel()
	_, err := ParseMemLimit("xyzGiB")
	assert.Error(t, err)
}

func TestGetEnvAsFloat64_ValidValue(t *testing.T) {
	key := "TEST_FLOAT64_VALID"
	t.Setenv(key, "0.95")
	defer os.Unsetenv(key)

	result := getEnvAsFloat64(key, 0.5)
	assert.Equal(t, 0.95, result)
}

func TestGetEnvAsFloat64_InvalidValue(t *testing.T) {
	key := "TEST_FLOAT64_INVALID"
	t.Setenv(key, "not-a-float")
	defer os.Unsetenv(key)

	result := getEnvAsFloat64(key, 0.8)
	assert.Equal(t, 0.8, result)
}

func TestGetEnvAsFloat64_NotSet(t *testing.T) {
	result := getEnvAsFloat64("TEST_FLOAT64_UNSET_KEY_XYZ", 0.75)
	assert.Equal(t, 0.75, result)
}
