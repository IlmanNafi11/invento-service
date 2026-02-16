package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"invento-service/internal/usecase/repo"

	"github.com/google/uuid"
	zlog "github.com/rs/zerolog/log"
)

type TusUploadUsecase interface {
	CheckUploadSlot(ctx context.Context, userID string) (*dto.TusUploadSlotResponse, error)
	ResetUploadQueue(ctx context.Context, userID string) error
	InitiateUpload(ctx context.Context, userID string, userEmail string, userRole string, fileSize int64, metadata dto.TusUploadInitRequest) (*dto.TusUploadResponse, error)
	HandleChunk(ctx context.Context, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetUploadInfo(ctx context.Context, uploadID string, userID string) (*dto.TusUploadInfoResponse, error)
	CancelUpload(ctx context.Context, uploadID string, userID string) error
	GetUploadStatus(ctx context.Context, uploadID string, userID string) (int64, int64, error)
	InitiateProjectUpdateUpload(ctx context.Context, projectID uint, userID string, fileSize int64, metadata dto.TusUploadInitRequest) (*dto.TusUploadResponse, error)
	HandleProjectUpdateChunk(ctx context.Context, projectID uint, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetProjectUpdateUploadStatus(ctx context.Context, projectID uint, uploadID string, userID string) (int64, int64, error)
	GetProjectUpdateUploadInfo(ctx context.Context, projectID uint, uploadID string, userID string) (*dto.TusUploadInfoResponse, error)
	CancelProjectUpdateUpload(ctx context.Context, projectID uint, uploadID string, userID string) error
}

type tusUploadUsecase struct {
	tusUploadRepo  repo.TusUploadRepository
	projectRepo    repo.ProjectRepository
	projectUsecase ProjectUsecase
	tusManager     *upload.TusManager
	fileManager    *storage.FileManager
	config         *config.Config
}

func NewTusUploadUsecase(
	tusUploadRepo repo.TusUploadRepository,
	projectRepo repo.ProjectRepository,
	projectUsecase ProjectUsecase,
	tusManager *upload.TusManager,
	fileManager *storage.FileManager,
	cfg *config.Config,
) TusUploadUsecase {
	return &tusUploadUsecase{
		tusUploadRepo:  tusUploadRepo,
		projectRepo:    projectRepo,
		projectUsecase: projectUsecase,
		tusManager:     tusManager,
		fileManager:    fileManager,
		config:         cfg,
	}
}

func (uc *tusUploadUsecase) CheckUploadSlot(ctx context.Context, userID string) (*dto.TusUploadSlotResponse, error) {
	response := uc.tusManager.CheckUploadSlot()

	return &dto.TusUploadSlotResponse{
		Available:     response.Available,
		Message:       response.Message,
		QueueLength:   response.QueueLength,
		ActiveUpload:  response.ActiveUpload,
		MaxConcurrent: response.MaxConcurrent,
	}, nil
}

func (uc *tusUploadUsecase) ResetUploadQueue(ctx context.Context, userID string) error {
	activeUploads, err := uc.tusUploadRepo.GetActiveByUserID(ctx, userID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.ResetUploadQueue: %w", err))
	}

	for _, upload := range activeUploads {
		projectID := upload.ProjectID
		if err := uc.cancelUpload(ctx, upload.ID, userID, projectID); err != nil {
			return err
		}
	}

	return nil
}

func (uc *tusUploadUsecase) InitiateUpload(ctx context.Context, userID string, userEmail string, userRole string, fileSize int64, metadata dto.TusUploadInitRequest) (*dto.TusUploadResponse, error) {
	return uc.initiateUpload(ctx, userID, fileSize, metadata, domain.UploadTypeProjectCreate, nil)
}

func (uc *tusUploadUsecase) InitiateProjectUpdateUpload(ctx context.Context, projectID uint, userID string, fileSize int64, metadata dto.TusUploadInitRequest) (*dto.TusUploadResponse, error) {
	project, err := uc.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Project")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.InitiateProjectUpdateUpload: %w", err))
	}

	if project.UserID != userID {
		return nil, apperrors.NewForbiddenError("Anda tidak memiliki akses ke project ini")
	}

	if metadata.NamaProject == "" {
		metadata.NamaProject = project.NamaProject
	}
	if metadata.Kategori == "" {
		metadata.Kategori = project.Kategori
	}
	if metadata.Semester == 0 {
		metadata.Semester = project.Semester
	}

	return uc.initiateUpload(ctx, userID, fileSize, metadata, domain.UploadTypeProjectUpdate, &projectID)
}

func (uc *tusUploadUsecase) initiateUpload(ctx context.Context, userID string, fileSize int64, metadata dto.TusUploadInitRequest, uploadType string, projectID *uint) (*dto.TusUploadResponse, error) {
	if fileSize > uc.config.Upload.MaxSizeProject {
		return nil, apperrors.NewPayloadTooLargeError(fmt.Sprintf("ukuran file melebihi batas maksimal %d MB", uc.config.Upload.MaxSizeProject/(1024*1024)))
	}

	if fileSize <= 0 {
		return nil, apperrors.NewValidationError("ukuran file tidak valid", nil)
	}

	existingUploads, err := uc.tusUploadRepo.GetActiveByUserID(ctx, userID)
	if err == nil {
		for _, existing := range existingUploads {
			if existing.Status == domain.UploadStatusPending ||
				existing.Status == domain.UploadStatusUploading ||
				existing.Status == domain.UploadStatusQueued {
				return nil, apperrors.NewConflictError("anda sudah memiliki upload yang sedang berjalan, selesaikan atau batalkan terlebih dahulu")
			}
		}
	}

	if !uc.tusManager.CanAcceptUpload() {
		return nil, apperrors.NewConflictError("slot upload tidak tersedia, silakan coba lagi nanti")
	}

	uploadID := uuid.New().String()
	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Second)
	uploadURL := fmt.Sprintf("/project/upload/%s", uploadID)
	if uploadType == domain.UploadTypeProjectUpdate && projectID != nil {
		uploadURL = fmt.Sprintf("/project/%d/update/%s", *projectID, uploadID)
	}

	tusUpload := &domain.TusUpload{
		ID:         uploadID,
		UserID:     userID,
		ProjectID:  projectID,
		UploadType: uploadType,
		UploadURL:  uploadURL,
		UploadMetadata: domain.TusUploadMetadata{
			NamaProject: metadata.NamaProject,
			Kategori:    metadata.Kategori,
			Semester:    metadata.Semester,
		},
		FileSize:      fileSize,
		CurrentOffset: 0,
		Status:        domain.UploadStatusPending,
		Progress:      0,
		ExpiresAt:     expiresAt,
	}

	if err := uc.tusUploadRepo.Create(ctx, tusUpload); err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.initiateUpload: create record: %w", err))
	}

	metadataMap := map[string]string{
		"nama_project": metadata.NamaProject,
		"kategori":     metadata.Kategori,
		"semester":     fmt.Sprintf("%d", metadata.Semester),
		"user_id":      userID,
	}
	if projectID != nil {
		metadataMap["project_id"] = fmt.Sprintf("%d", *projectID)
	}

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		_ = uc.tusUploadRepo.Delete(ctx, uploadID)
		return nil, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.initiateUpload: init storage: %w", err))
	}

	uc.tusManager.AddToQueue(uploadID)

	return &dto.TusUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusUploadUsecase) HandleChunk(ctx context.Context, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(ctx, uploadID, userID, offset, chunk, nil)
}

func (uc *tusUploadUsecase) HandleProjectUpdateChunk(ctx context.Context, projectID uint, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(ctx, uploadID, userID, offset, chunk, &projectID)
}

func (uc *tusUploadUsecase) handleChunk(ctx context.Context, uploadID string, userID string, offset int64, chunk io.Reader, projectID *uint) (int64, error) {
	upload, err := uc.getOwnedUpload(ctx, uploadID, userID, projectID)
	if err != nil {
		return 0, err
	}

	if upload.Status == domain.UploadStatusCompleted {
		return upload.FileSize, apperrors.NewTusCompletedError()
	}

	if upload.Status == domain.UploadStatusCancelled || upload.Status == domain.UploadStatusFailed {
		return 0, apperrors.NewTusInactiveError()
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return offset, err
	}

	if upload.Status == domain.UploadStatusPending && newOffset > 0 {
		if err := uc.tusUploadRepo.UpdateStatus(ctx, uploadID, domain.UploadStatusUploading); err != nil {
			return newOffset, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.handleChunk: update status: %w", err))
		}
		upload.Status = domain.UploadStatusUploading
	}

	progress := (float64(newOffset) / float64(upload.FileSize)) * 100
	if err := uc.tusUploadRepo.UpdateOffset(ctx, uploadID, newOffset, progress); err != nil {
		return newOffset, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.handleChunk: update offset: %w", err))
	}

	upload.CurrentOffset = newOffset
	upload.Progress = progress
	if newOffset >= upload.FileSize {
		if err := uc.completeUpload(ctx, upload); err != nil {
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusUploadUsecase) completeUpload(ctx context.Context, upload *domain.TusUpload) error {
	randomDir, err := uc.fileManager.GenerateRandomDirectory()
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.completeUpload: generate dir: %w", err))
	}

	finalFilePath := uc.fileManager.GetProjectFilePath(upload.UserID, randomDir, "project.zip")
	if err := uc.tusManager.FinalizeUpload(upload.ID, finalFilePath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.completeUpload: finalize: %w", err))
	}

	var projectID uint
	switch upload.UploadType {
	case domain.UploadTypeProjectCreate:
		projectID, err = uc.completeProjectCreate(ctx, upload, finalFilePath)
		if err != nil {
			return err
		}
	case domain.UploadTypeProjectUpdate:
		projectID, err = uc.completeProjectUpdate(ctx, upload, finalFilePath)
		if err != nil {
			return err
		}
	default:
		return apperrors.NewValidationError("tipe upload tidak didukung", nil)
	}

	if err := uc.tusUploadRepo.Complete(ctx, upload.ID, projectID, finalFilePath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.completeUpload: complete record: %w", err))
	}

	uc.tusManager.FinishUpload(upload.ID)
	return nil
}

func (uc *tusUploadUsecase) completeProjectCreate(ctx context.Context, upload *domain.TusUpload, finalFilePath string) (uint, error) {
	project := &domain.Project{
		UserID:      upload.UserID,
		NamaProject: upload.UploadMetadata.NamaProject,
		Kategori:    upload.UploadMetadata.Kategori,
		Semester:    upload.UploadMetadata.Semester,
		Ukuran:      storage.GetFileSizeFromPath(finalFilePath),
		PathFile:    finalFilePath,
	}

	if err := uc.projectRepo.Create(ctx, project); err != nil {
		return 0, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.completeProjectCreate: %w", err))
	}

	return project.ID, nil
}

func (uc *tusUploadUsecase) completeProjectUpdate(ctx context.Context, upload *domain.TusUpload, finalFilePath string) (uint, error) {
	if upload.ProjectID == nil {
		return 0, apperrors.NewValidationError("project ID tidak ditemukan", nil)
	}

	project, err := uc.projectRepo.GetByID(ctx, *upload.ProjectID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return 0, apperrors.NewNotFoundError("Project")
		}
		return 0, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.completeProjectUpdate: %w", err))
	}

	oldFilePath := project.PathFile
	project.NamaProject = upload.UploadMetadata.NamaProject
	project.Kategori = upload.UploadMetadata.Kategori
	project.Semester = upload.UploadMetadata.Semester
	project.Ukuran = storage.GetFileSizeFromPath(finalFilePath)
	project.PathFile = finalFilePath

	if err := uc.projectRepo.Update(ctx, project); err != nil {
		return 0, apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.completeProjectUpdate: %w", err))
	}

	if oldFilePath != "" && oldFilePath != finalFilePath {
		if err := storage.DeleteFile(oldFilePath); err != nil {
			// Old file deletion after successful update is critical but non-blocking
			zlog.Warn().Err(err).Str("file", oldFilePath).Msg("TusUploadUsecase.completeProjectUpdate: failed to delete old file")
		}
	}

	return project.ID, nil
}

func (uc *tusUploadUsecase) GetUploadInfo(ctx context.Context, uploadID string, userID string) (*dto.TusUploadInfoResponse, error) {
	return uc.getUploadInfo(ctx, uploadID, userID, nil)
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadInfo(ctx context.Context, projectID uint, uploadID string, userID string) (*dto.TusUploadInfoResponse, error) {
	return uc.getUploadInfo(ctx, uploadID, userID, &projectID)
}

func (uc *tusUploadUsecase) getUploadInfo(ctx context.Context, uploadID string, userID string, projectID *uint) (*dto.TusUploadInfoResponse, error) {
	upload, err := uc.getOwnedUpload(ctx, uploadID, userID, projectID)
	if err != nil {
		return nil, err
	}

	response := &dto.TusUploadInfoResponse{
		UploadID:    upload.ID,
		NamaProject: upload.UploadMetadata.NamaProject,
		Kategori:    upload.UploadMetadata.Kategori,
		Semester:    upload.UploadMetadata.Semester,
		Status:      upload.Status,
		Progress:    upload.Progress,
		Offset:      upload.CurrentOffset,
		Length:      upload.FileSize,
		CreatedAt:   upload.CreatedAt,
		UpdatedAt:   upload.UpdatedAt,
	}

	if upload.ProjectID != nil {
		response.ProjectID = *upload.ProjectID
	}

	return response, nil
}

func (uc *tusUploadUsecase) GetUploadStatus(ctx context.Context, uploadID string, userID string) (int64, int64, error) {
	return uc.getUploadStatus(ctx, uploadID, userID, nil)
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadStatus(ctx context.Context, projectID uint, uploadID string, userID string) (int64, int64, error) {
	return uc.getUploadStatus(ctx, uploadID, userID, &projectID)
}

func (uc *tusUploadUsecase) getUploadStatus(ctx context.Context, uploadID string, userID string, projectID *uint) (int64, int64, error) {
	upload, err := uc.getOwnedUpload(ctx, uploadID, userID, projectID)
	if err != nil {
		return 0, 0, err
	}
	return upload.CurrentOffset, upload.FileSize, nil
}

func (uc *tusUploadUsecase) CancelUpload(ctx context.Context, uploadID string, userID string) error {
	return uc.cancelUpload(ctx, uploadID, userID, nil)
}

func (uc *tusUploadUsecase) CancelProjectUpdateUpload(ctx context.Context, projectID uint, uploadID string, userID string) error {
	return uc.cancelUpload(ctx, uploadID, userID, &projectID)
}

func (uc *tusUploadUsecase) cancelUpload(ctx context.Context, uploadID string, userID string, projectID *uint) error {
	upload, err := uc.getOwnedUpload(ctx, uploadID, userID, projectID)
	if err != nil {
		return err
	}

	if upload.Status == domain.UploadStatusCompleted {
		return apperrors.NewTusCompletedError()
	}

	if err := uc.tusUploadRepo.UpdateStatus(ctx, uploadID, domain.UploadStatusCancelled); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusUploadUsecase.cancelUpload: %w", err))
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		zlog.Warn().Err(err).Str("upload_id", uploadID).Msg("failed to delete upload file")
	}

	uc.tusManager.FinishUpload(uploadID)
	return nil
}

func (uc *tusUploadUsecase) getOwnedUpload(ctx context.Context, uploadID string, userID string, projectID *uint) (*domain.TusUpload, error) {
	upload, err := uc.tusUploadRepo.GetByID(ctx, uploadID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Upload")
	}

	if upload.UserID != userID {
		return nil, apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	if upload.Status == domain.UploadStatusExpired || upload.Status == domain.UploadStatusCancelled || upload.Status == domain.UploadStatusFailed {
		return nil, apperrors.NewConflictError("upload sudah " + upload.Status + ", tidak dapat dilanjutkan")
	}

	if projectID != nil {
		if upload.ProjectID == nil || *upload.ProjectID != *projectID {
			return nil, apperrors.NewValidationError("project ID tidak cocok", nil)
		}
	}

	return upload, nil
}
