package usecase

import (
	"bytes"
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestTusUploadUsecase_InitiateUpload_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)
	userEmail := "test@example.com"
	userRole := "mahasiswa"
	fileSize := int64(1024 * 1024) // 1 MB
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	mockTusUploadRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil)

	result, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.UploadID)
	assert.NotEmpty(t, result.UploadURL)
	assert.Equal(t, int64(0), result.Offset)
	assert.Equal(t, fileSize, result.Length)

	// Cleanup
	tusManager.CancelUpload(result.UploadID)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestInitiateUpload_QueueFull tests when upload queue is full
func TestTusUploadUsecase_InitiateUpload_QueueFull(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)
	userEmail := "test@example.com"
	userRole := "mahasiswa"
	fileSize := int64(1024 * 1024) // 1 MB
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	// Setup mock expectations - First upload succeeds
	mockTusUploadRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil).Once()

	// Fill the queue by starting an upload
	firstUpload, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)
	assert.NoError(t, err)
	assert.NotNil(t, firstUpload)

	// Try to create another upload - should fail (queue full check happens before Create)
	result, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "slot upload tidak tersedia")

	// Cleanup
	tusManager.CancelUpload(firstUpload.UploadID)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestInitiateUpload_FileTooLarge tests file size validation
func TestTusUploadUsecase_InitiateUpload_FileTooLarge(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)
	userEmail := "test@example.com"
	userRole := "mahasiswa"
	fileSize := int64(600 * 1024 * 1024) // 600 MB - exceeds 500 MB limit
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	result, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
}

// TestUploadChunk_Success tests successful chunk processing
func TestTusUploadUsecase_UploadChunk_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunkData := []byte("test chunk data")
	chunk := bytes.NewReader(chunkData)

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        userID,
		FileSize:      int64(1024),
		CurrentOffset: 0,
		Status:        domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusUploadRepo.On("UpdateOffset", uploadID, int64(len(chunkData)), mock.AnythingOfType("float64")).Return(nil)

	// First, initiate the upload in the TusStore
	tusManager.InitiateUpload(uploadID, int64(1024), map[string]string{"user_id": "1"})

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunkData)), newOffset)

	mockTusUploadRepo.AssertExpectations(t)

	// Cleanup
	tusManager.CancelUpload(uploadID)
}

// TestUploadChunk_InvalidOffset tests offset validation
func TestTusUploadUsecase_UploadChunk_InvalidOffset(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(100) // Wrong offset
	chunk := bytes.NewReader([]byte("test data"))

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        userID,
		FileSize:      int64(1024),
		CurrentOffset: 0,
		Status:        domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(100), newOffset) // TusManager returns the provided offset on error

	mockTusUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_UploadNotFound tests when upload is not found
func TestTusUploadUsecase_UploadChunk_UploadNotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "non-existent-upload"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_UnauthorizedUser tests when user doesn't own the upload
func TestTusUploadUsecase_UploadChunk_UnauthorizedUser(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2) // Different user
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        ownerID,
		FileSize:      int64(1024),
		CurrentOffset: 0,
		Status:        domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadStatus_Success tests successful status retrieval
func TestTusUploadUsecase_GetUploadStatus_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	expectedOffset := int64(512)
	expectedFileSize := int64(1024)

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        userID,
		FileSize:      expectedFileSize,
		CurrentOffset: expectedOffset,
		Status:        domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	offset, fileSize, err := uc.GetUploadStatus(uploadID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedOffset, offset)
	assert.Equal(t, expectedFileSize, fileSize)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadStatus_Unauthorized tests unauthorized access
func TestTusUploadUsecase_GetUploadStatus_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	tusUpload := &domain.TusUpload{
		ID:     uploadID,
		UserID: ownerID,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	offset, fileSize, err := uc.GetUploadStatus(uploadID, userID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, int64(0), fileSize)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadInfo_Success tests successful upload info retrieval
func TestTusUploadUsecase_GetUploadInfo_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	tusUpload := &domain.TusUpload{
		ID:             uploadID,
		UserID:         userID,
		UploadMetadata: metadata,
		FileSize:       int64(1024),
		CurrentOffset:  int64(512),
		Status:         domain.UploadStatusUploading,
		Progress:       50.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	info, err := uc.GetUploadInfo(uploadID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, uploadID, info.UploadID)
	assert.Equal(t, "Test Project", info.NamaProject)
	assert.Equal(t, "website", info.Kategori)
	assert.Equal(t, 1, info.Semester)
	assert.Equal(t, domain.UploadStatusUploading, info.Status)
	assert.Equal(t, int64(512), info.Offset)
	assert.Equal(t, int64(1024), info.Length)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_Success tests successful upload cancellation
func TestTusUploadUsecase_CancelUpload_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	tusUpload := &domain.TusUpload{
		ID:     uploadID,
		UserID: userID,
		Status: domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusUploadRepo.On("UpdateStatus", uploadID, domain.UploadStatusCancelled).Return(nil)

	err := uc.CancelUpload(uploadID, userID)

	assert.NoError(t, err)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_AlreadyCompleted tests cancelling a completed upload
func TestTusUploadUsecase_CancelUpload_AlreadyCompleted(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	tusUpload := &domain.TusUpload{
		ID:     uploadID,
		UserID: userID,
		Status: domain.UploadStatusCompleted,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	err := uc.CancelUpload(uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload sudah selesai")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCheckUploadSlot_Success tests checking upload slot availability
func TestTusUploadUsecase_CheckUploadSlot_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)

	response, err := uc.CheckUploadSlot(userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Available)
	assert.Equal(t, "Slot upload tersedia", response.Message)
	assert.Equal(t, 0, response.QueueLength)
	assert.False(t, response.ActiveUpload)
	assert.Equal(t, 1, response.MaxConcurrent)
}

// TestResetUploadQueue_Success tests resetting the upload queue
func TestTusUploadUsecase_ResetUploadQueue_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)

	err := uc.ResetUploadQueue(userID)

	assert.NoError(t, err)
}

// TestInitiateProjectUpdateUpload_Success tests successful project update upload initiation
func TestTusUploadUsecase_InitiateProjectUpdateUpload_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	userID := uint(1)
	fileSize := int64(1024 * 1024) // 1 MB
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	project := &domain.Project{
		ID:       projectID,
		UserID:   userID,
		NamaProject: "Old Project",
		Kategori: "website",
		Semester: 1,
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	mockTusUploadRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil)

	result, err := uc.InitiateProjectUpdateUpload(projectID, userID, fileSize, metadata)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.UploadID)

	// Cleanup
	tusManager.CancelUpload(result.UploadID)

	mockProjectRepo.AssertExpectations(t)
	mockTusUploadRepo.AssertExpectations(t)
}

// TestInitiateProjectUpdateUpload_ProjectNotFound tests when project is not found
func TestTusUploadUsecase_InitiateProjectUpdateUpload_ProjectNotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(999)
	userID := uint(1)
	fileSize := int64(1024 * 1024)
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	result, err := uc.InitiateProjectUpdateUpload(projectID, userID, fileSize, metadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "project tidak ditemukan")

	mockProjectRepo.AssertExpectations(t)
}

// TestHandleProjectUpdateChunk_Success tests successful project update chunk handling
func TestTusUploadUsecase_HandleProjectUpdateChunk_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunkData := []byte("test chunk data")
	chunk := bytes.NewReader(chunkData)

	projectIDVal := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &projectIDVal,
		FileSize:  int64(1024),
		Status:    domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusUploadRepo.On("UpdateOffset", uploadID, int64(len(chunkData)), mock.AnythingOfType("float64")).Return(nil)

	// Initiate the upload in TusStore first
	tusManager.InitiateUpload(uploadID, int64(1024), map[string]string{"user_id": "1", "project_id": "1"})

	newOffset, err := uc.HandleProjectUpdateChunk(projectID, uploadID, userID, offset, chunk)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunkData)), newOffset)

	mockTusUploadRepo.AssertExpectations(t)

	// Cleanup
	tusManager.CancelUpload(uploadID)
}

// TestHandleProjectUpdateChunk_ProjectMismatch tests when project ID doesn't match
func TestTusUploadUsecase_HandleProjectUpdateChunk_ProjectMismatch(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(2)
	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &uploadProjectID,
		FileSize:  int64(1024),
		Status:    domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleProjectUpdateChunk(projectID, uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "project ID tidak cocok")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadStatus_Success tests successful project update upload status
func TestTusUploadUsecase_GetProjectUpdateUploadStatus_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	userID := uint(1)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &uploadProjectID,
		FileSize:  int64(1024),
		CurrentOffset: int64(512),
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	offset, fileSize, err := uc.GetProjectUpdateUploadStatus(projectID, uploadID, userID)

	assert.NoError(t, err)
	assert.Equal(t, int64(512), offset)
	assert.Equal(t, int64(1024), fileSize)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_Success tests successful project update upload cancellation
func TestTusUploadUsecase_CancelProjectUpdateUpload_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	userID := uint(1)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &uploadProjectID,
		Status:    domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusUploadRepo.On("UpdateStatus", uploadID, domain.UploadStatusCancelled).Return(nil)

	err := uc.CancelProjectUpdateUpload(projectID, uploadID, userID)

	assert.NoError(t, err)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_CompletedUpload tests uploading to a completed upload
func TestTusUploadUsecase_UploadChunk_CompletedUpload(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(1024)
	chunk := bytes.NewReader([]byte("more data"))

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        userID,
		FileSize:      int64(1024),
		CurrentOffset: int64(1024),
		Status:        domain.UploadStatusCompleted,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.NoError(t, err)
	assert.Equal(t, int64(1024), newOffset) // Returns file size when completed

	mockTusUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_CancelledUpload tests uploading to a cancelled upload
func TestTusUploadUsecase_UploadChunk_CancelledUpload(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)
	// ProjectUsecase is not used in these tests

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        userID,
		FileSize:      int64(1024),
		CurrentOffset: int64(0),
		Status:        domain.UploadStatusCancelled,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "upload sudah dibatalkan atau gagal")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCheckUploadSlot_QueueFull tests checking upload slot when queue is full
func TestTusUploadUsecase_CheckUploadSlot_QueueFull(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)
	userEmail := "test@example.com"
	userRole := "mahasiswa"
	fileSize := int64(1024 * 1024)
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	// Setup mock expectation for Create
	mockTusUploadRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil).Once()

	// Fill the queue by starting an upload (this becomes the active upload)
	_, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)
	assert.NoError(t, err)

	// Now check the slot - should be full because active upload exists
	response, err := uc.CheckUploadSlot(userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.False(t, response.Available)
	assert.Contains(t, response.Message, "tidak tersedia")
	assert.Equal(t, 0, response.QueueLength) // Queue length is 0 because first upload is active, not queued
	assert.True(t, response.ActiveUpload)

	// Cleanup
	tusManager.ResetUploadQueue()

	mockTusUploadRepo.AssertExpectations(t)
}

// TestInitiateUpload_NegativeFileSize tests negative file size validation
func TestTusUploadUsecase_InitiateUpload_NegativeFileSize(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)
	userEmail := "test@example.com"
	userRole := "mahasiswa"
	fileSize := int64(-100)
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	result, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ukuran file tidak valid")
}

// TestInitiateUpload_RepositoryError tests repository error handling
func TestTusUploadUsecase_InitiateUpload_RepositoryError(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	userID := uint(1)
	userEmail := "test@example.com"
	userRole := "mahasiswa"
	fileSize := int64(1024 * 1024)
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
	}

	mockTusUploadRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(errors.New("database error"))

	result, err := uc.InitiateUpload(userID, userEmail, userRole, fileSize, metadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "gagal membuat upload record")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_FailedUpload tests uploading to a failed upload
func TestTusUploadUsecase_UploadChunk_FailedUpload(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	tusUpload := &domain.TusUpload{
		ID:            uploadID,
		UserID:        userID,
		FileSize:      int64(1024),
		CurrentOffset: int64(0),
		Status:        domain.UploadStatusFailed,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "upload sudah dibatalkan atau gagal")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadInfo_NotFound tests getting info for non-existent upload
func TestTusUploadUsecase_GetUploadInfo_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "non-existent-upload"
	userID := uint(1)

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	info, err := uc.GetUploadInfo(uploadID, userID)

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadInfo_Unauthorized tests unauthorized access to upload info
func TestTusUploadUsecase_GetUploadInfo_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	tusUpload := &domain.TusUpload{
		ID:     uploadID,
		UserID: ownerID,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	info, err := uc.GetUploadInfo(uploadID, userID)

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_NotFound tests cancelling non-existent upload
func TestTusUploadUsecase_CancelUpload_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "non-existent-upload"
	userID := uint(1)

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.CancelUpload(uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_Unauthorized tests unauthorized cancellation
func TestTusUploadUsecase_CancelUpload_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	tusUpload := &domain.TusUpload{
		ID:     uploadID,
		UserID: ownerID,
		Status: domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	err := uc.CancelUpload(uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_AlreadyCancelled tests cancelling already cancelled upload
func TestTusUploadUsecase_CancelUpload_AlreadyCancelled(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	tusUpload := &domain.TusUpload{
		ID:     uploadID,
		UserID: userID,
		Status: domain.UploadStatusCancelled,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusUploadRepo.On("UpdateStatus", uploadID, domain.UploadStatusCancelled).Return(nil)

	err := uc.CancelUpload(uploadID, userID)

	// Should succeed (idempotent operation)
	assert.NoError(t, err)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadStatus_NotFound tests status for non-existent upload
func TestTusUploadUsecase_GetUploadStatus_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	uploadID := "non-existent-upload"
	userID := uint(1)

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	offset, fileSize, err := uc.GetUploadStatus(uploadID, userID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, int64(0), fileSize)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetUploadStatus_ProgressCalculation tests progress calculation for various statuses
func TestTusUploadUsecase_GetUploadStatus_ProgressCalculation(t *testing.T) {
	tests := []struct {
		name           string
		status         string
		offset         int64
		fileSize       int64
		expectedOffset int64
		expectedSize   int64
	}{
		{
			name:           "Pending status",
			status:         domain.UploadStatusPending,
			offset:         0,
			fileSize:       1024,
			expectedOffset: 0,
			expectedSize:   1024,
		},
		{
			name:           "Half uploaded",
			status:         domain.UploadStatusUploading,
			offset:         512,
			fileSize:       1024,
			expectedOffset: 512,
			expectedSize:   1024,
		},
		{
			name:           "Queued status",
			status:         domain.UploadStatusQueued,
			offset:         0,
			fileSize:       2048,
			expectedOffset: 0,
			expectedSize:   2048,
		},
		{
			name:           "Expired status",
			status:         domain.UploadStatusExpired,
			offset:         256,
			fileSize:       1024,
			expectedOffset: 256,
			expectedSize:   1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTusUploadRepo := new(MockTusUploadRepository)
			mockProjectRepo := new(MockProjectRepository)

			cfg := getTestTusUploadConfig()

			pathResolver := helper.NewPathResolver(cfg)
			tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
			tusQueue := helper.NewTusQueue(1)
			tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
			fileManager := helper.NewFileManager(cfg)

			uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

			uploadID := "test-upload-id"
			userID := uint(1)

			tusUpload := &domain.TusUpload{
				ID:            uploadID,
				UserID:        userID,
				FileSize:      tt.fileSize,
				CurrentOffset: tt.offset,
				Status:        tt.status,
			}

			mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

			offset, fileSize, err := uc.GetUploadStatus(uploadID, userID)

			assert.NoError(t, err)
			assert.Equal(t, tt.expectedOffset, offset)
			assert.Equal(t, tt.expectedSize, fileSize)

			mockTusUploadRepo.AssertExpectations(t)
		})
	}
}

// TestInitiateProjectUpdateUpload_UnauthorizedProject tests unauthorized project update
func TestTusUploadUsecase_InitiateProjectUpdateUpload_UnauthorizedProject(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	ownerID := uint(2)
	userID := uint(1)
	fileSize := int64(1024 * 1024)
	metadata := domain.TusUploadInitRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	project := &domain.Project{
		ID:       projectID,
		UserID:   ownerID,
		NamaProject: "Old Project",
		Kategori: "website",
		Semester: 1,
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)

	result, err := uc.InitiateProjectUpdateUpload(projectID, userID, fileSize, metadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tidak memiliki akses ke project ini")

	mockProjectRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_Success tests successful project update upload info retrieval
func TestTusUploadUsecase_GetProjectUpdateUploadInfo_Success(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	userID := uint(1)

	metadata := domain.TusUploadInitRequest{
		NamaProject: "Updated Project",
		Kategori:    "mobile",
		Semester:    2,
	}

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &uploadProjectID,
		UploadMetadata: metadata,
		FileSize:       int64(1024),
		CurrentOffset:  int64(512),
		Status:         domain.UploadStatusUploading,
		Progress:       50.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	info, err := uc.GetProjectUpdateUploadInfo(projectID, uploadID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, uploadID, info.UploadID)
	assert.Equal(t, projectID, info.ProjectID)
	assert.Equal(t, "Updated Project", info.NamaProject)
	assert.Equal(t, "mobile", info.Kategori)
	assert.Equal(t, 2, info.Semester)
	assert.Equal(t, domain.UploadStatusUploading, info.Status)
	assert.Equal(t, int64(512), info.Offset)
	assert.Equal(t, int64(1024), info.Length)

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_NotFound tests getting info for non-existent project update upload
func TestTusUploadUsecase_GetProjectUpdateUploadInfo_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "non-existent-upload"
	userID := uint(1)

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	info, err := uc.GetProjectUpdateUploadInfo(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_Unauthorized tests unauthorized access to project update upload info
func TestTusUploadUsecase_GetProjectUpdateUploadInfo_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    ownerID,
		ProjectID: &uploadProjectID,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	info, err := uc.GetProjectUpdateUploadInfo(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadInfo_ProjectMismatch tests project ID mismatch
func TestTusUploadUsecase_GetProjectUpdateUploadInfo_ProjectMismatch(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(2)
	uploadID := "test-upload-id"
	userID := uint(1)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &uploadProjectID,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	info, err := uc.GetProjectUpdateUploadInfo(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Nil(t, info)
	assert.Contains(t, err.Error(), "project ID tidak cocok")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_NotFound tests cancelling non-existent project update upload
func TestTusUploadUsecase_CancelProjectUpdateUpload_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "non-existent-upload"
	userID := uint(1)

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	err := uc.CancelProjectUpdateUpload(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_Unauthorized tests unauthorized cancellation of project update upload
func TestTusUploadUsecase_CancelProjectUpdateUpload_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    ownerID,
		ProjectID: &uploadProjectID,
		Status:    domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	err := uc.CancelProjectUpdateUpload(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestCancelProjectUpdateUpload_ProjectMismatch tests project ID mismatch when cancelling
func TestTusUploadUsecase_CancelProjectUpdateUpload_ProjectMismatch(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(2)
	uploadID := "test-upload-id"
	userID := uint(1)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    userID,
		ProjectID: &uploadProjectID,
		Status:    domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	err := uc.CancelProjectUpdateUpload(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "project ID tidak cocok")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestHandleProjectUpdateChunk_NotFound tests handling chunk for non-existent project update upload
func TestTusUploadUsecase_HandleProjectUpdateChunk_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "non-existent-upload"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	newOffset, err := uc.HandleProjectUpdateChunk(projectID, uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestHandleProjectUpdateChunk_Unauthorized tests unauthorized chunk handling
func TestTusUploadUsecase_HandleProjectUpdateChunk_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    ownerID,
		ProjectID: &uploadProjectID,
		FileSize:  int64(1024),
		Status:    domain.UploadStatusUploading,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleProjectUpdateChunk(projectID, uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadStatus_NotFound tests status for non-existent project update upload
func TestTusUploadUsecase_GetProjectUpdateUploadStatus_NotFound(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "non-existent-upload"
	userID := uint(1)

	mockTusUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	offset, fileSize, err := uc.GetProjectUpdateUploadStatus(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, int64(0), fileSize)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusUploadRepo.AssertExpectations(t)
}

// TestGetProjectUpdateUploadStatus_Unauthorized tests unauthorized status check
func TestTusUploadUsecase_GetProjectUpdateUploadStatus_Unauthorized(t *testing.T) {
	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(1)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg)

	projectID := uint(1)
	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	uploadProjectID := uint(1)
	tusUpload := &domain.TusUpload{
		ID:        uploadID,
		UserID:    ownerID,
		ProjectID: &uploadProjectID,
	}

	mockTusUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	offset, fileSize, err := uc.GetProjectUpdateUploadStatus(projectID, uploadID, userID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, int64(0), fileSize)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusUploadRepo.AssertExpectations(t)
}
