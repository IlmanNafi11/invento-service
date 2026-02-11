package usecase

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"path/filepath"
	"strconv"
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

func modulMetadataHeader(name, tipe string, semester int) string {
	return fmt.Sprintf("nama_file %s,tipe %s,semester %s", b64(name), b64(tipe), b64(strconv.Itoa(semester)))
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
			tusRepo.On("CountActiveByUserID", "u1").Return(2, nil).Once()

			res, err := uc.CheckModulUploadSlot("u1")
			require.NoError(t, err)
			assert.True(t, res.Available)
			assert.Equal(t, 2, res.QueueLength)
		})

		t.Run("exceeded", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(5, nil).Once()

			res, err := uc.CheckModulUploadSlot("u1")
			require.NoError(t, err)
			assert.False(t, res.Available)
			assert.Contains(t, res.Message, "Antrian penuh")
		})
	})

	t.Run("InitiateModulUpload", func(t *testing.T) {
		validMeta := modulMetadataHeader("modul-a", "pdf", 1)

		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(0, nil).Once()
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
			tusRepo.On("CountActiveByUserID", "u1").Return(5, nil).Once()

			res, err := uc.InitiateModulUpload("u1", 1024, validMeta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "antrian penuh")
		})

		t.Run("invalid file size", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(0, nil).Once()

			res, err := uc.InitiateModulUpload("u1", uc.config.Upload.MaxSizeModul+1, validMeta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
		})

		t.Run("invalid metadata bad base64", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(0, nil).Once()

			res, err := uc.InitiateModulUpload("u1", 100, "nama_file !!!,tipe cGRm,semester MQ==")
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "nama_file wajib diisi")
		})

		t.Run("invalid metadata missing fields", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("CountActiveByUserID", "u1").Return(0, nil).Once()

			res, err := uc.InitiateModulUpload("u1", 100, fmt.Sprintf("nama_file %s", b64("abc")))
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "tipe file wajib diisi")
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
				UploadType:     domain.ModulUploadTypeCreate,
				UploadMetadata: domain.TusModulUploadInitRequest{NamaFile: "f", Tipe: "pdf", Semester: 1},
				FileSize:       10,
				CurrentOffset:  0,
				Status:         domain.ModulUploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", uploadID, domain.ModulUploadStatusUploading).Return(nil).Once()
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
				UploadType:     domain.ModulUploadTypeCreate,
				UploadMetadata: domain.TusModulUploadInitRequest{NamaFile: "filex", Tipe: "pdf", Semester: 2},
				FileSize:       fileSize,
				CurrentOffset:  0,
				Status:         domain.ModulUploadStatusUploading,
			}

			tusRepo.On("GetByID", uploadID).Return(uploadObj, nil).Twice()
			tusRepo.On("UpdateOffset", uploadID, fileSize, mock.AnythingOfType("float64")).Return(nil).Once()
			modulRepo.On("Create", mock.AnythingOfType("*domain.Modul")).Run(func(args mock.Arguments) {
				m := args.Get(0).(*domain.Modul)
				m.ID = 77
			}).Return(nil).Once()
			tusRepo.On("Complete", uploadID, uint(77), mock.MatchedBy(func(path string) bool { return path != "" })).Return(nil).Once()

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
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.ModulUploadStatusUploading}, nil).Once()

			_, err := uc.HandleModulChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed becomes inactive error", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.ModulUploadStatusCompleted}, nil).Once()

			_, err := uc.HandleModulChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak aktif")
		})
	})

	t.Run("GetModulUploadInfo and GetModulUploadStatus", func(t *testing.T) {
		t.Run("GetModulUploadInfo found", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := uint(5)
			now := time.Now()
			tusRepo.On("GetByID", "info-id").Return(&domain.TusModulUpload{
				ID:             "info-id",
				UserID:         "u1",
				ModulID:        &modulID,
				UploadMetadata: domain.TusModulUploadInitRequest{NamaFile: "f", Tipe: "pdf", Semester: 3},
				Status:         domain.ModulUploadStatusUploading,
				Progress:       33,
				CurrentOffset:  3,
				FileSize:       9,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetModulUploadInfo("info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, uint(5), info.ModulID)
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

			tusRepo.On("GetByID", "cancel-id").Return(&domain.TusModulUpload{ID: "cancel-id", UserID: "u1", Status: domain.ModulUploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", "cancel-id", domain.ModulUploadStatusCancelled).Return(nil).Once()

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
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.ModulUploadStatusUploading}, nil).Once()

			err := uc.CancelModulUpload("id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.ModulUploadStatusCompleted}, nil).Once()

			err := uc.CancelModulUpload("id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("InitiateModulUpdateUpload", func(t *testing.T) {
		meta := modulMetadataHeader("new-name", "pptx", 4)

		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", uint(8)).Return(&domain.Modul{ID: 8, UserID: "u1"}, nil).Once()
			tusRepo.On("CountActiveByUserID", "u1").Return(0, nil).Once()
			tusRepo.On("Create", mock.AnythingOfType("*domain.TusModulUpload")).Return(nil).Once()

			res, err := uc.InitiateModulUpdateUpload(8, "u1", 2048, meta)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/modul/8/update/")
		})

		t.Run("modul not found", func(t *testing.T) {
			uc, _, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", uint(8)).Return(nil, gorm.ErrRecordNotFound).Once()

			res, err := uc.InitiateModulUpdateUpload(8, "u1", 2048, meta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "Modul tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			uc, _, modulRepo, _ := newTusModulTestDeps(t)
			modulRepo.On("GetByID", uint(8)).Return(&domain.Modul{ID: 8, UserID: "owner"}, nil).Once()

			res, err := uc.InitiateModulUpdateUpload(8, "u1", 2048, meta)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})
	})

	t.Run("HandleModulUpdateChunk", func(t *testing.T) {
		t.Run("happy path", func(t *testing.T) {
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			modulID := uint(9)
			uploadID := "update-chunk"

			tusRepo.On("GetByID", uploadID).Return(&domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				ModulID:        &modulID,
				UploadType:     domain.ModulUploadTypeUpdate,
				UploadMetadata: domain.TusModulUploadInitRequest{NamaFile: "n", Tipe: "pdf", Semester: 2},
				FileSize:       8,
				CurrentOffset:  0,
				Status:         domain.ModulUploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", uploadID, domain.ModulUploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 8, map[string]string{"user_id": "u1", "modul_id": "9"})
			offset, err := uc.HandleModulUpdateChunk(uploadID, "u1", 0, bytes.NewReader([]byte("abc")))
			require.NoError(t, err)
			assert.Equal(t, int64(3), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			uc, tusRepo, modulRepo, manager := newTusModulTestDeps(t)
			modulID := uint(11)
			uploadID := "update-complete"

			uploadObj := &domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				ModulID:        &modulID,
				UploadType:     domain.ModulUploadTypeUpdate,
				UploadMetadata: domain.TusModulUploadInitRequest{NamaFile: "rev", Tipe: "docx", Semester: 7},
				FileSize:       4,
				CurrentOffset:  0,
				Status:         domain.ModulUploadStatusUploading,
			}

			tusRepo.On("GetByID", uploadID).Return(uploadObj, nil).Twice()
			tusRepo.On("UpdateOffset", uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()
			modulRepo.On("GetByID", modulID).Return(&domain.Modul{ID: modulID, UserID: "u1", PathFile: ""}, nil).Once()
			modulRepo.On("Update", mock.AnythingOfType("*domain.Modul")).Return(nil).Once()
			tusRepo.On("Complete", uploadID, modulID, mock.MatchedBy(func(path string) bool { return path != "" })).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1", "modul_id": "11"})
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
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.ModulUploadStatusUploading}, nil).Once()

			_, err := uc.HandleModulUpdateChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed becomes inactive error", func(t *testing.T) {
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.ModulUploadStatusCompleted}, nil).Once()

			_, err := uc.HandleModulUpdateChunk("id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak aktif")
		})
	})
}
