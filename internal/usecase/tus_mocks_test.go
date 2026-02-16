package usecase

import (
	"context"
	"time"

	"invento-service/internal/domain"

	"github.com/stretchr/testify/mock"
)

// MockTusModulUploadRepository is a mock for TusModulUploadRepository
type MockTusModulUploadRepository struct {
	mock.Mock
}

func (m *MockTusModulUploadRepository) Create(ctx context.Context, upload *domain.TusModulUpload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) GetByID(ctx context.Context, id string) (*domain.TusModulUpload, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetByUserID(ctx context.Context, userID string) ([]domain.TusModulUpload, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error {
	args := m.Called(ctx, id, offset, progress)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Complete(ctx context.Context, id, modulID, filePath string) error {
	args := m.Called(ctx, id, modulID, filePath)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusModulUpload, error) {
	args := m.Called(ctx, before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusModulUpload, error) {
	args := m.Called(ctx, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) CountActiveByUserID(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveByUserID(ctx context.Context, userID string) ([]domain.TusModulUpload, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveUploadIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// MockTusUploadRepository is a mock for TusUploadRepository
type MockTusUploadRepository struct {
	mock.Mock
}

func (m *MockTusUploadRepository) Create(ctx context.Context, upload *domain.TusUpload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetByID(ctx context.Context, id string) (*domain.TusUpload, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetByUserID(ctx context.Context, userID string) ([]domain.TusUpload, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetActiveByUserID(ctx context.Context, userID string) ([]domain.TusUpload, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) CountActiveByUserID(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockTusUploadRepository) UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error {
	args := m.Called(ctx, id, offset, progress)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateStatus(ctx context.Context, id, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTusUploadRepository) Complete(ctx context.Context, id string, projectID uint, filePath string) error {
	args := m.Called(ctx, id, projectID, filePath)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusUpload, error) {
	args := m.Called(ctx, before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusUpload, error) {
	args := m.Called(ctx, timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTusUploadRepository) ListActive(ctx context.Context) ([]domain.TusUpload, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetActiveUploadIDs(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}
