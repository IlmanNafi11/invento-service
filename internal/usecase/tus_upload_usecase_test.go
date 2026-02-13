package usecase

import (
	"bytes"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTusUploadTestDeps(t *testing.T) (*tusUploadUsecase, *MockTusUploadRepository, *MockProjectRepository, *helper.TusManager) {
	t.Helper()

	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()
	baseDir := t.TempDir()
	cfg.App.Env = "development"
	cfg.Upload.PathDevelopment = filepath.Join(baseDir, "uploads")
	cfg.Upload.TempPathDevelopment = filepath.Join(baseDir, "temp")

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	tusQueue := helper.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg).(*tusUploadUsecase)

	return uc, mockTusUploadRepo, mockProjectRepo, tusManager
}

func seedTusUploadStore(t *testing.T, manager *helper.TusManager, uploadID string, size int64, metadata map[string]string) {
	t.Helper()
	require.NoError(t, manager.InitiateUpload(uploadID, size, metadata))
}

func TestTusUploadUsecase(t *testing.T) {
	t.Run("CheckUploadSlot available and not available", func(t *testing.T) {
		t.Run("available", func(t *testing.T) {
			uc, _, _, _ := newTusUploadTestDeps(t)

			res, err := uc.CheckUploadSlot("u1")
			require.NoError(t, err)
			assert.True(t, res.Available)
			assert.False(t, res.ActiveUpload)
			assert.Equal(t, 0, res.QueueLength)
		})

		t.Run("not available", func(t *testing.T) {
			uc, _, _, manager := newTusUploadTestDeps(t)
			manager.AddToQueue("busy-upload")

			res, err := uc.CheckUploadSlot("u1")
			require.NoError(t, err)
			assert.False(t, res.Available)
			assert.True(t, res.ActiveUpload)
		})
	})

	t.Run("ResetUploadQueue clears active upload", func(t *testing.T) {
		uc, _, _, manager := newTusUploadTestDeps(t)
		seedTusUploadStore(t, manager, "active-id", 16, map[string]string{"user_id": "u1"})
		manager.AddToQueue("active-id")

		resBefore, err := uc.CheckUploadSlot("u1")
		require.NoError(t, err)
		assert.False(t, resBefore.Available)

		require.NoError(t, uc.ResetUploadQueue("u1"))

		resAfter, err := uc.CheckUploadSlot("u1")
		require.NoError(t, err)
		assert.True(t, resAfter.Available)
		assert.False(t, resAfter.ActiveUpload)
	})

	t.Run("InitiateUpload", func(t *testing.T) {
		metadata := domain.TusUploadInitRequest{NamaProject: "Project Alpha", Kategori: "website", Semester: 2}

		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			tusRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil).Once()

			res, err := uc.InitiateUpload("u1", "u1@mail.com", "mahasiswa", 1024, metadata)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.NotEmpty(t, res.UploadID)
			assert.Equal(t, int64(1024), res.Length)

			_, _, err = manager.GetUploadStatus(res.UploadID)
			assert.NoError(t, err)
			tusRepo.AssertExpectations(t)
		})

		t.Run("file too large", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)

			res, err := uc.InitiateUpload("u1", "u1@mail.com", "mahasiswa", uc.config.Upload.MaxSizeProject+1, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
			tusRepo.AssertNotCalled(t, "Create", mock.Anything)
		})

		t.Run("no upload slot", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			manager.AddToQueue("active")

			res, err := uc.InitiateUpload("u1", "u1@mail.com", "mahasiswa", 256, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "slot upload tidak tersedia")
			tusRepo.AssertNotCalled(t, "Create", mock.Anything)
		})
	})

	t.Run("HandleChunk", func(t *testing.T) {
		t.Run("happy path updates offset", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			uploadID := "upload-happy"
			chunk := []byte("abcd")

			tusRepo.On("GetByID", uploadID).Return(&domain.TusUpload{
				ID:            uploadID,
				UserID:        "u1",
				FileSize:      10,
				CurrentOffset: 0,
				Status:        domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(len(chunk)), mock.MatchedBy(func(progress float64) bool {
				return progress > 0 && progress < 100
			})).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})
			newOffset, err := uc.HandleChunk(uploadID, "u1", 0, bytes.NewReader(chunk))
			require.NoError(t, err)
			assert.Equal(t, int64(4), newOffset)
			tusRepo.AssertExpectations(t)
		})

		t.Run("pending transitions to uploading", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			uploadID := "upload-pending"

			tusRepo.On("GetByID", uploadID).Return(&domain.TusUpload{
				ID:            uploadID,
				UserID:        "u1",
				FileSize:      10,
				CurrentOffset: 0,
				Status:        domain.UploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", uploadID, domain.UploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})
			newOffset, err := uc.HandleChunk(uploadID, "u1", 0, bytes.NewReader([]byte("abc")))
			require.NoError(t, err)
			assert.Equal(t, int64(3), newOffset)
			tusRepo.AssertExpectations(t)
		})

		t.Run("auto completion when offset reaches file size", func(t *testing.T) {
			uc, tusRepo, projectRepo, manager := newTusUploadTestDeps(t)
			uploadID := "upload-complete"

			tusRepo.On("GetByID", uploadID).Return(&domain.TusUpload{
				ID:             uploadID,
				UserID:         "u1",
				UploadType:     domain.UploadTypeProjectCreate,
				UploadMetadata: domain.TusUploadInitRequest{NamaProject: "Final", Kategori: "mobile", Semester: 3},
				FileSize:       4,
				CurrentOffset:  0,
				Status:         domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(4), mock.MatchedBy(func(progress float64) bool { return progress == 100 })).Return(nil).Once()
			projectRepo.On("Create", mock.AnythingOfType("*domain.Project")).Return(nil).Once()
			tusRepo.On("Complete", uploadID, mock.AnythingOfType("uint"), mock.AnythingOfType("string")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1"})
			newOffset, err := uc.HandleChunk(uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, int64(4), newOffset)
			tusRepo.AssertExpectations(t)
			projectRepo.AssertExpectations(t)
		})

		t.Run("upload not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			newOffset, err := uc.HandleChunk("missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Equal(t, int64(0), newOffset)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user forbidden", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "u").Return(&domain.TusUpload{ID: "u", UserID: "owner", FileSize: 10, Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleChunk("u", "intruder", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "u").Return(&domain.TusUpload{ID: "u", UserID: "u1", FileSize: 10, Status: domain.UploadStatusCompleted}, nil).Once()

			offset, err := uc.HandleChunk("u", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Equal(t, int64(10), offset)
			assert.Contains(t, err.Error(), "sudah selesai")
		})

		t.Run("inactive status", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "u").Return(&domain.TusUpload{ID: "u", UserID: "u1", FileSize: 10, Status: domain.UploadStatusCancelled}, nil).Once()

			_, err := uc.HandleChunk("u", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak aktif")
		})
	})

	t.Run("GetUploadInfo and GetUploadStatus", func(t *testing.T) {
		t.Run("GetUploadInfo found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(7)
			now := time.Now()
			tusRepo.On("GetByID", "info-id").Return(&domain.TusUpload{
				ID:             "info-id",
				UserID:         "u1",
				ProjectID:      &projectID,
				UploadMetadata: domain.TusUploadInitRequest{NamaProject: "P", Kategori: "iot", Semester: 4},
				Status:         domain.UploadStatusUploading,
				Progress:       50,
				CurrentOffset:  5,
				FileSize:       10,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetUploadInfo("info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, uint(7), info.ProjectID)
			assert.Equal(t, int64(5), info.Offset)
		})

		t.Run("GetUploadInfo not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetUploadInfo("missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetUploadStatus found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "status-id").Return(&domain.TusUpload{ID: "status-id", UserID: "u1", CurrentOffset: 8, FileSize: 16}, nil).Once()

			offset, length, err := uc.GetUploadStatus("status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(8), offset)
			assert.Equal(t, int64(16), length)
		})

		t.Run("GetUploadStatus not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetUploadStatus("missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})
	})

	t.Run("CancelUpload", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			seedTusUploadStore(t, manager, "cancel-id", 8, map[string]string{"user_id": "u1"})
			manager.AddToQueue("cancel-id")

			tusRepo.On("GetByID", "cancel-id").Return(&domain.TusUpload{ID: "cancel-id", UserID: "u1", Status: domain.UploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", "cancel-id", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelUpload("cancel-id", "u1")
			require.NoError(t, err)
			tusRepo.AssertExpectations(t)
		})

		t.Run("not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelUpload("missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			err := uc.CancelUpload("id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			err := uc.CancelUpload("id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("InitiateProjectUpdateUpload", func(t *testing.T) {
		metadata := domain.TusUploadInitRequest{NamaProject: "Update", Kategori: "mobile", Semester: 5}

		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", uint(9)).Return(&domain.Project{ID: 9, UserID: "u1", NamaProject: "Old", Kategori: "website", Semester: 1}, nil).Once()
			tusRepo.On("Create", mock.AnythingOfType("*domain.TusUpload")).Return(nil).Once()

			res, err := uc.InitiateProjectUpdateUpload(9, "u1", 512, metadata)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/project/9/update/")
		})

		t.Run("project not found", func(t *testing.T) {
			uc, _, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", uint(9)).Return(nil, gorm.ErrRecordNotFound).Once()

			res, err := uc.InitiateProjectUpdateUpload(9, "u1", 512, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "Project tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, _, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", uint(9)).Return(&domain.Project{ID: 9, UserID: "owner"}, nil).Once()

			res, err := uc.InitiateProjectUpdateUpload(9, "u1", 512, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("uses existing project metadata when request metadata empty", func(t *testing.T) {
			uc, tusRepo, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", uint(9)).Return(&domain.Project{
				ID:          9,
				UserID:      "u1",
				NamaProject: "Existing Name",
				Kategori:    "website",
				Semester:    6,
			}, nil).Once()
			tusRepo.On("Create", mock.MatchedBy(func(upload *domain.TusUpload) bool {
				return upload.UploadMetadata.NamaProject == "Existing Name" &&
					upload.UploadMetadata.Kategori == "website" &&
					upload.UploadMetadata.Semester == 6 &&
					upload.ProjectID != nil && *upload.ProjectID == 9 &&
					upload.UploadType == domain.UploadTypeProjectUpdate
			})).Return(nil).Once()

			res, err := uc.InitiateProjectUpdateUpload(9, "u1", 512, domain.TusUploadInitRequest{})
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/project/9/update/")
		})
	})

	t.Run("HandleProjectUpdateChunk", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			projectID := uint(1)
			uploadID := "proj-chunk"
			chunk := []byte("xyz")

			tusRepo.On("GetByID", uploadID).Return(&domain.TusUpload{
				ID:            uploadID,
				UserID:        "u1",
				ProjectID:     &projectID,
				FileSize:      10,
				CurrentOffset: 0,
				Status:        domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1", "project_id": "1"})
			offset, err := uc.HandleProjectUpdateChunk(projectID, uploadID, "u1", 0, bytes.NewReader(chunk))
			require.NoError(t, err)
			assert.Equal(t, int64(3), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			uc, tusRepo, projectRepo, manager := newTusUploadTestDeps(t)
			projectID := uint(2)
			uploadID := "proj-complete"
			projectRepo.On("GetByID", projectID).Return(&domain.Project{ID: projectID, UserID: "u1", PathFile: "", NamaProject: "Old", Kategori: "website", Semester: 1}, nil).Once()
			projectRepo.On("Update", mock.AnythingOfType("*domain.Project")).Return(nil).Once()

			tusRepo.On("GetByID", uploadID).Return(&domain.TusUpload{
				ID:             uploadID,
				UserID:         "u1",
				ProjectID:      &projectID,
				UploadType:     domain.UploadTypeProjectUpdate,
				UploadMetadata: domain.TusUploadInitRequest{NamaProject: "New", Kategori: "iot", Semester: 6},
				FileSize:       4,
				Status:         domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()
			tusRepo.On("Complete", uploadID, mock.AnythingOfType("uint"), mock.AnythingOfType("string")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1", "project_id": "2"})
			offset, err := uc.HandleProjectUpdateChunk(projectID, uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
		})

		t.Run("upload not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			_, err := uc.HandleProjectUpdateChunk(1, "missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "owner", ProjectID: &projectID, FileSize: 8, Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleProjectUpdateChunk(projectID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", ProjectID: &projectID, FileSize: 8, Status: domain.UploadStatusCompleted}, nil).Once()

			offset, err := uc.HandleProjectUpdateChunk(projectID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Equal(t, int64(8), offset)
			assert.Contains(t, err.Error(), "sudah selesai")
		})

		t.Run("not active", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", ProjectID: &projectID, FileSize: 8, Status: domain.UploadStatusCancelled}, nil).Once()

			_, err := uc.HandleProjectUpdateChunk(projectID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak aktif")
		})
	})

	t.Run("CancelProjectUpdateUpload", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			projectID := uint(10)
			seedTusUploadStore(t, manager, "cancel-proj", 8, map[string]string{"user_id": "u1", "project_id": fmt.Sprintf("%d", projectID)})
			manager.AddToQueue("cancel-proj")

			tusRepo.On("GetByID", "cancel-proj").Return(&domain.TusUpload{ID: "cancel-proj", UserID: "u1", ProjectID: &projectID, Status: domain.UploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", "cancel-proj", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelProjectUpdateUpload(projectID, "cancel-proj", "u1")
			require.NoError(t, err)
		})

		t.Run("not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelProjectUpdateUpload(1, "missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "owner", ProjectID: &projectID, Status: domain.UploadStatusUploading}, nil).Once()

			err := uc.CancelProjectUpdateUpload(projectID, "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", ProjectID: &projectID, Status: domain.UploadStatusCompleted}, nil).Once()

			err := uc.CancelProjectUpdateUpload(projectID, "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("Project update wrapper methods", func(t *testing.T) {
		t.Run("GetProjectUpdateUploadInfo found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(11)
			now := time.Now()
			tusRepo.On("GetByID", "project-info-id").Return(&domain.TusUpload{
				ID:             "project-info-id",
				UserID:         "u1",
				ProjectID:      &projectID,
				UploadMetadata: domain.TusUploadInitRequest{NamaProject: "P", Kategori: "iot", Semester: 4},
				Status:         domain.UploadStatusUploading,
				Progress:       60,
				CurrentOffset:  6,
				FileSize:       10,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetProjectUpdateUploadInfo(projectID, "project-info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, uint(11), info.ProjectID)
			assert.Equal(t, int64(6), info.Offset)
		})

		t.Run("GetProjectUpdateUploadInfo not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetProjectUpdateUploadInfo(11, "missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetProjectUpdateUploadStatus found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(12)
			tusRepo.On("GetByID", "project-status-id").Return(&domain.TusUpload{
				ID:            "project-status-id",
				UserID:        "u1",
				ProjectID:     &projectID,
				CurrentOffset: 9,
				FileSize:      18,
			}, nil).Once()

			offset, length, err := uc.GetProjectUpdateUploadStatus(projectID, "project-status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(9), offset)
			assert.Equal(t, int64(18), length)
		})

		t.Run("GetProjectUpdateUploadStatus not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetProjectUpdateUploadStatus(12, "missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})
	})
}

func TestUsecaseTestMocksHelpers(t *testing.T) {
	t.Run("uintPtr returns pointer value", func(t *testing.T) {
		ptr := uintPtr(42)
		require.NotNil(t, ptr)
		assert.Equal(t, uint(42), *ptr)
	})

	t.Run("MockTusUploadRepository uncovered list methods", func(t *testing.T) {
		repo := new(MockTusUploadRepository)
		uploads := []domain.TusUpload{{ID: "u1"}}

		repo.On("GetByUserID", "user-1").Return(uploads, nil).Once()
		items, err := repo.GetByUserID("user-1")
		require.NoError(t, err)
		assert.Len(t, items, 1)

		repo.On("GetByUserIDAndStatus", "user-1", domain.UploadStatusUploading).Return(uploads, nil).Once()
		itemsByStatus, err := repo.GetByUserIDAndStatus("user-1", domain.UploadStatusUploading)
		require.NoError(t, err)
		assert.Len(t, itemsByStatus, 1)

		repo.On("GetActiveByUserID", "user-1").Return(uploads, nil).Once()
		active, err := repo.GetActiveByUserID("user-1")
		require.NoError(t, err)
		assert.Len(t, active, 1)

		repo.On("ListActive").Return(uploads, nil).Once()
		listed, err := repo.ListActive()
		require.NoError(t, err)
		assert.Len(t, listed, 1)

		repo.On("GetActiveUploadIDs").Return([]string{"u1"}, nil).Once()
		ids, err := repo.GetActiveUploadIDs()
		require.NoError(t, err)
		assert.Equal(t, []string{"u1"}, ids)

		repo.AssertExpectations(t)
	})

	t.Run("MockTusUploadRepository nil list branches", func(t *testing.T) {
		repo := new(MockTusUploadRepository)

		repo.On("GetByUserID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()
		items, err := repo.GetByUserID("missing")
		require.Error(t, err)
		assert.Nil(t, items)

		repo.On("GetActiveByUserID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()
		active, err := repo.GetActiveByUserID("missing")
		require.Error(t, err)
		assert.Nil(t, active)

		repo.On("ListActive").Return(nil, gorm.ErrRecordNotFound).Once()
		listed, err := repo.ListActive()
		require.Error(t, err)
		assert.Nil(t, listed)

		repo.On("GetActiveUploadIDs").Return(nil, gorm.ErrRecordNotFound).Once()
		ids, err := repo.GetActiveUploadIDs()
		require.Error(t, err)
		assert.Nil(t, ids)

		repo.AssertExpectations(t)
	})
}
