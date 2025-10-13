package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
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
	HandleProjectUpdateChunk(projectID uint, uploadID string, userID uint, offset int64, chunk io.Reader, chunkSize int64) (int64, error)
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
	
	if err := uc.tusManager.ValidateMetadata(metadataMap); err != nil {
		uc.tusUploadRepo.Delete(uploadID)
		return nil, err
	}
	
	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
	
	if err := uc.tusManager.ValidateMetadata(metadataMap); err != nil {
	uc.tusUploadRepo.Delete(uploadID)
		return nil, err
	}
	
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
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("upload tidak ditemukan")
		}
		return 0, errors.New("gagal mengambil data upload")
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

	if !uc.tusManager.IsActiveUpload(uploadID) {
		return upload.CurrentOffset, errors.New("upload tidak aktif")
	}

	if upload.CurrentOffset != offset {
		return upload.CurrentOffset, errors.New("offset tidak valid")
	}

	upload.CurrentOffset = offset

	bytesWritten, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		uc.tusManager.RemoveFromQueue(uploadID)
		uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusFailed)
		return upload.CurrentOffset, err
	}

	newOffset := offset + bytesWritten
	progress := (float64(newOffset) / float64(upload.FileSize)) * 100

	if err := uc.tusUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, errors.New("gagal update offset upload")
	}

	if newOffset >= upload.FileSize {
		if err := uc.completeUpload(upload); err != nil {
			uc.tusManager.RemoveFromQueue(uploadID)
			uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusFailed)
			return newOffset, err
		}
		uc.tusQueue.FinishActiveUpload()
	}

	return newOffset, nil
}

func (uc *tusUploadUsecase) completeUpload(upload *domain.TusUpload) error {
	if err := uc.tusUploadRepo.UpdateStatus(upload.ID, domain.UploadStatusCompleted); err != nil {
		return errors.New("gagal update status upload")
	}

	srcPath := uc.tusManager.GetUploadInfo(upload.ID).ID

	username := fmt.Sprintf("user-%d", upload.UserID)
	destDir := uc.fileManager.GetUserUploadPath(upload.UserID)
	

	destPath := filepath.Join(destDir, fmt.Sprintf("%s.zip", upload.UploadMetadata.NamaProject))

	if err := helper.MoveFile(srcPath, destPath); err != nil {
		return errors.New("gagal memindahkan file")
	}

	ukuran := helper.FormatFileSize(upload.FileSize)

	project := &domain.Project{
		UserID:      upload.UserID,
		NamaProject: upload.UploadMetadata.NamaProject,
		Kategori:    upload.UploadMetadata.Kategori,
		Semester:    upload.UploadMetadata.Semester,
		Ukuran:      ukuran,
		PathFile:    destPath,
	}

	if err := uc.projectRepo.Create(project); err != nil {
		return errors.New("gagal menyimpan data project")
	}

	if err := uc.tusManager.CancelUpload(upload.ID); err != nil {
		return nil
	}

	return nil
}

func (uc *tusUploadUsecase) GetUploadInfo(uploadID string, userID uint) (*domain.TusUploadInfoResponse, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("upload tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke upload ini")
	}

	return &domain.TusUploadInfoResponse{
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
	}, nil
}

func (uc *tusUploadUsecase) CancelUpload(uploadID string, userID uint) error {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("upload tidak ditemukan")
		}
		return errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.Status == domain.UploadStatusCompleted {
		return errors.New("upload sudah selesai dan tidak bisa dibatalkan")
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		return errors.New("gagal menghapus file upload")
	}

	if err := uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusCancelled); err != nil {
		return errors.New("gagal update status upload")
	}

	uc.tusManager.RemoveFromQueue(uploadID)

	return nil
}

func (uc *tusUploadUsecase) GetUploadStatus(uploadID string, userID uint) (int64, int64, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, errors.New("upload tidak ditemukan")
		}
		return 0, 0, errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return 0, 0, errors.New("tidak memiliki akses ke upload ini")
	}

	return upload.CurrentOffset, upload.FileSize, nil
}

func (uc *tusUploadUsecase) InitiateProjectUpdateUpload(projectID uint, userID uint, fileSize int64, metadata domain.TusUploadInitRequest) (*domain.TusUploadResponse, error) {
	if fileSize > uc.config.Upload.MaxSize {
		return nil, fmt.Errorf("ukuran file melebihi batas maksimal %d MB", uc.config.Upload.MaxSize/(1024*1024))
	}

	if fileSize <= 0 {
		return nil, errors.New("ukuran file tidak valid")
	}

	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data project")
	}

	if project.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke project ini")
	}

	if !uc.tusManager.CanAcceptUpload() {
		return nil, errors.New("slot upload tidak tersedia, coba beberapa saat lagi")
	}

	uploadID := uuid.New().String()
	uploadURL := fmt.Sprintf("/api/v1/project/%d/update/%s", projectID, uploadID)

	upload := &domain.TusUpload{
		ID:             uploadID,
		UserID:         userID,
		ProjectID:      &projectID,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.UploadStatusPending,
		UploadType:     domain.UploadTypeProjectUpdate,
		UploadURL:      uploadURL,
		UploadMetadata: metadata,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := uc.tusUploadRepo.Create(upload); err != nil {
		return nil, errors.New("gagal membuat record upload")
	}

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadata); err != nil {
		uc.tusUploadRepo.Delete(uploadID)
		return nil, errors.New("gagal menginisiasi file upload")
	}

	uc.tusManager.AddToQueue(uploadID)

	response := &domain.TusUploadResponse{
		UploadID:   uploadID,
		UploadURL:  uploadURL,
		Offset:     0,
		Length:     fileSize,
	}

	return response, nil
}

func (uc *tusUploadUsecase) HandleProjectUpdateChunk(projectID uint, uploadID string, userID uint, offset int64, chunk io.Reader, chunkSize int64) (int64, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, errors.New("upload tidak ditemukan")
		}
		return 0, errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return 0, errors.New("upload tidak terkait dengan project ini")
	}

	if upload.Status != domain.UploadStatusUploading && upload.Status != domain.UploadStatusPending {
		return 0, errors.New("upload tidak aktif")
	}

	if !uc.tusManager.IsActiveUpload(uploadID) && !uc.tusManager.CanAcceptUpload() {
		return 0, errors.New("upload tidak aktif")
	}

	if upload.ProjectID == nil {
		return 0, errors.New("project tidak ditemukan")
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return 0, fmt.Errorf("gagal menulis chunk: %v", err)
	}

	upload.CurrentOffset = newOffset
	upload.UpdatedAt = time.Now()
	
	if err := uc.tusUploadRepo.UpdateOffsetOnly(uploadID, newOffset); err != nil {
		return 0, errors.New("gagal update offset")
	}

	if newOffset >= upload.FileSize {
		upload.Status = domain.UploadStatusCompleted
		upload.CompletedAt = &time.Time{}
		*upload.CompletedAt = time.Now()
		
		if err := uc.tusUploadRepo.UpdateUpload(upload); err != nil {
			return newOffset, errors.New("gagal update status upload")
		}

		if err := uc.completeProjectUpdateUpload(upload, projectID, userID); err != nil {
			upload.Status = domain.UploadStatusFailed
			uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusFailed)
			return newOffset, fmt.Errorf("gagal menyelesaikan upload project: %v", err)
		}

		uc.tusManager.RemoveFromQueue(uploadID)
	}

	return newOffset, nil
}

func (uc *tusUploadUsecase) completeProjectUpdateUpload(upload *domain.TusUpload, projectID uint, userID uint) error {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		return errors.New("gagal mengambil data project")
	}

	if project.UserID != userID {
		return errors.New("tidak memiliki akses ke project ini")
	}

	oldFilePath := project.PathFile

	tempFilePath := uc.tusStore.GetFilePath(upload.ID)

	userDir := filepath.Dir(oldFilePath)
	filename := filepath.Base(tempFilePath)
	finalFilePath := filepath.Join(userDir, filename)

	if err := uc.tusManager.FinalizeUpload(upload.ID, finalFilePath); err != nil {
		return errors.New("gagal memindahkan file ke lokasi final")
	}

	project.PathFile = finalFilePath
	project.Ukuran = helper.GetFileSizeFromPath(finalFilePath)

	if upload.UploadMetadata.NamaProject != "" {
		project.NamaProject = upload.UploadMetadata.NamaProject
	}

	if upload.UploadMetadata.Kategori != "" {
		project.Kategori = upload.UploadMetadata.Kategori
	}

	if upload.UploadMetadata.Semester > 0 {
		project.Semester = upload.UploadMetadata.Semester
	}

	project.UpdatedAt = time.Now()

	if err := uc.projectRepo.Update(project); err != nil {
		helper.DeleteFile(finalFilePath)
		return errors.New("gagal mengupdate data project")
	}

	helper.DeleteFile(oldFilePath)

	return nil
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadStatus(projectID uint, uploadID string, userID uint) (int64, int64, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, errors.New("upload tidak ditemukan")
		}
		return 0, 0, errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return 0, 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return 0, 0, errors.New("upload tidak terkait dengan project ini")
	}

	return upload.CurrentOffset, upload.FileSize, nil
}

func (uc *tusUploadUsecase) GetProjectUpdateUploadInfo(projectID uint, uploadID string, userID uint) (*domain.TusUploadInfoResponse, error) {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("upload tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return nil, errors.New("upload tidak terkait dengan project ini")
	}

	response := &domain.TusUploadInfoResponse{
		UploadID:   upload.ID,
		Status:     string(upload.Status),
		Offset:     upload.CurrentOffset,
		Length:     upload.FileSize,
		Progress:   float64(upload.CurrentOffset) / float64(upload.FileSize) * 100,
		CreatedAt:  upload.CreatedAt,
		UpdatedAt:  upload.UpdatedAt,
	}

	if upload.ProjectID != nil {
		response.ProjectID = *upload.ProjectID
	}

	metadata := upload.UploadMetadata
	response.NamaProject = metadata.NamaProject
	response.Kategori = metadata.Kategori
	response.Semester = metadata.Semester

	return response, nil
}

func (uc *tusUploadUsecase) CancelProjectUpdateUpload(projectID uint, uploadID string, userID uint) error {
	upload, err := uc.tusUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("upload tidak ditemukan")
		}
		return errors.New("gagal mengambil data upload")
	}

	if upload.UserID != userID {
		return errors.New("tidak memiliki akses ke upload ini")
	}

	if upload.ProjectID == nil || *upload.ProjectID != projectID {
		return errors.New("upload tidak terkait dengan project ini")
	}

	if upload.Status == domain.UploadStatusCompleted {
		return errors.New("upload sudah selesai dan tidak bisa dibatalkan")
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		return errors.New("gagal menghapus file upload")
	}

	if err := uc.tusUploadRepo.UpdateStatus(uploadID, domain.UploadStatusCancelled); err != nil {
		return errors.New("gagal update status upload")
	}

	uc.tusManager.RemoveFromQueue(uploadID)

	return nil
}
