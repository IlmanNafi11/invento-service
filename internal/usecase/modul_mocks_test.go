package usecase

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"github.com/stretchr/testify/mock"
)

// MockModulRepository is a mock for ModulRepository
type MockModulRepository struct {
	mock.Mock
}

func (m *MockModulRepository) Create(ctx context.Context, modul *domain.Modul) error {
	args := m.Called(ctx, modul)
	return args.Error(0)
}

func (m *MockModulRepository) GetByID(ctx context.Context, id string) (*domain.Modul, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByIDs(ctx context.Context, ids []string, userID string) ([]domain.Modul, error) {
	args := m.Called(ctx, ids, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByUserID(ctx context.Context, userID string, search string, filterType string, filterStatus string, page, limit int) ([]dto.ModulListItem, int, error) {
	args := m.Called(ctx, userID, search, filterType, filterStatus, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]dto.ModulListItem), args.Int(1), args.Error(2)
}

func (m *MockModulRepository) CountByUserID(ctx context.Context, userID string) (int, error) {
	args := m.Called(ctx, userID)
	return args.Int(0), args.Error(1)
}

func (m *MockModulRepository) Update(ctx context.Context, modul *domain.Modul) error {
	args := m.Called(ctx, modul)
	return args.Error(0)
}

func (m *MockModulRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockModulRepository) UpdateMetadata(ctx context.Context, modul *domain.Modul) error {
	args := m.Called(ctx, modul)
	return args.Error(0)
}
