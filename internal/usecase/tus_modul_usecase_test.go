package usecase

import (
	"bytes"
	"encoding/base64"
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

func b64(v string) string {
	return base64.StdEncoding.EncodeToString([]byte(v))
}

func modulMetadataHeader(judul, deskripsi string) string {
	return fmt.Sprintf("judul %s,deskripsi %s", b64(judul), b64(deskripsi))
}

func newTusModulTestDeps(t *testing.T) (*tusModulUsecase, *MockTusModulUploadRepository, *MockModulRepository, *helper.TusManager) {
	t.Helper()

	mockTusModulRepo := new(MockTusModulUploadRepository)
	mockModulRepo := new(MockModulRepository)

	cfg := getTestModulConfig()
	baseDir := t.TempDir()
	cfg.App.Env = "development"
	cfg.Upload.PathDevelopment = filepath.Join(baseDir, "uploads")
	cfg.Upload.TempPathDevelopment = filepath.Join(baseDir, "temp")

	pathResolver := helper.NewPathResolver(cfg)
	tusStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	tusQueue := helper.NewTusQueue(5)
	tusManager := helper.NewTusManager(tusStore, tusQueue, nil, cfg)
	fileManager := helper.NewFileManager(cfg)

	uc := NewTusModulUsecase(mockTusModulRepo, mockModulRepo, tusManager, fileManager, cfg).(*tusModulUsecase)
	return uc, mockTusModulRepo, mockModulRepo, tusManager
}

func seedTusModulStore(t *testing.T, manager *helper.TusManager, uploadID string, size int64, metadata map[string]string) {
	t.Helper()
	require.NoError(t, manager.InitiateUpload(uploadID, size, metadata))
}

func TestTusModulUsecase(t *testing.T) {
	t.Run("CheckModulUploadSlot available and exceeded", func(t *testing.T) {
		t.Run("available", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(2), nil).Once()

			res, err := uc.CheckModulUploadSlot("u1")
			require.NoError(t, err)
			assert.True(t, res.Available)
			assert.Equal(t, 2, res.QueueLength)
		})

		t.Run("exceeded", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(5), nil).Once()

			res, err := uc.CheckModulUploadSlot("u1")
			require.NoError(t, err)
			assert.False(t, res.Available)
			assert.Contains(t, res.Message, "Antrian penuh")
		})
	})

	t.Run("InitiateModulUpload", func(t *testing.T) {
		validMeta := modulMetadataHeader("modul-a", "deskripsi modul")

		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(0), nil).Once()
			tusRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil).Once()

			res, err := uc.InitiateModulUpload("u1", 1024, validMeta)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.NotEmpty(t, res.UploadID)

			_, _, err = manager.GetUploadStatus(res.UploadID)
			assert.NoError(t, err)
		})

		t.Run("no slot available", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(5), nil).Once()

			res, err := uc.InitiateModulUpload("u1", 1024, validMeta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "antrian penuh")
		})

		t.Run("invalid file size", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(0), nil).Once()

			res, err := uc.InitiateModulUpload("u1", uc.config.Upload.MaxSizeModul+1, validMeta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
		})

		t.Run("invalid metadata bad base64", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(0), nil).Once()

			res, err := uc.InitiateModulUpload("u1", 100, "judul !!!,deskripsi dGVzdA==")
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "judul wajib diisi")
		})

		t.Run("invalid metadata missing fields", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(0), nil).Once()

			res, err := uc.InitiateModulUpload("u1", 100, fmt.Sprintf("deskripsi %s", b64("abc")))
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "judul wajib diisi")
		})
	})

	t.Run("HandleModulChunk", func(t *testing.T) {
		t.Run("happy path + status transition pending to uploading", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			uploadID := "modul-chunk"
			chunk := []byte("abcd")

			tusRepo.On("GetByID", uploadID).Return(&domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				UploadType:     domain.UploadTypeModulCreate,
				UploadMetadata: domain.TusModulUploadInitRequest{Judul: "f", Deskripsi: "desc"},
				FileSize:       10,
				CurrentOffset:  0,
				Status:         domain.UploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", uploadID, domain.UploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})
			offset, err := uc.HandleModulChunk(uploadID, "u1", 0, bytes.NewReader(chunk))
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			uc, tusRepo, modulRepo, manager := newTusModulTestDeps(t)
			uploadID := "modul-complete"
			fileSize := int64(4)

			uploadObj := &domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				UploadType:     domain.UploadTypeModulCreate,
				UploadMetadata: domain.TusModulUploadInitRequest{Judul: "filex", Deskripsi: "desc"},
				FileSize:       fileSize,
				CurrentOffset:  0,
				Status:         domain.UploadStatusUploading,
			}

			tusRepo.On("GetByID", uploadID).Return(uploadObj, nil).Twice()
			tusRepo.On("UpdateOffset", uploadID, fileSize, mock.AnythingOfType("float64")).Return(nil).Once()
			modulRepo.On("Create", mock.AnythingOfType("*domain.Modul")).Run(func(args mock.Arguments) {
				m := args.Get(0).(*domain.Modul)
				m.ID = "550e8400-e29b-41d4-a716-446655440077"
			}).Return(nil).Once()
			tusRepo.On("Complete", uploadID, "550e8400-e29b-41d4-a716-446655440077", mock.MatchedBy(func(path string) bool { return path != "" })).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, fileSize, map[string]string{"user_id": "u1"})
			offset, err := uc.HandleModulChunk(uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, fileSize, offset)
		})

		t.Run("not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			_, err := uc.HandleModulChunk("missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleModulChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed becomes inactive error", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			_, err := uc.HandleModulChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("GetModulUploadInfo and GetModulUploadStatus", func(t *testing.T) {
		t.Run("GetModulUploadInfo found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440005"
			now := time.Now()
			tusRepo.On("GetByID", "info-id").Return(&domain.TusModulUpload{
				ID:             "info-id",
				UserID:         "u1",
				ModulID:        &modulID,
				UploadMetadata: domain.TusModulUploadInitRequest{Judul: "f", Deskripsi: "desc"},
				Status:         domain.UploadStatusUploading,
				Progress:       33,
				CurrentOffset:  3,
				FileSize:       9,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetModulUploadInfo("info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, "550e8400-e29b-41d4-a716-446655440005", info.ModulID)
			assert.Equal(t, int64(3), info.Offset)
		})

		t.Run("GetModulUploadInfo not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetModulUploadInfo("missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetModulUploadStatus found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "status-id").Return(&domain.TusModulUpload{ID: "status-id", UserID: "u1", CurrentOffset: 4, FileSize: 12}, nil).Once()

			offset, length, err := uc.GetModulUploadStatus("status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
			assert.Equal(t, int64(12), length)
		})

		t.Run("GetModulUploadStatus not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetModulUploadStatus("missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})
	})

	t.Run("CancelModulUpload", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			seedTusModulStore(t, manager, "cancel-id", 8, map[string]string{"user_id": "u1"})

			tusRepo.On("GetByID", "cancel-id").Return(&domain.TusModulUpload{ID: "cancel-id", UserID: "u1", Status: domain.UploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", "cancel-id", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelModulUpload("cancel-id", "u1")
			require.NoError(t, err)
		})

		t.Run("not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelModulUpload("missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			err := uc.CancelModulUpload("id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			err := uc.CancelModulUpload("id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("InitiateModulUpdateUpload", func(t *testing.T) {
		meta := modulMetadataHeader("new-name", "new deskripsi")
		modulID := "550e8400-e29b-41d4-a716-446655440008"

		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", modulID).Return(&domain.Modul{ID: modulID, UserID: "u1"}, nil).Once()
			tusRepo.On("CountActiveByUserID", "u1").Return(int64(0), nil).Once()
			tusRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil).Once()

			res, err := uc.InitiateModulUpdateUpload(modulID, "u1", 2048, meta)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/modul/"+modulID+"/update/")
		})

		t.Run("modul not found", func(t *testing.T) {
			uc, _, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", modulID).Return(nil, gorm.ErrRecordNotFound).Once()

			res, err := uc.InitiateModulUpdateUpload(modulID, "u1", 2048, meta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "Modul tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, _, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", modulID).Return(&domain.Modul{ID: modulID, UserID: "owner"}, nil).Once()

			res, err := uc.InitiateModulUpdateUpload(modulID, "u1", 2048, meta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})
	})

	t.Run("HandleModulUpdateChunk", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440009"
			uploadID := "update-chunk"

			tusRepo.On("GetByID", uploadID).Return(&domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				ModulID:        &modulID,
				UploadType:     domain.UploadTypeModulUpdate,
				UploadMetadata: domain.TusModulUploadInitRequest{Judul: "n", Deskripsi: "desc"},
				FileSize:       8,
				CurrentOffset:  0,
				Status:         domain.UploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", uploadID, domain.UploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 8, map[string]string{"user_id": "u1", "modul_id": modulID})
			offset, err := uc.HandleModulUpdateChunk(uploadID, "u1", 0, bytes.NewReader([]byte("abc")))
			require.NoError(t, err)
			assert.Equal(t, int64(3), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			uc, tusRepo, modulRepo, manager := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440011"
			uploadID := "update-complete"

			uploadObj := &domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				ModulID:        &modulID,
				UploadType:     domain.UploadTypeModulUpdate,
				UploadMetadata: domain.TusModulUploadInitRequest{Judul: "rev", Deskripsi: "new desc"},
				FileSize:       4,
				CurrentOffset:  0,
				Status:         domain.UploadStatusUploading,
			}

			tusRepo.On("GetByID", uploadID).Return(uploadObj, nil).Twice()
			tusRepo.On("UpdateOffset", uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()
			modulRepo.On("GetByID", modulID).Return(&domain.Modul{ID: modulID, UserID: "u1", FilePath: ""}, nil).Once()
			modulRepo.On("Update", mock.AnythingOfType("*domain.Modul")).Return(nil).Once()
			tusRepo.On("Complete", uploadID, modulID, mock.MatchedBy(func(path string) bool { return path != "" })).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1", "modul_id": modulID})
			offset, err := uc.HandleModulUpdateChunk(uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
		})

		t.Run("not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			_, err := uc.HandleModulUpdateChunk("missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleModulUpdateChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed becomes inactive error", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			_, err := uc.HandleModulUpdateChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("Modul update wrapper methods", func(t *testing.T) {
		t.Run("GetModulUpdateUploadInfo found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440031"
			now := time.Now()
			tusRepo.On("GetByID", "update-info-id").Return(&domain.TusModulUpload{
				ID:             "update-info-id",
				UserID:         "u1",
				ModulID:        &modulID,
				UploadMetadata: domain.TusModulUploadInitRequest{Judul: "judul", Deskripsi: "deskripsi"},
				Status:         domain.UploadStatusUploading,
				Progress:       44,
				CurrentOffset:  4,
				FileSize:       9,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetModulUpdateUploadInfo(modulID, "update-info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, modulID, info.ModulID)
			assert.Equal(t, int64(4), info.Offset)
		})

		t.Run("GetModulUpdateUploadInfo not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetModulUpdateUploadInfo("550e8400-e29b-41d4-a716-446655440031", "missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetModulUpdateUploadStatus found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440032"
			tusRepo.On("GetByID", "update-status-id").Return(&domain.TusModulUpload{
				ID:            "update-status-id",
				UserID:        "u1",
				ModulID:       &modulID,
				CurrentOffset: 7,
				FileSize:      14,
			}, nil).Once()

			offset, length, err := uc.GetModulUpdateUploadStatus(modulID, "update-status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(7), offset)
			assert.Equal(t, int64(14), length)
		})

		t.Run("GetModulUpdateUploadStatus not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetModulUpdateUploadStatus("550e8400-e29b-41d4-a716-446655440032", "missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})

		t.Run("CancelModulUpdateUpload success", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440033"
			seedTusModulStore(t, manager, "cancel-update-id", 8, map[string]string{"user_id": "u1", "modul_id": modulID})

			tusRepo.On("GetByID", "cancel-update-id").Return(&domain.TusModulUpload{
				ID:      "cancel-update-id",
				UserID:  "u1",
				ModulID: &modulID,
				Status:  domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateStatus", "cancel-update-id", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelModulUpdateUpload(modulID, "cancel-update-id", "u1")
			require.NoError(t, err)
		})

		t.Run("CancelModulUpdateUpload not found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelModulUpdateUpload("550e8400-e29b-41d4-a716-446655440033", "missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})
	})
}
