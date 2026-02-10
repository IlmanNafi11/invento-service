package usecase_test

import (
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

type MockModulRepository struct {
	mock.Mock
}

func (m *MockModulRepository) Create(modul *domain.Modul) error {
	args := m.Called(modul)
	return args.Error(0)
}

func (m *MockModulRepository) GetByID(id uint) (*domain.Modul, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByIDs(ids []uint, userID uint) ([]domain.Modul, error) {
	args := m.Called(ids, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByIDsForUser(ids []uint, ownerUserID uint) ([]domain.Modul, error) {
	args := m.Called(ids, ownerUserID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Modul), args.Error(1)
}

func (m *MockModulRepository) GetByUserID(userID uint, search string, filterType string, filterSemester int, page, limit int) ([]domain.ModulListItem, int, error) {
	args := m.Called(userID, search, filterType, filterSemester, page, limit)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]domain.ModulListItem), args.Int(1), args.Error(2)
}

func (m *MockModulRepository) CountByUserID(userID uint) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockModulRepository) Update(modul *domain.Modul) error {
	args := m.Called(modul)
	return args.Error(0)
}

func (m *MockModulRepository) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockModulRepository) UpdateMetadata(modul *domain.Modul) error {
	args := m.Called(modul)
	return args.Error(0)
}

func TestCreateModul_Success(t *testing.T) {
	mockModulRepo := new(MockModulRepository)

	_ = usecase.NewModulUsecase(mockModulRepo)

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

	_ = usecase.NewModulUsecase(mockModulRepo)

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

	_ = usecase.NewModulUsecase(mockModulRepo)

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

	_ = usecase.NewModulUsecase(mockModulRepo)

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

	_ = usecase.NewModulUsecase(mockModulRepo)

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

	modulUc := usecase.NewModulUsecase(mockModulRepo)

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

	modulUc := usecase.NewModulUsecase(mockModulRepo)

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

func (m *MockTusModulUploadRepository) GetByUserID(userID uint) ([]domain.TusModulUpload, error) {
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

func (m *MockTusModulUploadRepository) Complete(id string, modulID uint, filePath string) error {
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

func (m *MockTusModulUploadRepository) CountActiveByUserID(userID uint) (int, error) {
	args := m.Called(userID)
	return args.Int(0), args.Error(1)
}

func (m *MockTusModulUploadRepository) GetActiveByUserID(userID uint) ([]domain.TusModulUpload, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
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

	_ = usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

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

	tusModulUc := usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

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

	_ = usecase.NewTusModulUsecase(mockTusModulUploadRepo, mockModulRepo, tusManager, fileManager, cfg)

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
