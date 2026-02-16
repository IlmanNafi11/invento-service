package usecase

import (
	"invento-service/internal/rbac"
	"testing"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

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
	casbinEnforcer, err := rbac.NewCasbinEnforcer(casbinDB)
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
	casbinEnforcer, err := rbac.NewCasbinEnforcer(casbinDB)
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
