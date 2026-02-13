package usecase

import (
	"fmt"
	"io"
	"log"
	"time"

	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

	"github.com/google/uuid"
)

type TusUploadUsecase interface {
	CheckUploadSlot(userID string) (*domain.TusUploadSlotResponse, error)
	ResetUploadQueue(userID string) error
	InitiateUpload(userID string, userEmail string, userRole string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error)
	HandleChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetUploadInfo(uploadID string, userID string) (*domain.TusUploadInfoResponse, error)
	CancelUpload(uploadID string, userID string) error
	GetUploadStatus(uploadID string, userID string) (int64, int64, error)
	InitiateProjectUpdateUpload(projectID uint, userID string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error)
	HandleProjectUpdateChunk(projectID uint, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID string) (int64, int64, error)
	GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID string) (*domain.TusUploadInfoResponse, error)
	CancelProjectUpdateUpload(projectID uint, uploadID string, userID string) error
}

type tusUploadUsecase struct {
	tusUploadRepo  repo.TusUploadRepository
	projectRepo    repo.ProjectRepository
	projectUsecase ProjectUsecase
	tusManager     *helper.TusManager
	fileManager    *helper.FileManager
	config         *config.Config
}

func NewTusUploadUsecase(
	tusUploadRepo repo.TusUploadRepository,
	projectRepo repo.ProjectRepository,
	projectUsecase ProjectUsecase,
	tusManager *helper.TusManager,
	fileManager *helper.FileManager,
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

func (uc *tusUploadUsecase) CheckUploadSlot(userID string) (*domain.TusUploadSlotResponse, error) {
	response := uc.tusManager.CheckUploadSlot()

	return &domain.TusUploadSlotResponse{
		Available:     response.Available,
		Message:       response.Message,
		QueueLength:   response.QueueLength,
		ActiveUpload:  response.ActiveUpload,
		MaxConcurrent: response.MaxConcurrent,
	}, nil
}

func (uc *tusUploadUsecase) ResetUploadQueue(userID string) error {
	return uc.tusManager.ResetUploadQueue()
}

func (uc *tusUploadUsecase) InitiateUpload(userID string, userEmail string, userRole string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	return uc.initiateUpload(userID, fileSize, metadata, domain.UploadTypeProjectCreate, nil)
}

func (uc *tusUploadUsecase) InitiateProjectUpdateUpload(projectID uint, userID string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Project")
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

	return uc.initiateUpload(userID, fileSize, metadata, domain.UploadTypeProjectUpdate, &projectID)
}

func (uc *tusUploadUsecase) initiateUpload(userID string, fileSize int64, metadata domain.TusUploadInitRequest, uploadType string, projectID *uint) (*domain.TusUploadResponse, error) {
	if fileSize > uc.config.Upload.MaxSizeProject {
		return nil, apperrors.NewPayloadTooLargeError(fmt.Sprintf("ukuran file melebihi batas maksimal %d MB", uc.config.Upload.MaxSizeProject/(1024*1024)))
	}

	if fileSize <= 0 {
		return nil, apperrors.NewValidationError("ukuran file tidak valid", nil)
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
		ID:             uploadID,
		UserID:         userID,
		ProjectID:      projectID,
		UploadType:     uploadType,
		UploadURL:      uploadURL,
		UploadMetadata: metadata,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.UploadStatusPending,
		Progress:       0,
		ExpiresAt:      expiresAt,
	}

	if err := uc.tusUploadRepo.Create(tusUpload); err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal membuat upload record: %w", err))
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
		_ = uc.tusUploadRepo.Delete(uploadID)
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal menginisiasi upload storage: %w", err))
	}

	uc.tusManager.AddToQueue(uploadID)

	return &domain.TusUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusUploadUsecase) HandleChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(uploadID, userID, offset, chunk, nil)
}

func (uc *tusUploadUsecase) HandleProjectUpdateChunk(projectID uint, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(uploadID, userID, offset, chunk, &projectID)
}

func (uc *tusUploadUsecase) handleChunk(uploadID string, userID string, offset int64, chunk io.Reader, projectID *uint) (int64, error) {
	upload, err := uc.getOwnedUpload(uploadID, userID, projectID)
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
		if err := uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusUploading); err != nil {
			return newOffset, apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
		}
		upload.Status = domain.UploadStatusUploading
	}

	progress := (float64(newOffset) / float64(upload.FileSize)) * 100
	if err := uc.tusUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, apperrors.NewInternalError(fmt.Errorf("gagal update offset upload: %w", err))
	}

	upload.CurrentOffset = newOffset
	upload.Progress = progress
	if newOffset >= upload.FileSize {
		if err := uc.completeUpload(upload); err != nil {
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusUploadUsecase) completeUpload(upload *domain.TusUpload) error {
	randomDir, err := uc.fileManager.GenerateRandomDirectory()
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal generate random directory: %w", err))
	}

	finalFilePath := uc.fileManager.GetProjectFilePath(upload.UserID, randomDir, "project.zip")
	if err := uc.tusManager.FinalizeUpload(upload.ID, finalFilePath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal finalisasi upload: %w", err))
	}

	var projectID uint
	switch upload.UploadType {
	case domain.UploadTypeProjectCreate:
		projectID, err = uc.completeProjectCreate(upload, finalFilePath)
		if err != nil {
			return err
		}
	case domain.UploadTypeProjectUpdate:
		projectID, err = uc.completeProjectUpdate(upload, finalFilePath)
		if err != nil {
			return err
		}
	default:
		return apperrors.NewValidationError("tipe upload tidak didukung", nil)
	}

	if err := uc.tusUploadRepo.Complete(upload.ID, projectID, finalFilePath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
	}

	uc.tusManager.RemoveFromQueue(upload.ID)
	return nil
}

func (uc *tusUploadUsecase) completeProjectCreate(upload *domain.TusUpload, finalFilePath string) (uint, error) {
	project := &domain.Project{
		UserID:      upload.UserID,
		NamaProject: upload.UploadMetadata.NamaProject,
		Kategori:    upload.UploadMetadata.Kategori,
		Semester:    upload.UploadMetadata.Semester,
		Ukuran:      helper.GetFileSizeFromPath(finalFilePath),
		PathFile:    finalFilePath,
	}

	if err := uc.projectRepo.Create(project); err != nil {
		return 0, apperrors.NewInternalError(fmt.Errorf("gagal membuat project: %w", err))
	}

	return project.ID, nil
}

func (uc *tusUploadUsecase) completeProjectUpdate(upload *domain.TusUpload, finalFilePath string) (uint, error) {
	if upload.ProjectID == nil {
		return 0, apperrors.NewValidationError("project ID tidak ditemukan", nil)
	}

	project, err := uc.projectRepo.GetByID(*upload.ProjectID)
	if err != nil {
		return 0, apperrors.NewNotFoundError("Project")
	}

	if project.PathFile != "" {
		if err := helper.DeleteFile(project.PathFile); err != nil {
			log.Printf("Warning: gagal menghapus file lama: %v", err)
		}
	}

	project.NamaProject = upload.UploadMetadata.NamaProject
	project.Kategori = upload.UploadMetadata.Kategori
	project.Semester = upload.UploadMetadata.Semester
	project.Ukuran = helper.GetFileSizeFromPath(finalFilePath)
	project.PathFile = finalFilePath

	if err := uc.projectRepo.Update(project); err != nil {
		return 0, apperrors.NewInternalError(fmt.Errorf("gagal update project: %w", err))
	}

	return project.ID, nil
}

func (uc *tusUploadUsecase) GetUploadInfo(uploadID string, userID string) (*domain.TusUploadInfoResponse, error) {
	return uc.getUploadInfo(uploadID, userID, nil)
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID string) (*domain.TusUploadInfoResponse, error) {
	return uc.getUploadInfo(uploadID, userID, &projectID)
}

func (uc *tusUploadUsecase) getUploadInfo(uploadID string, userID string, projectID *uint) (*domain.TusUploadInfoResponse, error) {
	upload, err := uc.getOwnedUpload(uploadID, userID, projectID)
	if err != nil {
		return nil, err
	}

	response := &domain.TusUploadInfoResponse{
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

func (uc *tusUploadUsecase) GetUploadStatus(uploadID string, userID string) (int64, int64, error) {
	return uc.getUploadStatus(uploadID, userID, nil)
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID string) (int64, int64, error) {
	return uc.getUploadStatus(uploadID, userID, &projectID)
}

func (uc *tusUploadUsecase) getUploadStatus(uploadID string, userID string, projectID *uint) (int64, int64, error) {
	upload, err := uc.getOwnedUpload(uploadID, userID, projectID)
	if err != nil {
		return 0, 0, err
	}
	return upload.CurrentOffset, upload.FileSize, nil
}

func (uc *tusUploadUsecase) CancelUpload(uploadID string, userID string) error {
	return uc.cancelUpload(uploadID, userID, nil)
}

func (uc *tusUploadUsecase) CancelProjectUpdateUpload(projectID uint, uploadID string, userID string) error {
	return uc.cancelUpload(uploadID, userID, &projectID)
}

func (uc *tusUploadUsecase) cancelUpload(uploadID string, userID string, projectID *uint) error {
	upload, err := uc.getOwnedUpload(uploadID, userID, projectID)
	if err != nil {
		return err
	}

	if upload.Status == domain.UploadStatusCompleted {
		return apperrors.NewTusCompletedError()
	}

	if err := uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusCancelled); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal membatalkan upload: %w", err))
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		log.Printf("Warning: gagal menghapus file upload: %v", err)
	}

	uc.tusManager.RemoveFromQueue(uploadID)
	return nil
}

func (uc *tusUploadUsecase) getOwnedUpload(uploadID string, userID string, projectID *uint) (*domain.TusUpload, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return nil, apperrors.NewNotFoundError("Upload")
	}

	if upload.UserID != userID {
		return nil, apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	if projectID != nil {
		if upload.ProjectID == nil || *upload.ProjectID != *projectID {
			return nil, apperrors.NewValidationError("project ID tidak cocok", nil)
		}
	}

	return upload, nil
}
