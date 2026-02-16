package repo_test

import (
	"context"
	"invento-service/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func (m *MockTusUploadRepository) UpdateOffset(ctx context.Context, id string, offset int64, progress float64) error {
	args := m.Called(ctx, id, offset, progress)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateOffsetOnly(ctx context.Context, id string, offset int64) error {
	args := m.Called(ctx, id, offset)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateUpload(ctx context.Context, upload *domain.TusUpload) error {
	args := m.Called(ctx, upload)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	args := m.Called(ctx, id, status)
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

func (m *MockTusUploadRepository) GetByUserIDAndStatus(ctx context.Context, userID string, status string) ([]domain.TusUpload, error) {
	args := m.Called(ctx, userID, status)
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

// TusUploadRepository Tests

func TestTusUploadRepository_Create_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	upload := &domain.TusUpload{
		ID:         "test-upload-id",
		UserID:     "user-1",
		UploadType: domain.UploadTypeProjectCreate,
		UploadMetadata: domain.TusUploadMetadata{
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    3,
		},
		FileSize:      1024 * 1024,
		CurrentOffset: 0,
		Status:        domain.UploadStatusPending,
		Progress:      0,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	mockRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusUpload")).Return(nil)

	err := mockRepo.Create(context.Background(), upload)

	assert.NoError(t, err)
	assert.NotNil(t, upload)
	assert.Equal(t, "test-upload-id", upload.ID)
	assert.Equal(t, "user-1", upload.UserID)
	assert.Equal(t, domain.UploadTypeProjectCreate, upload.UploadType)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_GetByID_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"
	expectedUpload := &domain.TusUpload{
		ID:         uploadID,
		UserID:     "user-1",
		UploadType: domain.UploadTypeProjectCreate,
		UploadMetadata: domain.TusUploadMetadata{
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    3,
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
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_GetByUploadID_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	userID := "user-1"
	expectedUploads := []domain.TusUpload{
		{
			ID:            "upload-1",
			UserID:        "user-1",
			UploadType:    domain.UploadTypeProjectCreate,
			FileSize:      1024 * 1024,
			CurrentOffset: 512 * 1024,
			Status:        domain.UploadStatusUploading,
			Progress:      50.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:         "upload-2",
			UserID:     "user-2",
			UploadType: domain.UploadTypeProjectCreate,
			UploadMetadata: domain.TusUploadMetadata{
				NamaProject: "Project 1",
				Kategori:    "website",
				Semester:    3,
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 1024 * 1024,
			Status:        domain.UploadStatusCompleted,
			Progress:      100.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:         "upload-2",
			UserID:     userID,
			UploadType: domain.UploadTypeProjectUpdate,
			UploadMetadata: domain.TusUploadMetadata{
				NamaProject: "Project 2",
				Kategori:    "mobile",
				Semester:    4,
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
	assert.Equal(t, 3, len(uploads))
	assert.Equal(t, "upload-1", uploads[0].ID)
	assert.Equal(t, "upload-2", uploads[1].ID)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_Update_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	upload := &domain.TusUpload{
		ID:         "test-upload-id",
		UserID:     "user-1",
		UploadType: domain.UploadTypeProjectCreate,
		UploadMetadata: domain.TusUploadMetadata{
			NamaProject: "Updated Project",
			Kategori:    "website",
			Semester:    3,
		},
		FileSize:      1024 * 1024,
		CurrentOffset: 1024 * 1024,
		Status:        domain.UploadStatusCompleted,
		Progress:      100.0,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	mockRepo.On("UpdateUpload", mock.Anything, mock.AnythingOfType("*domain.TusUpload")).Return(nil)

	err := mockRepo.UpdateUpload(context.Background(), upload)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_UpdateOffset_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"
	newOffset := int64(768 * 1024)
	progress := 75.0

	mockRepo.On("UpdateOffset", mock.Anything, uploadID, newOffset, progress).Return(nil)

	err := mockRepo.UpdateOffset(context.Background(), uploadID, newOffset, progress)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"
	newStatus := domain.UploadStatusCompleted

	mockRepo.On("UpdateStatus", mock.Anything, uploadID, newStatus).Return(nil)

	err := mockRepo.UpdateStatus(context.Background(), uploadID, newStatus)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_Delete_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"

	mockRepo.On("Delete", mock.Anything, uploadID).Return(nil)

	err := mockRepo.Delete(context.Background(), uploadID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_ListActiveUploads_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	expectedUploads := []domain.TusUpload{
		{
			ID:            "upload-1",
			UserID:        "user-1",
			UploadType:    domain.UploadTypeProjectCreate,
			FileSize:      1024 * 1024,
			CurrentOffset: 512 * 1024,
			Status:        domain.UploadStatusUploading,
			Progress:      50.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:            "upload-2",
			UserID:        "user-2",
			UploadType:    domain.UploadTypeProjectCreate,
			FileSize:      2048 * 1024,
			CurrentOffset: 0,
			Status:        domain.UploadStatusQueued,
			Progress:      0.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.On("ListActive", mock.Anything).Return(expectedUploads, nil)

	uploads, err := mockRepo.ListActive(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 2, len(uploads))
	assert.Equal(t, domain.UploadStatusUploading, uploads[0].Status)
	assert.Equal(t, domain.UploadStatusQueued, uploads[1].Status)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_GetExpiredUploads_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	before := time.Now()
	expectedUploads := []domain.TusUpload{
		{
			ID:            "expired-upload-1",
			UserID:        "user-1",
			UploadType:    domain.UploadTypeProjectCreate,
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
	assert.Equal(t, "expired-upload-1", uploads[0].ID)
	mockRepo.AssertExpectations(t)
}
