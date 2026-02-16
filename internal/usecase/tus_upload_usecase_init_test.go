package usecase

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"invento-service/internal/domain"
	"invento-service/internal/storage"
	"invento-service/internal/upload"

	dto "invento-service/internal/dto"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func newTusUploadTestDeps(t *testing.T) (*tusUploadUsecase, *MockTusUploadRepository, *MockProjectRepository, *upload.TusManager) {
	t.Helper()

	mockTusUploadRepo := new(MockTusUploadRepository)
	mockProjectRepo := new(MockProjectRepository)

	cfg := getTestTusUploadConfig()
	baseDir := t.TempDir()
	cfg.App.Env = "development"
	cfg.Upload.PathDevelopment = filepath.Join(baseDir, "uploads")
	cfg.Upload.TempPathDevelopment = filepath.Join(baseDir, "temp")

	pathResolver := storage.NewPathResolver(cfg)
	tusStore := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	tusQueue := upload.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	tusManager := upload.NewTusManager(tusStore, tusQueue, nil, cfg, zerolog.Nop())
	fileManager := storage.NewFileManager(cfg)

	uc := NewTusUploadUsecase(mockTusUploadRepo, mockProjectRepo, nil, tusManager, fileManager, cfg).(*tusUploadUsecase)

	return uc, mockTusUploadRepo, mockProjectRepo, tusManager
}

func seedTusUploadStore(t *testing.T, manager *upload.TusManager, uploadID string, size int64, metadata map[string]string) {
	t.Helper()
	require.NoError(t, manager.InitiateUpload(uploadID, size, metadata))
}

func TestTusUploadUsecase_InitAndStatus(t *testing.T) {
	t.Parallel()
	t.Run("CheckUploadSlot available and not available", func(t *testing.T) {
		t.Parallel()
		t.Run("available", func(t *testing.T) {
			t.Parallel()
			uc, _, _, _ := newTusUploadTestDeps(t)

			res, err := uc.CheckUploadSlot(context.Background(), "u1")
			require.NoError(t, err)
			assert.True(t, res.Available)
			assert.False(t, res.ActiveUpload)
			assert.Equal(t, 0, res.QueueLength)
		})

		t.Run("not available", func(t *testing.T) {
			t.Parallel()
			uc, _, _, manager := newTusUploadTestDeps(t)
			manager.AddToQueue("busy-upload")

			res, err := uc.CheckUploadSlot(context.Background(), "u1")
			require.NoError(t, err)
			assert.False(t, res.Available)
			assert.True(t, res.ActiveUpload)
		})
	})

	t.Run("ResetUploadQueue clears active upload", func(t *testing.T) {
		t.Parallel()
		uc, tusRepo, _, manager := newTusUploadTestDeps(t)
		seedTusUploadStore(t, manager, "active-id", 16, map[string]string{"user_id": "u1"})
		manager.AddToQueue("active-id")

		resBefore, err := uc.CheckUploadSlot(context.Background(), "u1")
		require.NoError(t, err)
		assert.False(t, resBefore.Available)

		tusRepo.On("GetActiveByUserID", mock.Anything, "u1").Return([]domain.TusUpload{
			{ID: "active-id", UserID: "u1", Status: domain.UploadStatusUploading},
		}, nil).Once()
		tusRepo.On("GetByID", mock.Anything, "active-id").Return(&domain.TusUpload{ID: "active-id", UserID: "u1", Status: domain.UploadStatusUploading}, nil).Once()
		tusRepo.On("UpdateStatus", mock.Anything, "active-id", domain.UploadStatusCancelled).Return(nil).Once()

		require.NoError(t, uc.ResetUploadQueue(context.Background(), "u1"))

		resAfter, err := uc.CheckUploadSlot(context.Background(), "u1")
		require.NoError(t, err)
		assert.True(t, resAfter.Available)
		assert.False(t, resAfter.ActiveUpload)
		tusRepo.AssertExpectations(t)
	})

	t.Run("InitiateUpload", func(t *testing.T) {
		t.Parallel()
		metadata := dto.TusUploadInitRequest{NamaProject: "Project Alpha", Kategori: "website", Semester: 2}

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			tusRepo.On("GetActiveByUserID", mock.Anything, "u1").Return([]domain.TusUpload{}, nil).Once()
			tusRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusUpload")).Return(nil).Once()

			res, err := uc.InitiateUpload(context.Background(), "u1", "u1@mail.com", "mahasiswa", 1024, metadata)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.NotEmpty(t, res.UploadID)
			assert.Equal(t, int64(1024), res.Length)

			_, _, err = manager.GetUploadStatus(res.UploadID)
			assert.NoError(t, err)
			tusRepo.AssertExpectations(t)
		})

		t.Run("file too large", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetActiveByUserID", mock.Anything, "u1").Return([]domain.TusUpload{}, nil)

			res, err := uc.InitiateUpload(context.Background(), "u1", "u1@mail.com", "mahasiswa", uc.config.Upload.MaxSizeProject+1, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
			tusRepo.AssertNotCalled(t, "Create", mock.Anything)
		})

		t.Run("no upload slot", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			tusRepo.On("GetActiveByUserID", mock.Anything, "u1").Return([]domain.TusUpload{}, nil)
			manager.AddToQueue("active")

			res, err := uc.InitiateUpload(context.Background(), "u1", "u1@mail.com", "mahasiswa", 256, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "slot upload tidak tersedia")
			tusRepo.AssertNotCalled(t, "Create", mock.Anything)
		})
	})

	t.Run("GetUploadInfo and GetUploadStatus", func(t *testing.T) {
		t.Parallel()
		t.Run("GetUploadInfo found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(7)
			now := time.Now()
			tusRepo.On("GetByID", mock.Anything, "info-id").Return(&domain.TusUpload{
				ID:             "info-id",
				UserID:         "u1",
				ProjectID:      &projectID,
				UploadMetadata: domain.TusUploadMetadata{NamaProject: "P", Kategori: "iot", Semester: 4},
				Status:         domain.UploadStatusUploading,
				Progress:       50,
				CurrentOffset:  5,
				FileSize:       10,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetUploadInfo(context.Background(), "info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, uint(7), info.ProjectID)
			assert.Equal(t, int64(5), info.Offset)
		})

		t.Run("GetUploadInfo not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetUploadInfo(context.Background(), "missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetUploadStatus found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "status-id").Return(&domain.TusUpload{ID: "status-id", UserID: "u1", CurrentOffset: 8, FileSize: 16}, nil).Once()

			offset, length, err := uc.GetUploadStatus(context.Background(), "status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(8), offset)
			assert.Equal(t, int64(16), length)
		})

		t.Run("GetUploadStatus not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetUploadStatus(context.Background(), "missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})
	})
}

func TestUsecaseTestMocksHelpers(t *testing.T) {
	t.Parallel()
	t.Run("uintPtr returns pointer value", func(t *testing.T) {
		t.Parallel()
		ptr := uintPtr(42)
		require.NotNil(t, ptr)
		assert.Equal(t, uint(42), *ptr)
	})

	t.Run("MockTusUploadRepository uncovered list methods", func(t *testing.T) {
		t.Parallel()
		repo := new(MockTusUploadRepository)
		uploads := []domain.TusUpload{{ID: "u1"}}
		ctx := context.Background()

		repo.On("GetByUserID", mock.Anything, "user-1").Return(uploads, nil).Once()
		items, err := repo.GetByUserID(ctx, "user-1")
		require.NoError(t, err)
		assert.Len(t, items, 1)

		repo.On("GetActiveByUserID", mock.Anything, "user-1").Return(uploads, nil).Once()
		active, err := repo.GetActiveByUserID(ctx, "user-1")
		require.NoError(t, err)
		assert.Len(t, active, 1)

		repo.On("ListActive", mock.Anything).Return(uploads, nil).Once()
		listed, err := repo.ListActive(ctx)
		require.NoError(t, err)
		assert.Len(t, listed, 1)

		repo.On("GetActiveUploadIDs", mock.Anything).Return([]string{"u1"}, nil).Once()
		ids, err := repo.GetActiveUploadIDs(ctx)
		require.NoError(t, err)
		assert.Equal(t, []string{"u1"}, ids)

		repo.AssertExpectations(t)
	})

	t.Run("MockTusUploadRepository nil list branches", func(t *testing.T) {
		t.Parallel()
		repo := new(MockTusUploadRepository)
		ctx := context.Background()

		repo.On("GetByUserID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()
		items, err := repo.GetByUserID(ctx, "missing")
		require.Error(t, err)
		assert.Nil(t, items)

		repo.On("GetActiveByUserID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()
		active, err := repo.GetActiveByUserID(ctx, "missing")
		require.Error(t, err)
		assert.Nil(t, active)

		repo.On("ListActive", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
		listed, err := repo.ListActive(ctx)
		require.Error(t, err)
		assert.Nil(t, listed)

		repo.On("GetActiveUploadIDs", mock.Anything).Return(nil, gorm.ErrRecordNotFound).Once()
		ids, err := repo.GetActiveUploadIDs(ctx)
		require.Error(t, err)
		assert.Nil(t, ids)

		repo.AssertExpectations(t)
	})
}
