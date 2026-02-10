package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type Config struct {
	App     AppConfig
	Database DatabaseConfig
	JWT      JWTConfig
	Upload   UploadConfig
	Resend   ResendConfig
	OTP      OTPConfig
	Logging  LoggingConfig
	Swagger  SwaggerConfig
}

type AppConfig struct {
	Name           string
	Port           string
	Env            string
	CorsOriginDev  string
	CorsOriginProd string
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
	PrivateKeyPath          string
	PublicKeyPath           string
	PrivateKeyRotationPath  string
	PublicKeyRotationPath   string
	ExpireHours             int
	RefreshTokenExpireHours int
	KeyRotationHours        int
}

type ResendConfig struct {
	APIKey    string
	FromEmail string
	FromName  string
}

type OTPConfig struct {
	Length                int
	ExpiryMinutes         int
	MaxAttempts           int
	ResendCooldownSeconds int
	ResendMaxTimes        int
}

type UploadConfig struct {
	MaxSize              int64
	MaxSizeProject       int64
	MaxSizeModul         int64
	ChunkSize            int64
	MaxConcurrent        int
	MaxConcurrentProject int
	MaxConcurrentModul   int
	MaxQueueModulPerUser int
	IdleTimeout          int
	CleanupInterval      int
	PathProduction       string
	PathDevelopment      string
	TempPathProduction   string
	TempPathDevelopment  string
	TusVersion           string
	MaxResumeAttempts    int
}

type LoggingConfig struct {
	Level         string
	Format        string
	LogRequestBody bool
}

type SwaggerConfig struct {
	Enabled bool
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Tidak dapat memuat file .env, menggunakan environment variables")
	}

	viper.AutomaticEnv()

	config := &Config{
		App: AppConfig{
			Name:           getEnv("APP_NAME", "Fiber Boilerplate"),
			Port:           getEnv("APP_PORT", "3000"),
			Env:            getEnv("APP_ENV", "development"),
			CorsOriginDev:  getEnv("CORS_ORIGIN_DEVELOPMENT", "http://localhost:5173"),
			CorsOriginProd: getEnv("CORS_ORIGIN_PRODUCTION", "https://yourdomain.com"),
		},
		Database: DatabaseConfig{
			Host:           getEnv("DB_HOST", "localhost"),
			Port:           getEnv("DB_PORT", "3306"),
			User:           getEnv("DB_USER", "root"),
			Password:       getEnvAllowEmpty("DB_PASSWORD", "admin"),
			Name:           getEnv("DB_NAME", "fiber_boilerplate"),
			AutoMigrate:    getEnvAsBool("DB_AUTO_MIGRATE", true),
			RunSeeder:      getEnvAsBool("DB_RUN_SEEDER", false),
			SeedUsers:      getEnvAsBool("DB_SEED_USERS", false),
			MigrateOnStart: getEnvAsBool("DB_MIGRATE_ON_START", true),
		},
		JWT: JWTConfig{
			PrivateKeyPath:          getEnv("JWT_PRIVATE_KEY_PATH", "./keys/private.pem"),
			PublicKeyPath:           getEnv("JWT_PUBLIC_KEY_PATH", "./keys/public.pem"),
			PrivateKeyRotationPath:  getEnv("JWT_PRIVATE_KEY_ROTATION_PATH", "./keys/private_rotation.pem"),
			PublicKeyRotationPath:   getEnv("JWT_PUBLIC_KEY_ROTATION_PATH", "./keys/public_rotation.pem"),
			ExpireHours:             getEnvAsInt("JWT_EXPIRE_HOURS", 1),
			RefreshTokenExpireHours: getEnvAsInt("REFRESH_TOKEN_EXPIRE_HOURS", 168),
			KeyRotationHours:        getEnvAsInt("JWT_KEY_ROTATION_HOURS", 168),
		},
		Upload: UploadConfig{
			MaxSize:              getEnvAsInt64("UPLOAD_MAX_SIZE", 524288000),
			MaxSizeProject:       getEnvAsInt64("UPLOAD_MAX_SIZE_PROJECT", 524288000),
			MaxSizeModul:         getEnvAsInt64("UPLOAD_MAX_SIZE_MODUL", 52428800),
			ChunkSize:            getEnvAsInt64("UPLOAD_CHUNK_SIZE", 1048576),
			MaxConcurrent:        getEnvAsInt("UPLOAD_MAX_CONCURRENT", 1),
			MaxConcurrentProject: getEnvAsInt("UPLOAD_MAX_CONCURRENT_PROJECT", 1),
			MaxConcurrentModul:   getEnvAsInt("UPLOAD_MAX_CONCURRENT_MODUL", 1),
			MaxQueueModulPerUser: getEnvAsInt("UPLOAD_MAX_QUEUE_MODUL_PER_USER", 5),
			IdleTimeout:          getEnvAsInt("UPLOAD_IDLE_TIMEOUT", 600),
			CleanupInterval:      getEnvAsInt("UPLOAD_CLEANUP_INTERVAL", 300),
			PathProduction:       getEnv("UPLOAD_PATH_PRODUCTION", "/volume1/data-invento/"),
			PathDevelopment:      getEnv("UPLOAD_PATH_DEVELOPMENT", "./uploads/"),
			TempPathProduction:   getEnv("UPLOAD_TEMP_PATH_PRODUCTION", "/volume1/data-invento/temp/"),
			TempPathDevelopment:  getEnv("UPLOAD_TEMP_PATH_DEVELOPMENT", "./uploads/temp/"),
			TusVersion:           getEnv("TUS_RESUMABLE_VERSION", "1.0.0"),
			MaxResumeAttempts:    getEnvAsInt("TUS_MAX_RESUME_ATTEMPTS", 10),
		},
		Resend: ResendConfig{
			APIKey:    getEnv("RESEND_API_KEY", ""),
			FromEmail: getEnv("RESEND_FROM_EMAIL", "noreply@example.com"),
			FromName:  getEnv("RESEND_FROM_NAME", "Invento Service"),
		},
		OTP: OTPConfig{
			Length:                getEnvAsInt("OTP_LENGTH", 6),
			ExpiryMinutes:         getEnvAsInt("OTP_EXPIRY_MINUTES", 10),
			MaxAttempts:           getEnvAsInt("OTP_MAX_ATTEMPTS", 5),
			ResendCooldownSeconds: getEnvAsInt("OTP_RESEND_COOLDOWN_SECONDS", 60),
			ResendMaxTimes:        getEnvAsInt("OTP_RESEND_MAX_TIMES", 5),
		},
		Logging: LoggingConfig{
			Level:         getEnv("LOG_LEVEL", "INFO"),
			Format:        getEnv("LOG_FORMAT", "text"),
			LogRequestBody: getEnvAsBool("LOG_REQUEST_BODY", false),
		},
		Swagger: SwaggerConfig{
			Enabled: getEnvAsBool("SWAGGER_ENABLED", false),
		},
	}

	// Log config values for debugging
	log.Printf("[CONFIG] Resend loaded: APIKey=%s, FromEmail=%s, FromName=%s",
		maskToken(config.Resend.APIKey),
		config.Resend.FromEmail,
		config.Resend.FromName)
	log.Printf("[CONFIG] OTP: Length=%d, ExpiryMinutes=%d, MaxAttempts=%d, ResendCooldown=%d, ResendMaxTimes=%d",
		config.OTP.Length,
		config.OTP.ExpiryMinutes,
		config.OTP.MaxAttempts,
		config.OTP.ResendCooldownSeconds,
		config.OTP.ResendMaxTimes)

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

func getEnvAsInt64(key string, defaultValue int64) int64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseInt(valueStr, 10, 64); err == nil {
		return value
	}
	return defaultValue
}

func maskToken(token string) string {
	if len(token) <= 10 {
		return "***"
	}
	return token[:10] + "..."
}

func getEnvAllowEmpty(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}
