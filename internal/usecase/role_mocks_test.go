package usecase

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"github.com/stretchr/testify/mock"
)

// MockRoleRepository is a mock for RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) Create(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetByID(ctx context.Context, id uint) (*domain.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) GetByName(ctx context.Context, name string) (*domain.Role, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Role), args.Error(1)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *domain.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) GetAll(ctx context.Context, search string, page, limit int) ([]dto.RoleListItem, int, error) {
	args := m.Called(ctx, search, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]dto.RoleListItem), args.Int(1), args.Error(2)
}

// MockPermissionRepository is a mock for PermissionRepository
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) Create(ctx context.Context, permission *domain.Permission) error {
	args := m.Called(ctx, permission)
	return args.Error(0)
}

func (m *MockPermissionRepository) GetByID(ctx context.Context, id uint) (*domain.Permission, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetByResourceAndAction(ctx context.Context, resource, action string) (*domain.Permission, error) {
	args := m.Called(ctx, resource, action)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetAllByResourceActions(ctx context.Context, permissions map[string][]string) ([]domain.Permission, error) {
	args := m.Called(ctx, permissions)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetAll(ctx context.Context) ([]domain.Permission, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}

func (m *MockPermissionRepository) GetAvailablePermissions(ctx context.Context) ([]dto.ResourcePermissions, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.ResourcePermissions), args.Error(1)
}

func (m *MockPermissionRepository) BulkCreate(ctx context.Context, permissions []domain.Permission) error {
	args := m.Called(ctx, permissions)
	return args.Error(0)
}

// MockRolePermissionRepository is a mock for RolePermissionRepository
type MockRolePermissionRepository struct {
	mock.Mock
}

func (m *MockRolePermissionRepository) Create(ctx context.Context, rolePermission *domain.RolePermission) error {
	args := m.Called(ctx, rolePermission)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) BulkCreate(ctx context.Context, rolePermissions []domain.RolePermission) error {
	args := m.Called(ctx, rolePermissions)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) GetByRoleID(ctx context.Context, roleID uint) ([]domain.RolePermission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RolePermission), args.Error(1)
}

func (m *MockRolePermissionRepository) DeleteByRoleID(ctx context.Context, roleID uint) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

func (m *MockRolePermissionRepository) GetPermissionsForRole(ctx context.Context, roleID uint) ([]domain.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Permission), args.Error(1)
}
