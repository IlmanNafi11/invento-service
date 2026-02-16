package usecase

import (
	"context"
	"testing"

	"invento-service/internal/rbac"

	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStatisticUsecase_GetStatistics_ZeroCounts(t *testing.T) {
	t.Parallel()
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

	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(0, nil)
	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(0, nil)

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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
			t.Parallel()
			mockCasbin := new(MockCasbinEnforcer)
			mockProjectRepo := new(MockProjectRepository)
			mockModulRepo := new(MockModulRepository)

			userUC := newStatisticUsecaseWithInterface(mockUserRepo, mockProjectRepo, mockModulRepo, mockRoleRepo, mockCasbin, db)

			mockCasbin.On("CheckPermission", tc.userRole, "Project", "read").Return(true, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "Modul", "read").Return(true, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "User", "read").Return(true, nil)
			mockCasbin.On("CheckPermission", tc.userRole, "Role", "read").Return(true, nil)

			mockProjectRepo.On("CountByUserID", mock.Anything, tc.userID).Return(tc.projectCount, nil)
			mockModulRepo.On("CountByUserID", mock.Anything, tc.userID).Return(tc.modulCount, nil)

			result, err := userUC.GetStatistics(context.Background(), tc.userID, tc.userRole)

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
	t.Parallel()
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
	t.Parallel()
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
			t.Parallel()
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
				mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(tc.projectCount, nil)
			}
			if tc.hasModulRead {
				mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(tc.modulCount, nil)
			}

			result, err := userUC.GetStatistics(context.Background(), userID, tc.userRole)

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
	t.Parallel()
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
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	mockRoleRepo := new(MockRoleRepository)

	// Create a real Casbin enforcer for testing
	casbinDB, err := gorm.Open(sqlite.Open("file:stat_user_real?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open casbin test database: %v", err)
	}
	casbinDB.AutoMigrate(&gormadapter.CasbinRule{})
	casbinEnforcer, err := rbac.NewCasbinEnforcer(casbinDB)
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

	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(10, nil)
	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(25, nil)

	result, err := usecase.GetStatistics(context.Background(), userID, userRole)

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
