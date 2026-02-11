package usecase

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
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
		ID:       1,
		UserID:   1,
		NamaFile: "Test Modul",
		Tipe:     "pdf",
		Ukuran:   "1.5 MB",
		Semester: 3,
		PathFile: "/uploads/test_modul.pdf",
	}

	mockModulRepo.On("Create", mock.AnythingOfType("*domain.Modul")).Return(nil)

	err := mockModulRepo.Create(modul)

	assert.NoError(t, err)
	assert.NotNil(t, modul)
	assert.Equal(t, uint(1), modul.ID)
	assert.Equal(t, "Test Modul", modul.NamaFile)
	mockModulRepo.AssertExpectations(t)
}

func TestGetModulByID_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	modulID := uint(1)
	userID := uint(1)

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		NamaFile:  "Test Modul",
		Tipe:      "pdf",
		Ukuran:    "1.5 MB",
		Semester:  3,
		PathFile:  "/uploads/test_modul.pdf",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(expectedModul, nil)

	modul, err := mockModulRepo.GetByID(modulID)

	assert.NoError(t, err)
	assert.NotNil(t, modul)
	assert.Equal(t, modulID, modul.ID)
	assert.Equal(t, userID, modul.UserID)
	assert.Equal(t, "Test Modul", modul.NamaFile)
	mockModulRepo.AssertExpectations(t)
}

func TestGetModulByID_NotFound(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	modulID := uint(999)

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

	userID := uint(1)
	search := ""
	filterType := ""
	filterSemester := 0
	page := 1
	limit := 10

	expectedModuls := []domain.ModulListItem{
		{
			ID:                 1,
			NamaFile:           "Test Modul 1",
			Tipe:               "pdf",
			Ukuran:             "1.5 MB",
			Semester:           3,
			PathFile:           "/uploads/test1.pdf",
			TerakhirDiperbarui: time.Now(),
		},
		{
			ID:                 2,
			NamaFile:           "Test Modul 2",
			Tipe:               "docx",
			Ukuran:             "2.0 MB",
			Semester:           4,
			PathFile:           "/uploads/test2.docx",
			TerakhirDiperbarui: time.Now(),
		},
	}

	total := int64(2)

	mockModulRepo.On("GetByUserID", userID, search, filterType, filterSemester, page, limit).
		Return(expectedModuls, int(total), nil)

	moduls, count, err := mockModulRepo.GetByUserID(userID, search, filterType, filterSemester, page, limit)

	assert.NoError(t, err)
	assert.NotNil(t, moduls)
	assert.Equal(t, 2, len(moduls))
	assert.Equal(t, int(total), count)
	assert.Equal(t, "Test Modul 1", moduls[0].NamaFile)
	assert.Equal(t, "Test Modul 2", moduls[1].NamaFile)
	mockModulRepo.AssertExpectations(t)
}

func TestListModulsByProject_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = NewModulUsecase(mockModulRepo)

	userID := uint(1)
	ids := []uint{1, 2, 3}

	expectedModuls := []domain.Modul{
		{
			ID:        1,
			UserID:    userID,
			NamaFile:  "Project Modul 1",
			Tipe:      "pdf",
			Ukuran:    "1.5 MB",
			Semester:  3,
			PathFile:  "/uploads/project1.pdf",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        2,
			UserID:    userID,
			NamaFile:  "Project Modul 2",
			Tipe:      "docx",
			Ukuran:    "2.0 MB",
			Semester:  4,
			PathFile:  "/uploads/project2.docx",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockModulRepo.On("GetByIDs", ids, userID).Return(expectedModuls, nil)

	moduls, err := mockModulRepo.GetByIDs(ids, userID)

	assert.NoError(t, err)
	assert.NotNil(t, moduls)
	assert.Equal(t, 2, len(moduls))
	assert.Equal(t, "Project Modul 1", moduls[0].NamaFile)
	assert.Equal(t, "Project Modul 2", moduls[1].NamaFile)
	mockModulRepo.AssertExpectations(t)
}

func TestUpdateModul_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	modulUc := NewModulUsecase(mockModulRepo)

	existingModul := &domain.Modul{
		ID:        1,
		UserID:    1,
		NamaFile:  "Old Name",
		Tipe:      "pdf",
		Ukuran:    "1.5 MB",
		Semester:  3,
		PathFile:  "/uploads/test.pdf",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	req := domain.ModulUpdateRequest{
		NamaFile:  "Updated Name",
		Semester:  5,
	}

	mockModulRepo.On("GetByID", uint(1)).Return(existingModul, nil)
	mockModulRepo.On("Update", mock.AnythingOfType("*domain.Modul")).Return(nil)

	err := modulUc.UpdateMetadata(1, 1, req)

	assert.NoError(t, err)
	mockModulRepo.AssertExpectations(t)
}

func TestDeleteModul_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	modulUc := NewModulUsecase(mockModulRepo)

	modulID := uint(1)
	userID := uint(1)

	existingModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		NamaFile:  "Test Modul",
		Tipe:      "pdf",
		Ukuran:    "1.5 MB",
		Semester:  3,
		PathFile:  "/uploads/test.pdf",
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

	userID := uint(1)

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(1, nil)

	response, err := mockTusModulUploadRepo.CountActiveByUserID(userID)

	assert.NoError(t, err)
	assert.Equal(t, 1, response)
	mockTusModulUploadRepo.AssertExpectations(t)
}

func TestInitiateUpload_Success(t *testing.T) {
	mockTusModulUploadRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxSize:             524288000, // 500 MB (required by TusManager)
			MaxSizeModul:        52428800,  // 50 MB
			MaxQueueModulPerUser: 3,
			IdleTimeout:         30,
		},
	}

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 52428800)
	tusQueue := helper.NewTusQueue(3)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	tusModulUc := NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	userID := uint(1)
	fileSize := int64(1024 * 1024) // 1 MB
	// TUS metadata format: "key base64value,key2 base64value2"
	// nama_file testfile,tipe pdf,semester 1
	metadata := "nama_file dGVzdGZpbGU=,tipe cGRm,semester MQ=="

	mockTusModulUploadRepo.On("CountActiveByUserID", userID).Return(0, nil)
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
			MaxSizeModul:        10 * 1024 * 1024,
			MaxQueueModulPerUser: 3,
			IdleTimeout:         30,
		},
	}

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, 10*1024*1024)
	tusQueue := helper.NewTusQueue(3)
	fileManager := helper.NewFileManager(cfg)
	tusManager := helper.NewTusManager(tusStore, tusQueue, fileManager, cfg)

	_ = NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

	uploadID := "test-upload-id"
	userID := uint(1)

	existingUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		UploadType:     domain.ModulUploadTypeCreate,
		UploadURL:      "/modul/upload/" + uploadID,
		UploadMetadata: domain.TusModulUploadInitRequest{NamaFile: "Test Modul", Tipe: "pdf", Semester: 3},
		FileSize:       1024 * 1024,
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusPending,
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

	userID := uint(1)
	search := ""
	filterType := ""
	filterSemester := 0
	page := 1
	limit := 10

	expectedModuls := []domain.ModulListItem{
		{
			ID:                 1,
			NamaFile:           "Test Modul 1",
			Tipe:               "pdf",
			Ukuran:             "1.5 MB",
			Semester:           3,
			PathFile:           "/uploads/test1.pdf",
			TerakhirDiperbarui: time.Now(),
		},
		{
			ID:                 2,
			NamaFile:           "Test Modul 2",
			Tipe:               "docx",
			Ukuran:             "2.0 MB",
			Semester:           4,
			PathFile:           "/uploads/test2.docx",
			TerakhirDiperbarui: time.Now(),
		},
	}

	mockModulRepo.On("GetByUserID", userID, search, filterType, filterSemester, page, limit).
		Return(expectedModuls, 2, nil)

	result, err := modulUC.GetList(userID, search, filterType, filterSemester, page, limit)

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

	mockModulRepo.On("GetByUserID", uint(1), "", "", 0, 1, 10).
		Return(nil, 0, assert.AnError)

	result, err := modulUC.GetList(1, "", "", 0, 1, 10)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "gagal mengambil data modul")

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetList_WithFilters tests list retrieval with filters
func TestModulUsecase_GetList_WithFilters(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	expectedModuls := []domain.ModulListItem{
		{
			ID:                 1,
			NamaFile:           "Test Modul",
			Tipe:               "pdf",
			Ukuran:             "1.5 MB",
			Semester:           3,
			PathFile:           "/uploads/test.pdf",
			TerakhirDiperbarui: time.Now(),
		},
	}

	mockModulRepo.On("GetByUserID", uint(1), "test", "pdf", 3, 1, 10).
		Return(expectedModuls, 1, nil)

	result, err := modulUC.GetList(1, "test", "pdf", 3, 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_Success tests successful modul retrieval
func TestModulUsecase_GetByID_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := uint(1)
	userID := uint(1)

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		NamaFile:  "Test Modul",
		Tipe:      "pdf",
		Ukuran:    "1.5 MB",
		Semester:  3,
		PathFile:  "/uploads/test.pdf",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(expectedModul, nil)

	result, err := modulUC.GetByID(modulID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, modulID, result.ID)
	assert.Equal(t, "Test Modul", result.NamaFile)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_NotFound tests modul retrieval when not found
func TestModulUsecase_GetByID_NotFound(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := uint(999)
	userID := uint(1)

	mockModulRepo.On("GetByID", modulID).Return(nil, gorm.ErrRecordNotFound)

	result, err := modulUC.GetByID(modulID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "modul tidak ditemukan")

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_Unauthorized tests access denial for different user
func TestModulUsecase_GetByID_Unauthorized(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := uint(1)
	userID := uint(2)

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    1,
		NamaFile:  "Test Modul",
		Tipe:      "pdf",
		Ukuran:    "1.5 MB",
		Semester:  3,
		PathFile:  "/uploads/test.pdf",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", modulID).Return(expectedModul, nil)

	result, err := modulUC.GetByID(modulID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "tidak memiliki akses")

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_Download_SingleFile tests single file download
func TestModulUsecase_Download_SingleFile(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := uint(1)
	modulIDs := []uint{1}

	expectedModuls := []domain.Modul{
		{
			ID:       1,
			UserID:   userID,
			NamaFile: "Test Modul.pdf",
			PathFile: "/uploads/test.pdf",
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

	userID := uint(1)
	modulIDs := []uint{}

	result, err := modulUC.Download(userID, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "id modul tidak boleh kosong")
}

// TestModulUsecase_Download_NotFound tests download when moduls not found
func TestModulUsecase_Download_NotFound(t *testing.T) {
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := uint(1)
	modulIDs := []uint{1, 2}

	mockModulRepo.On("GetByIDs", modulIDs, userID).Return([]domain.Modul{}, nil)

	result, err := modulUC.Download(userID, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Contains(t, err.Error(), "modul tidak ditemukan")

	mockModulRepo.AssertExpectations(t)
}
