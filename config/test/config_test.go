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
	os.Setenv("JWT_SECRET", "testsecret")

	cfg := config.LoadConfig()

	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "8080", cfg.App.Port)
	assert.Equal(t, "test", cfg.App.Env)
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.Equal(t, "testsecret", cfg.JWT.Secret)
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
	os.Setenv("JWT_SECRET", "test-jwt-secret")
	os.Setenv("JWT_EXPIRE_HOURS", "2")
	os.Setenv("JWT_REFRESH_TOKEN_EXPIRE_HOURS", "48")

	cfg := config.LoadConfig()

	assert.Equal(t, "test-jwt-secret", cfg.JWT.Secret)
	assert.Equal(t, 2, cfg.JWT.ExpireHours)
	assert.Equal(t, 168, cfg.JWT.RefreshTokenExpireHours)
}

func TestLoadConfig_MailConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("MAIL_HOST", "smtp.test.com")
	os.Setenv("MAIL_PORT", "587")
	os.Setenv("MAIL_USERNAME", "test@test.com")
	os.Setenv("MAIL_PASSWORD", "mailpass")
	os.Setenv("MAIL_FROM", "noreply@test.com")

	cfg := config.LoadConfig()

	assert.Equal(t, "smtp.test.com", cfg.Mail.Host)
	assert.Equal(t, "587", cfg.Mail.Port)
	assert.Equal(t, "test@test.com", cfg.Mail.Username)
	assert.Equal(t, "mailpass", cfg.Mail.Password)
	assert.Equal(t, "noreply@test.com", cfg.Mail.From)
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
			Secret:                  "secret",
			ExpireHours:             1,
			RefreshTokenExpireHours: 24,
		},
		Mail: config.MailConfig{
			Host:     "smtp.gmail.com",
			Port:     "587",
			Username: "test@gmail.com",
			Password: "password",
			From:     "noreply@test.com",
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

	assert.Equal(t, "secret", cfg.JWT.Secret)
	assert.Equal(t, 1, cfg.JWT.ExpireHours)
	assert.Equal(t, 24, cfg.JWT.RefreshTokenExpireHours)

	assert.Equal(t, "smtp.gmail.com", cfg.Mail.Host)
	assert.Equal(t, "587", cfg.Mail.Port)
}
