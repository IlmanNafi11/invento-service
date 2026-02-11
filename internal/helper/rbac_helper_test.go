package helper

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockCasbinEnforcerForRBAC is a mock implementation of CasbinEnforcerInterface for testing RBACHelper
type MockCasbinEnforcerForRBAC struct {
	mock.Mock
}

func (m *MockCasbinEnforcerForRBAC) AddPermissionForRole(roleName, resource, action string) error {
	args := m.Called(roleName, resource, action)
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) RemovePermissionForRole(roleName, resource, action string) error {
	args := m.Called(roleName, resource, action)
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) RemoveAllPermissionsForRole(roleName string) error {
	args := m.Called(roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) GetPermissionsForRole(roleName string) ([][]string, error) {
	args := m.Called(roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]string), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) CheckPermission(roleName, resource, action string) (bool, error) {
	args := m.Called(roleName, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) SavePolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) LoadPolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) DeleteRole(roleName string) error {
	args := m.Called(roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) GetEnforcer() *casbin.Enforcer {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*casbin.Enforcer)
}

func (m *MockCasbinEnforcerForRBAC) GetAllRoles() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) GetAllPolicies() ([][]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]string), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) HasPolicy(roleName, resource, action string) (bool, error) {
	args := m.Called(roleName, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) AddRoleForUser(userID, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) RemoveRoleForUser(userID, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcerForRBAC) GetRolesForUser(userID string) ([]string, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) GetUsersForRole(roleName string) ([]string, error) {
	args := m.Called(roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCasbinEnforcerForRBAC) DeleteAllRolesForUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

// MockPermissionRepoForRBAC is a mock implementation of the permission repository
type MockPermissionRepoForRBAC struct {
	mock.Mock
}

func (m *MockPermissionRepoForRBAC) GetByResourceAndAction(resource, action string) (*domain.Permission, error) {
	args := m.Called(resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

// MockRolePermissionRepoForRBAC is a mock implementation of the role permission repository
type MockRolePermissionRepoForRBAC struct {
	mock.Mock
}

func (m *MockRolePermissionRepoForRBAC) DeleteByRoleID(roleID uint) error {
	args := m.Called(roleID)
	return args.Error(0)
}

func (m *MockRolePermissionRepoForRBAC) Create(rolePermission *domain.RolePermission) error {
	args := m.Called(rolePermission)
	return args.Error(0)
}

// =============================================================================
// Tests for NewRBACHelper
// =============================================================================

func TestNewRBACHelper(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	assert.NotNil(t, rh)
	assert.Equal(t, mockCasbin, rh.casbinEnforcer)
}

func TestNewRBACHelper_NilEnforcer(t *testing.T) {
	rh := NewRBACHelper(nil)

	assert.NotNil(t, rh)
	assert.Nil(t, rh.casbinEnforcer)
}

// =============================================================================
// Tests for ValidatePermissionFormat
// =============================================================================

func TestRBACHelper_ValidatePermissionFormat_EmptyPermissions(t *testing.T) {
	rh := NewRBACHelper(nil)

	err := rh.ValidatePermissionFormat(map[string][]string{})

	assert.Error(t, err)
	assert.Equal(t, "permission tidak boleh kosong", err.Error())
}

func TestRBACHelper_ValidatePermissionFormat_NilPermissions(t *testing.T) {
	rh := NewRBACHelper(nil)

	err := rh.ValidatePermissionFormat(nil)

	assert.Error(t, err)
	assert.Equal(t, "permission tidak boleh kosong", err.Error())
}

func TestRBACHelper_ValidatePermissionFormat_EmptyResourceName(t *testing.T) {
	rh := NewRBACHelper(nil)

	permissions := map[string][]string{
		"": {"read", "write"},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.Error(t, err)
	assert.Equal(t, "nama resource tidak boleh kosong", err.Error())
}

func TestRBACHelper_ValidatePermissionFormat_EmptyActionsForResource(t *testing.T) {
	rh := NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.Error(t, err)
	assert.Equal(t, "action untuk resource users tidak boleh kosong", err.Error())
}

func TestRBACHelper_ValidatePermissionFormat_EmptyActionString(t *testing.T) {
	rh := NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {"read", ""},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.Error(t, err)
	assert.Equal(t, "action tidak boleh kosong untuk resource users", err.Error())
}

func TestRBACHelper_ValidatePermissionFormat_ValidPermissions(t *testing.T) {
	rh := NewRBACHelper(nil)

	permissions := map[string][]string{
		"users":    {"read", "write", "delete"},
		"projects": {"read", "create"},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.NoError(t, err)
}

func TestRBACHelper_ValidatePermissionFormat_SingleResourceSingleAction(t *testing.T) {
	rh := NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {"read"},
	}

	err := rh.ValidatePermissionFormat(permissions)

	assert.NoError(t, err)
}

// =============================================================================
// Tests for SyncPermissionsToRole
// =============================================================================

func TestRBACHelper_SyncPermissionsToRole_InvalidPermissionRepo(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read"},
	}

	_, _, err := rh.SyncPermissionsToRole("admin", permissions, "invalid")

	assert.Error(t, err)
	assert.Equal(t, "invalid permission repository", err.Error())
}

func TestRBACHelper_SyncPermissionsToRole_PermissionNotFound(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read"},
	}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(nil, errors.New("not found"))

	details, count, err := rh.SyncPermissionsToRole("admin", permissions, mockPermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 0, count)
	assert.Empty(t, details)
	mockPermRepo.AssertExpectations(t)
}

func TestRBACHelper_SyncPermissionsToRole_CasbinAddError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

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

func TestRBACHelper_SyncPermissionsToRole_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

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

func TestRBACHelper_SyncPermissionsToRole_MultipleResources(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

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

func TestRBACHelper_SyncPermissionsToRole_PartialSuccess(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

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
// Tests for CreateRolePermissions
// =============================================================================

func TestRBACHelper_CreateRolePermissions_InvalidPermissionRepo(t *testing.T) {
	rh := NewRBACHelper(nil)

	permissions := map[string][]string{
		"users": {"read"},
	}

	_, _, err := rh.CreateRolePermissions(1, permissions, "invalid", nil)

	assert.Error(t, err)
	assert.Equal(t, "invalid permission repository", err.Error())
}

func TestRBACHelper_CreateRolePermissions_InvalidRolePermissionRepo(t *testing.T) {
	rh := NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepoForRBAC)

	permissions := map[string][]string{
		"users": {"read"},
	}

	_, _, err := rh.CreateRolePermissions(1, permissions, mockPermRepo, "invalid")

	assert.Error(t, err)
	assert.Equal(t, "invalid role permission repository", err.Error())
}

func TestRBACHelper_CreateRolePermissions_PermissionNotFound(t *testing.T) {
	rh := NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)

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

func TestRBACHelper_CreateRolePermissions_CreateError(t *testing.T) {
	rh := NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)

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

func TestRBACHelper_CreateRolePermissions_Success(t *testing.T) {
	rh := NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)

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

func TestRBACHelper_CreateRolePermissions_MultipleResources(t *testing.T) {
	rh := NewRBACHelper(nil)
	mockPermRepo := new(MockPermissionRepoForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)

	permissions := map[string][]string{
		"users":    {"read"},
		"projects": {"create"},
	}

	permission1 := &domain.Permission{ID: 1, Resource: "users", Action: "read"}
	permission2 := &domain.Permission{ID: 2, Resource: "projects", Action: "create"}

	mockPermRepo.On("GetByResourceAndAction", "users", "read").Return(permission1, nil)
	mockPermRepo.On("GetByResourceAndAction", "projects", "create").Return(permission2, nil)
	mockRolePermRepo.On("Create", mock.AnythingOfType("*domain.RolePermission")).Return(nil).Twice()

	details, count, err := rh.CreateRolePermissions(1, permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 2)
	mockPermRepo.AssertExpectations(t)
	mockRolePermRepo.AssertExpectations(t)
}

// =============================================================================
// Tests for RemoveAllRolePermissions
// =============================================================================

func TestRBACHelper_RemoveAllRolePermissions_InvalidRepo(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	err := rh.RemoveAllRolePermissions(1, "admin", "invalid")

	assert.Error(t, err)
	assert.Equal(t, "invalid role permission repository", err.Error())
}

func TestRBACHelper_RemoveAllRolePermissions_DeleteByRoleIDError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockRolePermRepo.On("DeleteByRoleID", uint(1)).Return(errors.New("db error"))

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus permission lama", err.Error())
	mockRolePermRepo.AssertExpectations(t)
}

func TestRBACHelper_RemoveAllRolePermissions_CasbinRemoveError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockRolePermRepo.On("DeleteByRoleID", uint(1)).Return(nil)
	mockCasbin.On("RemoveAllPermissionsForRole", "admin").Return(errors.New("casbin error"))

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus policy casbin lama", err.Error())
	mockRolePermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_RemoveAllRolePermissions_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	mockRolePermRepo := new(MockRolePermissionRepoForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockRolePermRepo.On("DeleteByRoleID", uint(1)).Return(nil)
	mockCasbin.On("RemoveAllPermissionsForRole", "admin").Return(nil)

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.NoError(t, err)
	mockRolePermRepo.AssertExpectations(t)
	mockCasbin.AssertExpectations(t)
}

// =============================================================================
// Tests for BuildRoleDetailResponse
// =============================================================================

func TestRBACHelper_BuildRoleDetailResponse_WithPermissions(t *testing.T) {
	rh := NewRBACHelper(nil)

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

func TestRBACHelper_BuildRoleDetailResponse_EmptyPermissions(t *testing.T) {
	rh := NewRBACHelper(nil)

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

func TestRBACHelper_BuildRoleDetailResponse_SingleResource(t *testing.T) {
	rh := NewRBACHelper(nil)

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
// Tests for CheckUserPermission
// =============================================================================

func TestRBACHelper_CheckUserPermission_Allowed(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "admin", "users", "read").Return(true, nil)

	allowed, err := rh.CheckUserPermission("admin", "users", "read")

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_CheckUserPermission_NotAllowed(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "user", "admin", "delete").Return(false, nil)

	allowed, err := rh.CheckUserPermission("user", "admin", "delete")

	assert.NoError(t, err)
	assert.False(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_CheckUserPermission_Error(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "admin", "users", "read").Return(false, errors.New("casbin error"))

	allowed, err := rh.CheckUserPermission("admin", "users", "read")

	assert.Error(t, err)
	assert.False(t, allowed)
	mockCasbin.AssertExpectations(t)
}

// =============================================================================
// Tests for SavePolicy
// =============================================================================

func TestRBACHelper_SavePolicy_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("SavePolicy").Return(nil)

	err := rh.SavePolicy()

	assert.NoError(t, err)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_SavePolicy_Error(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("SavePolicy").Return(errors.New("save error"))

	err := rh.SavePolicy()

	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
	mockCasbin.AssertExpectations(t)
}
