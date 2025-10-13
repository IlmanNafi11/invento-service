package usecase

import (
	"errors"
	"fmt"
	"io"
	"log"
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
		return nil, fmt.Errorf("gagal membuat upload record: %v", err)
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

	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return 0, errors.New("upload tidak ditemukan")
	}
	if upload.UserID != userID {
		return 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.Status == domain.UploadStatusCompleted {
		return upload.FileSize, nil
	}

	if upload.Status == domain.UploadStatusCancelled || upload.Status == domain.UploadStatusFailed {
		return 0, errors.New("upload sudah dibatalkan atau gagal")
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return offset, err
	}

	progress := (float64(newOffset) / float64(upload.FileSize)) * 100
	if err := uc.tusUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, errors.New("gagal update offset upload")
	}
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
		return errors.New("gagal generate random directory")
	}

	finalFilePath := uc.fileManager.GetProjectFilePath(upload.UserID, randomDir, "project.zip")

	if err := uc.tusManager.FinalizeUpload(upload.ID, finalFilePath); err != nil {
		return errors.New("gagal finalisasi upload")
	}

	if err := uc.tusUploadRepo.UpdateStatus(upload.ID, domain.UploadStatusCompleted); err != nil {
		return errors.New("gagal update status upload")
	}

	if upload.UploadType == domain.UploadTypeProjectCreate {
		if err := uc.completeProjectCreate(upload, finalFilePath); err != nil {
			return err
		}
	}

	uc.tusManager.RemoveFromQueue(upload.ID)

	return nil
}

func (uc *tusUploadUsecase) completeProjectCreate(upload *domain.TusUpload, finalFilePath string) error {

	project := &domain.Project{
		UserID:      upload.UserID,
		NamaProject: upload.UploadMetadata.NamaProject,
		Kategori:    upload.UploadMetadata.Kategori,
		Semester:    upload.UploadMetadata.Semester,
		Ukuran:      helper.GetFileSizeFromPath(finalFilePath),
		PathFile:    finalFilePath,
	}

	if err := uc.projectRepo.Create(project); err != nil {
		return errors.New("gagal membuat project")
	}

	return nil
}

func (uc *tusUploadUsecase) GetUploadInfo(uploadID string, userID uint) (*domain.TusUploadInfoResponse, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return nil, errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke upload ini")
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

func (uc *tusUploadUsecase) CancelUpload(uploadID string, userID uint) error {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.Status == domain.UploadStatusCompleted {
		return errors.New("upload sudah selesai dan tidak bisa dibatalkan")
	}

	if err := uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusCancelled); err != nil {
		return errors.New("gagal membatalkan upload")
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		log.Printf("Warning: gagal menghapus file upload: %v", err)
	}

	uc.tusManager.RemoveFromQueue(uploadID)

	return nil
}

func (uc *tusUploadUsecase) GetUploadStatus(uploadID string, userID uint) (int64, int64, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return 0, 0, errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return 0, 0, errors.New("tidak memiliki akses ke upload ini")
	}

	return upload.CurrentOffset, upload.FileSize, nil
}

func (uc *tusUploadUsecase) InitiateProjectUpdateUpload(projectID uint, userID uint, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return nil, errors.New("project tidak ditemukan")
	}

	if project.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke project ini")
	}

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

	if metadata.NamaProject == "" {
		metadata.NamaProject = project.NamaProject
	}
	if metadata.Kategori == "" {
		metadata.Kategori = project.Kategori
	}
	if metadata.Semester == 0 {
		metadata.Semester = project.Semester
	}

	tusUpload := &domain.TusUpload{
		ID:             uploadID,
		UserID:         userID,
		ProjectID:      &projectID,
		UploadType:     domain.UploadTypeProjectUpdate,
		UploadURL:      fmt.Sprintf("/api/v1/project/%d/update/%s", projectID, uploadID),
		UploadMetadata: metadata,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.UploadStatusUploading,
		Progress:       0,
		ExpiresAt:      expiresAt,
	}

	if err := uc.tusUploadRepo.Create(tusUpload); err != nil {
		return nil, fmt.Errorf("gagal membuat upload record: %v", err)
	}

	metadataMap := make(map[string]string)
	metadataMap["nama_project"] = metadata.NamaProject
	metadataMap["kategori"] = metadata.Kategori
	metadataMap["semester"] = fmt.Sprintf("%d", metadata.Semester)
	metadataMap["user_id"] = fmt.Sprintf("%d", userID)
	metadataMap["project_id"] = fmt.Sprintf("%d", projectID)

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		uc.tusUploadRepo.Delete(uploadID)
		return nil, errors.New("gagal menginisiasi upload storage")
	}

	uc.tusManager.AddToQueue(uploadID)

	uploadURL := fmt.Sprintf("/api/v1/project/%d/update/%s", projectID, uploadID)

	return &domain.TusUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusUploadUsecase) HandleProjectUpdateChunk(projectID uint, uploadID string, userID uint, offset int64, chunk io.Reader) (int64, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return 0, errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return 0, errors.New("project ID tidak cocok")
	}

	if upload.Status == domain.UploadStatusCompleted {
		return upload.FileSize, nil
	}

	if upload.Status == domain.UploadStatusCancelled || upload.Status == domain.UploadStatusFailed {
		return 0, errors.New("upload sudah dibatalkan atau gagal")
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return offset, err
	}

	progress := (float64(newOffset) / float64(upload.FileSize)) * 100
	if err := uc.tusUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, errors.New("gagal update offset upload")
	}

	if newOffset >= upload.FileSize {
		if err := uc.completeProjectUpdate(upload); err != nil {
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID uint) (int64, int64, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return 0, 0, errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return 0, 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return 0, 0, errors.New("project ID tidak cocok")
	}

	return upload.CurrentOffset, upload.FileSize, nil
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID uint) (*domain.TusUploadInfoResponse, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return nil, errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return nil, errors.New("project ID tidak cocok")
	}

	response := &domain.TusUploadInfoResponse{
		UploadID:    upload.ID,
		ProjectID:   *upload.ProjectID,
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

	return response, nil
}

func (uc *tusUploadUsecase) CancelProjectUpdateUpload(projectID uint, uploadID string, userID uint) error {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		return errors.New("upload tidak ditemukan")
	}

	if upload.UserID != userID {
		return errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return errors.New("project ID tidak cocok")
	}

	if upload.Status == domain.UploadStatusCompleted {
		return errors.New("upload sudah selesai dan tidak bisa dibatalkan")
	}

	if err := uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusCancelled); err != nil {
		return errors.New("gagal membatalkan upload")
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		log.Printf("Warning: gagal menghapus file upload: %v", err)
	}

	uc.tusManager.RemoveFromQueue(uploadID)

	return nil
}

func (uc *tusUploadUsecase) completeProjectUpdate(upload *domain.TusUpload) error {

	randomDir, err := uc.fileManager.GenerateRandomDirectory()
	if err != nil {
		return errors.New("gagal generate random directory")
	}

	finalFilePath := uc.fileManager.GetProjectFilePath(upload.UserID, randomDir, "project.zip")

	if err := uc.tusManager.FinalizeUpload(upload.ID, finalFilePath); err != nil {
		return errors.New("gagal finalisasi upload")
	}

	if err := uc.tusUploadRepo.UpdateStatus(upload.ID, domain.UploadStatusCompleted); err != nil {
		return errors.New("gagal update status upload")
	}

	project, err := uc.projectRepo.GetByID(*upload.ProjectID)
	if err != nil {
		return errors.New("project tidak ditemukan")
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
		return errors.New("gagal update project")
	}

	uc.tusManager.RemoveFromQueue(upload.ID)

	return nil
}
