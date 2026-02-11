package helper_test

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCasbinEnforcer is a mock implementation for testing RBACHelper
type MockCasbinEnforcer struct {
	mock.Mock
}

func (m *MockCasbinEnforcer) AddPermissionForRole(roleName, resource, action string) error {
	args := m.Called(roleName, resource, action)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) RemoveAllPermissionsForRole(roleName string) error {
	args := m.Called(roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	args := m.Called(roleName, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockCasbinEnforcer) SavePolicy() error {
	args := m.Called()
	return args.Error(0)
}

// CasbinEnforcerInterface defines the interface for testing
type CasbinEnforcerInterface interface {
	AddPermissionForRole(roleName, resource, action string) error
	RemoveAllPermissionsForRole(roleName string) error
	CheckPermission(roleName, resource, action string) (bool, error)
	SavePolicy() error
}

// TestableRBACHelper is a test version of RBACHelper that uses an interface
type TestableRBACHelper struct {
	casbinEnforcer CasbinEnforcerInterface
}

func NewTestableRBACHelper(casbinEnforcer CasbinEnforcerInterface) *TestableRBACHelper {
	return &TestableRBACHelper{
		casbinEnforcer: casbinEnforcer,
	}
}

func (rh *TestableRBACHelper) SyncPermissionsToRole(roleName string, permissions map[string][]string, permissionRepo interface{}) ([]domain.RolePermissionDetail, int, error) {
	type PermissionRepo interface {
		GetByResourceAndAction(resource, action string) (*domain.Permission, error)
	}

	repo, ok := permissionRepo.(PermissionRepo)
	if !ok {
		return nil, 0, errors.New("invalid permission repository")
	}

	var permissionDetails []domain.RolePermissionDetail
	permissionCount := 0

	for resource, actions := range permissions {
		var resourceActions []string
		for _, action := range actions {
			_, err := repo.GetByResourceAndAction(resource, action)
			if err != nil {
				continue
			}

			if err := rh.casbinEnforcer.AddPermissionForRole(roleName, resource, action); err != nil {
				continue
			}

			resourceActions = append(resourceActions, action)
			permissionCount++
		}

		if len(resourceActions) > 0 {
			permissionDetails = append(permissionDetails, domain.RolePermissionDetail{
				Resource: resource,
				Actions:  resourceActions,
			})
		}
	}

	return permissionDetails, permissionCount, nil
}

func (rh *TestableRBACHelper) RemoveAllRolePermissions(roleID uint, roleName string, rolePermissionRepo interface{}) error {
	type RolePermissionRepo interface {
		DeleteByRoleID(roleID uint) error
	}

	repo, ok := rolePermissionRepo.(RolePermissionRepo)
	if !ok {
		return errors.New("invalid role permission repository")
	}

	if err := repo.DeleteByRoleID(roleID); err != nil {
		return errors.New("gagal menghapus permission lama")
	}

	if err := rh.casbinEnforcer.RemoveAllPermissionsForRole(roleName); err != nil {
		return errors.New("gagal menghapus policy casbin lama")
	}

	return nil
}

func (rh *TestableRBACHelper) CheckUserPermission(roleName, resource, action string) (bool, error) {
	return rh.casbinEnforcer.CheckPermission(roleName, resource, action)
}

func (rh *TestableRBACHelper) SavePolicy() error {
	return rh.casbinEnforcer.SavePolicy()
}

// MockPermissionRepo is a mock implementation of the permission repository
type MockPermissionRepo struct {
	mock.Mock
}

func (m *MockPermissionRepo) GetByResourceAndAction(resource, action string) (*domain.Permission, error) {
	args := m.Called(resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

// MockRolePermissionRepo is a mock implementation of the role permission repository
type MockRolePermissionRepo struct {
	mock.Mock
}

func (m *MockRolePermissionRepo) DeleteByRoleID(roleID uint) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func (m *MockRolePermissionRepo) Create(rolePermission *domain.RolePermission) error {
	args := m.Called(rolePermission)
	return args.Error(0)
}

// =============================================================================
// Tests for ValidatePermissionFormat
// =============================================================================

func TestValidatePermissionFormat_EmptyPermissions(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	err := rh.ValidatePermissionFormat(map[string][]string{})

	assert.Error(t, err)
	assert.Equal(t, "permission tidak boleh kosong", err.Error())
}

func TestValidatePermissionFormat_NilPermissions(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	err := rh.ValidatePermissionFormat(nil)

	assert.Error(t, err)
	assert.Equal(t, "permission tidak boleh kosong", err.Error())
}

func TestValidatePermissionFormat_EmptyResourceName(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	permissions := map[string][]string{
		"": {"read", "write"},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.Error(t, err)
	assert.Equal(t, "nama resource tidak boleh kosong", err.Error())
}

func TestValidatePermissionFormat_EmptyActionsForResource(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.Error(t, err)
	assert.Equal(t, "action untuk resource users tidak boleh kosong", err.Error())
}

func TestValidatePermissionFormat_EmptyActionString(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {"read", ""},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.Error(t, err)
	assert.Equal(t, "action tidak boleh kosong untuk resource users", err.Error())
}

func TestValidatePermissionFormat_ValidPermissions(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	permissions := map[string][]string{
		"users":    {"read", "write", "delete"},
		"projects": {"read", "create"},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.NoError(t, err)
}

func TestValidatePermissionFormat_SingleResourceSingleAction(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {"read"},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.NoError(t, err)
}

// =============================================================================
// Tests for SyncPermissionsToRole (using TestableRBACHelper with mock)
// =============================================================================

func TestSyncPermissionsToRole_InvalidPermissionRepo(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read"},
	}

	// Pass an invalid repository (not implementing the required interface)
	_, _, err := rh.SyncPermissionsToRole("admin", permissions, "invalid")

	assert.Error(t, err)
	assert.Equal(t, "invalid permission repository", err.Error())
}

func TestSyncPermissionsToRole_PermissionNotFound(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockPermRepo := new(MockPermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read"},
	}

	// Permission not found in repository
	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(nil, errors.New("not found"))

	details, count, err := rh.SyncPermissionsToRole("admin", permissions, mockPermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, details)
	mockPermRepo.AssertExpectations(t)
}

func TestSyncPermissionsToRole_CasbinAddError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockPermRepo := new(MockPermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read"},
	}

	permission := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission, nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(errors.New("casbin error"))

	details, count, err := rh.SyncPermissionsToRole("admin", permissions, mockPermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, details)
	mockPermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

func TestSyncPermissionsToRole_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockPermRepo := new(MockPermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read", "write"},
	}

	permission1 := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	permission2 := &domain.Permission{ID: 2, Resource: "users", Action: "write"}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission1, nil)
	mockPermRepo.On("GetByResourceAndAction", "users", "write").Return(permission2, nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "write").Return(nil)

	details, count, err := rh.SyncPermissionsToRole("admin", permissions, mockPermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 1)
	assert.Equal(t, "users", details[0].Resource)
	assert.Len(t, details[0].Actions, 2)
	mockPermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

func TestSyncPermissionsToRole_MultipleResources(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockPermRepo := new(MockPermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users":    {"read"},
		"projects": {"create"},
	}

	permission1 := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	permission2 := &domain.Permission{ID: 2, Resource: "projects", Action: "create"}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission1, nil)
	mockPermRepo.On("GetByResourceAndAction", "projects", "create").Return(permission2, nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "projects", "create").Return(nil)

	details, count, err := rh.SyncPermissionsToRole("admin", permissions, mockPermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 2)
	mockPermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

func TestSyncPermissionsToRole_PartialSuccess(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockPermRepo := new(MockPermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read", "write", "delete"},
	}

	permission1 := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	permission2 := &domain.Permission{ID: 2, Resource: "users", Action: "write"}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission1, nil)
	mockPermRepo.On("GetByResourceAndAction", "users", "write").Return(permission2, nil)
	mockPermRepo.On("GetByResourceAndAction", "users", "delete").Return(nil, errors.New("not found"))

	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "write").Return(nil)

	details, count, err := rh.SyncPermissionsToRole("admin", permissions, mockPermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 1)
	assert.Equal(t, "users", details[0].Resource)
	assert.Len(t, details[0].Actions, 2)
	mockPermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

// =============================================================================
// Tests for RemoveAllRolePermissions (using TestableRBACHelper with mock)
// =============================================================================

func TestRemoveAllRolePermissions_InvalidRepo(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	err := rh.RemoveAllRolePermissions(1, "admin", "invalid")

	assert.Error(t, err)
	assert.Equal(t, "invalid role permission repository", err.Error())
}

func TestRemoveAllRolePermissions_DeleteByRoleIDError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockRolePermRepo := new(MockRolePermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	mockRolePermRepo.On("DeleteByRoleID", uint(1)).Return(errors.New("db error"))

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus permission lama", err.Error())
	mockRolePermRepo.AssertExpectations(t)
}

func TestRemoveAllRolePermissions_CasbinRemoveError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockRolePermRepo := new(MockRolePermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	mockRolePermRepo.On("DeleteByRoleID", uint(1)).Return(nil)
	mockCasbin.On("RemoveAllPermissionsForRole", "admin").Return(errors.New("casbin error"))

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus policy casbin lama", err.Error())
	mockRolePermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

func TestRemoveAllRolePermissions_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	mockRolePermRepo := new(MockRolePermissionRepo)
	rh := NewTestableRBACHelper(mockCasbin)

	mockRolePermRepo.On("DeleteByRoleID", uint(1)).Return(nil)
	mockCasbin.On("RemoveAllPermissionsForRole", "admin").Return(nil)

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.NoError(t, err)
	mockRolePermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

// =============================================================================
// Tests for CheckUserPermission (using TestableRBACHelper with mock)
// =============================================================================

func TestCheckUserPermission_Allowed(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "admin", "users", "read").Return(true, nil)

	allowed, err := rh.CheckUserPermission("admin", "users", "read")

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestCheckUserPermission_NotAllowed(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "user", "admin", "delete").Return(false, nil)

	allowed, err := rh.CheckUserPermission("user", "admin", "delete")

	assert.NoError(t, err)
	assert.False(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestCheckUserPermission_Error(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "admin", "users", "read").Return(false, errors.New("casbin error"))

	allowed, err := rh.CheckUserPermission("admin", "users", "read")

	assert.Error(t, err)
	assert.False(t, allowed)
	mockCasbin.AssertExpectations(t)
}

// =============================================================================
// Tests for SavePolicy (using TestableRBACHelper with mock)
// =============================================================================

func TestSavePolicy_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	mockCasbin.On("SavePolicy").Return(nil)

	err := rh.SavePolicy()

	assert.NoError(t, err)
	mockCasbin.AssertExpectations(t)
}

func TestSavePolicy_Error(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcer)
	rh := NewTestableRBACHelper(mockCasbin)

	mockCasbin.On("SavePolicy").Return(errors.New("save error"))

	err := rh.SavePolicy()

	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
	mockCasbin.AssertExpectations(t)
}

// =============================================================================
// Tests for BuildRoleDetailResponse
// =============================================================================

func TestBuildRoleDetailResponse_WithPermissions(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	role := &domain.Role{
		ID:        1,
		NamaRole:  "admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	permissions := []domain.Permission{
		{ID: 1, Resource: "users", Action: "read"},
		{ID: 2, Resource: "users", Action: "write"},
		{ID: 3, Resource: "projects", Action: "create"},
	}

	response := rh.BuildRoleDetailResponse(role, permissions)

	assert.Equal(t, role.ID, response.ID)
	assert.Equal(t, role.NamaRole, response.NamaRole)
	assert.Equal(t, 3, response.JumlahPermission)
	assert.Len(t, response.Permissions, 2) // 2 unique resources
	assert.Equal(t, role.CreatedAt, response.CreatedAt)
	assert.Equal(t, role.UpdatedAt, response.UpdatedAt)
}

func TestBuildRoleDetailResponse_EmptyPermissions(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	role := &domain.Role{
		ID:        1,
		NamaRole:  "guest",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	permissions := []domain.Permission{}

	response := rh.BuildRoleDetailResponse(role, permissions)

	assert.Equal(t, role.ID, response.ID)
	assert.Equal(t, role.NamaRole, response.NamaRole)
	assert.Equal(t, 0, response.JumlahPermission)
	assert.Empty(t, response.Permissions)
}

func TestBuildRoleDetailResponse_SingleResource(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	role := &domain.Role{
		ID:       1,
		NamaRole: "viewer",
	}

	permissions := []domain.Permission{
		{ID: 1, Resource: "reports", Action: "read"},
		{ID: 2, Resource: "reports", Action: "export"},
	}

	response := rh.BuildRoleDetailResponse(role, permissions)

	assert.Equal(t, role.ID, response.ID)
	assert.Equal(t, role.NamaRole, response.NamaRole)
	assert.Equal(t, 2, response.JumlahPermission)
	assert.Len(t, response.Permissions, 1)
	assert.Equal(t, "reports", response.Permissions[0].Resource)
	assert.Len(t, response.Permissions[0].Actions, 2)
}

// =============================================================================
// Tests for CreateRolePermissions (using real RBACHelper)
// =============================================================================

func TestCreateRolePermissions_InvalidPermissionRepo(t *testing.T) {
	rh := helper.NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {"read"},
	}

	_, _, err := rh.CreateRolePermissions(1, permissions, "invalid", nil)

	assert.Error(t, err)
	assert.Equal(t, "invalid permission repository", err.Error())
}

func TestCreateRolePermissions_InvalidRolePermissionRepo(t *testing.T) {
	rh := helper.NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepo)

	permissions := map[string][]string{
		"users": {"read"},
	}

	_, _, err := rh.CreateRolePermissions(1, permissions, mockPermRepo, "invalid")

	assert.Error(t, err)
	assert.Equal(t, "invalid role permission repository", err.Error())
}

func TestCreateRolePermissions_PermissionNotFound(t *testing.T) {
	rh := helper.NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepo)
	mockRolePermRepo := new(MockRolePermissionRepo)

	permissions := map[string][]string{
		"users": {"read"},
	}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(nil, errors.New("not found"))

	details, count, err := rh.CreateRolePermissions(1, permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, details)
	mockPermRepo.AssertExpectations(t)
}

func TestCreateRolePermissions_CreateError(t *testing.T) {
	rh := helper.NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepo)
	mockRolePermRepo := new(MockRolePermissionRepo)

	permissions := map[string][]string{
		"users": {"read"},
	}

	permission := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission, nil)
	mockRolePermRepo.On("Create", mock.AnythingOfType("*domain.RolePermission")).Return(errors.New("db error"))

	details, count, err := rh.CreateRolePermissions(1, permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, details)
	mockPermRepo.AssertExpectations(t)
	mockRolePermRepo.AssertExpectations(t)
}

func TestCreateRolePermissions_Success(t *testing.T) {
	rh := helper.NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepo)
	mockRolePermRepo := new(MockRolePermissionRepo)

	permissions := map[string][]string{
		"users": {"read", "write"},
	}

	permission1 := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	permission2 := &domain.Permission{ID: 2, Resource: "users", Action: "write"}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission1, nil)
	mockPermRepo.On("GetByResourceAndAction", "users", "write").Return(permission2, nil)
	mockRolePermRepo.On("Create", mock.AnythingOfType("*domain.RolePermission")).Return(nil).Twice()

	details, count, err := rh.CreateRolePermissions(1, permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 1)
	assert.Equal(t, "users", details[0].Resource)
	assert.Len(t, details[0].Actions, 2)
	mockPermRepo.AssertExpectations(t)
	mockRolePermRepo.AssertExpectations(t)
}
