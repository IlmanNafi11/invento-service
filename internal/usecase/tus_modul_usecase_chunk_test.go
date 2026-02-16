package usecase

import (
	"bytes"
	"context"
	"testing"
	"time"

	"invento-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTusModulUsecase_ChunkAndUpdate(t *testing.T) {
	t.Parallel()
	t.Run("HandleModulChunk", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path + status transition pending to uploading", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			uploadID := "modul-chunk"
			chunk := []byte("abcd")

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				UploadType:     domain.UploadTypeModulCreate,
				UploadMetadata: domain.TusModulUploadMetadata{Judul: "f", Deskripsi: "desc"},
				FileSize:       10,
				CurrentOffset:  0,
				Status:         domain.UploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, uploadID, domain.UploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})
			offset, err := uc.HandleModulChunk(context.Background(), uploadID, "u1", 0, bytes.NewReader(chunk))
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, modulRepo, manager := newTusModulTestDeps(t)
			uploadID := "modul-complete"
			fileSize := int64(4)

			uploadObj := &domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				UploadType:     domain.UploadTypeModulCreate,
				UploadMetadata: domain.TusModulUploadMetadata{Judul: "filex", Deskripsi: "desc"},
				FileSize:       fileSize,
				CurrentOffset:  0,
				Status:         domain.UploadStatusUploading,
			}

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(uploadObj, nil).Twice()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, fileSize, mock.AnythingOfType("float64")).Return(nil).Once()
			modulRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Modul")).Run(func(args mock.Arguments) {
				m := args.Get(1).(*domain.Modul)
				m.ID = "550e8400-e29b-41d4-a716-446655440077"
			}).Return(nil).Once()
			tusRepo.On("Complete", mock.Anything, uploadID, "550e8400-e29b-41d4-a716-446655440077", mock.MatchedBy(func(path string) bool { return path != "" })).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, fileSize, map[string]string{"user_id": "u1"})
			offset, err := uc.HandleModulChunk(context.Background(), uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, fileSize, offset)
		})

		t.Run("not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			_, err := uc.HandleModulChunk(context.Background(), "missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleModulChunk(context.Background(), "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed becomes inactive error", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			_, err := uc.HandleModulChunk(context.Background(), "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("HandleModulUpdateChunk", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440009"
			uploadID := "update-chunk"

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				ModulID:        &modulID,
				UploadType:     domain.UploadTypeModulUpdate,
				UploadMetadata: domain.TusModulUploadMetadata{Judul: "n", Deskripsi: "desc"},
				FileSize:       8,
				CurrentOffset:  0,
				Status:         domain.UploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, uploadID, domain.UploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 8, map[string]string{"user_id": "u1", "modul_id": modulID})
			offset, err := uc.HandleModulUpdateChunk(context.Background(), modulID, uploadID, "u1", 0, bytes.NewReader([]byte("abc")))
			require.NoError(t, err)
			assert.Equal(t, int64(3), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, modulRepo, manager := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440011"
			uploadID := "update-complete"

			uploadObj := &domain.TusModulUpload{
				ID:             uploadID,
				UserID:         "u1",
				ModulID:        &modulID,
				UploadType:     domain.UploadTypeModulUpdate,
				UploadMetadata: domain.TusModulUploadMetadata{Judul: "rev", Deskripsi: "new desc"},
				FileSize:       4,
				CurrentOffset:  0,
				Status:         domain.UploadStatusUploading,
			}

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(uploadObj, nil).Twice()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()
			modulRepo.On("GetByID", mock.Anything, modulID).Return(&domain.Modul{ID: modulID, UserID: "u1", FilePath: ""}, nil).Once()
			modulRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Modul")).Return(nil).Once()
			tusRepo.On("Complete", mock.Anything, uploadID, modulID, mock.MatchedBy(func(path string) bool { return path != "" })).Return(nil).Once()

			seedTusModulStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1", "modul_id": modulID})
			offset, err := uc.HandleModulUpdateChunk(context.Background(), modulID, uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
		})

		t.Run("not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			_, err := uc.HandleModulUpdateChunk(context.Background(), "some-modul-id", "missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusModulUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleModulUpdateChunk(context.Background(), "some-modul-id", "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed becomes inactive error", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440099"
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusModulUpload{ID: "id", UserID: "u1", ModulID: &modulID, Status: domain.UploadStatusCompleted}, nil).Once()

			_, err := uc.HandleModulUpdateChunk(context.Background(), modulID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("Modul update wrapper methods", func(t *testing.T) {
		t.Parallel()
		t.Run("GetModulUpdateUploadInfo found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440031"
			now := time.Now()
			tusRepo.On("GetByID", mock.Anything, "update-info-id").Return(&domain.TusModulUpload{
				ID:             "update-info-id",
				UserID:         "u1",
				ModulID:        &modulID,
				UploadMetadata: domain.TusModulUploadMetadata{Judul: "judul", Deskripsi: "deskripsi"},
				Status:         domain.UploadStatusUploading,
				Progress:       44,
				CurrentOffset:  4,
				FileSize:       9,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetModulUpdateUploadInfo(context.Background(), modulID, "update-info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, modulID, info.ModulID)
			assert.Equal(t, int64(4), info.Offset)
		})

		t.Run("GetModulUpdateUploadInfo not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetModulUpdateUploadInfo(context.Background(), "550e8400-e29b-41d4-a716-446655440031", "missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetModulUpdateUploadStatus found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440032"
			tusRepo.On("GetByID", mock.Anything, "update-status-id").Return(&domain.TusModulUpload{
				ID:            "update-status-id",
				UserID:        "u1",
				ModulID:       &modulID,
				CurrentOffset: 7,
				FileSize:      14,
			}, nil).Once()

			offset, length, err := uc.GetModulUpdateUploadStatus(context.Background(), modulID, "update-status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(7), offset)
			assert.Equal(t, int64(14), length)
		})

		t.Run("GetModulUpdateUploadStatus not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetModulUpdateUploadStatus(context.Background(), "550e8400-e29b-41d4-a716-446655440032", "missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})

		t.Run("CancelModulUpdateUpload success", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusModulTestDeps(t)
			modulID := "550e8400-e29b-41d4-a716-446655440033"
			seedTusModulStore(t, manager, "cancel-update-id", 8, map[string]string{"user_id": "u1", "modul_id": modulID})

			tusRepo.On("GetByID", mock.Anything, "cancel-update-id").Return(&domain.TusModulUpload{
				ID:      "cancel-update-id",
				UserID:  "u1",
				ModulID: &modulID,
				Status:  domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, "cancel-update-id", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelModulUpdateUpload(context.Background(), modulID, "cancel-update-id", "u1")
			require.NoError(t, err)
		})

		t.Run("CancelModulUpdateUpload not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusModulTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelModulUpdateUpload(context.Background(), "550e8400-e29b-41d4-a716-446655440033", "missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})
	})
}

func TestTusModulUsecase_HandleModulChunk_OffsetMismatch(t *testing.T) {
	t.Parallel()
	uc, tusRepo, _, manager := newTusModulTestDeps(t)
	uploadID := "offset-mismatch"

	tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusModulUpload{
		ID:            uploadID,
		UserID:        "u1",
		FileSize:      10,
		CurrentOffset: 4,
		Status:        domain.UploadStatusUploading,
	}, nil).Once()

	seedTusModulStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})

	offset, err := uc.HandleModulChunk(context.Background(), uploadID, "u1", 0, bytes.NewReader([]byte("abc")))
	require.Error(t, err)
	assert.Equal(t, int64(4), offset)
	assert.Contains(t, err.Error(), "offset")

	tusRepo.AssertExpectations(t)
}
