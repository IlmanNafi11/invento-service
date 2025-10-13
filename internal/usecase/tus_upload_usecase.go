package usecase

import (
	"errors"
	"fmt"
	"io"
	"time"

	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

	"github.com/google/uuid"
)

type TusUploadUsecase interface {
	CheckUploadSlot(userID uint) (*domain.TusUploadSlotResponse, error)
	ResetUploadQueue(userID uint) error
	InitiateUpload(userID uint, userEmail string, userRole string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error)
	HandleChunk(uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error)
	GetUploadInfo(uploadID string, userID uint) (*domain.TusUploadInfoResponse, error)
	CancelUpload(uploadID string, userID uint) error
	GetUploadStatus(uploadID string, userID uint) (int64, int64, error)
	InitiateProjectUpdateUpload(projectID uint, userID uint, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error)
	HandleProjectUpdateChunk(projectID uint, uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error)
	GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID uint) (int64, int64, error)
	GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID uint) (*domain.TusUploadInfoResponse, error)
	CancelProjectUpdateUpload(projectID uint, uploadID string, userID uint) error
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

func (uc *tusUploadUsecase) CheckUploadSlot(userID uint) (*domain.TusUploadSlotResponse, error) {
	response := uc.tusManager.CheckUploadSlot()
	
	return &domain.TusUploadSlotResponse{
		Available:     response.Available,
		Message:       response.Message,
		QueueLength:   response.QueueLength,
		ActiveUpload:  response.ActiveUpload,
		MaxConcurrent: response.MaxConcurrent,
	}, nil
}

func (uc *tusUploadUsecase) ResetUploadQueue(userID uint) error {
	return uc.tusManager.ResetUploadQueue()
}

func (uc *tusUploadUsecase) InitiateUpload(userID uint, userEmail string, userRole string, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	if fileSize > uc.config.Upload.MaxSize {
		return nil, fmt.Errorf("ukuran file melebihi batas maksimal %d MB", uc.config.Upload.MaxSize/(1024*1024))
	}

	if fileSize <= 0 {
		return nil, errors.New("ukuran file tidak valid")
	}

	if !uc.tusManager.CanAcceptUpload() {
		return nil, errors.New("slot upload tidak tersedia, silakan coba lagi nanti")
	}

	uploadID := uuid.New().String()

	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Second)

	tusUpload := &domain.TusUpload{
		ID:             uploadID,
		UserID:         userID,
		UploadType:     domain.UploadTypeProjectCreate,
		UploadURL:      fmt.Sprintf("/api/v1/project/upload/%s", uploadID),
		UploadMetadata: metadata,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.UploadStatusUploading,
		Progress:       0,
		ExpiresAt:      expiresAt,
	}

	if err := uc.tusUploadRepo.Create(tusUpload); err != nil {
		return nil, errors.New("gagal membuat upload record")
	}

	metadataMap := make(map[string]string)
	metadataMap["nama_project"] = metadata.NamaProject
	metadataMap["kategori"] = metadata.Kategori
	metadataMap["semester"] = fmt.Sprintf("%d", metadata.Semester)
	metadataMap["user_id"] = fmt.Sprintf("%d", userID)

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		uc.tusUploadRepo.Delete(uploadID)
		return nil, errors.New("gagal menginisiasi upload storage")
	}

	uc.tusManager.AddToQueue(uploadID)

	uploadURL := fmt.Sprintf("/api/v1/project/upload/%s", uploadID)

	return &domain.TusUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusUploadUsecase) HandleChunk(uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error) {
	return uc.tusManager.HandleChunk(uploadID, offset, chunk)
}

// Placeholder implementations for other methods
func (uc *tusUploadUsecase) GetUploadInfo(uploadID string, userID uint) (*domain.TusUploadInfoResponse, error) {
	// TODO: Implement
	return nil, nil
}

func (uc *tusUploadUsecase) CancelUpload(uploadID string, userID uint) error {
	// TODO: Implement
	return nil
}

func (uc *tusUploadUsecase) GetUploadStatus(uploadID string, userID uint) (int64, int64, error) {
	// TODO: Implement
	return 0, 0, nil
}

func (uc *tusUploadUsecase) InitiateProjectUpdateUpload(projectID uint, userID uint, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	// TODO: Implement
	return nil, nil
}

func (uc *tusUploadUsecase) HandleProjectUpdateChunk(projectID uint, uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error) {
	// TODO: Implement
	return 0, nil
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID uint) (int64, int64, error) {
	// TODO: Implement
	return 0, 0, nil
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID uint) (*domain.TusUploadInfoResponse, error) {
	// TODO: Implement
	return nil, nil
}

func (uc *tusUploadUsecase) CancelProjectUpdateUpload(projectID uint, uploadID string, userID uint) error {
	// TODO: Implement
	return nil
}
