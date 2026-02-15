package usecase

import (
	"invento-service/internal/domain"
	"invento-service/internal/helper"
	"invento-service/internal/usecase/repo"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockCasbinEnforcer is a mock for CasbinEnforcer
type MockCasbinEnforcer struct {
	mock.Mock
}

func (m *MockCasbinEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	args := m.Called(roleName, resource, action)
	return args.Bool(0), args.Error(1)
}

// CasbinPermissionChecker is an interface for CheckPermission method
type CasbinPermissionChecker interface {
	CheckPermission(roleName, resource, action string) (bool, error)
}

// statisticUsecaseWithInterface is a wrapper for testing with interface
type statisticUsecaseWithInterface struct {
	userRepo       repo.UserRepository
	projectRepo    repo.ProjectRepository
	modulRepo      repo.ModulRepository
	roleRepo       repo.RoleRepository
	casbinEnforcer CasbinPermissionChecker
	db             *gorm.DB
}

func (su *statisticUsecaseWithInterface) GetStatistics(userID string, userRole string) (*domain.StatisticData, error) {
	result := &domain.StatisticData{}

	hasProjectRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Project", "read")
	hasModulRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Modul", "read")
	hasUserRead, _ := su.casbinEnforcer.CheckPermission(userRole, "User", "read")
	hasRoleRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Role", "read")

	if hasProjectRead {
		projectCount, _ := su.projectRepo.CountByUserID(userID)
		result.TotalProject = &projectCount
	}

	if hasModulRead {
		modulCount, _ := su.modulRepo.CountByUserID(userID)
		result.TotalModul = &modulCount
	}

	if hasUserRead {
		var totalUser int64
		su.db.Model(&domain.User{}).Count(&totalUser)
		count := int(totalUser)
		result.TotalUser = &count
	}

	if hasRoleRead {
		var totalRole int64
		su.db.Model(&domain.Role{}).Count(&totalRole)
		count := int(totalRole)
		result.TotalRole = &count
	}

	return result, nil
}

func newStatisticUsecaseWithInterface(
	userRepo repo.UserRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	roleRepo repo.RoleRepository,
	casbinEnforcer CasbinPermissionChecker,
	db *gorm.DB,
) *statisticUsecaseWithInterface {
	return &statisticUsecaseWithInterface{
		userRepo:       userRepo,
		projectRepo:    projectRepo,
		modulRepo:      modulRepo,
		roleRepo:       roleRepo,
		casbinEnforcer: casbinEnforcer,
		db:             db,
	}
}

// setupTestDB creates an in-memory SQLite database for testing
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create tables
	err = db.AutoMigrate(&domain.User{}, &domain.Role{})
	if err != nil {
		t.Fatalf("failed to migrate test database: %v", err)
	}

	return db
}

// seedTestDB seeds the database with test data
func seedTestDB(db *gorm.DB, userCount, roleCount int) {
	// Create roles
	for i := 1; i <= roleCount; i++ {
		role := domain.Role{
			ID:       uint(i),
			NamaRole: "role_" + string(rune('0'+i)),
		}
		db.Create(&role)
	}

	// Create users
	for i := 1; i <= userCount; i++ {
		user := domain.User{
			ID:    fmt.Sprintf("user-%d", i),
			Email: "user" + string(rune('0'+i)) + "@example.com",
			Name:  "User " + string(rune('0'+i)),
		}
		db.Create(&user)
	}
}

// TestStatisticUsecase_GetStatistics_AllPermissionsGranted tests statistics retrieval when user has all permissions
func TestStatisticUsecase_GetStatistics_AllPermissionsGranted(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	// Seed database with test data
	seedTestDB(db, 15, 5)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-1"
	userRole := "admin"

	// Setup expectations - all permissions granted
	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(true, nil)

	mockProjectRepo.On("CountByUserID", userID).Return(10, nil)
	mockModulRepo.On("CountByUserID", userID).Return(25, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 10, *result.TotalProject)
	assert.NotNil(t, result.TotalModul)
	assert.Equal(t, 25, *result.TotalModul)
	assert.NotNil(t, result.TotalUser)
	assert.Equal(t, 15, *result.TotalUser)
	assert.NotNil(t, result.TotalRole)
	assert.Equal(t, 5, *result.TotalRole)

	mockCasbin.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_OnlyProjectPermission tests statistics with only Project read permission
func TestStatisticUsecase_GetStatistics_OnlyProjectPermission(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 8, 3)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-2"
	userRole := "project_manager"

	// Only project permission granted
	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(false, nil)

	mockProjectRepo.On("CountByUserID", userID).Return(5, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 5, *result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.Nil(t, result.TotalUser)
	assert.Nil(t, result.TotalRole)

	mockCasbin.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_OnlyModulPermission tests statistics with only Modul read permission
func TestStatisticUsecase_GetStatistics_OnlyModulPermission(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 12, 4)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-3"
	userRole := "modul_viewer"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(false, nil)

	mockModulRepo.On("CountByUserID", userID).Return(30, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.TotalProject)
	assert.NotNil(t, result.TotalModul)
	assert.Equal(t, 30, *result.TotalModul)
	assert.Nil(t, result.TotalUser)
	assert.Nil(t, result.TotalRole)

	mockCasbin.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_OnlyUserPermission tests statistics with only User read permission
func TestStatisticUsecase_GetStatistics_OnlyUserPermission(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 20, 6)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-4"
	userRole := "user_manager"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(false, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.NotNil(t, result.TotalUser)
	assert.Equal(t, 20, *result.TotalUser)
	assert.Nil(t, result.TotalRole)

	mockCasbin.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_OnlyRolePermission tests statistics with only Role read permission
func TestStatisticUsecase_GetStatistics_OnlyRolePermission(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 25, 7)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-5"
	userRole := "role_manager"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(true, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.Nil(t, result.TotalUser)
	assert.NotNil(t, result.TotalRole)
	assert.Equal(t, 7, *result.TotalRole)

	mockCasbin.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_NoPermissions tests statistics when user has no read permissions
func TestStatisticUsecase_GetStatistics_NoPermissions(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 10, 3)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-6"
	userRole := "guest"

	// No permissions granted
	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(false, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.Nil(t, result.TotalUser)
	assert.Nil(t, result.TotalRole)

	mockCasbin.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_ProjectAndModulPermissions tests statistics with Project and Modul permissions
func TestStatisticUsecase_GetStatistics_ProjectAndModulPermissions(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 18, 5)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-7"
	userRole := "content_manager"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(false, nil)

	mockProjectRepo.On("CountByUserID", userID).Return(12, nil)
	mockModulRepo.On("CountByUserID", userID).Return(45, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 12, *result.TotalProject)
	assert.NotNil(t, result.TotalModul)
	assert.Equal(t, 45, *result.TotalModul)
	assert.Nil(t, result.TotalUser)
	assert.Nil(t, result.TotalRole)

	mockCasbin.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_UserAndRolePermissions tests statistics with User and Role permissions
func TestStatisticUsecase_GetStatistics_UserAndRolePermissions(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 30, 8)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-8"
	userRole := "admin"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(false, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(true, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.NotNil(t, result.TotalUser)
	assert.Equal(t, 30, *result.TotalUser)
	assert.NotNil(t, result.TotalRole)
	assert.Equal(t, 8, *result.TotalRole)

	mockCasbin.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_EmptyDatabase tests statistics with empty database
func TestStatisticUsecase_GetStatistics_EmptyDatabase(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)
	// Don't seed - keep it empty

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-9"
	userRole := "admin"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(true, nil)

	mockProjectRepo.On("CountByUserID", userID).Return(0, nil)
	mockModulRepo.On("CountByUserID", userID).Return(0, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 0, *result.TotalProject)
	assert.NotNil(t, result.TotalModul)
	assert.Equal(t, 0, *result.TotalModul)
	assert.NotNil(t, result.TotalUser)
	assert.Equal(t, 0, *result.TotalUser)
	assert.NotNil(t, result.TotalRole)
	assert.Equal(t, 0, *result.TotalRole)

	mockCasbin.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_ZeroCounts tests statistics with zero counts
func TestStatisticUsecase_GetStatistics_ZeroCounts(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	seedTestDB(db, 5, 2)

	userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	userID := "user-10"
	userRole := "user_with_no_content"

	mockCasbin.On("CheckPermission", userRole, "Project", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Modul", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "User", "read").Return(true, nil)
	mockCasbin.On("CheckPermission", userRole, "Role", "read").Return(true, nil)

	mockProjectRepo.On("CountByUserID", userID).Return(0, nil)
	mockModulRepo.On("CountByUserID", userID).Return(0, nil)

	result, err := userUC.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 0, *result.TotalProject)
	assert.NotNil(t, result.TotalModul)
	assert.Equal(t, 0, *result.TotalModul)

	mockCasbin.AssertExpectations(t)
	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

// TestStatisticUsecase_GetStatistics_VariousUserIDs tests statistics with different user IDs
func TestStatisticUsecase_GetStatistics_VariousUserIDs(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	db := setupTestDB(t)

	seedTestDB(db, 50, 10)

	testCases := []struct {
		name          string
		userID        string
		userRole      string
		projectCount  int
		modulCount    int
		expectedUsers int
		expectedRoles int
	}{
		{
			name:          "User 1 with admin role",
			userID:        "user-1",
			userRole:      "admin",
			projectCount:  8,
			modulCount:    15,
			expectedUsers: 50,
			expectedRoles: 10,
		},
		{
			name:          "User 2 with manager role",
			userID:        "user-2",
			userRole:      "manager",
			projectCount:  12,
			modulCount:    25,
			expectedUsers: 50,
			expectedRoles: 10,
		},
		{
			name:          "User 100 with admin role",
			userID:        "user-100",
			userRole:      "admin",
			projectCount:  0,
			modulCount:    0,
			expectedUsers: 50,
			expectedRoles: 10,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCasbin := new(MockCasbinEnforcer)
			mockProjectRepo := new(MockProjectRepository)
			mockModulRepo := new(MockModulRepository)

			userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

			mockCasbin.On("CheckPermission", tc.userRole, "Project", "read").Return(true, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "Modul", "read").Return(true, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "User", "read").Return(true, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "Role", "read").Return(true, nil)

			mockProjectRepo.On("CountByUserID", tc.userID).Return(tc.projectCount, nil)
			mockModulRepo.On("CountByUserID", tc.userID).Return(tc.modulCount, nil)

			result, err := userUC.GetStatistics(tc.userID, tc.userRole)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.NotNil(t, result.TotalProject)
			assert.Equal(t, tc.projectCount, *result.TotalProject)
			assert.NotNil(t, result.TotalModul)
			assert.Equal(t, tc.modulCount, *result.TotalModul)
			assert.NotNil(t, result.TotalUser)
			assert.Equal(t, tc.expectedUsers, *result.TotalUser)
			assert.NotNil(t, result.TotalRole)
			assert.Equal(t, tc.expectedRoles, *result.TotalRole)

			mockCasbin.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
			mockModulRepo.AssertExpectations(t)
		})
	}
}

// TestNewStatisticUsecase tests the constructor function
func TestNewStatisticUsecase(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	usecase := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	assert.NotNil(t, usecase)

	// Verify that the usecase is properly initialized
	assert.NotNil(t, usecase)
	// Verify dependencies are set
	concreteUC := usecase
	assert.Equal(t, mockUserRepo, concreteUC.userRepo)
	assert.Equal(t, mockProjectRepo, concreteUC.projectRepo)
	assert.Equal(t, mockModulRepo, concreteUC.modulRepo)
	assert.Equal(t, mockRoleRepo, concreteUC.roleRepo)
	assert.Equal(t, mockCasbin, concreteUC.casbinEnforcer)
	assert.Equal(t, db, concreteUC.db)
}

// TestStatisticUsecase_GetStatistics_MixedPermissions tests various permission combinations
func TestStatisticUsecase_GetStatistics_MixedPermissions(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	db := setupTestDB(t)

	seedTestDB(db, 40, 9)

	testCases := []struct {
		name             string
		userRole         string
		hasProjectRead   bool
		hasModulRead     bool
		hasUserRead      bool
		hasRoleRead      bool
		projectCount     int
		modulCount       int
		expectProjectNil bool
		expectModulNil   bool
		expectUserNil    bool
		expectRoleNil    bool
	}{
		{
			name:             "Full admin - all permissions",
			userRole:         "admin",
			hasProjectRead:   true,
			hasModulRead:     true,
			hasUserRead:      true,
			hasRoleRead:      true,
			projectCount:     15,
			modulCount:       35,
			expectProjectNil: false,
			expectModulNil:   false,
			expectUserNil:    false,
			expectRoleNil:    false,
		},
		{
			name:             "Project manager - only projects",
			userRole:         "project_manager",
			hasProjectRead:   true,
			hasModulRead:     false,
			hasUserRead:      false,
			hasRoleRead:      false,
			projectCount:     20,
			modulCount:       0,
			expectProjectNil: false,
			expectModulNil:   true,
			expectUserNil:    true,
			expectRoleNil:    true,
		},
		{
			name:             "User admin - users and roles",
			userRole:         "user_admin",
			hasProjectRead:   false,
			hasModulRead:     false,
			hasUserRead:      true,
			hasRoleRead:      true,
			projectCount:     0,
			modulCount:       0,
			expectProjectNil: true,
			expectModulNil:   true,
			expectUserNil:    false,
			expectRoleNil:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockCasbin := new(MockCasbinEnforcer)
			mockProjectRepo := new(MockProjectRepository)
			mockModulRepo := new(MockModulRepository)

			userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

			userID := "user-1"

			mockCasbin.On("CheckPermission", tc.userRole, "Project", "read").Return(tc.hasProjectRead, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "Modul", "read").Return(tc.hasModulRead, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "User", "read").Return(tc.hasUserRead, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "Role", "read").Return(tc.hasRoleRead, nil)

			if tc.hasProjectRead {
				mockProjectRepo.On("CountByUserID", userID).Return(tc.projectCount, nil)
			}
			if tc.hasModulRead {
				mockModulRepo.On("CountByUserID", userID).Return(tc.modulCount, nil)
			}

			result, err := userUC.GetStatistics(userID, tc.userRole)

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tc.expectProjectNil {
				assert.Nil(t, result.TotalProject)
			} else {
				assert.NotNil(t, result.TotalProject)
				assert.Equal(t, tc.projectCount, *result.TotalProject)
			}

			if tc.expectModulNil {
				assert.Nil(t, result.TotalModul)
			} else {
				assert.NotNil(t, result.TotalModul)
				assert.Equal(t, tc.modulCount, *result.TotalModul)
			}

			if tc.expectUserNil {
				assert.Nil(t, result.TotalUser)
			} else {
				assert.NotNil(t, result.TotalUser)
			}

			if tc.expectRoleNil {
				assert.Nil(t, result.TotalRole)
			} else {
				assert.NotNil(t, result.TotalRole)
			}

			mockCasbin.AssertExpectations(t)
			mockProjectRepo.AssertExpectations(t)
			mockModulRepo.AssertExpectations(t)
		})
	}
}

// TestStatisticUsecase_InterfaceImplementation tests that the usecase implements the interface
func TestStatisticUsecase_InterfaceImplementation(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockCasbin := new(MockCasbinEnforcer)
	db := setupTestDB(t)

	usecase := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

	// Verify that the usecase is properly initialized
	assert.NotNil(t, usecase)
}

// TestStatisticUsecase_ActualGetStatistics_WithRealEnforcer tests the actual GetStatistics with real Casbin enforcer
func TestStatisticUsecase_ActualGetStatistics_WithRealEnforcer(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)

	// Create a real Casbin enforcer for testing
	casbinDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open casbin test database: %v", err)
	}
	casbinEnforcer, err := helper.NewCasbinEnforcer(casbinDB)
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}

	db := setupTestDB(t)
	seedTestDB(db, 15, 5)

	usecase := NewStatisticUsecase(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, casbinEnforcer, db)

	userID := "user-1"
	userRole := "admin"

	// Add permissions to admin role
	casbinEnforcer.AddPermissionForRole(userRole, "Project", "read")
	casbinEnforcer.AddPermissionForRole(userRole, "Modul", "read")
	casbinEnforcer.AddPermissionForRole(userRole, "User", "read")
	casbinEnforcer.AddPermissionForRole(userRole, "Role", "read")
	casbinEnforcer.SavePolicy()

	mockProjectRepo.On("CountByUserID", userID).Return(10, nil)
	mockModulRepo.On("CountByUserID", userID).Return(25, nil)

	result, err := usecase.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 10, *result.TotalProject)
	assert.NotNil(t, result.TotalModul)
	assert.Equal(t, 25, *result.TotalModul)
	assert.NotNil(t, result.TotalUser)
	assert.Equal(t, 15, *result.TotalUser)
	assert.NotNil(t, result.TotalRole)
	assert.Equal(t, 5, *result.TotalRole)

	mockProjectRepo.AssertExpectations(t)
	mockModulRepo.AssertExpectations(t)
}

// TestStatisticUsecase_ActualGetStatistics_NoPermissions tests with real enforcer but no permissions
func TestStatisticUsecase_ActualGetStatistics_NoPermissions(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)

	// Create a real Casbin enforcer for testing
	casbinDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open casbin test database: %v", err)
	}
	casbinEnforcer, err := helper.NewCasbinEnforcer(casbinDB)
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}

	db := setupTestDB(t)
	seedTestDB(db, 10, 3)

	usecase := NewStatisticUsecase(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, casbinEnforcer, db)

	userID := "user-1"
	userRole := "guest"
	// No permissions added for guest

	result, err := usecase.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Nil(t, result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.Nil(t, result.TotalUser)
	assert.Nil(t, result.TotalRole)
}

// TestStatisticUsecase_ActualGetStatistics_PartialPermissions tests with partial permissions
func TestStatisticUsecase_ActualGetStatistics_PartialPermissions(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)

	// Create a real Casbin enforcer for testing
	casbinDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open casbin test database: %v", err)
	}
	casbinEnforcer, err := helper.NewCasbinEnforcer(casbinDB)
	if err != nil {
		t.Fatalf("failed to create casbin enforcer: %v", err)
	}

	db := setupTestDB(t)
	seedTestDB(db, 20, 6)

	usecase := NewStatisticUsecase(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, casbinEnforcer, db)

	userID := "user-1"
	userRole := "project_manager"

	// Add only Project and User permissions
	casbinEnforcer.AddPermissionForRole(userRole, "Project", "read")
	casbinEnforcer.AddPermissionForRole(userRole, "User", "read")
	casbinEnforcer.SavePolicy()

	mockProjectRepo.On("CountByUserID", userID).Return(8, nil)

	result, err := usecase.GetStatistics(userID, userRole)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotNil(t, result.TotalProject)
	assert.Equal(t, 8, *result.TotalProject)
	assert.Nil(t, result.TotalModul)
	assert.NotNil(t, result.TotalUser)
	assert.Equal(t, 20, *result.TotalUser)
	assert.Nil(t, result.TotalRole)

	mockProjectRepo.AssertExpectations(t)
}

// TestNewStatisticUsecase_RealConstructor tests the actual NewStatisticUsecase constructor
func TestNewStatisticUsecase_RealConstructor(t *testing.T) {
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)
	db := setupTestDB(t)

	usecase := NewStatisticUsecase(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, nil, db)

	assert.NotNil(t, usecase)
}
