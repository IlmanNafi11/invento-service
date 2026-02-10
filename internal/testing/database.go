package testing

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fmt"
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDatabase creates an in-memory SQLite database for testing
func SetupTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to test database: %w", err)
	}

	// Auto migrate all models
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
		&domain.RefreshToken{},
		&domain.Project{},
		&domain.Modul{},
		&domain.PasswordResetToken{},
		&domain.OTP{},
		&domain.TusUpload{},
		&domain.TusModulUpload{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate test database: %w", err)
	}

	return db, nil
}

// SetupTestDatabaseWithData creates an in-memory database and seeds it with test data
func SetupTestDatabaseWithData() (*gorm.DB, error) {
	db, err := SetupTestDatabase()
	if err != nil {
		return nil, err
	}

	// Seed test data
	err = SeedTestDatabase(db)
	if err != nil {
		return nil, fmt.Errorf("failed to seed test database: %w", err)
	}

	return db, nil
}

// SeedTestDatabase seeds the database with test data
func SeedTestDatabase(db *gorm.DB) error {
	// Create roles
	roleID1 := uint(1)
	roleID2 := uint(2)
	roles := []domain.Role{
		{
			ID:       roleID1,
			NamaRole: "user",
		},
		{
			ID:       roleID2,
			NamaRole: "admin",
		},
	}

	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return fmt.Errorf("failed to create role: %w", err)
		}
	}

	// Create users
	users := []domain.User{
		{
			Name:     "Test User",
			Email:    "test@example.com",
			Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password123"
			RoleID:   &roleID1,
		},
		{
			Name:     "Admin User",
			Email:    "admin@example.com",
			Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password123"
			RoleID:   &roleID2,
		},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}
	}

	// Create permissions
	permissions := []domain.Permission{
		{Resource: "users", Action: "read", Label: "Read users"},
		{Resource: "users", Action: "write", Label: "Write users"},
		{Resource: "users", Action: "delete", Label: "Delete users"},
		{Resource: "projects", Action: "read", Label: "Read projects"},
		{Resource: "projects", Action: "write", Label: "Write projects"},
		{Resource: "projects", Action: "delete", Label: "Delete projects"},
		{Resource: "moduls", Action: "read", Label: "Read moduls"},
		{Resource: "moduls", Action: "write", Label: "Write moduls"},
		{Resource: "moduls", Action: "delete", Label: "Delete moduls"},
	}

	for _, permission := range permissions {
		if err := db.Create(&permission).Error; err != nil {
			return fmt.Errorf("failed to create permission: %w", err)
		}
	}

	// Create projects
	userID1 := uint(1)
	userID2 := uint(2)
	projects := []domain.Project{
		{
			NamaProject: "Test Project",
			UserID:      userID1,
			Kategori:    "website",
			Semester:    1,
			Ukuran:      "small",
			PathFile:    "/test/path1",
		},
		{
			NamaProject: "Admin Project",
			UserID:      userID2,
			Kategori:    "mobile",
			Semester:    2,
			Ukuran:      "medium",
			PathFile:    "/test/path2",
		},
	}

	for _, project := range projects {
		if err := db.Create(&project).Error; err != nil {
			return fmt.Errorf("failed to create project: %w", err)
		}
	}

	// Create moduls
	moduls := []domain.Modul{
		{
			NamaFile: "Test Modul",
			UserID:   userID1,
			Tipe:     "pdf",
			Ukuran:   "small",
			Semester: 1,
			PathFile: "/test/modul1",
		},
		{
			NamaFile: "Another Modul",
			UserID:   userID1,
			Tipe:     "video",
			Ukuran:   "medium",
			Semester: 1,
			PathFile: "/test/modul2",
		},
	}

	for _, modul := range moduls {
		if err := db.Create(&modul).Error; err != nil {
			return fmt.Errorf("failed to create modul: %w", err)
		}
	}

	return nil
}

// TeardownTestDatabase closes the database connection
func TeardownTestDatabase(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	return sqlDB.Close()
}

// CleanupTestDatabase truncates all tables in the test database
func CleanupTestDatabase(db *gorm.DB) error {
	// Get all table names
	var tables []string
	err := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name NOT LIKE 'sqlite_%'").Scan(&tables).Error
	if err != nil {
		return fmt.Errorf("failed to get table names: %w", err)
	}

	// Truncate each table
	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		if err != nil {
			return fmt.Errorf("failed to truncate table %s: %w", table, err)
		}
	}

	return nil
}

// SetupTestRepositories creates test repositories with an in-memory database
// Note: This function returns the database instance. Repository constructors should be called individually as needed.
func SetupTestRepositories() (*gorm.DB, error) {
	db, err := SetupTestDatabaseWithData()
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TeardownTestRepositories cleans up test repositories
func TeardownTestRepositories(db *gorm.DB) error {
	return TeardownTestDatabase(db)
}

// SetupTestConfig creates a test configuration
func SetupTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Name:           "Test App",
			Port:           "3000",
			Env:            "test",
			CorsOriginDev:  "http://localhost:5173",
			CorsOriginProd: "https://example.com",
		},
		Database: config.DatabaseConfig{
			Host:           ":memory:",
			Port:           "0",
			User:           "test",
			Password:       "test",
			Name:           "test",
			AutoMigrate:    true,
			RunSeeder:      false,
			SeedUsers:      false,
			MigrateOnStart: false,
		},
		JWT: config.JWTConfig{
			PrivateKeyPath:          "/tmp/test_private.pem",
			PublicKeyPath:           "/tmp/test_public.pem",
			PrivateKeyRotationPath:  "/tmp/test_private_rotation.pem",
			PublicKeyRotationPath:   "/tmp/test_public_rotation.pem",
			ExpireHours:             24,
			RefreshTokenExpireHours: 168,
			KeyRotationHours:        168,
		},
		Upload: config.UploadConfig{
			MaxSize:              524288000,
			MaxSizeProject:       524288000,
			MaxSizeModul:         52428800,
			ChunkSize:            1048576,
			MaxConcurrent:        1,
			MaxConcurrentProject: 1,
			MaxConcurrentModul:   1,
			MaxQueueModulPerUser: 5,
			IdleTimeout:          600,
			CleanupInterval:      300,
			PathProduction:       "/tmp/uploads/",
			PathDevelopment:      "/tmp/uploads/",
			TempPathProduction:   "/tmp/uploads/temp/",
			TempPathDevelopment:  "/tmp/uploads/temp/",
			TusVersion:           "1.0.0",
			MaxResumeAttempts:    10,
		},
		OTP: config.OTPConfig{
			Length:                6,
			ExpiryMinutes:         10,
			MaxAttempts:           5,
			ResendCooldownSeconds: 60,
			ResendMaxTimes:        5,
		},
	}
}

// GetTestDBLogger returns a silent logger for testing
func GetTestDBLogger() logger.Interface {
	return logger.Default.LogMode(logger.Silent)
}

// CreateTestUser creates a test user in the database
func CreateTestUser(db *gorm.DB, email, name string, roleID uint) (*domain.User, error) {
	user := &domain.User{
		Name:     name,
		Email:    email,
		Password: "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy", // "password123"
		RoleID:   &roleID,
	}

	if err := db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create test user: %w", err)
	}

	return user, nil
}

// CreateTestProject creates a test project in the database
func CreateTestProject(db *gorm.DB, name string, userID uint) (*domain.Project, error) {
	project := &domain.Project{
		NamaProject: name,
		UserID:      userID,
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
	}

	if err := db.Create(project).Error; err != nil {
		return nil, fmt.Errorf("failed to create test project: %w", err)
	}

	return project, nil
}

// CreateTestModul creates a test modul in the database
func CreateTestModul(db *gorm.DB, name string, userID uint) (*domain.Modul, error) {
	modul := &domain.Modul{
		NamaFile: name,
		UserID:   userID,
		Tipe:     "pdf",
		Ukuran:   "small",
		Semester: 1,
		PathFile: "/test/modul",
	}

	if err := db.Create(modul).Error; err != nil {
		return nil, fmt.Errorf("failed to create test modul: %w", err)
	}

	return modul, nil
}

// TruncateTable truncates a specific table
func TruncateTable(db *gorm.DB, model interface{}) error {
	return db.Exec(fmt.Sprintf("DELETE FROM %s", db.Statement.Table)).Error
}

// WaitForDatabase waits for the database to be ready
func WaitForDatabase(db *gorm.DB, maxAttempts int) error {
	for i := 0; i < maxAttempts; i++ {
		sqlDB, err := db.DB()
		if err != nil {
			log.Printf("Attempt %d: failed to get database instance: %v", i+1, err)
			continue
		}

		if err := sqlDB.Ping(); err == nil {
			return nil
		}

		log.Printf("Attempt %d: database not ready: %v", i+1, err)
	}

	return fmt.Errorf("database not ready after %d attempts", maxAttempts)
}
