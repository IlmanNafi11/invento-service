package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Supabase    SupabaseConfig
	Upload      UploadConfig
	Logging     LoggingConfig
	Swagger     SwaggerConfig
	Performance PerformanceConfig
}

type AppConfig struct {
	Name           string
	Port           string
	Env            string
	CorsOriginDev  string
	CorsOriginProd string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
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
	Level          string
	Format         string
	LogRequestBody bool
}

type SwaggerConfig struct {
	Enabled bool
}

type SupabaseConfig struct {
	URL        string
	ServiceKey string
	AnonKey    string
	DBURL      string
	JWTSecret  string
}

type PerformanceConfig struct {
	// Fiber settings
	FiberConcurrency       int  // FIBER_CONCURRENCY, default 1024
	FiberReduceMemory      bool // FIBER_REDUCE_MEMORY_USAGE, default false
	FiberStreamRequestBody bool // FIBER_STREAM_REQUEST_BODY, default true
	FiberReadBufferSize    int  // FIBER_READ_BUFFER_SIZE, default 16384

	// Database pool
	DBMaxOpenConns    int // DB_MAX_OPEN_CONNS, default 10
	DBMaxIdleConns    int // DB_MAX_IDLE_CONNS, default 3
	DBConnMaxLifetime int // DB_CONN_MAX_LIFETIME, default 1800 (seconds)
	DBConnMaxIdleTime int // DB_CONN_MAX_IDLE_TIME, default 300 (seconds)

	// Go runtime
	GoMemLimit string // GOMEMLIMIT, default "350MiB"
	GoGC       int    // GOGC, default 100

	// Profiling
	EnablePprof bool // ENABLE_PPROF, default false

	// Memory monitoring
	MemoryWarningThreshold float64 // MEMORY_WARNING_THRESHOLD, default 0.8 (80%)
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Tidak dapat memuat file .env, menggunakan environment variables\n")
	}

	config := &Config{
		App: AppConfig{
			Name:           getEnv("APP_NAME", "invento-service"),
			Port:           getEnv("APP_PORT", "3000"),
			Env:            getEnv("APP_ENV", EnvDevelopment),
			CorsOriginDev:  getEnv("CORS_ORIGIN_DEVELOPMENT", "http://localhost:5173"),
			CorsOriginProd: getEnv("CORS_ORIGIN_PRODUCTION", "https://yourdomain.com"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnvAllowEmpty("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "postgres"),
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
		Logging: LoggingConfig{
			Level:          getEnv("LOG_LEVEL", "INFO"),
			Format:         getEnv("LOG_FORMAT", "text"),
			LogRequestBody: getEnvAsBool("LOG_REQUEST_BODY", false),
		},
		Swagger: SwaggerConfig{
			Enabled: getEnvAsBool("SWAGGER_ENABLED", false),
		},
		Supabase: SupabaseConfig{
			URL:        getEnv("SUPABASE_URL", ""),
			ServiceKey: getEnv("SUPABASE_SERVICE_ROLE_KEY", ""),
			AnonKey:    getEnv("SUPABASE_ANON_KEY", ""),
			DBURL:      getEnv("SUPABASE_DB_URL", ""),
			JWTSecret:  getEnv("SUPABASE_JWT_SECRET", ""),
		},
		Performance: PerformanceConfig{
			FiberConcurrency:       getEnvAsInt("FIBER_CONCURRENCY", 1024),
			FiberReduceMemory:      getEnvAsBool("FIBER_REDUCE_MEMORY_USAGE", false),
			FiberStreamRequestBody: getEnvAsBool("FIBER_STREAM_REQUEST_BODY", true),
			FiberReadBufferSize:    getEnvAsInt("FIBER_READ_BUFFER_SIZE", 16384),
			DBMaxOpenConns:         getEnvAsInt("DB_MAX_OPEN_CONNS", 10),
			DBMaxIdleConns:         getEnvAsInt("DB_MAX_IDLE_CONNS", 3),
			DBConnMaxLifetime:      getEnvAsInt("DB_CONN_MAX_LIFETIME", 1800),
			DBConnMaxIdleTime:      getEnvAsInt("DB_CONN_MAX_IDLE_TIME", 300),
			GoMemLimit:             getEnv("GOMEMLIMIT", "350MiB"),
			GoGC:                   getEnvAsInt("GOGC", 100),
			EnablePprof:            getEnvAsBool("ENABLE_PPROF", false),
			MemoryWarningThreshold: getEnvAsFloat64("MEMORY_WARNING_THRESHOLD", 0.8),
		},
	}

	return config, nil
}

// Validate checks that all critical environment variables are set.
// Call this explicitly in main() so tests can skip validation.
func (c *Config) Validate() error {
	var missing []string
	if c.Supabase.URL == "" {
		missing = append(missing, "SUPABASE_URL")
	}
	if c.Supabase.ServiceKey == "" {
		missing = append(missing, "SUPABASE_SERVICE_ROLE_KEY")
	}
	if len(missing) > 0 {
		return fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}
	return nil
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

func getEnvAllowEmpty(key, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvAsFloat64(key string, defaultValue float64) float64 {
	valueStr := getEnv(key, "")
	if value, err := strconv.ParseFloat(valueStr, 64); err == nil {
		return value
	}
	return defaultValue
}

// ParseMemLimit parses memory limit strings like "350MiB" or "1GiB" to bytes.
func ParseMemLimit(s string) (int64, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "MiB") {
		val, err := strconv.ParseInt(strings.TrimSuffix(s, "MiB"), 10, 64)
		return val * 1024 * 1024, err
	}
	if strings.HasSuffix(s, "GiB") {
		val, err := strconv.ParseInt(strings.TrimSuffix(s, "GiB"), 10, 64)
		return val * 1024 * 1024 * 1024, err
	}
	return strconv.ParseInt(s, 10, 64)
}
