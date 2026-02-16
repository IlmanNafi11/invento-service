package usecase

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"github.com/stretchr/testify/mock"
)

// MockProjectRepository is a mock for ProjectRepository
type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByUserID(ctx context.Context, userID string, search string, filterSemester int, filterKategori string, page, limit int) ([]dto.ProjectListItem, int, error) {
	args := m.Called(ctx, userID, search, filterSemester, filterKategori, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]dto.ProjectListItem), args.Int(1), args.Error(2)
}

func (m *MockProjectRepository) GetByID(ctx context.Context, id uint) (*domain.Project, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Project), args.Error(1)
}

func (m *MockProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	args := m.Called(ctx, project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockProjectRepository) GetByIDs(ctx context.Context, projectIDs []uint, userID string) ([]domain.Project, error) {
	args := m.Called(ctx, projectIDs, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Project), args.Error(1)
}

func (m *MockProjectRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}
