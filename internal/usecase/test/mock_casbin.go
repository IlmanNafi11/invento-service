package mocks

import (
	"fiber-boiler-plate/internal/helper"

	"github.com/casbin/casbin/v2"
	"github.com/stretchr/testify/mock"
)

type MockCasbinEnforcer struct {
	mock.Mock
	MockEnforcer *casbin.Enforcer
}

func (m *MockCasbinEnforcer) AddPermissionForRole(roleName, resource, action string) error {
	args := m.Called(roleName, resource, action)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) RemovePermissionForRole(roleName, resource, action string) error {
	args := m.Called(roleName, resource, action)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) RemoveAllPermissionsForRole(roleName string) error {
	args := m.Called(roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) GetPermissionsForRole(roleName string) ([][]string, error) {
	args := m.Called(roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]string), args.Error(1)
}

func (m *MockCasbinEnforcer) CheckPermission(roleName, resource, action string) (bool, error) {
	args := m.Called(roleName, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockCasbinEnforcer) SavePolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasbinEnforcer) LoadPolicy() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockCasbinEnforcer) DeleteRole(roleName string) error {
	args := m.Called(roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) GetEnforcer() *casbin.Enforcer {
	if m.MockEnforcer != nil {
		return m.MockEnforcer
	}
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*casbin.Enforcer)
}

func (m *MockCasbinEnforcer) GetAllRoles() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCasbinEnforcer) GetAllPolicies() ([][]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([][]string), args.Error(1)
}

func (m *MockCasbinEnforcer) HasPolicy(roleName, resource, action string) (bool, error) {
	args := m.Called(roleName, resource, action)
	return args.Bool(0), args.Error(1)
}

func (m *MockCasbinEnforcer) AddRoleForUser(userID, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) RemoveRoleForUser(userID, roleName string) error {
	args := m.Called(userID, roleName)
	return args.Error(0)
}

func (m *MockCasbinEnforcer) GetRolesForUser(userID string) ([]string, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCasbinEnforcer) GetUsersForRole(roleName string) ([]string, error) {
	args := m.Called(roleName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockCasbinEnforcer) DeleteAllRolesForUser(userID string) error {
	args := m.Called(userID)
	return args.Error(0)
}

func NewMockCasbinEnforcer() *MockCasbinEnforcer {
	return &MockCasbinEnforcer{}
}

var _ helper.CasbinEnforcerInterface = (*MockCasbinEnforcer)(nil)
