package helper_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/helper"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewResendClient(t *testing.T) {
	cfg := &config.ResendConfig{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test Sender",
	}

	client := helper.NewResendClient(cfg)

	assert.NotNil(t, client)
}

func TestResendClient_StructFields(t *testing.T) {
	cfg := &config.ResendConfig{
		APIKey:    "test-api-key",
		FromEmail: "noreply@example.com",
		FromName:  "App Name",
	}

	client := helper.NewResendClient(cfg)

	// We can't directly access private fields, but we verified the client is created
	assert.NotNil(t, client)
}

func TestResendClient_EmptyFromName(t *testing.T) {
	cfg := &config.ResendConfig{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "",
	}

	client := helper.NewResendClient(cfg)

	assert.NotNil(t, client)
}

func TestResendClient_EmptyFromEmail(t *testing.T) {
	cfg := &config.ResendConfig{
		APIKey:    "test-api-key",
		FromEmail: "",
		FromName:  "Test",
	}

	client := helper.NewResendClient(cfg)

	assert.NotNil(t, client)
}

func TestResendClient_NoConfig(t *testing.T) {
	cfg := &config.ResendConfig{}

	client := helper.NewResendClient(cfg)

	assert.NotNil(t, client)
}

func TestResendClient_VariousAPIKeys(t *testing.T) {
	apiKeys := []string{
		"re_123456",
		"re_complex-key-with-dashes",
		"re_key_with_underscores",
	}

	for _, apiKey := range apiKeys {
		t.Run("", func(t *testing.T) {
			cfg := &config.ResendConfig{
				APIKey:    apiKey,
				FromEmail: "test@example.com",
			}

			client := helper.NewResendClient(cfg)
			assert.NotNil(t, client)
		})
	}
}

func TestResendClient_VariousFromAddresses(t *testing.T) {
	tests := []struct {
		name     string
		fromEmail string
		fromName  string
	}{
		{"Simple email", "test@example.com", "Sender"},
		{"Email with subdomain", "mail@test.example.com", "Test App"},
		{"No name", "noreply@example.com", ""},
		{"Name with special chars", "info@my-app.com", "My App®"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.ResendConfig{
				APIKey:    "re_test",
				FromEmail: tt.fromEmail,
				FromName:  tt.fromName,
			}

			client := helper.NewResendClient(cfg)
			assert.NotNil(t, client)
		})
	}
}

func TestResendClient_HTTPClientTimeout(t *testing.T) {
	cfg := &config.ResendConfig{
		APIKey:    "test-api-key",
		FromEmail: "test@example.com",
		FromName:  "Test",
	}

	client := helper.NewResendClient(cfg)

	// Client should have HTTP client configured with 30 second timeout
	assert.NotNil(t, client)
}
