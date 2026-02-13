package usecase

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TusModulUsecase interface {
	InitiateModulUpload(userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error)
	HandleModulChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetModulUploadInfo(uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error)
	GetModulUploadStatus(uploadID string, userID string) (int64, int64, error)
	CancelModulUpload(uploadID string, userID string) error
	CheckModulUploadSlot(userID string) (*domain.TusModulUploadSlotResponse, error)
	InitiateModulUpdateUpload(modulID string, userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error)
	HandleModulUpdateChunk(modulID string, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetModulUpdateUploadStatus(modulID string, uploadID string, userID string) (int64, int64, error)
	GetModulUpdateUploadInfo(modulID string, uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error)
	CancelModulUpdateUpload(modulID string, uploadID string, userID string) error
}

type tusModulUsecase struct {
	tusModulUploadRepo repo.TusModulUploadRepository
	modulRepo          repo.ModulRepository
	tusManager         *helper.TusManager
	fileManager        *helper.FileManager
	config             *config.Config
}

func NewTusModulUsecase(
	tusModulUploadRepo repo.TusModulUploadRepository,
	modulRepo repo.ModulRepository,
	tusManager *helper.TusManager,
	fileManager *helper.FileManager,
	config *config.Config,
) TusModulUsecase {
	return &tusModulUsecase{
		tusModulUploadRepo: tusModulUploadRepo,
		modulRepo:          modulRepo,
		tusManager:         tusManager,
		fileManager:        fileManager,
		config:             config,
	}
}

func (uc *tusModulUsecase) CheckModulUploadSlot(userID string) (*domain.TusModulUploadSlotResponse, error) {
	activeCount, err := uc.tusModulUploadRepo.CountActiveByUserID(userID)
	if err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal mengecek slot upload: %w", err))
	}

	maxQueue := int64(uc.config.Upload.MaxQueueModulPerUser)
	available := activeCount < maxQueue

	response := &domain.TusModulUploadSlotResponse{
		Available:   available,
		QueueLength: int(activeCount),
		MaxQueue:    uc.config.Upload.MaxQueueModulPerUser,
	}

	if available {
		response.Message = fmt.Sprintf("Slot tersedia (%d/%d)", activeCount, maxQueue)
	} else {
		response.Message = fmt.Sprintf("Antrian penuh (%d/%d), silakan tunggu upload selesai", activeCount, maxQueue)
	}

	return response, nil
}

func (uc *tusModulUsecase) InitiateModulUpload(userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	return uc.initiateUpload(userID, fileSize, uploadMetadata, domain.UploadTypeModulCreate, nil)
}

func (uc *tusModulUsecase) InitiateModulUpdateUpload(modulID string, userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Modul")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal mengambil data modul: %w", err))
	}

	if modul.UserID != userID {
		return nil, apperrors.NewForbiddenError("Anda tidak memiliki akses ke modul ini")
	}

	return uc.initiateUpload(userID, fileSize, uploadMetadata, domain.UploadTypeModulUpdate, &modulID)
}

func (uc *tusModulUsecase) initiateUpload(userID string, fileSize int64, uploadMetadata string, uploadType string, modulID *string) (*domain.TusModulUploadResponse, error) {
	slotCheck, err := uc.CheckModulUploadSlot(userID)
	if err != nil {
		return nil, err
	}
	if !slotCheck.Available {
		return nil, apperrors.NewConflictError(fmt.Sprintf("antrian penuh: %s", slotCheck.Message))
	}

	if err := uc.validateModulFileSize(fileSize); err != nil {
		return nil, err
	}

	metadata, err := uc.parseModulMetadata(uploadMetadata)
	if err != nil {
		return nil, err
	}

	uploadID := uuid.New().String()
	uploadURL := fmt.Sprintf("/modul/upload/%s", uploadID)
	if uploadType == domain.UploadTypeModulUpdate && modulID != nil {
		uploadURL = fmt.Sprintf("/modul/%s/update/%s", *modulID, uploadID)
	}

	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Second)
	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		ModulID:        modulID,
		UploadType:     uploadType,
		UploadURL:      uploadURL,
		UploadMetadata: *metadata,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.UploadStatusPending,
		Progress:       0,
		ExpiresAt:      expiresAt,
	}

	if err := uc.tusModulUploadRepo.Create(tusUpload); err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal membuat upload record: %w", err))
	}

	metadataMap := map[string]string{
		"judul":     metadata.Judul,
		"deskripsi": metadata.Deskripsi,
		"user_id":   userID,
	}
	if modulID != nil {
		metadataMap["modul_id"] = *modulID
	}

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		_ = uc.tusModulUploadRepo.Delete(uploadID)
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal inisiasi TUS upload: %w", err))
	}

	return &domain.TusModulUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusModulUsecase) HandleModulChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(uploadID, userID, offset, chunk, nil)
}

func (uc *tusModulUsecase) HandleModulUpdateChunk(modulID string, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(uploadID, userID, offset, chunk, &modulID)
}

func (uc *tusModulUsecase) handleChunk(uploadID string, userID string, offset int64, chunk io.Reader, modulID *string) (int64, error) {
	tusUpload, err := uc.getOwnedUpload(uploadID, userID, modulID)
	if err != nil {
		return 0, err
	}

	if tusUpload.Status != domain.UploadStatusPending && tusUpload.Status != domain.UploadStatusUploading {
		if tusUpload.Status == domain.UploadStatusCompleted {
			return tusUpload.FileSize, apperrors.NewTusCompletedError()
		}
		return 0, apperrors.NewTusInactiveError()
	}

	if offset != tusUpload.CurrentOffset {
		return tusUpload.CurrentOffset, apperrors.NewTusOffsetError(tusUpload.CurrentOffset, offset)
	}

	if tusUpload.Status == domain.UploadStatusPending {
		if err := uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.UploadStatusUploading); err != nil {
			return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
		}
		tusUpload.Status = domain.UploadStatusUploading
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("gagal menulis chunk: %w", err))
	}

	progress := float64(newOffset) / float64(tusUpload.FileSize) * 100
	if err := uc.tusModulUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, apperrors.NewInternalError(fmt.Errorf("gagal update offset: %w", err))
	}

	tusUpload.CurrentOffset = newOffset
	tusUpload.Progress = progress
	if newOffset >= tusUpload.FileSize {
		if err := uc.completeUpload(tusUpload, userID); err != nil {
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusModulUsecase) completeUpload(tusUpload *domain.TusModulUpload, userID string) error {
	dirPath, randomDir, err := uc.fileManager.CreateModulUploadDirectory(userID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal membuat direktori modul: %w", err))
	}

	fileName := fmt.Sprintf("%s.dat", tusUpload.ID)
	finalPath := filepath.Join(dirPath, fileName)
	if err := uc.tusManager.FinalizeUpload(tusUpload.ID, finalPath); err != nil {
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return apperrors.NewInternalError(fmt.Errorf("gagal finalisasi upload: %w", err))
	}

	var modulID string
	switch tusUpload.UploadType {
	case domain.UploadTypeModulCreate:
		modulID, err = uc.completeModulCreate(tusUpload, userID, fileName, finalPath, randomDir)
		if err != nil {
			return err
		}
	case domain.UploadTypeModulUpdate:
		modulID, err = uc.completeModulUpdate(tusUpload, userID, fileName, finalPath, randomDir)
		if err != nil {
			return err
		}
	default:
		return apperrors.NewValidationError("tipe upload tidak didukung", nil)
	}

	if err := uc.tusModulUploadRepo.Complete(tusUpload.ID, modulID, finalPath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
	}

	return nil
}

func (uc *tusModulUsecase) completeModulCreate(tusUpload *domain.TusModulUpload, userID string, fileName string, finalPath string, randomDir string) (string, error) {
	modul := &domain.Modul{
		UserID:    userID,
		Judul:     tusUpload.UploadMetadata.Judul,
		Deskripsi: tusUpload.UploadMetadata.Deskripsi,
		FileName:  fileName,
		FilePath:  finalPath,
		FileSize:  tusUpload.FileSize,
		MimeType:  detectMimeType(finalPath),
		Status:    "completed",
	}

	if err := uc.modulRepo.Create(modul); err != nil {
		_ = helper.DeleteFile(finalPath)
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return "", apperrors.NewInternalError(fmt.Errorf("gagal menyimpan data modul: %w", err))
	}

	return modul.ID, nil
}

func (uc *tusModulUsecase) completeModulUpdate(tusUpload *domain.TusModulUpload, userID string, fileName string, finalPath string, randomDir string) (string, error) {
	if tusUpload.ModulID == nil {
		return "", apperrors.NewValidationError("modul_id tidak ditemukan dalam upload record", nil)
	}

	modul, err := uc.modulRepo.GetByID(*tusUpload.ModulID)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("gagal mengambil data modul: %w", err))
	}

	oldFilePath := modul.FilePath
	modul.Judul = tusUpload.UploadMetadata.Judul
	modul.Deskripsi = tusUpload.UploadMetadata.Deskripsi
	modul.FilePath = finalPath
	modul.FileName = fileName
	modul.FileSize = tusUpload.FileSize
	modul.MimeType = detectMimeType(finalPath)

	if err := uc.modulRepo.Update(modul); err != nil {
		_ = helper.DeleteFile(finalPath)
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return "", apperrors.NewInternalError(fmt.Errorf("gagal update data modul: %w", err))
	}

	_ = helper.DeleteFile(oldFilePath)
	return modul.ID, nil
}

func (uc *tusModulUsecase) GetModulUploadInfo(uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error) {
	return uc.getUploadInfo(uploadID, userID, nil)
}

func (uc *tusModulUsecase) GetModulUpdateUploadInfo(modulID string, uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error) {
	return uc.getUploadInfo(uploadID, userID, &modulID)
}

func (uc *tusModulUsecase) getUploadInfo(uploadID string, userID string, modulID *string) (*domain.TusModulUploadInfoResponse, error) {
	tusUpload, err := uc.getOwnedUpload(uploadID, userID, modulID)
	if err != nil {
		return nil, err
	}

	response := &domain.TusModulUploadInfoResponse{
		UploadID:  tusUpload.ID,
		Judul:     tusUpload.UploadMetadata.Judul,
		Deskripsi: tusUpload.UploadMetadata.Deskripsi,
		Status:    tusUpload.Status,
		Progress:  tusUpload.Progress,
		Offset:    tusUpload.CurrentOffset,
		Length:    tusUpload.FileSize,
		CreatedAt: tusUpload.CreatedAt,
		UpdatedAt: tusUpload.UpdatedAt,
	}

	if tusUpload.ModulID != nil {
		response.ModulID = *tusUpload.ModulID
	}

	return response, nil
}

func (uc *tusModulUsecase) GetModulUploadStatus(uploadID string, userID string) (int64, int64, error) {
	return uc.getUploadStatus(uploadID, userID, nil)
}

func (uc *tusModulUsecase) GetModulUpdateUploadStatus(modulID string, uploadID string, userID string) (int64, int64, error) {
	return uc.getUploadStatus(uploadID, userID, &modulID)
}

func (uc *tusModulUsecase) getUploadStatus(uploadID string, userID string, modulID *string) (int64, int64, error) {
	tusUpload, err := uc.getOwnedUpload(uploadID, userID, modulID)
	if err != nil {
		return 0, 0, err
	}

	return tusUpload.CurrentOffset, tusUpload.FileSize, nil
}

func (uc *tusModulUsecase) CancelModulUpload(uploadID string, userID string) error {
	return uc.cancelUpload(uploadID, userID, nil)
}

func (uc *tusModulUsecase) CancelModulUpdateUpload(modulID string, uploadID string, userID string) error {
	return uc.cancelUpload(uploadID, userID, &modulID)
}

func (uc *tusModulUsecase) cancelUpload(uploadID string, userID string, modulID *string) error {
	tusUpload, err := uc.getOwnedUpload(uploadID, userID, modulID)
	if err != nil {
		return err
	}

	if tusUpload.Status == domain.UploadStatusCompleted {
		return apperrors.NewTusCompletedError()
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal membatalkan upload: %w", err))
	}

	if err := uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.UploadStatusCancelled); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal update status: %w", err))
	}

	return nil
}

func (uc *tusModulUsecase) getOwnedUpload(uploadID string, userID string, modulID *string) (*domain.TusModulUpload, error) {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Upload")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	if tusUpload.UserID != userID {
		return nil, apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	if modulID != nil {
		if tusUpload.ModulID == nil || *tusUpload.ModulID != *modulID {
			return nil, apperrors.NewValidationError("modul ID tidak cocok", nil)
		}
	}

	return tusUpload, nil
}

func (uc *tusModulUsecase) parseModulMetadata(metadataHeader string) (*domain.TusModulUploadInitRequest, error) {
	if metadataHeader == "" {
		return nil, apperrors.NewValidationError("metadata wajib diisi", nil)
	}

	metadataMap := helper.ParseTusMetadata(metadataHeader)
	judul, ok := metadataMap["judul"]
	if !ok || judul == "" {
		return nil, apperrors.NewValidationError("judul wajib diisi", nil)
	}

	if len(judul) < 3 || len(judul) > 255 {
		return nil, apperrors.NewValidationError("judul harus antara 3-255 karakter", nil)
	}

	return &domain.TusModulUploadInitRequest{
		Judul:     judul,
		Deskripsi: metadataMap["deskripsi"],
	}, nil
}

func (uc *tusModulUsecase) validateModulFileSize(fileSize int64) error {
	if fileSize <= 0 {
		return apperrors.NewValidationError("ukuran file tidak valid", nil)
	}

	maxSize := uc.config.Upload.MaxSizeModul
	if fileSize > maxSize {
		return apperrors.NewPayloadTooLargeError(fmt.Sprintf("ukuran file melebihi batas maksimal %d MB", maxSize/(1024*1024)))
	}

	return nil
}

func detectMimeType(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return "application/octet-stream"
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil {
		return "application/octet-stream"
	}

	return http.DetectContentType(buffer[:n])
}
