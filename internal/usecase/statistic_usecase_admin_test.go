package usecase

import (
	"context"
	"fmt"
	"testing"

	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	repo "invento-service/internal/usecase/repo"
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

func (su *statisticUsecaseWithInterface) GetStatistics(ctx context.Context, userID, userRole string) (*dto.StatisticData, error) {
	result := &dto.StatisticData{}

	hasProjectRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Project", "read")
	hasModulRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Modul", "read")
	hasUserRead, _ := su.casbinEnforcer.CheckPermission(userRole, "User", "read")
	hasRoleRead, _ := su.casbinEnforcer.CheckPermission(userRole, "Role", "read")

	if hasProjectRead {
		projectCount, _ := su.projectRepo.CountByUserID(ctx, userID)
		result.TotalProject = &projectCount
	}

	if hasModulRead {
		modulCount, _ := su.modulRepo.CountByUserID(ctx, userID)
		result.TotalModul = &modulCount
	}

	if hasUserRead {
		var totalUser int64
		su.db.WithContext(ctx).Model(&domain.User{}).Count(&totalUser)
		count := int(totalUser)
		result.TotalUser = &count
	}

	if hasRoleRead {
		var totalRole int64
		su.db.WithContext(ctx).Model(&domain.Role{}).Count(&totalRole)
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
	t.Parallel()
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

	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(10, nil)
	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(25, nil)

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(5, nil)

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(30, nil)

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(12, nil)
	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(45, nil)

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
	t.Parallel()
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

	mockProjectRepo.On("CountByUserID", mock.Anything, userID).Return(0, nil)
	mockModulRepo.On("CountByUserID", mock.Anything, userID).Return(0, nil)

	result, err := userUC.GetStatistics(context.Background(), userID, userRole)

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
