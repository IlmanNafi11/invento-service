package testing

import (
	"invento-service/internal/domain"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestDatabase creates an in-memory SQLite database for testing
// Returns the GORM DB instance and any error encountered
func SetupTestDatabase() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	// Auto-migrate all domain models
	err = db.AutoMigrate(
		&domain.User{},
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
		&domain.Project{},
		&domain.Modul{},
		&domain.TusUpload{},
		&domain.TusModulUpload{},
	)
	if err != nil {
		return nil, err
	}

	return db, nil
}

// TeardownTestDatabase closes the database connection and performs cleanup
func TeardownTestDatabase(db *gorm.DB) {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

// SetupTestDatabaseWithData creates a test database with seeded data
func SetupTestDatabaseWithData() (*gorm.DB, error) {
	db, err := SetupTestDatabase()
	if err != nil {
		return nil, err
	}

	// Seed roles
	roles := GetTestRoles()
	for _, role := range roles {
		if err := db.Create(&role).Error; err != nil {
			return nil, err
		}
	}

	// Seed users
	users := GetTestUsers()
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return nil, err
		}
	}

	return db, nil
}

// CleanupTestDatabase removes all data from all tables
func CleanupTestDatabase(db *gorm.DB) error {
	tables := []interface{}{
		&domain.TusModulUpload{},
		&domain.TusUpload{},
		&domain.Modul{},
		&domain.Project{},
		&domain.RolePermission{},
		&domain.Permission{},
		&domain.User{},
		&domain.Role{},
	}

	for _, table := range tables {
		if err := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			return err
		}
	}

	return nil
}

// TruncateTable removes all data from a specific table
func TruncateTable(db *gorm.DB, model interface{}) error {
	return db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(model).Error
}

// CreateTestUser creates a test user in the database
func CreateTestUser(db *gorm.DB, email, name string, roleID int) (*domain.User, error) {
	user := &domain.User{
		ID:     "test-user-" + email,
		Email:  email,
		Name:   name,
		RoleID: &roleID,
	}
	if err := db.Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

// CreateTestProject creates a test project in the database
func CreateTestProject(db *gorm.DB, name string, userID string) (*domain.Project, error) {
	project := &domain.Project{
		NamaProject: name,
		UserID:      userID,
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
	}
	if err := db.Create(project).Error; err != nil {
		return nil, err
	}
	return project, nil
}

// CreateTestModul creates a test modul in the database
func CreateTestModul(db *gorm.DB, name string, userID string) (*domain.Modul, error) {
	modul := &domain.Modul{
		Judul:     name,
		Deskripsi: "Test deskripsi",
		UserID:    userID,
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
	}
	if err := db.Create(modul).Error; err != nil {
		return nil, err
	}
	return modul, nil
}

// WaitForDatabase waits for the database to be ready (useful for integration tests)
func WaitForDatabase(db *gorm.DB, maxRetries int) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	for i := 0; i < maxRetries; i++ {
		if err := sqlDB.Ping(); err == nil {
			return nil
		}
	}

	return sqlDB.Ping()
}
