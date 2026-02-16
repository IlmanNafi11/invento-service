package rbac

import (
	"errors"
	"invento-service/internal/domain"
	"testing"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

type MockPermissionRepoForRBAC struct {
	GetAllByResourceActionsFunc func(permissions map[string][]string) ([]domain.Permission, error)
}

func (m *MockPermissionRepoForRBAC) GetAllByResourceActions(permissions map[string][]string) ([]domain.Permission, error) {
	if m.GetAllByResourceActionsFunc == nil {
		return nil, nil
	}
	return m.GetAllByResourceActionsFunc(permissions)
}

type MockRolePermissionRepoForRBAC struct {
	BulkCreateFunc     func(rolePermissions []domain.RolePermission) error
	DeleteByRoleIDFunc func(roleID uint) error
}

func (m *MockRolePermissionRepoForRBAC) BulkCreate(rolePermissions []domain.RolePermission) error {
	if m.BulkCreateFunc == nil {
		return nil
	}
	return m.BulkCreateFunc(rolePermissions)
}

func (m *MockRolePermissionRepoForRBAC) DeleteByRoleID(roleID uint) error {
	if m.DeleteByRoleIDFunc == nil {
		return nil
	}
	return m.DeleteByRoleIDFunc(roleID)
}

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

func TestRBACHelper_SetRolePermissions_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read", "write"},
	}

	mockPermRepo := &MockPermissionRepoForRBAC{
		GetAllByResourceActionsFunc: func(input map[string][]string) ([]domain.Permission, error) {
			assert.Equal(t, permissions, input)
			return []domain.Permission{
				{ID: 1, Resource: "users", Action: "read"},
				{ID: 2, Resource: "users", Action: "write"},
			}, nil
		},
	}

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		BulkCreateFunc: func(rolePermissions []domain.RolePermission) error {
			assert.Len(t, rolePermissions, 2)
			assert.Equal(t, uint(10), rolePermissions[0].RoleID)
			assert.Equal(t, uint(10), rolePermissions[1].RoleID)
			return nil
		},
	}

	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "write").Return(nil)

	details, count, err := rh.SetRolePermissions(10, "admin", permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 1)
	assert.Equal(t, "users", details[0].Resource)
	assert.ElementsMatch(t, []string{"read", "write"}, details[0].Actions)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_SetRolePermissions_PermissionRepoError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockPermRepo := &MockPermissionRepoForRBAC{
		GetAllByResourceActionsFunc: func(input map[string][]string) ([]domain.Permission, error) {
			return nil, errors.New("db error")
		},
	}

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{}

	details, count, err := rh.SetRolePermissions(1, "admin", map[string][]string{"users": {"read"}}, mockPermRepo, mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal mengambil data permission", err.Error())
	assert.Nil(t, details)
	assert.Equal(t, 0, count)
}

func TestRBACHelper_SetRolePermissions_BulkCreateError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	permissions := map[string][]string{"users": {"read"}}
	mockPermRepo := &MockPermissionRepoForRBAC{
		GetAllByResourceActionsFunc: func(input map[string][]string) ([]domain.Permission, error) {
			return []domain.Permission{{ID: 1, Resource: "users", Action: "read"}}, nil
		},
	}
	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		BulkCreateFunc: func(rolePermissions []domain.RolePermission) error {
			return errors.New("insert error")
		},
	}

	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)

	details, count, err := rh.SetRolePermissions(1, "admin", permissions, mockPermRepo, mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal membuat role permission", err.Error())
	assert.Nil(t, details)
	assert.Equal(t, 0, count)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_SetRolePermissions_EmptyPermissions(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockPermRepo := &MockPermissionRepoForRBAC{
		GetAllByResourceActionsFunc: func(input map[string][]string) ([]domain.Permission, error) {
			assert.Empty(t, input)
			return nil, nil
		},
	}
	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		BulkCreateFunc: func(rolePermissions []domain.RolePermission) error {
			t.Fatalf("BulkCreate tidak boleh dipanggil untuk permission kosong")
			return nil
		},
	}

	details, count, err := rh.SetRolePermissions(1, "admin", map[string][]string{}, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Empty(t, details)
	assert.Equal(t, 0, count)
}

func TestRBACHelper_SetRolePermissions_MultipleResourcesMultipleActions(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users":    {"read", "write"},
		"projects": {"create", "delete"},
	}

	mockPermRepo := &MockPermissionRepoForRBAC{
		GetAllByResourceActionsFunc: func(input map[string][]string) ([]domain.Permission, error) {
			return []domain.Permission{
				{ID: 1, Resource: "users", Action: "read"},
				{ID: 2, Resource: "users", Action: "write"},
				{ID: 3, Resource: "projects", Action: "create"},
				{ID: 4, Resource: "projects", Action: "delete"},
			}, nil
		},
	}

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		BulkCreateFunc: func(rolePermissions []domain.RolePermission) error {
			assert.Len(t, rolePermissions, 4)
			return nil
		},
	}

	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "write").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "projects", "create").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "projects", "delete").Return(nil)

	details, count, err := rh.SetRolePermissions(1, "admin", permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 4, count)
	assert.Len(t, details, 2)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_SetRolePermissions_PermissionsNotFoundInDBAreSkipped(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	permissions := map[string][]string{
		"users": {"read", "write", "delete"},
	}

	mockPermRepo := &MockPermissionRepoForRBAC{
		GetAllByResourceActionsFunc: func(input map[string][]string) ([]domain.Permission, error) {
			return []domain.Permission{
				{ID: 1, Resource: "users", Action: "read"},
				{ID: 2, Resource: "users", Action: "write"},
			}, nil
		},
	}

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		BulkCreateFunc: func(rolePermissions []domain.RolePermission) error {
			assert.Len(t, rolePermissions, 2)
			return nil
		},
	}

	mockCasbin.On("AddPermissionForRole", "admin", "users", "read").Return(nil)
	mockCasbin.On("AddPermissionForRole", "admin", "users", "write").Return(nil)

	details, count, err := rh.SetRolePermissions(1, "admin", permissions, mockPermRepo, mockRolePermRepo)

	assert.NoError(t, err)
	assert.Equal(t, 2, count)
	assert.Len(t, details, 1)
	assert.ElementsMatch(t, []string{"read", "write"}, details[0].Actions)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_RemoveAllRolePermissions_DeleteByRoleIDError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		DeleteByRoleIDFunc: func(roleID uint) error {
			assert.Equal(t, uint(1), roleID)
			return errors.New("db error")
		},
	}

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus permission lama", err.Error())
}

func TestRBACHelper_RemoveAllRolePermissions_CasbinRemoveError(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		DeleteByRoleIDFunc: func(roleID uint) error {
			return nil
		},
	}
	mockCasbin.On("RemoveAllPermissionsForRole", "admin").Return(errors.New("casbin error"))

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.Error(t, err)
	assert.Equal(t, "gagal menghapus policy casbin lama", err.Error())
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_RemoveAllRolePermissions_Success(t *testing.T) {
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockRolePermRepo := &MockRolePermissionRepoForRBAC{
		DeleteByRoleIDFunc: func(roleID uint) error {
			assert.Equal(t, uint(1), roleID)
			return nil
		},
	}
	mockCasbin.On("RemoveAllPermissionsForRole", "admin").Return(nil)

	err := rh.RemoveAllRolePermissions(1, "admin", mockRolePermRepo)

	assert.NoError(t, err)
	mockCasbin.AssertExpectations(t)
}

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
	assert.Len(t, response.Permissions, 2)
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
