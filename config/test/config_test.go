package config_test

import (
	"fiber-boiler-plate/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_ENV", "test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("JWT_PRIVATE_KEY_PATH", "./keys/test_private.pem")
	os.Setenv("JWT_PUBLIC_KEY_PATH", "./keys/test_public.pem")
	os.Setenv("JWT_PRIVATE_KEY_ROTATION_PATH", "./keys/test_private_rotation.pem")
	os.Setenv("JWT_PUBLIC_KEY_ROTATION_PATH", "./keys/test_public_rotation.pem")
	os.Setenv("JWT_EXPIRE_HOURS", "2")

	cfg := config.LoadConfig()

	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "8080", cfg.App.Port)
	assert.Equal(t, "test", cfg.App.Env)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "./keys/test_private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "./keys/test_public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, "./keys/test_private_rotation.pem", cfg.JWT.PrivateKeyRotationPath)
	assert.Equal(t, "./keys/test_public_rotation.pem", cfg.JWT.PublicKeyRotationPath)
	assert.Equal(t, 2, cfg.JWT.ExpireHours)
}

func TestLoadConfig_DatabaseBooleanValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("DB_AUTO_MIGRATE", "true")
	os.Setenv("DB_RUN_SEEDER", "false")
	os.Setenv("DB_SEED_USERS", "true")
	os.Setenv("DB_MIGRATE_ON_START", "false")

	cfg := config.LoadConfig()

	assert.True(t, cfg.Database.AutoMigrate)
	assert.False(t, cfg.Database.RunSeeder)
	assert.True(t, cfg.Database.SeedUsers)
	assert.False(t, cfg.Database.MigrateOnStart)
}

func TestLoadConfig_JWTConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("JWT_PRIVATE_KEY_PATH", "./keys/test_private.pem")
	os.Setenv("JWT_PUBLIC_KEY_PATH", "./keys/test_public.pem")
	os.Setenv("JWT_PRIVATE_KEY_ROTATION_PATH", "./keys/test_private_rotation.pem")
	os.Setenv("JWT_PUBLIC_KEY_ROTATION_PATH", "./keys/test_public_rotation.pem")
	os.Setenv("JWT_EXPIRE_HOURS", "2")
	os.Setenv("REFRESH_TOKEN_EXPIRE_HOURS", "48")

	cfg := config.LoadConfig()

	assert.Equal(t, "./keys/test_private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "./keys/test_public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, "./keys/test_private_rotation.pem", cfg.JWT.PrivateKeyRotationPath)
	assert.Equal(t, "./keys/test_public_rotation.pem", cfg.JWT.PublicKeyRotationPath)
	assert.Equal(t, 2, cfg.JWT.ExpireHours)
	assert.Equal(t, 48, cfg.JWT.RefreshTokenExpireHours)
}

func TestLoadConfig_ResendConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("RESEND_API_KEY", "re_test_api_key")
	os.Setenv("RESEND_FROM_EMAIL", "noreply@test.com")
	os.Setenv("RESEND_FROM_NAME", "Test Service")

	cfg := config.LoadConfig()

	assert.Equal(t, "re_test_api_key", cfg.Resend.APIKey)
	assert.Equal(t, "noreply@test.com", cfg.Resend.FromEmail)
	assert.Equal(t, "Test Service", cfg.Resend.FromName)
}

func TestConfig_StructureValidation(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Name: "test-app",
			Port: "3000",
			Env:  "development",
		},
		Database: config.DatabaseConfig{
			Host:           "localhost",
			Port:           "3306",
			User:           "root",
			Password:       "admin",
			Name:           "testdb",
			AutoMigrate:    true,
			RunSeeder:      false,
			SeedUsers:      true,
			MigrateOnStart: true,
		},
		JWT: config.JWTConfig{
			PrivateKeyPath:          "./keys/private.pem",
			PublicKeyPath:           "./keys/public.pem",
			PrivateKeyRotationPath:  "./keys/private_rotation.pem",
			PublicKeyRotationPath:   "./keys/public_rotation.pem",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		Resend: config.ResendConfig{
			APIKey:    "re_test_api_key",
			FromEmail: "noreply@test.com",
			FromName:  "Test Service",
		},
	}

	assert.NotNil(t, cfg)
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "3000", cfg.App.Port)
	assert.Equal(t, "development", cfg.App.Env)

	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "3306", cfg.Database.Port)
	assert.True(t, cfg.Database.AutoMigrate)
	assert.False(t, cfg.Database.RunSeeder)

	assert.Equal(t, "./keys/private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "./keys/public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, "./keys/private_rotation.pem", cfg.JWT.PrivateKeyRotationPath)
	assert.Equal(t, "./keys/public_rotation.pem", cfg.JWT.PublicKeyRotationPath)
	assert.Equal(t, 1, cfg.JWT.ExpireHours)
	assert.Equal(t, 24, cfg.JWT.RefreshTokenExpireHours)

	assert.Equal(t, "re_test_api_key", cfg.Resend.APIKey)
	assert.Equal(t, "noreply@test.com", cfg.Resend.FromEmail)
	assert.Equal(t, "Test Service", cfg.Resend.FromName)
}
