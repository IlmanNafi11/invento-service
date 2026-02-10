package usecase_test

import (
	"bytes"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

// Helper function to create test config for modul
func getTestModulConfig() *config.Config {
	return &config.Config{
		Upload: config.UploadConfig{
			MaxSize:             524288000, // 500 MB (required by TusManager)
			MaxSizeModul:        52428800,  // 50 MB
			MaxQueueModulPerUser: 5,
			IdleTimeout:         600, // 10 minutes
		},
	}
}

// TestInitiateUpload_Success tests successful modul upload creation
func TestTusModulUsecase_InitiateUpload_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)
	fileSize := int64(1024 * 1024) // 1 MB
	// Base64 encoded metadata: "nama_file dGVzdGZpbGU=,tipe cGRm,semester MQ=="
	uploadMetadata := "nama_file dGVzdGZpbGU=,tipe cGRm,semester MQ=="

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(0, nil)
	mockTusModulUploadRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil)

	result, err := uc.InitiateModulUpload(userID, fileSize, uploadMetadata)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.UploadID)
	assert.NotEmpty(t, result.UploadURL)
	assert.Equal(t, int64(0), result.Offset)
	assert.Equal(t, fileSize, result.Length)

	// Cleanup
	tusManager.CancelUpload(result.UploadID)

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestInitiateUpload_QueueFull tests when modul upload queue is full
func TestTusModulUsecase_InitiateUpload_QueueFull(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)
	fileSize := int64(1024 * 1024) // 1 MB
	uploadMetadata := "nama_file dGVzdGZpbGU=,tipe cGRm,semester MQ=="

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(5, nil) // Max queue reached

	result, err := uc.InitiateModulUpload(userID, fileSize, uploadMetadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "antrian penuh")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestInitiateUpload_FileTooLarge tests file size validation for modul
func TestTusModulUsecase_InitiateUpload_FileTooLarge(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)
	fileSize := int64(100 * 1024 * 1024) // 100 MB - exceeds 50 MB limit
	uploadMetadata := "nama_file dGVzdGZpbGU=,tipe cGRm,semester MQ=="

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(0, nil)

	result, err := uc.InitiateModulUpload(userID, fileSize, uploadMetadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
}

// TestInitiateUpload_InvalidMetadata tests invalid metadata
func TestTusModulUsecase_InitiateUpload_InvalidMetadata(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)
	fileSize := int64(1024 * 1024)
	uploadMetadata := "" // Empty metadata

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(0, nil)

	result, err := uc.InitiateModulUpload(userID, fileSize, uploadMetadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "metadata wajib diisi")
}

// TestInitiateUpload_InvalidFileType tests invalid file type
func TestTusModulUsecase_InitiateUpload_InvalidFileType(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)
	fileSize := int64(1024 * 1024)
	// Invalid file type: "exe"
	uploadMetadata := "nama_file dGVzdGZpbGU=,tipe ZXhl,semester MQ=="

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(0, nil)

	result, err := uc.InitiateModulUpload(userID, fileSize, uploadMetadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tipe file harus salah satu dari")
}

// TestUploadChunk_Success tests successful chunk processing for modul
func TestTusModulUsecase_UploadChunk_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunkData := []byte("test chunk data")
	chunk := bytes.NewReader(chunkData)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		FileSize:       int64(1024),
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusPending,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusModulUploadRepo.On("UpdateStatus", uploadID, domain.ModulUploadStatusUploading).Return(nil)
	mockTusModulUploadRepo.On("UpdateOffset", uploadID, int64(len(chunkData)), mock.AnythingOfType("float64")).Return(nil)

	// Initiate the upload in TusStore first
	tusManager.InitiateUpload(uploadID, int64(1024), map[string]string{"user_id": "1"})

	newOffset, err := uc.HandleModulChunk(uploadID, userID, offset, chunk)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunkData)), newOffset)

	mockTusModulUploadRepo.AssertExpectations(t)

	// Cleanup
	tusManager.CancelUpload(uploadID)
}

// TestUploadChunk_InvalidOffset tests offset validation for modul
func TestTusModulUsecase_UploadChunk_InvalidOffset(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(100) // Wrong offset
	chunk := bytes.NewReader([]byte("test data"))

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		FileSize:       int64(1024),
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusUploading,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleModulChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "offset tidak valid")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_UploadNotFound tests when upload is not found
func TestTusModulUsecase_UploadChunk_UploadNotFound(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "non-existent-upload"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(nil, gorm.ErrRecordNotFound)

	newOffset, err := uc.HandleModulChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "upload tidak ditemukan")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestUploadChunk_UnauthorizedUser tests when user doesn't own the upload
func TestTusModulUsecase_UploadChunk_UnauthorizedUser(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2) // Different user
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         ownerID,
		FileSize:       int64(1024),
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusUploading,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleModulChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestGetUploadStatus_Success tests successful status retrieval for modul
func TestTusModulUsecase_GetUploadStatus_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	expectedOffset := int64(512)
	expectedFileSize := int64(1024)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		FileSize:       expectedFileSize,
		CurrentOffset:  expectedOffset,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	offset, fileSize, err := uc.GetModulUploadStatus(uploadID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedOffset, offset)
	assert.Equal(t, expectedFileSize, fileSize)

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestGetUploadStatus_Unauthorized tests unauthorized access
func TestTusModulUsecase_GetUploadStatus_Unauthorized(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	ownerID := uint(1)
	userID := uint(2)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         ownerID,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	offset, fileSize, err := uc.GetModulUploadStatus(uploadID, userID)

	assert.Error(t, err)
	assert.Equal(t, int64(0), offset)
	assert.Equal(t, int64(0), fileSize)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestGetUploadInfo_Success tests successful upload info retrieval
func TestTusModulUsecase_GetUploadInfo_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	modulID := uint(123)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		ModulID:        &modulID,
		UploadMetadata: metadata,
		FileSize:       int64(1024),
		CurrentOffset:  int64(512),
		Status:         domain.ModulUploadStatusUploading,
		Progress:       50.0,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	info, err := uc.GetModulUploadInfo(uploadID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, info)
	assert.Equal(t, uploadID, info.UploadID)
	assert.Equal(t, "testfile", info.NamaFile)
	assert.Equal(t, "pdf", info.Tipe)
	assert.Equal(t, 1, info.Semester)
	assert.Equal(t, domain.ModulUploadStatusUploading, info.Status)
	assert.Equal(t, int64(512), info.Offset)
	assert.Equal(t, int64(1024), info.Length)
	assert.Equal(t, modulID, info.ModulID)

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_Success tests successful upload cancellation
func TestTusModulUsecase_CancelUpload_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		Status:         domain.ModulUploadStatusUploading,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusModulUploadRepo.On("UpdateStatus", uploadID, domain.ModulUploadStatusCancelled).Return(nil)

	err := uc.CancelModulUpload(uploadID, userID)

	assert.NoError(t, err)

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestCancelUpload_AlreadyCompleted tests cancelling a completed upload
func TestTusModulUsecase_CancelUpload_AlreadyCompleted(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		Status:         domain.ModulUploadStatusCompleted,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	err := uc.CancelModulUpload(uploadID, userID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "upload sudah selesai")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestCheckUploadSlot_Success tests checking modul upload slot availability
func TestTusModulUsecase_CheckUploadSlot_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(2, nil)

	response, err := uc.CheckModulUploadSlot(userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.True(t, response.Available)
	assert.Equal(t, 2, response.QueueLength)
	assert.Equal(t, 5, response.MaxQueue)
	assert.Contains(t, response.Message, "Slot tersedia")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestCheckUploadSlot_QueueFull tests when modul upload queue is full
func TestTusModulUsecase_CheckUploadSlot_QueueFull(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(5, nil) // Max queue reached

	response, err := uc.CheckModulUploadSlot(userID)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.False(t, response.Available)
	assert.Equal(t, 5, response.QueueLength)
	assert.Equal(t, 5, response.MaxQueue)
	assert.Contains(t, response.Message, "Antrian penuh")

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestInitiateModulUpdateUpload_Success tests successful modul update upload initiation
func TestTusModulUsecase_InitiateModulUpdateUpload_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	modulID := uint(1)
	userID := uint(1)
	fileSize := int64(1024 * 1024) // 1 MB
	uploadMetadata := "nama_file dXBkYXRlZGZpbGU=,tipe cGRm,semester Mg=="

	modul := &domain.Modul{
		ID:       modulID,
		UserID:   userID,
		NamaFile: "oldfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	mockModulRepo.On("GetByID", modulID).Return(modul, nil)
	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(0, nil)
	mockTusModulUploadRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil)

	result, err := uc.InitiateModulUpdateUpload(modulID, userID, fileSize, uploadMetadata)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.NotEmpty(t, result.UploadID)

	// Cleanup
	tusManager.CancelUpload(result.UploadID)

	mockModulRepo.AssertExpectations(t)
	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestInitiateModulUpdateUpload_ModulNotFound tests when modul is not found
func TestTusModulUsecase_InitiateModulUpdateUpload_ModulNotFound(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	modulID := uint(999)
	userID := uint(1)
	fileSize := int64(1024 * 1024)
	uploadMetadata := "nama_file dXBkYXRlZGZpbGU=,tipe cGRm,semester Mg=="

	mockModulRepo.On("GetByID", modulID).Return(nil, gorm.ErrRecordNotFound)

	result, err := uc.InitiateModulUpdateUpload(modulID, userID, fileSize, uploadMetadata)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "modul tidak ditemukan")

	mockModulRepo.AssertExpectations(t)
}

// TestHandleModulUpdateChunk_Success tests successful modul update chunk handling
func TestTusModulUsecase_HandleModulUpdateChunk_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunkData := []byte("test chunk data")
	chunk := bytes.NewReader(chunkData)

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "updatedfile",
		Tipe:     "pdf",
		Semester: 2,
	}

	modulID := uint(1)
	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		ModulID:        &modulID,
		FileSize:       int64(1024),
		Status:         domain.ModulUploadStatusPending,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)
	mockTusModulUploadRepo.On("UpdateStatus", uploadID, domain.ModulUploadStatusUploading).Return(nil)
	mockTusModulUploadRepo.On("UpdateOffset", uploadID, int64(len(chunkData)), mock.AnythingOfType("float64")).Return(nil)

	// Initiate the upload in TusStore first
	tusManager.InitiateUpload(uploadID, int64(1024), map[string]string{"user_id": "1", "modul_id": "1"})

	newOffset, err := uc.HandleModulUpdateChunk(uploadID, userID, offset, chunk)

	assert.NoError(t, err)
	assert.Equal(t, int64(len(chunkData)), newOffset)

	mockTusModulUploadRepo.AssertExpectations(t)

	// Cleanup
	tusManager.CancelUpload(uploadID)
}

// TestUploadChunk_InactiveStatus tests uploading to an inactive upload
func TestTusModulUsecase_UploadChunk_InactiveStatus(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)
	offset := int64(0)
	chunk := bytes.NewReader([]byte("test data"))

	metadata := domain.TusModulUploadInitRequest{
		NamaFile: "testfile",
		Tipe:     "pdf",
		Semester: 1,
	}

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		FileSize:       int64(1024),
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusCompleted,
		UploadMetadata: metadata,
	}

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(tusUpload, nil)

	newOffset, err := uc.HandleModulChunk(uploadID, userID, offset, chunk)

	assert.Error(t, err)
	assert.Equal(t, int64(0), newOffset)
	assert.Contains(t, err.Error(), "upload tidak aktif")

	mockTusModulUploadRepo.AssertExpectations(t)
}
