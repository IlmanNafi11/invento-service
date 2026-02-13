package repo_test

import (
	"fiber-boiler-plate/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTusUploadRepository struct {
	mock.Mock
}

func (m *MockTusUploadRepository) Create(upload *domain.TusUpload) error {
	args := m.Called(upload)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetByID(id string) (*domain.TusUpload, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetByUserID(userID string) ([]domain.TusUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) UpdateOffset(id string, offset int64, progress float64) error {
	args := m.Called(id, offset, progress)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateOffsetOnly(id string, offset int64) error {
	args := m.Called(id, offset)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateUpload(upload *domain.TusUpload) error {
	args := m.Called(upload)
	return args.Error(0)
}

func (m *MockTusUploadRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockTusUploadRepository) GetExpired(before time.Time) ([]domain.TusUpload, error) {
	args := m.Called(before)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) GetByUserIDAndStatus(userID string, status string) ([]domain.TusUpload, error) {
	args := m.Called(userID, status)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *MockTusUploadRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTusUploadRepository) ListActive() ([]domain.TusUpload, error) {
	args := m.Called()
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
		UploadMetadata: domain.TusUploadInitRequest{
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

	mockRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil)

	err := mockRepo.Create(upload)

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
		UploadMetadata: domain.TusUploadInitRequest{
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

	mockRepo.On("GetByID", uploadID).Return(expectedUpload, nil)

	upload, err := mockRepo.GetByID(uploadID)

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
			UploadMetadata: domain.TusUploadInitRequest{
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
			UploadMetadata: domain.TusUploadInitRequest{
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

	mockRepo.On("GetByUserID", userID).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetByUserID(userID)

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
		UploadMetadata: domain.TusUploadInitRequest{
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

	mockRepo.On("UpdateUpload", mock.AnythingOfType("*domain.TusUpload")).Return(nil)

	err := mockRepo.UpdateUpload(upload)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_UpdateOffset_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"
	newOffset := int64(768 * 1024)
	progress := 75.0

	mockRepo.On("UpdateOffset", uploadID, newOffset, progress).Return(nil)

	err := mockRepo.UpdateOffset(uploadID, newOffset, progress)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"
	newStatus := domain.UploadStatusCompleted

	mockRepo.On("UpdateStatus", uploadID, newStatus).Return(nil)

	err := mockRepo.UpdateStatus(uploadID, newStatus)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_Delete_Success(t *testing.T) {
	mockRepo := new(MockTusUploadRepository)

	uploadID := "test-upload-id"

	mockRepo.On("Delete", uploadID).Return(nil)

	err := mockRepo.Delete(uploadID)

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

	mockRepo.On("ListActive").Return(expectedUploads, nil)

	uploads, err := mockRepo.ListActive()

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 2, len(uploads))
	assert.Equal(t, domain.UploadStatusUploading, uploads[0].Status)
	assert.Equal(t, domain.UploadStatusQueued, uploads[1].Status)
	mockRepo.AssertExpectations(t)
}

func TestTusUploadRepository_GetExpired_Success(t *testing.T) {
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

	mockRepo.On("GetExpired", mock.AnythingOfType("time.Time")).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetExpired(before)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 1, len(uploads))
	assert.Equal(t, "expired-upload-1", uploads[0].ID)
	mockRepo.AssertExpectations(t)
}

type MockTusModulUploadRepository struct {
	mock.Mock
}

func (m *MockTusModulUploadRepository) Create(upload *domain.TusModulUpload) error {
	args := m.Called(upload)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) GetByID(id string) (*domain.TusModulUpload, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetByUserID(userID string) ([]domain.TusModulUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) UpdateOffset(id string, offset int64, progress float64) error {
	args := m.Called(id, offset, progress)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Complete(id string, modulID string, filePath string) error {
	args := m.Called(id, modulID, filePath)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTusModulUploadRepository) GetExpiredUploads() ([]domain.TusModulUpload, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error) {
	args := m.Called(timeout)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) CountActiveByUserID(userID string) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveByUserID(userID string) ([]domain.TusModulUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveUploadIDs() ([]string, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

// TusModulUploadRepository Tests

func TestTusModulUploadRepository_Create_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	upload := &domain.TusModulUpload{
		ID:         "test-modul-upload-id",
		UserID:     "user-1",
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			Judul:     "Test Modul",
			Deskripsi: "Test Deskripsi",
		},
		FileSize:      1024 * 1024,
		CurrentOffset: 0,
		Status:        domain.ModulUploadStatusPending,
		Progress:      0,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
	}

	mockRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil)

	err := mockRepo.Create(upload)

	assert.NoError(t, err)
	assert.NotNil(t, upload)
	assert.Equal(t, "test-modul-upload-id", upload.ID)
	assert.Equal(t, "user-1", upload.UserID)
	assert.Equal(t, domain.ModulUploadTypeCreate, upload.UploadType)
	assert.Equal(t, "Test Modul", upload.UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetByID_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"
	expectedUpload := &domain.TusModulUpload{
		ID:         uploadID,
		UserID:     "user-1",
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			Judul:     "Test Modul",
			Deskripsi: "Test Deskripsi",
		},
		FileSize:      1024 * 1024,
		CurrentOffset: 512 * 1024,
		Status:        domain.ModulUploadStatusUploading,
		Progress:      50.0,
		ExpiresAt:     time.Now().Add(1 * time.Hour),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	mockRepo.On("GetByID", uploadID).Return(expectedUpload, nil)

	upload, err := mockRepo.GetByID(uploadID)

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
	mockRepo := new(MockTusModulUploadRepository)

	userID := "user-1"
	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "modul-upload-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				Judul:     "Modul 1",
				Deskripsi: "Deskripsi 1",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 1024 * 1024,
			Status:        domain.ModulUploadStatusCompleted,
			Progress:      100.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:         "modul-upload-2",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeUpdate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				Judul:     "Modul 2",
				Deskripsi: "Deskripsi 2",
			},
			FileSize:      2048 * 1024,
			CurrentOffset: 1024 * 1024,
			Status:        domain.ModulUploadStatusUploading,
			Progress:      50.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.On("GetByUserID", userID).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetByUserID(userID)

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
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"
	newOffset := int64(768 * 1024)
	progress := 75.0

	mockRepo.On("UpdateOffset", uploadID, newOffset, progress).Return(nil)

	err := mockRepo.UpdateOffset(uploadID, newOffset, progress)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_UpdateStatus_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"
	newStatus := domain.ModulUploadStatusCompleted

	mockRepo.On("UpdateStatus", uploadID, newStatus).Return(nil)

	err := mockRepo.UpdateStatus(uploadID, newStatus)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_Delete_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-modul-upload-id"

	mockRepo.On("Delete", uploadID).Return(nil)

	err := mockRepo.Delete(uploadID)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_ListActiveUploads_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	userID := "user-1"
	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "active-upload-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				Judul:     "Active Modul 1",
				Deskripsi: "Deskripsi 1",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 512 * 1024,
			Status:        domain.ModulUploadStatusUploading,
			Progress:      50.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
		{
			ID:         "active-upload-2",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				Judul:     "Active Modul 2",
				Deskripsi: "Deskripsi 2",
			},
			FileSize:      2048 * 1024,
			CurrentOffset: 0,
			Status:        domain.ModulUploadStatusQueued,
			Progress:      0.0,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		},
	}

	mockRepo.On("GetActiveByUserID", userID).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetActiveByUserID(userID)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 2, len(uploads))
	assert.Equal(t, domain.ModulUploadStatusUploading, uploads[0].Status)
	assert.Equal(t, domain.ModulUploadStatusQueued, uploads[1].Status)
	assert.Equal(t, "Active Modul 1", uploads[0].UploadMetadata.Judul)
	assert.Equal(t, "Active Modul 2", uploads[1].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetExpiredUploads_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "expired-modul-upload-1",
			UserID:     "user-1",
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				Judul:     "Expired Modul",
				Deskripsi: "Expired Deskripsi",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 256 * 1024,
			Status:        domain.ModulUploadStatusPending,
			Progress:      25.0,
			ExpiresAt:     time.Now().Add(-1 * time.Hour),
			CreatedAt:     time.Now().Add(-2 * time.Hour),
			UpdatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	mockRepo.On("GetExpiredUploads").Return(expectedUploads, nil)

	uploads, err := mockRepo.GetExpiredUploads()

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 1, len(uploads))
	assert.Equal(t, "expired-modul-upload-1", uploads[0].ID)
	assert.Equal(t, "Expired Modul", uploads[0].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_CountActiveByUserID_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	userID := "user-1"
	expectedCount := 3

	mockRepo.On("CountActiveByUserID", userID).Return(expectedCount, nil)

	count, err := mockRepo.CountActiveByUserID(userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedCount, count)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_Complete_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	uploadID := "test-upload-id"
	modulID := "550e8400-e29b-41d4-a716-446655440001"
	filePath := "/uploads/moduls/test_modul.pdf"

	mockRepo.On("Complete", uploadID, modulID, filePath).Return(nil)

	err := mockRepo.Complete(uploadID, modulID, filePath)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestTusModulUploadRepository_GetAbandonedUploads_Success(t *testing.T) {
	mockRepo := new(MockTusModulUploadRepository)

	timeout := 30 * time.Minute
	expectedUploads := []domain.TusModulUpload{
		{
			ID:         "abandoned-upload-1",
			UserID:     "user-1",
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				Judul:     "Abandoned Modul",
				Deskripsi: "Abandoned Deskripsi",
			},
			FileSize:      1024 * 1024,
			CurrentOffset: 128 * 1024,
			Status:        domain.ModulUploadStatusUploading,
			Progress:      12.5,
			UpdatedAt:     time.Now().Add(-1 * time.Hour),
			CreatedAt:     time.Now().Add(-2 * time.Hour),
		},
	}

	mockRepo.On("GetAbandonedUploads", timeout).Return(expectedUploads, nil)

	uploads, err := mockRepo.GetAbandonedUploads(timeout)

	assert.NoError(t, err)
	assert.NotNil(t, uploads)
	assert.Equal(t, 1, len(uploads))
	assert.Equal(t, "abandoned-upload-1", uploads[0].ID)
	assert.Equal(t, "Abandoned Modul", uploads[0].UploadMetadata.Judul)
	mockRepo.AssertExpectations(t)
}
