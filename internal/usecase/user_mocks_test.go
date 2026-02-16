package usecase

import (
	"context"

	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock for UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) GetProfileWithCounts(ctx context.Context, userID string) (user *domain.User, projectCount int, modulCount int, err error) {
	args := m.Called(ctx, userID)
	if args.Get(0) != nil {
		return args.Get(0).(*domain.User), args.Int(1), args.Int(2), args.Error(3)
	}
	return nil, args.Int(1), args.Int(2), args.Error(3)
}

func (m *MockUserRepository) GetUserFiles(ctx context.Context, userID string, search string, page, limit int) ([]dto.UserFileItem, int, error) {
	args := m.Called(ctx, userID, search, page, limit)
	if args.Get(0) != nil {
		return args.Get(0).([]dto.UserFileItem), args.Int(1), args.Error(2)
	}
	return nil, args.Int(1), args.Error(2)
}

func (m *MockUserRepository) GetByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	args := m.Called(ctx, userIDs)
	if args.Get(0) != nil {
		return args.Get(0).([]*domain.User), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, userID string, name string, jenisKelamin *string, fotoProfil *string) error {
	args := m.Called(ctx, userID, name, jenisKelamin, fotoProfil)
	return args.Error(0)
}

func (m *MockUserRepository) GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]dto.UserListItem, int, error) {
	args := m.Called(ctx, search, filterRole, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]dto.UserListItem), args.Int(1), args.Error(2)
}

func (m *MockUserRepository) UpdateRole(ctx context.Context, userID string, roleID *int) error {
	args := m.Called(ctx, userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserRepository) GetByRoleID(ctx context.Context, roleID uint) ([]dto.UserListItem, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.UserListItem), args.Error(1)
}

func (m *MockUserRepository) BulkUpdateRole(ctx context.Context, userIDs []string, roleID uint) error {
	args := m.Called(ctx, userIDs, roleID)
	return args.Error(0)
}
