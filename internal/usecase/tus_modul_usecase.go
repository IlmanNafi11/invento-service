package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"io"
	"path/filepath"
	"time"

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
	HandleModulUpdateChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
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

	maxQueue := uc.config.Upload.MaxQueueModulPerUser
	available := activeCount < maxQueue

	response := &domain.TusModulUploadSlotResponse{
		Available:   available,
		QueueLength: activeCount,
		MaxQueue:    maxQueue,
	}

	if available {
		response.Message = fmt.Sprintf("Slot tersedia (%d/%d)", activeCount, maxQueue)
	} else {
		response.Message = fmt.Sprintf("Antrian penuh (%d/%d), silakan tunggu upload selesai", activeCount, maxQueue)
	}

	return response, nil
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

	deskripsi := metadataMap["deskripsi"]

	return &domain.TusModulUploadInitRequest{
		Judul:     judul,
		Deskripsi: deskripsi,
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

func (uc *tusModulUsecase) InitiateModulUpload(userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
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
	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Second)

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		UploadType:     domain.ModulUploadTypeCreate,
		UploadURL:      uploadURL,
		UploadMetadata: *metadata,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusPending,
		Progress:       0,
		ExpiresAt:      expiresAt,
	}

	if err := uc.tusModulUploadRepo.Create(tusUpload); err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal membuat upload record: %w", err))
	}

	metadataMap := make(map[string]string)
	metadataMap["judul"] = metadata.Judul
	metadataMap["deskripsi"] = metadata.Deskripsi
	metadataMap["user_id"] = userID

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		uc.tusModulUploadRepo.Delete(uploadID)
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
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.NewNotFoundError("Upload")
		}
		return 0, apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	if tusUpload.UserID != userID {
		return 0, apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	if tusUpload.Status != domain.ModulUploadStatusPending && tusUpload.Status != domain.ModulUploadStatusUploading {
		return 0, apperrors.NewTusInactiveError()
	}

	if offset != tusUpload.CurrentOffset {
		return tusUpload.CurrentOffset, apperrors.NewTusOffsetMismatchError(tusUpload.CurrentOffset, offset)
	}

	if tusUpload.Status == domain.ModulUploadStatusPending {
		if err := uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.ModulUploadStatusUploading); err != nil {
			return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
		}
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("gagal menulis chunk: %w", err))
	}

	progress := float64(newOffset) / float64(tusUpload.FileSize) * 100

	if err := uc.tusModulUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, apperrors.NewInternalError(fmt.Errorf("gagal update offset: %w", err))
	}

	if newOffset >= tusUpload.FileSize {
		if err := uc.completeModulUpload(uploadID, userID); err != nil {
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusModulUsecase) completeModulUpload(uploadID string, userID string) error {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	dirPath, randomDir, err := uc.fileManager.CreateModulUploadDirectory(userID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal membuat direktori modul: %w", err))
	}

	// Generate a unique filename using the upload ID
	fileName := fmt.Sprintf("%s.dat", uploadID)
	finalPath := filepath.Join(dirPath, fileName)

	if err := uc.tusManager.FinalizeUpload(uploadID, finalPath); err != nil {
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return apperrors.NewInternalError(fmt.Errorf("gagal finalisasi upload: %w", err))
	}

	modul := &domain.Modul{
		UserID:    userID,
		Judul:     tusUpload.UploadMetadata.Judul,
		Deskripsi: tusUpload.UploadMetadata.Deskripsi,
		FileName:  fileName,
		FilePath:  finalPath,
		FileSize:  tusUpload.FileSize,
		MimeType:  "application/octet-stream", // Default MIME type, can be updated later
		Status:    "completed",
	}

	if err := uc.modulRepo.Create(modul); err != nil {
		helper.DeleteFile(finalPath)
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return apperrors.NewInternalError(fmt.Errorf("gagal menyimpan data modul: %w", err))
	}

	if err := uc.tusModulUploadRepo.Complete(uploadID, modul.ID, finalPath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
	}

	return nil
}

func (uc *tusModulUsecase) GetModulUploadInfo(uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error) {
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
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, apperrors.NewNotFoundError("Upload")
		}
		return 0, 0, apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	if tusUpload.UserID != userID {
		return 0, 0, apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	return tusUpload.CurrentOffset, tusUpload.FileSize, nil
}

func (uc *tusModulUsecase) CancelModulUpload(uploadID string, userID string) error {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Upload")
		}
		return apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	if tusUpload.UserID != userID {
		return apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	if tusUpload.Status == domain.ModulUploadStatusCompleted {
		return apperrors.NewTusAlreadyCompletedError()
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal membatalkan upload: %w", err))
	}

	if err := uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.ModulUploadStatusCancelled); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal update status: %w", err))
	}

	return nil
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
	uploadURL := fmt.Sprintf("/modul/%s/update/%s", modulID, uploadID)
	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Second)

	tusUpload := &domain.TusModulUpload{
		ID:             uploadID,
		UserID:         userID,
		ModulID:        &modulID,
		UploadType:     domain.ModulUploadTypeUpdate,
		UploadURL:      uploadURL,
		UploadMetadata: *metadata,
		FileSize:       fileSize,
		CurrentOffset:  0,
		Status:         domain.ModulUploadStatusPending,
		Progress:       0,
		ExpiresAt:      expiresAt,
	}

	if err := uc.tusModulUploadRepo.Create(tusUpload); err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal membuat upload record: %w", err))
	}

	metadataMap := make(map[string]string)
	metadataMap["judul"] = metadata.Judul
	metadataMap["deskripsi"] = metadata.Deskripsi
	metadataMap["user_id"] = userID
	metadataMap["modul_id"] = modulID

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		uc.tusModulUploadRepo.Delete(uploadID)
		return nil, apperrors.NewInternalError(fmt.Errorf("gagal inisiasi TUS upload: %w", err))
	}

	return &domain.TusModulUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusModulUsecase) HandleModulUpdateChunk(uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, apperrors.NewNotFoundError("Upload")
		}
		return 0, apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	if tusUpload.UserID != userID {
		return 0, apperrors.NewForbiddenError("Anda tidak memiliki akses ke upload ini")
	}

	if tusUpload.Status != domain.ModulUploadStatusPending && tusUpload.Status != domain.ModulUploadStatusUploading {
		return 0, apperrors.NewTusInactiveError()
	}

	if offset != tusUpload.CurrentOffset {
		return tusUpload.CurrentOffset, apperrors.NewTusOffsetMismatchError(tusUpload.CurrentOffset, offset)
	}

	if tusUpload.Status == domain.ModulUploadStatusPending {
		if err := uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.ModulUploadStatusUploading); err != nil {
			return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
		}
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("gagal menulis chunk: %w", err))
	}

	progress := float64(newOffset) / float64(tusUpload.FileSize) * 100

	if err := uc.tusModulUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, apperrors.NewInternalError(fmt.Errorf("gagal update offset: %w", err))
	}

	if newOffset >= tusUpload.FileSize {
		if err := uc.completeModulUpdate(uploadID, userID); err != nil {
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusModulUsecase) completeModulUpdate(uploadID string, userID string) error {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal mengambil data upload: %w", err))
	}

	if tusUpload.ModulID == nil {
		return apperrors.NewValidationError("modul_id tidak ditemukan dalam upload record", nil)
	}

	modul, err := uc.modulRepo.GetByID(*tusUpload.ModulID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal mengambil data modul: %w", err))
	}

	oldFilePath := modul.FilePath

	dirPath, randomDir, err := uc.fileManager.CreateModulUploadDirectory(userID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal membuat direktori modul: %w", err))
	}

	fileName := fmt.Sprintf("%s.dat", uploadID)
	finalPath := filepath.Join(dirPath, fileName)

	if err := uc.tusManager.FinalizeUpload(uploadID, finalPath); err != nil {
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return apperrors.NewInternalError(fmt.Errorf("gagal finalisasi upload: %w", err))
	}

	modul.Judul = tusUpload.UploadMetadata.Judul
	modul.Deskripsi = tusUpload.UploadMetadata.Deskripsi
	modul.FilePath = finalPath
	modul.FileName = fileName
	modul.FileSize = tusUpload.FileSize

	if err := uc.modulRepo.Update(modul); err != nil {
		helper.DeleteFile(finalPath)
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return apperrors.NewInternalError(fmt.Errorf("gagal update data modul: %w", err))
	}

	if err := helper.DeleteFile(oldFilePath); err != nil {
		return nil
	}

	if err := uc.tusModulUploadRepo.Complete(uploadID, modul.ID, finalPath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("gagal update status upload: %w", err))
	}

	return nil
}
