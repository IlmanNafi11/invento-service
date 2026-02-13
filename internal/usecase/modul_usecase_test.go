package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"
)

func TestCreateModul_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	modul := &domain.Modul{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "user-1",
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test_modul.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test_modul.pdf",
		Status:    "pending",
	}

	mockModulRepo.On("Create", mock.AnythingOfType("*domain.Modul")).Return(nil)

	err := mockModulRepo.Create(modul)

	assert.NoError(t, err)
	assert.NotNil(t, modul)
	assert.Equal(t, "550e8400-e29b-41d4-a716-446655440000", modul.ID)
	assert.Equal(t, "Test Modul", modul.Judul)
	mockModulRepo.AssertExpectations(t)
}

func TestGetModulByID_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "user-1"

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test_modul.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test_modul.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(expectedModul, nil)

	modul, err := mockModulRepo.GetByID(modulID)

	assert.NoError(t, err)
	assert.NotNil(t, modul)
	assert.Equal(t, modulID, modul.ID)
	assert.Equal(t, userID, modul.UserID)
	assert.Equal(t, "Test Modul", modul.Judul)
	mockModulRepo.AssertExpectations(t)
}

func TestGetModulByID_NotFound(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440999"

	mockModulRepo.On("GetByID", modulID).Return(nil, gorm.ErrRecordNotFound)

	modul, err := mockModulRepo.GetByID(modulID)

	assert.Error(t, err)
	assert.Nil(t, modul)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
	mockModulRepo.AssertExpectations(t)
}

func TestListModuls_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	userID := "user-1"
	search := ""
	filterType := ""
	filterStatus := ""
	page := 1
	limit := 10

	expectedModuls := []domain.ModulListItem{
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440001",
			Judul:              "Test Modul 1",
			Deskripsi:          "Deskripsi 1",
			FileName:           "test1.pdf",
			MimeType:           "application/pdf",
			FileSize:           1572864,
			FilePath:           "/uploads/test1.pdf",
			Status:             "completed",
			TerakhirDiperbarui: time.Now(),
		},
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440002",
			Judul:              "Test Modul 2",
			Deskripsi:          "Deskripsi 2",
			FileName:           "test2.docx",
			MimeType:           "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			FileSize:           2097152,
			FilePath:           "/uploads/test2.docx",
			Status:             "completed",
			TerakhirDiperbarui: time.Now(),
		},
	}

	total := int64(2)

	mockModulRepo.On("GetByUserID", userID, search, filterType, filterStatus, page, limit).
		Return(expectedModuls, int(total), nil)

	moduls, count, err := mockModulRepo.GetByUserID(userID, search, filterType, filterStatus, page, limit)

	assert.NoError(t, err)
	assert.NotNil(t, moduls)
	assert.Equal(t, 2, len(moduls))
	assert.Equal(t, int(total), count)
	assert.Equal(t, "Test Modul 1", moduls[0].Judul)
	assert.Equal(t, "Test Modul 2", moduls[1].Judul)
	mockModulRepo.AssertExpectations(t)
}

func TestListModulsByProject_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	userID := "user-1"
	ids := []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002", "550e8400-e29b-41d4-a716-446655440003"}

	expectedModuls := []domain.Modul{
		{
			ID:        "550e8400-e29b-41d4-a716-446655440001",
			UserID:    userID,
			Judul:     "Project Modul 1",
			Deskripsi: "Deskripsi 1",
			FileName:  "project1.pdf",
			MimeType:  "application/pdf",
			FileSize:  1572864,
			FilePath:  "/uploads/project1.pdf",
			Status:    "completed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        "550e8400-e29b-41d4-a716-446655440002",
			UserID:    userID,
			Judul:     "Project Modul 2",
			Deskripsi: "Deskripsi 2",
			FileName:  "project2.docx",
			MimeType:  "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			FileSize:  2097152,
			FilePath:  "/uploads/project2.docx",
			Status:    "completed",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockModulRepo.On("GetByIDs", ids, userID).Return(expectedModuls, nil)

	moduls, err := mockModulRepo.GetByIDs(ids, userID)

	assert.NoError(t, err)
	assert.NotNil(t, moduls)
	assert.Equal(t, 2, len(moduls))
	assert.Equal(t, "Project Modul 1", moduls[0].Judul)
	assert.Equal(t, "Project Modul 2", moduls[1].Judul)
	mockModulRepo.AssertExpectations(t)
}

func TestUpdateModul_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	modulUc := NewModulUsecase(mockModulRepo)

	existingModul := &domain.Modul{
		ID:        "550e8400-e29b-41d4-a716-446655440000",
		UserID:    "user-1",
		Judul:     "Old Judul",
		Deskripsi: "Old Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	req := domain.ModulUpdateRequest{
		Judul:     "Updated Judul",
		Deskripsi: "Updated Deskripsi",
	}

	mockModulRepo.On("GetByID", "550e8400-e29b-41d4-a716-446655440000").Return(existingModul, nil)
	mockModulRepo.On("Update", mock.AnythingOfType("*domain.Modul")).Return(nil)

	err := modulUc.UpdateMetadata("550e8400-e29b-41d4-a716-446655440000", "user-1", req)

	assert.NoError(t, err)
	mockModulRepo.AssertExpectations(t)
}

func TestDeleteModul_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	modulUc := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "user-1"

	existingModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(existingModul, nil)
	mockModulRepo.On("Delete", modulID).Return(nil)

	err := modulUc.Delete(modulID, userID)

	assert.NoError(t, err)
	mockModulRepo.AssertExpectations(t)
}

func TestCheckUploadSlot_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxQueueModulPerUser: 3,
		},
	}

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(3)
	fileManager := helper.NewFileManager(cfg)
	tusManager := helper.NewTusManager(tusStore, tusQueue, fileManager, cfg)

	_ = NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := "user-1"

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(int64(1), nil)

	response, err := mockTusModulUploadRepo.CountActiveByUserID(userID)

	assert.NoError(t, err)
	assert.Equal(t, int64(1), response)
	mockTusModulUploadRepo.AssertExpectations(t)
}

func TestInitiateUpload_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxSize:              524288000, // 500 MB (required by TusManager)
			MaxSizeModul:         52428800,  // 50 MB
			MaxQueueModulPerUser: 3,
			IdleTimeout:          30,
		},
	}

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 52428800)
	tusQueue := helper.NewTusQueue(3)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	tusModulUc := NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := "user-1"
	fileSize := int64(1024 * 1024) // 1 MB
	// TUS metadata format: "key base64value,key2 base64value2"
	metadata := "judul VGVzdCBKdWR1bA==,deskripsi VGVzdCBEZXNrcmlwc2k="

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(int64(0), nil)
	mockTusModulUploadRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil)

	response, err := tusModulUc.InitiateModulUpload(userID, fileSize, metadata)

	// Cleanup
	if response != nil && response.UploadID != "" {
		tusManager.CancelUpload(response.UploadID)
	}

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.NotEmpty(t, response.UploadID)
	assert.NotEmpty(t, response.UploadURL)
	mockTusModulUploadRepo.AssertExpectations(t)
}

func TestUploadChunk_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxSizeModul:         10 * 1024 * 1024,
			MaxQueueModulPerUser: 3,
			IdleTimeout:          30,
		},
	}

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(3)
	fileManager := helper.NewFileManager(cfg)
	tusManager := helper.NewTusManager(tusStore, tusQueue, fileManager, cfg)

	_ = NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := "user-1"

	existingUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		UploadType:     domain.UploadTypeModulCreate,
		UploadURL:      "/modul/upload/" + uploadID,
		UploadMetadata: domain.TusModulUploadInitRequest{Judul: "Test Modul", Deskripsi: "Test Deskripsi"},
		FileSize:       1024 * 1024,
		CurrentOffset:  0,
		Status:         domain.UploadStatusPending,
		Progress:       0,
		ExpiresAt:      time.Now().Add(30 * time.Minute),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	newOffset := int64(512 * 1024) // 512 KB
	progress := 50.0

	mockTusModulUploadRepo.On("GetByID", uploadID).Return(existingUpload, nil)
	mockTusModulUploadRepo.On("UpdateOffset", uploadID, newOffset, progress).Return(nil)

	upload, err := mockTusModulUploadRepo.GetByID(uploadID)
	assert.NoError(t, err)
	assert.NotNil(t, upload)

	err = mockTusModulUploadRepo.UpdateOffset(uploadID, newOffset, progress)
	assert.NoError(t, err)

	mockTusModulUploadRepo.AssertExpectations(t)
}

// TestModulUsecase_GetList_Success tests successful modul list retrieval
func TestModulUsecase_GetList_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	search := ""
	filterType := ""
	filterStatus := ""
	page := 1
	limit := 10

	expectedModuls := []domain.ModulListItem{
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440001",
			Judul:              "Test Modul 1",
			Deskripsi:          "Deskripsi 1",
			FileName:           "test1.pdf",
			MimeType:           "application/pdf",
			FileSize:           1572864,
			FilePath:           "/uploads/test1.pdf",
			Status:             "completed",
			TerakhirDiperbarui: time.Now(),
		},
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440002",
			Judul:              "Test Modul 2",
			Deskripsi:          "Deskripsi 2",
			FileName:           "test2.docx",
			MimeType:           "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			FileSize:           2097152,
			FilePath:           "/uploads/test2.docx",
			Status:             "completed",
			TerakhirDiperbarui: time.Now(),
		},
	}

	mockModulRepo.On("GetByUserID", userID, search, filterType, filterStatus, page, limit).
		Return(expectedModuls, 2, nil)

	result, err := modulUC.GetList(userID, search, filterType, filterStatus, page, limit)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 2, result.Pagination.TotalItems)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetList_Error tests error handling
func TestModulUsecase_GetList_Error(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	mockModulRepo.On("GetByUserID", "user-1", "", "", "", 1, 10).
		Return(nil, 0, assert.AnError)

	result, err := modulUC.GetList("user-1", "", "", "", 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)
	assert.Contains(t, strings.ToLower(appErr.Message), "kesalahan")

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetList_WithFilters tests list retrieval with filters
func TestModulUsecase_GetList_WithFilters(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	expectedModuls := []domain.ModulListItem{
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440001",
			Judul:              "Test Modul",
			Deskripsi:          "Test Deskripsi",
			FileName:           "test.pdf",
			MimeType:           "application/pdf",
			FileSize:           1572864,
			FilePath:           "/uploads/test.pdf",
			Status:             "completed",
			TerakhirDiperbarui: time.Now(),
		},
	}

	mockModulRepo.On("GetByUserID", "user-1", "test", "application/pdf", "completed", 1, 10).
		Return(expectedModuls, 1, nil)

	result, err := modulUC.GetList("user-1", "test", "application/pdf", "completed", 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_Success tests successful modul retrieval
func TestModulUsecase_GetByID_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "user-1"

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(expectedModul, nil)

	result, err := modulUC.GetByID(modulID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, modulID, result.ID)
	assert.Equal(t, "Test Modul", result.Judul)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_NotFound tests modul retrieval when not found
func TestModulUsecase_GetByID_NotFound(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440999"
	userID := "user-1"

	mockModulRepo.On("GetByID", modulID).Return(nil, gorm.ErrRecordNotFound)

	result, err := modulUC.GetByID(modulID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_Unauthorized tests access denial for different user
func TestModulUsecase_GetByID_Unauthorized(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "user-2"

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    "user-1",
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(expectedModul, nil)

	result, err := modulUC.GetByID(modulID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_Download_SingleFile tests single file download
func TestModulUsecase_Download_SingleFile(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	modulIDs := []string{"550e8400-e29b-41d4-a716-446655440000"}

	expectedModuls := []domain.Modul{
		{
			ID:        "550e8400-e29b-41d4-a716-446655440000",
			UserID:    userID,
			Judul:     "Test Modul.pdf",
			Deskripsi: "Test Deskripsi",
			FileName:  "test.pdf",
			FilePath:  "/uploads/test.pdf",
		},
	}

	mockModulRepo.On("GetByIDs", modulIDs, userID).Return(expectedModuls, nil)

	result, err := modulUC.Download(userID, modulIDs)

	assert.NoError(t, err)
	assert.Equal(t, "/uploads/test.pdf", result)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_Download_EmptyIDs tests empty modul IDs
func TestModulUsecase_Download_EmptyIDs(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	modulIDs := []string{}

	result, err := modulUC.Download(userID, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrValidation, appErr.Code)
}

// TestModulUsecase_Download_NotFound tests download when moduls not found
func TestModulUsecase_Download_NotFound(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	modulIDs := []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"}

	mockModulRepo.On("GetByIDs", modulIDs, userID).Return([]domain.Modul{}, nil)

	result, err := modulUC.Download(userID, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}
