package usecase

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"invento-service/internal/domain"
	dto "invento-service/internal/dto"
	apperrors "invento-service/internal/errors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestTusUploadUsecase_ChunkAndCancel(t *testing.T) {
	t.Parallel()
	t.Run("HandleChunk", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path updates offset", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			uploadID := "upload-happy"
			chunk := []byte("abcd")

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusUpload{
				ID:            uploadID,
				UserID:        "u1",
				FileSize:      10,
				CurrentOffset: 0,
				Status:        domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(len(chunk)), mock.MatchedBy(func(progress float64) bool {
				return progress > 0 && progress < 100
			})).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})
			newOffset, err := uc.HandleChunk(context.Background(), uploadID, "u1", 0, bytes.NewReader(chunk))
			require.NoError(t, err)
			assert.Equal(t, int64(4), newOffset)
			tusRepo.AssertExpectations(t)
		})

		t.Run("pending transitions to uploading", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			uploadID := "upload-pending"

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusUpload{
				ID:            uploadID,
				UserID:        "u1",
				FileSize:      10,
				CurrentOffset: 0,
				Status:        domain.UploadStatusPending,
			}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, uploadID, domain.UploadStatusUploading).Return(nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1"})
			newOffset, err := uc.HandleChunk(context.Background(), uploadID, "u1", 0, bytes.NewReader([]byte("abc")))
			require.NoError(t, err)
			assert.Equal(t, int64(3), newOffset)
			tusRepo.AssertExpectations(t)
		})

		t.Run("auto completion when offset reaches file size", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, projectRepo, manager := newTusUploadTestDeps(t)
			uploadID := "upload-complete"

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusUpload{
				ID:             uploadID,
				UserID:         "u1",
				UploadType:     domain.UploadTypeProjectCreate,
				UploadMetadata: domain.TusUploadMetadata{NamaProject: "Final", Kategori: "mobile", Semester: 3},
				FileSize:       4,
				CurrentOffset:  0,
				Status:         domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(4), mock.MatchedBy(func(progress float64) bool { return progress == 100 })).Return(nil).Once()
			projectRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.Project")).Return(nil).Once()
			tusRepo.On("Complete", mock.Anything, uploadID, mock.AnythingOfType("uint"), mock.AnythingOfType("string")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1"})
			newOffset, err := uc.HandleChunk(context.Background(), uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, int64(4), newOffset)
			tusRepo.AssertExpectations(t)
			projectRepo.AssertExpectations(t)
		})

		t.Run("upload not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			newOffset, err := uc.HandleChunk(context.Background(), "missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Equal(t, int64(0), newOffset)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user forbidden", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "u").Return(&domain.TusUpload{ID: "u", UserID: "owner", FileSize: 10, Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleChunk(context.Background(), "u", "intruder", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "u").Return(&domain.TusUpload{ID: "u", UserID: "u1", FileSize: 10, Status: domain.UploadStatusCompleted}, nil).Once()

			offset, err := uc.HandleChunk(context.Background(), "u", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Equal(t, int64(10), offset)
			assert.Contains(t, err.Error(), "sudah selesai")
		})

		t.Run("inactive status", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "u").Return(&domain.TusUpload{ID: "u", UserID: "u1", FileSize: 10, Status: domain.UploadStatusCancelled}, nil).Once()

			_, err := uc.HandleChunk(context.Background(), "u", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak dapat dilanjutkan")
		})
	})

	t.Run("CancelUpload", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			seedTusUploadStore(t, manager, "cancel-id", 8, map[string]string{"user_id": "u1"})
			manager.AddToQueue("cancel-id")

			tusRepo.On("GetByID", mock.Anything, "cancel-id").Return(&domain.TusUpload{ID: "cancel-id", UserID: "u1", Status: domain.UploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, "cancel-id", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelUpload(context.Background(), "cancel-id", "u1")
			require.NoError(t, err)
			tusRepo.AssertExpectations(t)
		})

		t.Run("not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelUpload(context.Background(), "missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "owner", Status: domain.UploadStatusUploading}, nil).Once()

			err := uc.CancelUpload(context.Background(), "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", Status: domain.UploadStatusCompleted}, nil).Once()

			err := uc.CancelUpload(context.Background(), "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("InitiateProjectUpdateUpload", func(t *testing.T) {
		t.Parallel()
		metadata := dto.TusUploadInitRequest{NamaProject: "Update", Kategori: "mobile", Semester: 5}

		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", mock.Anything, uint(9)).Return(&domain.Project{ID: 9, UserID: "u1", NamaProject: "Old", Kategori: "website", Semester: 1}, nil).Once()
			tusRepo.On("GetActiveByUserID", mock.Anything, "u1").Return([]domain.TusUpload{}, nil).Once()
			tusRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.TusUpload")).Return(nil).Once()

			res, err := uc.InitiateProjectUpdateUpload(context.Background(), 9, "u1", 512, metadata)
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/project/9/update/")
		})

		t.Run("project not found", func(t *testing.T) {
			t.Parallel()
			uc, _, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", mock.Anything, uint(9)).Return(nil, apperrors.ErrRecordNotFound).Once()

			res, err := uc.InitiateProjectUpdateUpload(context.Background(), 9, "u1", 512, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "Project tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, _, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", mock.Anything, uint(9)).Return(&domain.Project{ID: 9, UserID: "owner"}, nil).Once()

			res, err := uc.InitiateProjectUpdateUpload(context.Background(), 9, "u1", 512, metadata)
			require.Error(t, err)
			assert.Nil(t, res)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("uses existing project metadata when request metadata empty", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, projectRepo, _ := newTusUploadTestDeps(t)
			projectRepo.On("GetByID", mock.Anything, uint(9)).Return(&domain.Project{
				ID:          9,
				UserID:      "u1",
				NamaProject: "Existing Name",
				Kategori:    "website",
				Semester:    6,
			}, nil).Once()
			tusRepo.On("GetActiveByUserID", mock.Anything, "u1").Return([]domain.TusUpload{}, nil).Once()
			tusRepo.On("Create", mock.Anything, mock.MatchedBy(func(upload *domain.TusUpload) bool {
				return upload.UploadMetadata.NamaProject == "Existing Name" &&
					upload.UploadMetadata.Kategori == "website" &&
					upload.UploadMetadata.Semester == 6 &&
					upload.ProjectID != nil && *upload.ProjectID == 9 &&
					upload.UploadType == domain.UploadTypeProjectUpdate
			})).Return(nil).Once()

			res, err := uc.InitiateProjectUpdateUpload(context.Background(), 9, "u1", 512, dto.TusUploadInitRequest{})
			require.NoError(t, err)
			require.NotNil(t, res)
			assert.Contains(t, res.UploadURL, "/project/9/update/")
		})
	})

	t.Run("HandleProjectUpdateChunk", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			projectID := uint(1)
			uploadID := "proj-chunk"
			chunk := []byte("xyz")

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusUpload{
				ID:            uploadID,
				UserID:        "u1",
				ProjectID:     &projectID,
				FileSize:      10,
				CurrentOffset: 0,
				Status:        domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(3), mock.AnythingOfType("float64")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 10, map[string]string{"user_id": "u1", "project_id": "1"})
			offset, err := uc.HandleProjectUpdateChunk(context.Background(), projectID, uploadID, "u1", 0, bytes.NewReader(chunk))
			require.NoError(t, err)
			assert.Equal(t, int64(3), offset)
		})

		t.Run("auto completion", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, projectRepo, manager := newTusUploadTestDeps(t)
			projectID := uint(2)
			uploadID := "proj-complete"
			projectRepo.On("GetByID", mock.Anything, projectID).Return(&domain.Project{ID: projectID, UserID: "u1", PathFile: "", NamaProject: "Old", Kategori: "website", Semester: 1}, nil).Once()
			projectRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Project")).Return(nil).Once()

			tusRepo.On("GetByID", mock.Anything, uploadID).Return(&domain.TusUpload{
				ID:             uploadID,
				UserID:         "u1",
				ProjectID:      &projectID,
				UploadType:     domain.UploadTypeProjectUpdate,
				UploadMetadata: domain.TusUploadMetadata{NamaProject: "New", Kategori: "iot", Semester: 6},
				FileSize:       4,
				Status:         domain.UploadStatusUploading,
			}, nil).Once()
			tusRepo.On("UpdateOffset", mock.Anything, uploadID, int64(4), mock.AnythingOfType("float64")).Return(nil).Once()
			tusRepo.On("Complete", mock.Anything, uploadID, mock.AnythingOfType("uint"), mock.AnythingOfType("string")).Return(nil).Once()

			seedTusUploadStore(t, manager, uploadID, 4, map[string]string{"user_id": "u1", "project_id": "2"})
			offset, err := uc.HandleProjectUpdateChunk(context.Background(), projectID, uploadID, "u1", 0, bytes.NewReader([]byte("done")))
			require.NoError(t, err)
			assert.Equal(t, int64(4), offset)
		})

		t.Run("upload not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			_, err := uc.HandleProjectUpdateChunk(context.Background(), 1, "missing", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "owner", ProjectID: &projectID, FileSize: 8, Status: domain.UploadStatusUploading}, nil).Once()

			_, err := uc.HandleProjectUpdateChunk(context.Background(), projectID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", ProjectID: &projectID, FileSize: 8, Status: domain.UploadStatusCompleted}, nil).Once()

			offset, err := uc.HandleProjectUpdateChunk(context.Background(), projectID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Equal(t, int64(8), offset)
			assert.Contains(t, err.Error(), "sudah selesai")
		})

		t.Run("not active", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", ProjectID: &projectID, FileSize: 8, Status: domain.UploadStatusCancelled}, nil).Once()

			_, err := uc.HandleProjectUpdateChunk(context.Background(), projectID, "id", "u1", 0, bytes.NewReader([]byte("x")))
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak dapat dilanjutkan")
		})
	})

	t.Run("CancelProjectUpdateUpload", func(t *testing.T) {
		t.Parallel()
		t.Run("happy path", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, manager := newTusUploadTestDeps(t)
			projectID := uint(10)
			seedTusUploadStore(t, manager, "cancel-proj", 8, map[string]string{"user_id": "u1", "project_id": fmt.Sprintf("%d", projectID)})
			manager.AddToQueue("cancel-proj")

			tusRepo.On("GetByID", mock.Anything, "cancel-proj").Return(&domain.TusUpload{ID: "cancel-proj", UserID: "u1", ProjectID: &projectID, Status: domain.UploadStatusUploading}, nil).Once()
			tusRepo.On("UpdateStatus", mock.Anything, "cancel-proj", domain.UploadStatusCancelled).Return(nil).Once()

			err := uc.CancelProjectUpdateUpload(context.Background(), projectID, "cancel-proj", "u1")
			require.NoError(t, err)
		})

		t.Run("not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			err := uc.CancelProjectUpdateUpload(context.Background(), 1, "missing", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("wrong user", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "owner", ProjectID: &projectID, Status: domain.UploadStatusUploading}, nil).Once()

			err := uc.CancelProjectUpdateUpload(context.Background(), projectID, "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "tidak memiliki akses")
		})

		t.Run("already completed", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(1)
			tusRepo.On("GetByID", mock.Anything, "id").Return(&domain.TusUpload{ID: "id", UserID: "u1", ProjectID: &projectID, Status: domain.UploadStatusCompleted}, nil).Once()

			err := uc.CancelProjectUpdateUpload(context.Background(), projectID, "id", "u1")
			require.Error(t, err)
			assert.Contains(t, err.Error(), "sudah selesai")
		})
	})

	t.Run("Project update wrapper methods", func(t *testing.T) {
		t.Parallel()
		t.Run("GetProjectUpdateUploadInfo found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(11)
			now := time.Now()
			tusRepo.On("GetByID", mock.Anything, "project-info-id").Return(&domain.TusUpload{
				ID:             "project-info-id",
				UserID:         "u1",
				ProjectID:      &projectID,
				UploadMetadata: domain.TusUploadMetadata{NamaProject: "P", Kategori: "iot", Semester: 4},
				Status:         domain.UploadStatusUploading,
				Progress:       60,
				CurrentOffset:  6,
				FileSize:       10,
				CreatedAt:      now,
				UpdatedAt:      now,
			}, nil).Once()

			info, err := uc.GetProjectUpdateUploadInfo(context.Background(), projectID, "project-info-id", "u1")
			require.NoError(t, err)
			require.NotNil(t, info)
			assert.Equal(t, uint(11), info.ProjectID)
			assert.Equal(t, int64(6), info.Offset)
		})

		t.Run("GetProjectUpdateUploadInfo not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			info, err := uc.GetProjectUpdateUploadInfo(context.Background(), 11, "missing", "u1")
			require.Error(t, err)
			assert.Nil(t, info)
			assert.Contains(t, err.Error(), "tidak ditemukan")
		})

		t.Run("GetProjectUpdateUploadStatus found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			projectID := uint(12)
			tusRepo.On("GetByID", mock.Anything, "project-status-id").Return(&domain.TusUpload{
				ID:            "project-status-id",
				UserID:        "u1",
				ProjectID:     &projectID,
				CurrentOffset: 9,
				FileSize:      18,
			}, nil).Once()

			offset, length, err := uc.GetProjectUpdateUploadStatus(context.Background(), projectID, "project-status-id", "u1")
			require.NoError(t, err)
			assert.Equal(t, int64(9), offset)
			assert.Equal(t, int64(18), length)
		})

		t.Run("GetProjectUpdateUploadStatus not found", func(t *testing.T) {
			t.Parallel()
			uc, tusRepo, _, _ := newTusUploadTestDeps(t)
			tusRepo.On("GetByID", mock.Anything, "missing").Return(nil, gorm.ErrRecordNotFound).Once()

			offset, length, err := uc.GetProjectUpdateUploadStatus(context.Background(), 12, "missing", "u1")
			require.Error(t, err)
			assert.Equal(t, int64(0), offset)
			assert.Equal(t, int64(0), length)
		})
	})
}
