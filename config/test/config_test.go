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

func TestLoadConfig_OTPConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("OTP_LENGTH", "8")
	os.Setenv("OTP_EXPIRY_MINUTES", "15")
	os.Setenv("OTP_MAX_ATTEMPTS", "3")
	os.Setenv("OTP_RESEND_COOLDOWN_SECONDS", "30")
	os.Setenv("OTP_RESEND_MAX_TIMES", "3")

	cfg := config.LoadConfig()

	assert.Equal(t, 8, cfg.OTP.Length)
	assert.Equal(t, 15, cfg.OTP.ExpiryMinutes)
	assert.Equal(t, 3, cfg.OTP.MaxAttempts)
	assert.Equal(t, 30, cfg.OTP.ResendCooldownSeconds)
	assert.Equal(t, 3, cfg.OTP.ResendMaxTimes)
}

func TestLoadConfig_OTPConfiguration_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.Equal(t, 6, cfg.OTP.Length)
	assert.Equal(t, 10, cfg.OTP.ExpiryMinutes)
	assert.Equal(t, 5, cfg.OTP.MaxAttempts)
	assert.Equal(t, 60, cfg.OTP.ResendCooldownSeconds)
	assert.Equal(t, 5, cfg.OTP.ResendMaxTimes)
}

func TestLoadConfig_UploadConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("UPLOAD_MAX_SIZE", "1048576000")
	os.Setenv("UPLOAD_MAX_SIZE_PROJECT", "2097152000")
	os.Setenv("UPLOAD_MAX_SIZE_MODUL", "104857600")
	os.Setenv("UPLOAD_CHUNK_SIZE", "2097152")
	os.Setenv("UPLOAD_MAX_CONCURRENT", "2")
	os.Setenv("UPLOAD_MAX_CONCURRENT_PROJECT", "3")
	os.Setenv("UPLOAD_MAX_CONCURRENT_MODUL", "2")
	os.Setenv("UPLOAD_MAX_QUEUE_MODUL_PER_USER", "10")
	os.Setenv("UPLOAD_IDLE_TIMEOUT", "300")
	os.Setenv("UPLOAD_CLEANUP_INTERVAL", "600")
	os.Setenv("UPLOAD_PATH_PRODUCTION", "/data/uploads/")
	os.Setenv("UPLOAD_PATH_DEVELOPMENT", "./dev/uploads/")
	os.Setenv("UPLOAD_TEMP_PATH_PRODUCTION", "/data/temp/")
	os.Setenv("UPLOAD_TEMP_PATH_DEVELOPMENT", "./dev/temp/")
	os.Setenv("TUS_RESUMABLE_VERSION", "1.0.0")
	os.Setenv("TUS_MAX_RESUME_ATTEMPTS", "5")

	cfg := config.LoadConfig()

	assert.Equal(t, int64(1048576000), cfg.Upload.MaxSize)
	assert.Equal(t, int64(2097152000), cfg.Upload.MaxSizeProject)
	assert.Equal(t, int64(104857600), cfg.Upload.MaxSizeModul)
	assert.Equal(t, int64(2097152), cfg.Upload.ChunkSize)
	assert.Equal(t, 2, cfg.Upload.MaxConcurrent)
	assert.Equal(t, 3, cfg.Upload.MaxConcurrentProject)
	assert.Equal(t, 2, cfg.Upload.MaxConcurrentModul)
	assert.Equal(t, 10, cfg.Upload.MaxQueueModulPerUser)
	assert.Equal(t, 300, cfg.Upload.IdleTimeout)
	assert.Equal(t, 600, cfg.Upload.CleanupInterval)
	assert.Equal(t, "/data/uploads/", cfg.Upload.PathProduction)
	assert.Equal(t, "./dev/uploads/", cfg.Upload.PathDevelopment)
	assert.Equal(t, "/data/temp/", cfg.Upload.TempPathProduction)
	assert.Equal(t, "./dev/temp/", cfg.Upload.TempPathDevelopment)
	assert.Equal(t, "1.0.0", cfg.Upload.TusVersion)
	assert.Equal(t, 5, cfg.Upload.MaxResumeAttempts)
}

func TestLoadConfig_UploadConfiguration_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.Equal(t, int64(524288000), cfg.Upload.MaxSize)
	assert.Equal(t, int64(524288000), cfg.Upload.MaxSizeProject)
	assert.Equal(t, int64(52428800), cfg.Upload.MaxSizeModul)
	assert.Equal(t, int64(1048576), cfg.Upload.ChunkSize)
	assert.Equal(t, 1, cfg.Upload.MaxConcurrent)
	assert.Equal(t, 1, cfg.Upload.MaxConcurrentProject)
	assert.Equal(t, 1, cfg.Upload.MaxConcurrentModul)
	assert.Equal(t, 5, cfg.Upload.MaxQueueModulPerUser)
	assert.Equal(t, 600, cfg.Upload.IdleTimeout)
	assert.Equal(t, 300, cfg.Upload.CleanupInterval)
	assert.Equal(t, "/volume1/data-invento/", cfg.Upload.PathProduction)
	assert.Equal(t, "./uploads/", cfg.Upload.PathDevelopment)
	assert.Equal(t, "/volume1/data-invento/temp/", cfg.Upload.TempPathProduction)
	assert.Equal(t, "./uploads/temp/", cfg.Upload.TempPathDevelopment)
	assert.Equal(t, "1.0.0", cfg.Upload.TusVersion)
	assert.Equal(t, 10, cfg.Upload.MaxResumeAttempts)
}

func TestLoadConfig_LoggingConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_REQUEST_BODY", "true")

	cfg := config.LoadConfig()

	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.True(t, cfg.Logging.LogRequestBody)
}

func TestLoadConfig_LoggingConfiguration_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.Equal(t, "INFO", cfg.Logging.Level)
	assert.Equal(t, "text", cfg.Logging.Format)
	assert.False(t, cfg.Logging.LogRequestBody)
}

func TestLoadConfig_LoggingConfiguration_BooleanVariations(t *testing.T) {
	tests := []struct {
		name           string
		logRequestBody string
		expectedResult bool
	}{
		{
			name:           "LOG_REQUEST_BODY with lowercase true",
			logRequestBody: "true",
			expectedResult: true,
		},
		{
			name:           "LOG_REQUEST_BODY with uppercase TRUE",
			logRequestBody: "TRUE",
			expectedResult: true,
		},
		{
			name:           "LOG_REQUEST_BODY with 1",
			logRequestBody: "1",
			expectedResult: true,
		},
		{
			name:           "LOG_REQUEST_BODY with lowercase false",
			logRequestBody: "false",
			expectedResult: false,
		},
		{
			name:           "LOG_REQUEST_BODY with uppercase FALSE",
			logRequestBody: "FALSE",
			expectedResult: false,
		},
		{
			name:           "LOG_REQUEST_BODY with 0",
			logRequestBody: "0",
			expectedResult: false,
		},
		{
			name:           "LOG_REQUEST_BODY with invalid value (uses default)",
			logRequestBody: "invalid",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			os.Setenv("APP_NAME", "test-app")
			os.Setenv("LOG_REQUEST_BODY", tt.logRequestBody)

			cfg := config.LoadConfig()

			assert.Equal(t, tt.expectedResult, cfg.Logging.LogRequestBody)
		})
	}
}

func TestLoadConfig_SwaggerConfiguration(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("SWAGGER_ENABLED", "true")

	cfg := config.LoadConfig()

	assert.True(t, cfg.Swagger.Enabled)
}

func TestLoadConfig_SwaggerConfiguration_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.False(t, cfg.Swagger.Enabled)
}

func TestLoadConfig_SwaggerConfiguration_BooleanVariations(t *testing.T) {
	tests := []struct {
		name           string
		swaggerEnabled string
		expectedResult bool
	}{
		{
			name:           "SWAGGER_ENABLED with lowercase true",
			swaggerEnabled: "true",
			expectedResult: true,
		},
		{
			name:           "SWAGGER_ENABLED with uppercase TRUE",
			swaggerEnabled: "TRUE",
			expectedResult: true,
		},
		{
			name:           "SWAGGER_ENABLED with 1",
			swaggerEnabled: "1",
			expectedResult: true,
		},
		{
			name:           "SWAGGER_ENABLED with t",
			swaggerEnabled: "t",
			expectedResult: true,
		},
		{
			name:           "SWAGGER_ENABLED with lowercase false",
			swaggerEnabled: "false",
			expectedResult: false,
		},
		{
			name:           "SWAGGER_ENABLED with uppercase FALSE",
			swaggerEnabled: "FALSE",
			expectedResult: false,
		},
		{
			name:           "SWAGGER_ENABLED with 0",
			swaggerEnabled: "0",
			expectedResult: false,
		},
		{
			name:           "SWAGGER_ENABLED with f",
			swaggerEnabled: "f",
			expectedResult: false,
		},
		{
			name:           "SWAGGER_ENABLED with invalid value (uses default)",
			swaggerEnabled: "invalid",
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()

			os.Setenv("APP_NAME", "test-app")
			os.Setenv("SWAGGER_ENABLED", tt.swaggerEnabled)

			cfg := config.LoadConfig()

			assert.Equal(t, tt.expectedResult, cfg.Swagger.Enabled)
		})
	}
}

func TestLoadConfig_AllConfigurations(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("APP_PORT", "8080")
	os.Setenv("APP_ENV", "test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("DB_AUTO_MIGRATE", "true")
	os.Setenv("JWT_PRIVATE_KEY_PATH", "./keys/test_private.pem")
	os.Setenv("JWT_PUBLIC_KEY_PATH", "./keys/test_public.pem")
	os.Setenv("JWT_EXPIRE_HOURS", "2")
	os.Setenv("RESEND_API_KEY", "re_test_api_key")
	os.Setenv("RESEND_FROM_EMAIL", "noreply@test.com")
	os.Setenv("OTP_LENGTH", "8")
	os.Setenv("OTP_EXPIRY_MINUTES", "15")
	os.Setenv("UPLOAD_MAX_SIZE", "1048576000")
	os.Setenv("UPLOAD_CHUNK_SIZE", "2097152")
	os.Setenv("LOG_LEVEL", "DEBUG")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("LOG_REQUEST_BODY", "true")
	os.Setenv("SWAGGER_ENABLED", "true")

	cfg := config.LoadConfig()

	// App config
	assert.Equal(t, "test-app", cfg.App.Name)
	assert.Equal(t, "8080", cfg.App.Port)
	assert.Equal(t, "test", cfg.App.Env)

	// Database config
	assert.Equal(t, "localhost", cfg.Database.Host)
	assert.Equal(t, "5432", cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpass", cfg.Database.Password)
	assert.Equal(t, "testdb", cfg.Database.Name)
	assert.True(t, cfg.Database.AutoMigrate)

	// JWT config
	assert.Equal(t, "./keys/test_private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "./keys/test_public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, 2, cfg.JWT.ExpireHours)

	// Resend config
	assert.Equal(t, "re_test_api_key", cfg.Resend.APIKey)
	assert.Equal(t, "noreply@test.com", cfg.Resend.FromEmail)

	// OTP config
	assert.Equal(t, 8, cfg.OTP.Length)
	assert.Equal(t, 15, cfg.OTP.ExpiryMinutes)

	// Upload config
	assert.Equal(t, int64(1048576000), cfg.Upload.MaxSize)
	assert.Equal(t, int64(2097152), cfg.Upload.ChunkSize)

	// Logging config
	assert.Equal(t, "DEBUG", cfg.Logging.Level)
	assert.Equal(t, "json", cfg.Logging.Format)
	assert.True(t, cfg.Logging.LogRequestBody)

	// Swagger config
	assert.True(t, cfg.Swagger.Enabled)
}

func TestLoadConfig_AppConfig_CorsOrigins(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("CORS_ORIGIN_DEVELOPMENT", "http://localhost:3000")
	os.Setenv("CORS_ORIGIN_PRODUCTION", "https://example.com")

	cfg := config.LoadConfig()

	assert.Equal(t, "http://localhost:3000", cfg.App.CorsOriginDev)
	assert.Equal(t, "https://example.com", cfg.App.CorsOriginProd)
}

func TestLoadConfig_AppConfig_CorsOrigins_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.Equal(t, "http://localhost:5173", cfg.App.CorsOriginDev)
	assert.Equal(t, "https://yourdomain.com", cfg.App.CorsOriginProd)
}

func TestLoadConfig_JWT_KeyRotation(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("JWT_KEY_ROTATION_HOURS", "336")

	cfg := config.LoadConfig()

	assert.Equal(t, 336, cfg.JWT.KeyRotationHours)
}

func TestLoadConfig_JWT_KeyRotation_DefaultValue(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.Equal(t, 168, cfg.JWT.KeyRotationHours)
}

func TestLoadConfig_JWT_AllPaths(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("JWT_PRIVATE_KEY_PATH", "/custom/path/private.pem")
	os.Setenv("JWT_PUBLIC_KEY_PATH", "/custom/path/public.pem")
	os.Setenv("JWT_PRIVATE_KEY_ROTATION_PATH", "/custom/path/private_rot.pem")
	os.Setenv("JWT_PUBLIC_KEY_ROTATION_PATH", "/custom/path/public_rot.pem")

	cfg := config.LoadConfig()

	assert.Equal(t, "/custom/path/private.pem", cfg.JWT.PrivateKeyPath)
	assert.Equal(t, "/custom/path/public.pem", cfg.JWT.PublicKeyPath)
	assert.Equal(t, "/custom/path/private_rot.pem", cfg.JWT.PrivateKeyRotationPath)
	assert.Equal(t, "/custom/path/public_rot.pem", cfg.JWT.PublicKeyRotationPath)
}

func TestLoadConfig_Database_AllFields(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("DB_HOST", "testhost")
	os.Setenv("DB_PORT", "5433")
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpassword")
	os.Setenv("DB_NAME", "testdbname")
	os.Setenv("DB_AUTO_MIGRATE", "false")
	os.Setenv("DB_RUN_SEEDER", "true")
	os.Setenv("DB_SEED_USERS", "false")
	os.Setenv("DB_MIGRATE_ON_START", "false")

	cfg := config.LoadConfig()

	assert.Equal(t, "testhost", cfg.Database.Host)
	assert.Equal(t, "5433", cfg.Database.Port)
	assert.Equal(t, "testuser", cfg.Database.User)
	assert.Equal(t, "testpassword", cfg.Database.Password)
	assert.Equal(t, "testdbname", cfg.Database.Name)
	assert.False(t, cfg.Database.AutoMigrate)
	assert.True(t, cfg.Database.RunSeeder)
	assert.False(t, cfg.Database.SeedUsers)
	assert.False(t, cfg.Database.MigrateOnStart)
}

func TestLoadConfig_Resend_AllFields(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")
	os.Setenv("RESEND_API_KEY", "re_test_key_123")
	os.Setenv("RESEND_FROM_EMAIL", "test@example.com")
	os.Setenv("RESEND_FROM_NAME", "Test App")

	cfg := config.LoadConfig()

	assert.Equal(t, "re_test_key_123", cfg.Resend.APIKey)
	assert.Equal(t, "test@example.com", cfg.Resend.FromEmail)
	assert.Equal(t, "Test App", cfg.Resend.FromName)
}

func TestLoadConfig_Resend_DefaultValues(t *testing.T) {
	os.Clearenv()

	os.Setenv("APP_NAME", "test-app")

	cfg := config.LoadConfig()

	assert.Equal(t, "", cfg.Resend.APIKey)
	assert.Equal(t, "noreply@example.com", cfg.Resend.FromEmail)
	assert.Equal(t, "Invento Service", cfg.Resend.FromName)
}
