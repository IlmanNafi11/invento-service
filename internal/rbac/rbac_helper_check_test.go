package rbac

import (
	"errors"
	"invento-service/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRBACHelper_CheckUserPermission_Allowed(t *testing.T) {
	t.Parallel()
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "admin", "users", "read").Return(true, nil)

	allowed, err := rh.CheckUserPermission("admin", "users", "read")

	assert.NoError(t, err)
	assert.True(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_CheckUserPermission_NotAllowed(t *testing.T) {
	t.Parallel()
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "user", "admin", "delete").Return(false, nil)

	allowed, err := rh.CheckUserPermission("user", "admin", "delete")

	assert.NoError(t, err)
	assert.False(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_CheckUserPermission_Error(t *testing.T) {
	t.Parallel()
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("CheckPermission", "admin", "users", "read").Return(false, errors.New("casbin error"))

	allowed, err := rh.CheckUserPermission("admin", "users", "read")

	assert.Error(t, err)
	assert.False(t, allowed)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_SavePolicy_Success(t *testing.T) {
	t.Parallel()
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("SavePolicy").Return(nil)

	err := rh.SavePolicy()

	assert.NoError(t, err)
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_SavePolicy_Error(t *testing.T) {
	t.Parallel()
	mockCasbin := new(MockCasbinEnforcerForRBAC)
	rh := NewRBACHelper(mockCasbin)

	mockCasbin.On("SavePolicy").Return(errors.New("save error"))

	err := rh.SavePolicy()

	assert.Error(t, err)
	assert.Equal(t, "save error", err.Error())
	mockCasbin.AssertExpectations(t)
}

func TestRBACHelper_BuildRoleDetailResponse_WithPermissions(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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
	t.Parallel()
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
