package usecase

import (
	"context"
	"encoding/base64"
	"fmt"
	"invento-service/internal/domain"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"path/filepath"
	"testing"
	"time"

	apperrors "invento-service/internal/errors"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func b64(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

func modulMetadataHeader(judul, deskripsi string) string {
	return fmt.Sprintf("judul %s,deskripsi %s", b64(judul), b64(deskripsi))
}

func newTusModulTestDeps(t *testing.T) (*tusModulUsecase, *MockTusModulUploadRepository, *MockModulRepository, *upload.TusManager) {
	t.Helper()

	mockTusModulRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()
	baseDir := t.TempDir()
	cfg.App.Env = "development"
	cfg.Upload.PathDevelopment = filepath.Join(baseDir, "uploads")
	cfg.Upload.TempPathDevelopment = filepath.Join(baseDir, "temp")

	pathResolver := storage.NewPathResolver(cfg)
	tusStore := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	tusQueue := upload.NewTusQueue(5)
	tusManager := upload.NewTusManager(tusStore, tusQueue, nil, cfg, zerolog.Nop())
	fileManager := storage.NewFileManager(cfg)

	uc := NewTusModulUsecase(mockTusModulRepo, mockModulRepo, tusManager, fileManager, cfg).(*tusModulUsecase)
	return uc, mockTusModulRepo, mockModulRepo, tusManager
}

func seedTusModulStore(t *testing.T, manager *upload.TusManager, uploadID string, size int64, metadata map[string]string) {
	t.Helper()
	require.NoError(t, manager.InitiateUpload(uploadID, size, metadata))
}

func TestTusModulUsecase_InitAndStatus(t *testing.T) {
	t.Parallel()
	t.Run("CheckModulUploadSlot available and exceeded", func(t *testing.T) {
		t.Parallel()
		t.Run("available", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(2), nil).Once()

			res, err := uc.CheckModulUploadSlot(context.Background(), "u1")
			require.NoError(t, err)
			assert.True(t, res.Available)
			assert.Equal(t, 2, res.QueueLength)
		})

		t.Run("exceeded", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(5), nil).Once()

			res, err := uc.CheckModulUploadSlot(context.Background(), "u1")
			require.NoError(t, err)
			assert.False(t, res.Available)
			assert.Contains(t, res.Message, "Antrian penuh")
		})
	})

	t.Run("InitiateModulUpload", func(t *testing.T) {
		t.Parallel()
		validMeta := modulMetadataHeader("modul-a", "deskripsi modul")

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), nil).Once()
			tusRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusModulUpload")).Return(nil).Once()

			res, err := uc.InitiateModulUpload(context.Background(), "u1", 1024, validMeta)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.NotEmpty(t, res.UploadID)

			_, _, err = manager.GetUploadStatus(res.UploadID)
			assert.NoError(t, err)
		})

		t.Run("no slot available", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(5), nil).Once()

			res, err := uc.InitiateModulUpload(context.Background(), "u1", 1024, validMeta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "antrian penuh")
		})

		t.Run("invalid file size", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), nil).Once()

			res, err := uc.InitiateModulUpload(context.Background(), "u1", uc.config.Upload.MaxSizeModul+1, validMeta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
		})

		t.Run("invalid metadata bad base64", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), nil).Once()

			res, err := uc.InitiateModulUpload(context.Background(), "u1", 100, "judul !!!,deskripsi dGVzdA==")
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "judul wajib diisi")
		})

		t.Run("invalid metadata missing fields", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), nil).Once()

			res, err := uc.InitiateModulUpload(context.Background(), "u1", 100, fmt.Sprintf("deskripsi %s", b64("abc")))
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "judul wajib diisi")
		})
	})

	t.Run("GetModulUploadInfo and GetModulUploadStatus", func(t *testing.T) {
		t.Parallel()
		t.Run("GetModulUploadInfo found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440005"
			now := time.Now()
			tusRepo.On("GetByID", mock.Anything, "info-id").Return(&domain.TusModulUpload{
				ID:             "info-id",
				UserID:         "u1",
				ModulID:        &modulID,
				UploadMetadata: domain.TusModulUploadMetadata{Judul: "f", Deskripsi: "desc"},
				Status:         domain.UploadStatusUploading,
				Progress:       33,
				CurrentOffset:  3,
				FileSize:       9,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetModulUploadInfo(context.Background(), "info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, "550e8400-e29b-41d4-a716-446655440005", info.ModulID)
			assert.Equal(t, int64(3), info.Offset)
		})

		t.Run("GetModulUploadInfo not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetModulUploadInfo(context.Background(), "missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetModulUploadStatus found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "status-id").Return(&domain.TusModulUpload{ID: "status-id", UserID: "u1", CurrentOffset: 4, FileSize: 12}, nil).Once()

			offset, length, err := uc.GetModulUploadStatus(context.Background(), "status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
			assert.Equal(t, int64(12), length)
		})

		t.Run("GetModulUploadStatus not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetModulUploadStatus(context.Background(), "missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})
	})

	t.Run("CancelModulUpload", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			seedTusModulStore(t, manager, "cancel-id", 8, map[string]string{"user_id": "u1"})

			tusRepo.On("GetByID", mock.Anything, "cancel-id").Return(&domain.TusModulUpload{ID: "cancel-id", UserID: "u1", Status: domain.UploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, "cancel-id", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelModulUpload(context.Background(), "cancel-id", "u1")
			require.NoError(t, err)
		})

		t.Run("not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelModulUpload(context.Background(), "missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			err := uc.CancelModulUpload(context.Background(), "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			err := uc.CancelModulUpload(context.Background(), "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("InitiateModulUpdateUpload", func(t *testing.T) {
		t.Parallel()
		meta := modulMetadataHeader("new-name", "new deskripsi")
		modulID := "550e8400-e29b-41d4-a716-446655440008"

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", mock.Anything, modulID).Return(&domain.Modul{ID: modulID, UserID: "u1"}, nil).Once()
			tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), nil).Once()
			tusRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusModulUpload")).Return(nil).Once()

			res, err := uc.InitiateModulUpdateUpload(context.Background(), modulID, "u1", 2048, meta)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/modul/"+modulID+"/update/")
		})

		t.Run("modul not found", func(t *testing.T) {
			t.Parallel()
			uc, _, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", mock.Anything, modulID).Return(nil, apperrors.ErrRecordNotFound).Once()

			res, err := uc.InitiateModulUpdateUpload(context.Background(), modulID, "u1", 2048, meta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "Modul tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, _, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", mock.Anything, modulID).Return(&domain.Modul{ID: modulID, UserID: "owner"}, nil).Once()

			res, err := uc.InitiateModulUpdateUpload(context.Background(), modulID, "u1", 2048, meta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})
	})
}

func TestTusModulUsecase_InitiateModulUpload_RepoCreateError(t *testing.T) {
	t.Parallel()
	uc, tusRepo, _, _ := newTusModulTestDeps(t)
	meta := modulMetadataHeader("modul-a", "deskripsi")

	tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), nil).Once()
	tusRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusModulUpload")).Return(assert.AnError).Once()

	res, err := uc.InitiateModulUpload(context.Background(), "u1", 1024, meta)
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "create record")

	tusRepo.AssertExpectations(t)
}

func TestTusModulUsecase_CheckModulUploadSlot_RepoError(t *testing.T) {
	t.Parallel()
	uc, tusRepo, _, _ := newTusModulTestDeps(t)

	tusRepo.On("CountActiveByUserID", mock.Anything, "u1").Return(int64(0), assert.AnError).Once()

	res, err := uc.CheckModulUploadSlot(context.Background(), "u1")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Contains(t, err.Error(), "TusModulUsecase.CheckModulUploadSlot")

	tusRepo.AssertExpectations(t)
}
