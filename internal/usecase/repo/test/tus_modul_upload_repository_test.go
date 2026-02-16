package repo_test

import (
	"context"
	"invento-service/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockTusModulUploadRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Complete(ctx context.Context, id string, modulID string, filePath string) error {
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

// TusModulUploadRepository Tests

func TestTusModulUploadRepository_Create_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	upload := &domain.TusModulUpload{
		ID:         "test-modul-upload-id",
		UserID:     "user-1",
		UploadType: domain.UploadTypeModulCreate,
		UploadMetadata: domain.TusModulUploadMetadata{
			Judul:     "Test Modul",
			Deskripsi: "Test Deskripsi",
		},
		FileSize:      1024 * 1024,
		CurrentOffset: 0,
		Status:        domain.UploadStatusPending,
		Progress:      0,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusModulUpload")).Return(nil)

	err := mockRepo.Create(context.Background(), upload)

	assert.NoError(t, err)
	assert.NotNil(t, upload)
	assert.Equal(t, "test-modul-upload-id", upload.ID)
	assert.Equal(t, "user-1", upload.UserID)
	assert.Equal(t, domain.UploadTypeModulCreate, upload.UploadType)
	assert.Equal(t, "Test Modul", upload.UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"
	expectedUpload := &domain.TusModulUpload{
		ID:         uploadID,
		UserID:     "user-1",
		UploadType: domain.UploadTypeModulCreate,
		UploadMetadata: domain.TusModulUploadMetadata{
			Judul:     "Test Modul",
			Deskripsi: "Test Deskripsi",
		},
		FileSize:      1024 * 1024,
		CurrentOffset: 512 * 1024,
		Status:        domain.UploadStatusUploading,
		Progress:      50.0,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.On("GetByID", mock.Anything, uploadID).Return(expectedUpload, nil)

	upload, err := mockRepo.GetByID(context.Background(), uploadID)

	assert.NoError(t, err)
	assert.NotNil(t, upload)
	assert.Equal(t, uploadID, upload.ID)
	assert.Equal(t, "user-1", upload.UserID)
	assert.Equal(t, int64(512*1024), upload.CurrentOffset)
	assert.Equal(t, 50.0, upload.Progress)
	assert.Equal(t, "Test Modul", upload.UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetByModulID_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	userID := "user-1"
	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "modul-upload-1",
			UserID:     userID,
			UploadType: domain.UploadTypeModulCreate,
			UploadMetadata: domain.TusModulUploadMetadata{
				Judul:     "Modul 1",
				Deskripsi: "Deskripsi 1",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 1024 * 1024,
			Status:        domain.UploadStatusCompleted,
			Progress:      100.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:         "modul-upload-2",
			UserID:     userID,
			UploadType: domain.UploadTypeModulUpdate,
			UploadMetadata: domain.TusModulUploadMetadata{
				Judul:     "Modul 2",
				Deskripsi: "Deskripsi 2",
			},
			FileSize:      2048 * 1024,
			CurrentOffset: 1024 * 1024,
			Status:        domain.UploadStatusUploading,
			Progress:      50.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.On("GetByUserID", mock.Anything, userID).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 2, len(uploads))
	assert.Equal(t, "modul-upload-1", uploads[0].ID)
	assert.Equal(t, "modul-upload-2", uploads[1].ID)
	assert.Equal(t, "Modul 1", uploads[0].UploadMetadata.Judul)
	assert.Equal(t, "Modul 2", uploads[1].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_Update_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"
	newOffset := int64(768 * 1024)
	progress := 75.0

	mockRepo.On("UpdateOffset", mock.Anything, uploadID, newOffset, progress).Return(nil)

	err := mockRepo.UpdateOffset(context.Background(), uploadID, newOffset, progress)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_UpdateStatus_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"
	newStatus := domain.UploadStatusCompleted

	mockRepo.On("UpdateStatus", mock.Anything, uploadID, newStatus).Return(nil)

	err := mockRepo.UpdateStatus(context.Background(), uploadID, newStatus)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_Delete_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"

	mockRepo.On("Delete", mock.Anything, uploadID).Return(nil)

	err := mockRepo.Delete(context.Background(), uploadID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_ListActiveUploads_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	userID := "user-1"
	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "active-upload-1",
			UserID:     userID,
			UploadType: domain.UploadTypeModulCreate,
			UploadMetadata: domain.TusModulUploadMetadata{
				Judul:     "Active Modul 1",
				Deskripsi: "Deskripsi 1",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 512 * 1024,
			Status:        domain.UploadStatusUploading,
			Progress:      50.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:         "active-upload-2",
			UserID:     userID,
			UploadType: domain.UploadTypeModulCreate,
			UploadMetadata: domain.TusModulUploadMetadata{
				Judul:     "Active Modul 2",
				Deskripsi: "Deskripsi 2",
			},
			FileSize:      2048 * 1024,
			CurrentOffset: 0,
			Status:        domain.UploadStatusQueued,
			Progress:      0.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.On("GetActiveByUserID", mock.Anything, userID).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetActiveByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 2, len(uploads))
	assert.Equal(t, domain.UploadStatusUploading, uploads[0].Status)
	assert.Equal(t, domain.UploadStatusQueued, uploads[1].Status)
	assert.Equal(t, "Active Modul 1", uploads[0].UploadMetadata.Judul)
	assert.Equal(t, "Active Modul 2", uploads[1].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetExpiredUploads_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)
	before := time.Now()

	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "expired-modul-upload-1",
			UserID:     "user-1",
			UploadType: domain.UploadTypeModulCreate,
			UploadMetadata: domain.TusModulUploadMetadata{
				Judul:     "Expired Modul",
				Deskripsi: "Expired Deskripsi",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 256 * 1024,
			Status:        domain.UploadStatusPending,
			Progress:      25.0,
			ExpiresAt:     time.Now().Add(-1 * time.Hour),
			CreatedAt:     time.Now().Add(-2 * time.Hour),
			UpdatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	mockRepo.On("GetExpiredUploads", mock.Anything, mock.AnythingOfType("time.Time")).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetExpiredUploads(context.Background(), before)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 1, len(uploads))
	assert.Equal(t, "expired-modul-upload-1", uploads[0].ID)
	assert.Equal(t, "Expired Modul", uploads[0].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_CountActiveByUserID_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	userID := "user-1"
	expectedCount := int64(3)

	mockRepo.On("CountActiveByUserID", mock.Anything, userID).Return(expectedCount, nil)

	count, err := mockRepo.CountActiveByUserID(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_Complete_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-upload-id"
	modulID := "550e8400-e29b-41d4-a716-446655440001"
	filePath := "/uploads/moduls/test_modul.pdf"

	mockRepo.On("Complete", mock.Anything, uploadID, modulID, filePath).Return(nil)

	err := mockRepo.Complete(context.Background(), uploadID, modulID, filePath)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetAbandonedUploads_Success(t *testing.T) {
	t.Parallel()
	mockRepo := new(MockTusModulUploadRepository)

	timeout := 30 * time.Minute
	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "abandoned-upload-1",
			UserID:     "user-1",
			UploadType: domain.UploadTypeModulCreate,
			UploadMetadata: domain.TusModulUploadMetadata{
				Judul:     "Abandoned Modul",
				Deskripsi: "Abandoned Deskripsi",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 128 * 1024,
			Status:        domain.UploadStatusUploading,
			Progress:      12.5,
			UpdatedAt:     time.Now().Add(-1 * time.Hour),
			CreatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	mockRepo.On("GetAbandonedUploads", mock.Anything, timeout).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetAbandonedUploads(context.Background(), timeout)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 1, len(uploads))
	assert.Equal(t, "abandoned-upload-1", uploads[0].ID)
	assert.Equal(t, "Abandoned Modul", uploads[0].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}
