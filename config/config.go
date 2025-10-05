package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Mail     MailConfig
}

type AppConfig struct {
	Name string
	Port string
	Env  string
}

type DatabaseConfig struct {
	Host           string
	Port           string
	User           string
	Password       string
	Name           string
	AutoMigrate    bool
	RunSeeder      bool
	SeedUsers      bool
	MigrateOnStart bool
}

type JWTConfig struct {
	Secret                  string
	ExpireHours             int
	RefreshTokenExpireHours int
}

type MailConfig struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
}


func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak dapat memuat file .env, menggunakan environment variables")
	}

	viper.AutomaticEnv()

	config := &Config{
		App: AppConfig{
			Name: getEnv("APP_NAME", "Fiber Boilerplate"),
			Port: getEnv("APP_PORT", "3000"),
			Env:  getEnv("APP_ENV", "development"),
		},
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "3306"),
			User:           getEnv("DB_USER", "root"),
			Password:       getEnv("DB_PASSWORD", "admin"),
			Name:           getEnv("DB_NAME", "fiber_boilerplate"),
			AutoMigrate:    getEnvAsBool("DB_AUTO_MIGRATE", true),
			RunSeeder:      getEnvAsBool("DB_RUN_SEEDER", false),
			SeedUsers:      getEnvAsBool("DB_SEED_USERS", false),
			MigrateOnStart: getEnvAsBool("DB_MIGRATE_ON_START", true),
		},
		JWT: JWTConfig{
			Secret:                  getEnv("JWT_SECRET", "your-jwt-secret"),
			ExpireHours:             getEnvAsInt("JWT_EXPIRE_HOURS", 24),
			RefreshTokenExpireHours: getEnvAsInt("REFRESH_TOKEN_EXPIRE_HOURS", 168),
		},
		Mail: MailConfig{
			Host:     getEnv("MAIL_HOST", "localhost"),
			Port:     getEnv("MAIL_PORT", "587"),
			Username: getEnv("MAIL_USERNAME", ""),
			Password: getEnv("MAIL_PASSWORD", ""),
			From:     getEnv("MAIL_FROM", "noreply@example.com"),
		},
	}

	return config
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if value, err := strconv.Atoi(valueStr); err == nil {
		return value
	}
	return defaultValue
}

func getEnvAsBool(key string, defaultValue bool) bool {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseBool(valueStr); err == nil {
		return value
	}
	return defaultValue
}
