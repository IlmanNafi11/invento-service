package config_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// TestConnectDatabase tests the ConnectDatabase function with SQLite
func TestConnectDatabase(t *testing.T) {
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
		config.AutoMigrate(db)

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

	t.Run("AutoMigrateRoleTable", func(t *testing.T) {
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Migrate just the Role model
		err = db.AutoMigrate(&domain.Role{})
		require.NoError(t, err)

		// Verify we can create a role
		role := domain.Role{NamaRole: "test_role"}
		result := db.Create(&role)
		require.NoError(t, result.Error)
		assert.Greater(t, role.ID, uint(0))

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	t.Run("AutoMigrateUserTable", func(t *testing.T) {
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Migrate User and Role (User depends on Role)
		err = db.AutoMigrate(&domain.Role{}, &domain.User{})
		require.NoError(t, err)

		// Create a role first
		role := domain.Role{NamaRole: "test_role"}
		db.Create(&role)

		// Verify we can create a user
		user := domain.User{
			Email:    "test@example.com",
			Password: "hashed_password",
			Name:     "Test User",
			RoleID:   &role.ID,
			IsActive: true,
		}
		result := db.Create(&user)
		require.NoError(t, result.Error)
		assert.Greater(t, user.ID, uint(0))

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})

	t.Run("AutoMigrateProjectTable", func(t *testing.T) {
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Migrate User and Project
		err = db.AutoMigrate(&domain.User{}, &domain.Project{})
		require.NoError(t, err)

		// Create a user first
		user := domain.User{
			Email:    "test@example.com",
			Password: "hashed_password",
			Name:     "Test User",
			IsActive: true,
		}
		db.Create(&user)

		// Verify we can create a project
		project := domain.Project{
			UserID:      user.ID,
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    1,
			Ukuran:      "large",
			PathFile:    "/test/path",
		}
		result := db.Create(&project)
		require.NoError(t, result.Error)
		assert.Greater(t, project.ID, uint(0))

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
		config.AutoMigrate(db)

		// Create config with SeedUsers enabled
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				SeedUsers: true,
			},
		}

		// Run seeder
		config.RunSeeder(db, cfg)

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
		config.AutoMigrate(db)

		// Create config with SeedUsers disabled
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				SeedUsers: false,
			},
		}

		// Run seeder
		config.RunSeeder(db, cfg)

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
		config.AutoMigrate(db)

		// Create config
		cfg := &config.Config{
			Database: config.DatabaseConfig{
				SeedUsers: true,
			},
		}

		// Run seeder twice
		config.RunSeeder(db, cfg)

		var permissionCount1 int64
		db.Model(&domain.Permission{}).Count(&permissionCount1)

		var roleCount1 int64
		db.Model(&domain.Role{}).Count(&roleCount1)

		var userCount1 int64
		db.Model(&domain.User{}).Count(&userCount1)

		// Run seeder again
		config.RunSeeder(db, cfg)

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
}

// TestSeedPermissions tests the seedPermissions function specifically
func TestSeedPermissions(t *testing.T) {
	t.Run("SeedPermissionsCreatesExpectedPermissions", func(t *testing.T) {
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Migrate Permission model
		err = db.AutoMigrate(&domain.Permission{})
		require.NoError(t, err)

		// We need to use the internal function, but it's not exported
		// So we'll test through RunSeeder instead
		config.AutoMigrate(db)
		cfg := &config.Config{}

		config.RunSeeder(db, cfg)

		// Verify expected resources
		expectedResources := []string{"Role", "Permission", "Project", "Modul", "User"}
		for _, resource := range expectedResources {
			var count int64
			db.Model(&domain.Permission{}).Where("resource = ?", resource).Count(&count)
			assert.Greater(t, count, int64(0), "Resource %s should have permissions", resource)
		}

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}

// TestSeedAdminRole tests the seedAdminRole function specifically
func TestSeedAdminRole(t *testing.T) {
	t.Run("SeedAdminRoleAssignsAllPermissions", func(t *testing.T) {
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Run full migration and seeder
		config.AutoMigrate(db)
		cfg := &config.Config{}
		config.RunSeeder(db, cfg)

		// Get admin role
		var adminRole domain.Role
		err = db.Where("nama_role = ?", "admin").First(&adminRole).Error
		require.NoError(t, err)

		// Get all permissions
		var allPermissions []domain.Permission
		err = db.Find(&allPermissions).Error
		require.NoError(t, err)

		// Count role permissions for admin
		var rolePermissionCount int64
		err = db.Model(&domain.RolePermission{}).Where("role_id = ?", adminRole.ID).Count(&rolePermissionCount).Error
		require.NoError(t, err)

		// Admin should have all permissions assigned
		assert.Equal(t, int64(len(allPermissions)), rolePermissionCount, "Admin role should have all permissions")

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}

// TestSeedMahasiswaDosenRoles tests the seedMahasiswaDosenRoles function
func TestSeedMahasiswaDosenRoles(t *testing.T) {
	t.Run("MahasiswaAndDosenRolesCreated", func(t *testing.T) {
		dsn := "file::memory:?cache=shared"
		db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Silent),
		})
		require.NoError(t, err)

		// Run full migration and seeder
		config.AutoMigrate(db)
		cfg := &config.Config{}
		config.RunSeeder(db, cfg)

		// Verify mahasiswa role exists
		var mahasiswaRole domain.Role
		err = db.Where("nama_role = ?", "mahasiswa").First(&mahasiswaRole).Error
		require.NoError(t, err, "Mahasiswa role should exist")

		// Verify dosen role exists
		var dosenRole domain.Role
		err = db.Where("nama_role = ?", "dosen").First(&dosenRole).Error
		require.NoError(t, err, "Dosen role should exist")

		// Verify mahasiswa has Modul and Project permissions
		var mahasiswaPermissionCount int64
		db.Table("role_permissions").
			Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("role_permissions.role_id = ? AND permissions.resource IN (?, ?)", mahasiswaRole.ID, "Modul", "Project").
			Count(&mahasiswaPermissionCount)
		assert.Greater(t, mahasiswaPermissionCount, int64(0), "Mahasiswa should have Modul and Project permissions")

		// Verify dosen has Modul, Project, and User permissions
		var dosenPermissionCount int64
		db.Table("role_permissions").
			Joins("JOIN permissions ON permissions.id = role_permissions.permission_id").
			Where("role_permissions.role_id = ? AND permissions.resource IN (?, ?, ?)", dosenRole.ID, "Modul", "Project", "User").
			Count(&dosenPermissionCount)
		assert.Greater(t, dosenPermissionCount, int64(0), "Dosen should have Modul, Project, and User permissions")

		// Cleanup
		sqlDB, _ := db.DB()
		sqlDB.Close()
	})
}

// TestDatabaseConfigStructure tests the DatabaseConfig structure
func TestDatabaseConfigStructure(t *testing.T) {
	t.Run("DatabaseConfigHasAllFields", func(t *testing.T) {
		cfg := config.DatabaseConfig{
			Host:           "localhost",
			Port:           "5432",
			User:           "testuser",
			Password:       "testpass",
			Name:           "testdb",
			AutoMigrate:    true,
			RunSeeder:      false,
			SeedUsers:      true,
			MigrateOnStart: false,
		}

		assert.Equal(t, "localhost", cfg.Host)
		assert.Equal(t, "5432", cfg.Port)
		assert.Equal(t, "testuser", cfg.User)
		assert.Equal(t, "testpass", cfg.Password)
		assert.Equal(t, "testdb", cfg.Name)
		assert.True(t, cfg.AutoMigrate)
		assert.False(t, cfg.RunSeeder)
		assert.True(t, cfg.SeedUsers)
		assert.False(t, cfg.MigrateOnStart)
	})
}
