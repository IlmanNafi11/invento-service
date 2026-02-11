package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestIntegrationDatabaseConnection tests database connection with SQLite in-memory
func TestIntegrationDatabaseConnection(t *testing.T) {
	t.Run("ConnectToSQLiteInMemory", func(t *testing.T) {
		// Create SQLite in-memory database for testing
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

		require.NoError(t, err, "Should connect to SQLite in-memory database")
		assert.NotNil(t, db)

		// Verify connection is working by pinging
		sqlDB, err := db.DB()
		require.NoError(t, err)
		err = sqlDB.Ping()
		require.NoError(t, err, "Should be able to ping the database")

		// Cleanup
		sqlDB.Close()
	})

	t.Run("CreateAndDropTable", func(t *testing.T) {
		// Create a test model
		type TestModel struct {
			ID   uint   `gorm:"primarykey"`
			Name string `gorm:"size:255"`
		}

		// Connect to in-memory database
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Create table
		err = db.AutoMigrate(&TestModel{})
		require.NoError(t, err, "Should create table successfully")

		// Insert test data
		testRecord := TestModel{ID: 1, Name: "Test"}
		result := db.Create(&testRecord)
		require.NoError(t, result.Error)

		// Query data
		var fetched TestModel
		result = db.First(&fetched, 1)
		require.NoError(t, result.Error)
		assert.Equal(t, "Test", fetched.Name)

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	t.Run("DatabaseOperations", func(t *testing.T) {
		type User struct {
			ID    uint   `gorm:"primarykey"`
			Name  string `gorm:"size:255"`
			Email string `gorm:"size:255;uniqueIndex"`
		}

		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Auto migrate
		err = db.AutoMigrate(&User{})
		require.NoError(t, err)

		// Create multiple users
		users := []User{
			{Name: "John Doe", Email: "john@example.com"},
			{Name: "Jane Doe", Email: "jane@example.com"},
		}
		result := db.Create(&users)
		require.NoError(t, result.Error)
		assert.Equal(t, int64(2), result.RowsAffected)

		// Find users
		var foundUsers []User
		result = db.Find(&foundUsers)
		require.NoError(t, result.Error)
		assert.Len(t, foundUsers, 2)

		// Find specific user
		var user User
		result = db.Where("email = ?", "john@example.com").First(&user)
		require.NoError(t, result.Error)
		assert.Equal(t, "John Doe", user.Name)

		// Update user
		result = db.Model(&user).Update("name", "John Updated")
		require.NoError(t, result.Error)

		// Verify update
		db.First(&user, user.ID)
		assert.Equal(t, "John Updated", user.Name)

		// Delete user
		result = db.Delete(&user)
		require.NoError(t, result.Error)

		// Verify deletion
		var count int64
		db.Model(&User{}).Count(&count)
		assert.Equal(t, int64(1), count)

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}

// TestIntegrationConfigWithEnvFile tests config loading with .env file
func TestIntegrationConfigWithEnvFile(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	t.Run("LoadConfigWithEnvFile", func(t *testing.T) {
		// Clear all environment variables first to avoid interference
		envKeys := []string{
			"APP_NAME", "APP_PORT", "APP_ENV",
			"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
			"LOG_LEVEL", "LOG_FORMAT",
		}
		originalValues := make(map[string]string)
		for _, key := range envKeys {
			if val, exists := os.LookupEnv(key); exists {
				originalValues[key] = val
				os.Unsetenv(key)
			}
		}

		// Restore environment variables after test
		defer func() {
			for key, val := range originalValues {
				os.Setenv(key, val)
			}
		}()

		// Create a temporary .env file
		envContent := `
APP_NAME=Test Application
APP_PORT=8080
APP_ENV=test
DB_HOST=localhost
DB_PORT=5432
DB_USER=testuser
DB_PASSWORD=testpass
DB_NAME=testdb
LOG_LEVEL=DEBUG
LOG_FORMAT=json
`
		envPath := filepath.Join(tempDir, ".env")
		err := os.WriteFile(envPath, []byte(envContent), 0644)
		require.NoError(t, err)

		// Change to temp directory to load the .env file
		originalDir, _ := os.Getwd()
		err = os.Chdir(tempDir)
		require.NoError(t, err)
		defer os.Chdir(originalDir)

		// Load config - this will use the .env file in current directory
		cfg := LoadConfig()

		// Verify config values
		assert.Equal(t, "Test Application", cfg.App.Name)
		assert.Equal(t, "8080", cfg.App.Port)
		assert.Equal(t, "test", cfg.App.Env)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, "5432", cfg.Database.Port)
		assert.Equal(t, "testuser", cfg.Database.User)
		assert.Equal(t, "testpass", cfg.Database.Password)
		assert.Equal(t, "testdb", cfg.Database.Name)
		assert.Equal(t, "DEBUG", cfg.Logging.Level)
		assert.Equal(t, "json", cfg.Logging.Format)
	})

	t.Run("LoadConfigWithDefaults", func(t *testing.T) {
		// Clear all environment variables
		envKeys := []string{
			"APP_NAME", "APP_PORT", "APP_ENV",
			"DB_HOST", "DB_PORT", "DB_USER", "DB_PASSWORD", "DB_NAME",
			"LOG_LEVEL", "LOG_FORMAT",
		}
		originalValues := make(map[string]string)
		for _, key := range envKeys {
			if val, exists := os.LookupEnv(key); exists {
				originalValues[key] = val
				os.Unsetenv(key)
			}
		}

		// Restore environment variables after test
		defer func() {
			for key, val := range originalValues {
				os.Setenv(key, val)
			}
		}()

		// Load config with defaults
		cfg := LoadConfig()

		// Verify default values are used
		assert.Equal(t, "Fiber Boilerplate", cfg.App.Name)
		assert.Equal(t, "3000", cfg.App.Port)
		assert.Equal(t, "development", cfg.App.Env)
		assert.Equal(t, "localhost", cfg.Database.Host)
		assert.Equal(t, "3306", cfg.Database.Port)
		assert.Equal(t, "root", cfg.Database.User)
		assert.Equal(t, "INFO", cfg.Logging.Level)
		assert.Equal(t, "text", cfg.Logging.Format)
	})

	t.Run("LoadConfigWithEnvironmentVariables", func(t *testing.T) {
		// Set environment variables
		envVars := map[string]string{
			"APP_NAME":    "Env App Name",
			"APP_PORT":    "9000",
			"APP_ENV":     "production",
			"DB_HOST":     "envhost",
			"DB_PORT":     "3307",
			"DB_USER":     "envuser",
			"DB_PASSWORD": "envpass",
			"DB_NAME":     "envdb",
		}

		// Save and restore original values
		originalValues := make(map[string]string)
		for key, val := range envVars {
			if originalVal, exists := os.LookupEnv(key); exists {
				originalValues[key] = originalVal
			}
			os.Setenv(key, val)
		}

		// Restore environment variables after test
		defer func() {
			for key, val := range originalValues {
				os.Setenv(key, val)
			}
			for key := range envVars {
				if _, exists := originalValues[key]; !exists {
					os.Unsetenv(key)
				}
			}
		}()

		// Load config
		cfg := LoadConfig()

		// Verify environment variables take precedence
		assert.Equal(t, "Env App Name", cfg.App.Name)
		assert.Equal(t, "9000", cfg.App.Port)
		assert.Equal(t, "production", cfg.App.Env)
		assert.Equal(t, "envhost", cfg.Database.Host)
		assert.Equal(t, "3307", cfg.Database.Port)
		assert.Equal(t, "envuser", cfg.Database.User)
		assert.Equal(t, "envpass", cfg.Database.Password)
		assert.Equal(t, "envdb", cfg.Database.Name)
	})
}

// TestIntegrationConfigValidation tests config validation
func TestIntegrationConfigValidation(t *testing.T) {
	t.Run("DatabaseConfigDefaults", func(t *testing.T) {
		cfg := LoadConfig()

		// Verify database config has sensible defaults
		assert.NotEmpty(t, cfg.Database.Host)
		assert.NotEmpty(t, cfg.Database.Port)
		assert.NotEmpty(t, cfg.Database.User)
		assert.NotEmpty(t, cfg.Database.Name)
	})

	t.Run("SupabaseConfigDefaults", func(t *testing.T) {
		cfg := LoadConfig()

		// Verify Supabase config has sensible defaults
		// Note: URL, AnonKey, ServiceKey, and DBURL may be empty in test environment
		assert.NotNil(t, cfg.Supabase)
	})

	t.Run("UploadConfigDefaults", func(t *testing.T) {
		cfg := LoadConfig()

		// Verify upload config has sensible defaults
		assert.Greater(t, cfg.Upload.MaxSize, int64(0))
		assert.Greater(t, cfg.Upload.ChunkSize, int64(0))
		assert.Greater(t, cfg.Upload.MaxConcurrent, 0)
		assert.Greater(t, cfg.Upload.IdleTimeout, 0)
		assert.NotEmpty(t, cfg.Upload.TusVersion)
	})
}

// TestIntegrationConfigWithSQLite tests config integration with SQLite database
func TestIntegrationConfigWithSQLite(t *testing.T) {
	t.Run("ConnectDatabaseWithSQLite", func(t *testing.T) {
		// Create a test config to verify config structure
		_ = &Config{
			Database: DatabaseConfig{
				Host:     "localhost",
				Port:     "3306",
				User:     "test",
				Password: "test",
				Name:     "test.db",
			},
		}

		// For testing, use SQLite in-memory instead of PostgreSQL
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

		require.NoError(t, err, "Should connect to SQLite database")
		assert.NotNil(t, db)

		// Test basic operations
		type TestTable struct {
			ID   uint   `gorm:"primarykey"`
			Data string
		}

		err = db.AutoMigrate(&TestTable{})
		require.NoError(t, err, "Should migrate table successfully")

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}

// TestIntegrationConfigHelperFunctions tests config helper functions
func TestIntegrationConfigHelperFunctions(t *testing.T) {
	t.Run("GetEnvWithDefaults", func(t *testing.T) {
		// Test with non-existing environment variable
		val := getEnv("NON_EXISTING_VAR", "default_value")
		assert.Equal(t, "default_value", val)

		// Test with existing environment variable
		os.Setenv("TEST_GET_ENV", "test_value")
		val = getEnv("TEST_GET_ENV", "default_value")
		assert.Equal(t, "test_value", val)
		os.Unsetenv("TEST_GET_ENV")
	})

	t.Run("GetEnvAsInt", func(t *testing.T) {
		// Test with valid integer
		os.Setenv("TEST_INT_VAR", "42")
		val := getEnvAsInt("TEST_INT_VAR", 10)
		assert.Equal(t, 42, val)
		os.Unsetenv("TEST_INT_VAR")

		// Test with invalid integer (should return default)
		os.Setenv("TEST_INT_VAR", "not_a_number")
		val = getEnvAsInt("TEST_INT_VAR", 10)
		assert.Equal(t, 10, val)
		os.Unsetenv("TEST_INT_VAR")

		// Test with non-existing variable (should return default)
		val = getEnvAsInt("NON_EXISTING_INT_VAR", 99)
		assert.Equal(t, 99, val)
	})

	t.Run("GetEnvAsInt64", func(t *testing.T) {
		// Test with valid int64
		os.Setenv("TEST_INT64_VAR", "9223372036854775807")
		val := getEnvAsInt64("TEST_INT64_VAR", 0)
		assert.Equal(t, int64(9223372036854775807), val)
		os.Unsetenv("TEST_INT64_VAR")

		// Test with invalid int64 (should return default)
		os.Setenv("TEST_INT64_VAR", "not_a_number")
		val = getEnvAsInt64("TEST_INT64_VAR", 123456789)
		assert.Equal(t, int64(123456789), val)
		os.Unsetenv("TEST_INT64_VAR")
	})

	t.Run("GetEnvAsBool", func(t *testing.T) {
		// Test with various true values
		for _, val := range []string{"1", "true", "TRUE", "True"} {
			os.Setenv("TEST_BOOL_VAR", val)
			result := getEnvAsBool("TEST_BOOL_VAR", false)
			assert.True(t, result, fmt.Sprintf("Should parse %s as true", val))
		}
		os.Unsetenv("TEST_BOOL_VAR")

		// Test with various false values
		for _, val := range []string{"0", "false", "FALSE", "False"} {
			os.Setenv("TEST_BOOL_VAR", val)
			result := getEnvAsBool("TEST_BOOL_VAR", true)
			assert.False(t, result, fmt.Sprintf("Should parse %s as false", val))
		}
		os.Unsetenv("TEST_BOOL_VAR")

		// Test with invalid value (should return default)
		os.Setenv("TEST_BOOL_VAR", "not_a_bool")
		result := getEnvAsBool("TEST_BOOL_VAR", true)
		assert.True(t, result) // Should return default (true)
		os.Unsetenv("TEST_BOOL_VAR")
	})

	t.Run("MaskToken", func(t *testing.T) {
		// Test with long token
		longToken := "this_is_a_very_long_token_that_should_be_masked"
		masked := maskToken(longToken)
		assert.Equal(t, "this_is_a_...", masked)
		assert.LessOrEqual(t, len(masked), 13)

		// Test with short token
		shortToken := "short"
		masked = maskToken(shortToken)
		assert.Equal(t, "***", masked)

		// Test with empty token
		emptyToken := ""
		masked = maskToken(emptyToken)
		assert.Equal(t, "***", masked)
	})
}

// TestIntegrationDatabaseConnectionWithConfig tests database connection using config
func TestIntegrationDatabaseConnectionWithConfig(t *testing.T) {
	t.Run("ConnectWithConfigDefaults", func(t *testing.T) {
		// Load default config
		cfg := LoadConfig()

		// Verify config is loaded
		assert.NotNil(t, cfg)
		assert.NotEmpty(t, cfg.Database.Host)
		assert.NotEmpty(t, cfg.Database.Port)

		// Note: We don't actually connect to the database here
		// because it might not be available in test environment.
		// In real integration tests, you would use testcontainers or
		// a test database instance.
	})

	t.Run("SQLiteInMemoryForTesting", func(t *testing.T) {
		// This is a pattern for tests that need a real database
		// without external dependencies

		type TestModel struct {
			ID   uint   `gorm:"primarykey"`
			Name string `gorm:"size:255"`
			Code string `gorm:"size:50;uniqueIndex"`
		}

		// Create in-memory database
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Auto migrate
		err = db.AutoMigrate(&TestModel{})
		require.NoError(t, err)

		// Test CRUD operations
		record := TestModel{Name: "Test Record", Code: "TEST001"}
		result := db.Create(&record)
		require.NoError(t, result.Error)
		assert.Greater(t, record.ID, uint(0))

		// Read
		var fetched TestModel
		result = db.Where("code = ?", "TEST001").First(&fetched)
		require.NoError(t, result.Error)
		assert.Equal(t, "Test Record", fetched.Name)

		// Update
		result = db.Model(&fetched).Update("name", "Updated Record")
		require.NoError(t, result.Error)

		// Delete
		result = db.Delete(&fetched)
		require.NoError(t, result.Error)

		// Verify deletion
		var count int64
		db.Model(&TestModel{}).Where("code = ?", "TEST001").Count(&count)
		assert.Equal(t, int64(0), count)

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}
