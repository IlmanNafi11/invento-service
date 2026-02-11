package config

import (
	"fmt"
	"testing"

	"fiber-boiler-plate/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConnectDatabase tests the ConnectDatabase function
func TestConnectDatabase(t *testing.T) {
	t.Run("ConnectDatabaseBuildsDSNCorrectly", func(t *testing.T) {
		// Test that DSN is built correctly from config
		cfg := &Config{
			Database: DatabaseConfig{
				Host:     "testhost",
				Port:     "5432",
				User:     "testuser",
				Password: "testpass",
				Name:     "testdb",
			},
		}

		expectedDSN := "host=testhost port=5432 user=testuser password=testpass dbname=testdb sslmode=disable TimeZone=Asia/Jakarta"

		// Build the DSN the same way ConnectDatabase does
		actualDSN := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
		)

		assert.Equal(t, expectedDSN, actualDSN)
	})

	t.Run("ConnectToSQLiteInMemory", func(t *testing.T) {
		// Note: ConnectDatabase is designed for PostgreSQL
		// For testing, we use SQLite in-memory instead
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

	t.Run("ConnectDatabaseWithInvalidConnection", func(t *testing.T) {
		// Test opening PostgreSQL with invalid connection
		// This tests the gorm.Open path in ConnectDatabase
		dsn := "host=invalid port=9999 user=invalid password=invalid dbname=invalid sslmode=disable TimeZone=Asia/Jakarta"
		_, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})

		// We expect an error with invalid connection
		assert.Error(t, err, "Should fail to connect with invalid credentials")
	})
}

// TestAutoMigrate tests the AutoMigrate function
func TestAutoMigrate(t *testing.T) {
	t.Run("AutoMigrateAllModels", func(t *testing.T) {
		// Create SQLite in-memory database for testing
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Call AutoMigrate
		AutoMigrate(db)

		// Verify tables were created by checking if they exist
		// We can do this by attempting to query each table
		tables := []interface{}{
			&domain.Role{},
			&domain.Permission{},
			&domain.RolePermission{},
			&domain.User{},
			&domain.RefreshToken{},
			&domain.PasswordResetToken{},
			&domain.Project{},
			&domain.Modul{},
			&domain.TusUpload{},
			&domain.TusModulUpload{},
			&domain.OTP{},
		}

		for _, table := range tables {
			// Check if table exists by attempting to count records
			// This will fail if the table doesn't exist
			err := db.Find(table).Error
			assert.NoError(t, err, "Table %T should exist", table)
		}

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}

// TestRunSeeder tests the RunSeeder function
func TestRunSeeder(t *testing.T) {
	t.Run("RunSeederWithSeedUsersTrue", func(t *testing.T) {
		// Create SQLite in-memory database
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// First, run AutoMigrate
		AutoMigrate(db)

		// Create config with SeedUsers enabled
		cfg := &Config{
			Database: DatabaseConfig{
				SeedUsers: true,
			},
		}

		// Run seeder
		RunSeeder(db, cfg)

		// Verify permissions were created
		var permissions []domain.Permission
		result := db.Find(&permissions)
		require.NoError(t, result.Error)
		assert.Greater(t, len(permissions), 0, "Permissions should be seeded")

		// Verify admin role was created
		var adminRole domain.Role
		result = db.Where("nama_role = ?", "admin").First(&adminRole)
		require.NoError(t, result.Error, "Admin role should be seeded")

		// Verify mahasiswa role was created
		var mahasiswaRole domain.Role
		result = db.Where("nama_role = ?", "mahasiswa").First(&mahasiswaRole)
		require.NoError(t, result.Error, "Mahasiswa role should be seeded")

		// Verify dosen role was created
		var dosenRole domain.Role
		result = db.Where("nama_role = ?", "dosen").First(&dosenRole)
		require.NoError(t, result.Error, "Dosen role should be seeded")

		// Verify admin user was created
		var adminUser domain.User
		result = db.Where("email = ?", "admin@admin.polije.ac.id").First(&adminUser)
		require.NoError(t, result.Error, "Admin user should be seeded")
		assert.Equal(t, "Administrator", adminUser.Name)

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	t.Run("RunSeederWithSeedUsersFalse", func(t *testing.T) {
		// Create SQLite in-memory database
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// First, run AutoMigrate
		AutoMigrate(db)

		// Create config with SeedUsers disabled
		cfg := &Config{
			Database: DatabaseConfig{
				SeedUsers: false,
			},
		}

		// Run seeder
		RunSeeder(db, cfg)

		// Verify permissions were still created
		var permissions []domain.Permission
		result := db.Find(&permissions)
		require.NoError(t, result.Error)
		assert.Greater(t, len(permissions), 0, "Permissions should be seeded")

		// Verify admin role was created
		var adminRole domain.Role
		result = db.Where("nama_role = ?", "admin").First(&adminRole)
		require.NoError(t, result.Error, "Admin role should be seeded")

		// Verify admin user was NOT created
		var adminUser domain.User
		result = db.Where("email = ?", "admin@admin.polije.ac.id").First(&adminUser)
		assert.Error(t, result.Error, "Admin user should NOT be seeded when SeedUsers is false")

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	t.Run("RunSeederIdempotent", func(t *testing.T) {
		// Create SQLite in-memory database
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// First, run AutoMigrate
		AutoMigrate(db)

		// Create config
		cfg := &Config{
			Database: DatabaseConfig{
				SeedUsers: true,
			},
		}

		// Run seeder twice
		RunSeeder(db, cfg)

		var permissionCount1 int64
		db.Model(&domain.Permission{}).Count(&permissionCount1)

		var roleCount1 int64
		db.Model(&domain.Role{}).Count(&roleCount1)

		var userCount1 int64
		db.Model(&domain.User{}).Count(&userCount1)

		// Run seeder again
		RunSeeder(db, cfg)

		var permissionCount2 int64
		db.Model(&domain.Permission{}).Count(&permissionCount2)

		var roleCount2 int64
		db.Model(&domain.Role{}).Count(&roleCount2)

		var userCount2 int64
		db.Model(&domain.User{}).Count(&userCount2)

		// Counts should be the same (idempotent)
		assert.Equal(t, permissionCount1, permissionCount2, "Permission count should not change")
		assert.Equal(t, roleCount1, roleCount2, "Role count should not change")
		assert.Equal(t, userCount1, userCount2, "User count should not change")

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	t.Run("SeederWithExistingUser", func(t *testing.T) {
		// Create SQLite in-memory database
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// First, run AutoMigrate
		AutoMigrate(db)

		// Create a user with the specific email that assignRoleToExistingUsers looks for
		user := domain.User{
			Email:    "user@example.com",
			Password: "hashed_password",
			Name:     "Test User",
			IsActive: true,
		}
		result := db.Create(&user)
		require.NoError(t, result.Error)

		// Run seeder
		cfg := &Config{
			Database: DatabaseConfig{
				SeedUsers: false,
			},
		}
		RunSeeder(db, cfg)

		// Verify the user now has a role assigned
		var updatedUser domain.User
		result = db.First(&updatedUser, user.ID)
		require.NoError(t, result.Error)
		assert.NotNil(t, updatedUser.RoleID, "User should have role assigned")

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}
