package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
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
	"gorm.io/gorm"
)

type TusModulUsecase interface {
	InitiateModulUpload(ctx context.Context, userID string, fileSize int64, uploadMetadata string) (*dto.TusModulUploadResponse, error)
	HandleModulChunk(ctx context.Context, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetModulUploadInfo(ctx context.Context, uploadID string, userID string) (*dto.TusModulUploadInfoResponse, error)
	GetModulUploadStatus(ctx context.Context, uploadID string, userID string) (int64, int64, error)
	CancelModulUpload(ctx context.Context, uploadID string, userID string) error
	CheckModulUploadSlot(ctx context.Context, userID string) (*dto.TusModulUploadSlotResponse, error)
	InitiateModulUpdateUpload(ctx context.Context, modulID string, userID string, fileSize int64, uploadMetadata string) (*dto.TusModulUploadResponse, error)
	HandleModulUpdateChunk(ctx context.Context, modulID string, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error)
	GetModulUpdateUploadStatus(ctx context.Context, modulID string, uploadID string, userID string) (int64, int64, error)
	GetModulUpdateUploadInfo(ctx context.Context, modulID string, uploadID string, userID string) (*dto.TusModulUploadInfoResponse, error)
	CancelModulUpdateUpload(ctx context.Context, modulID string, uploadID string, userID string) error
}

type tusModulUsecase struct {
	tusModulUploadRepo repo.TusModulUploadRepository
	modulRepo          repo.ModulRepository
	tusManager         *upload.TusManager
	fileManager        *storage.FileManager
	config             *config.Config
}

func NewTusModulUsecase(
	tusModulUploadRepo repo.TusModulUploadRepository,
	modulRepo repo.ModulRepository,
	tusManager *upload.TusManager,
	fileManager *storage.FileManager,
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

func (uc *tusModulUsecase) CheckModulUploadSlot(ctx context.Context, userID string) (*dto.TusModulUploadSlotResponse, error) {
	activeCount, err := uc.tusModulUploadRepo.CountActiveByUserID(ctx, userID)
	if err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.CheckModulUploadSlot: %w", err))
	}

	maxQueue := int64(uc.config.Upload.MaxQueueModulPerUser)
	available := activeCount < maxQueue

	response := &dto.TusModulUploadSlotResponse{
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

func (uc *tusModulUsecase) InitiateModulUpload(ctx context.Context, userID string, fileSize int64, uploadMetadata string) (*dto.TusModulUploadResponse, error) {
	return uc.initiateUpload(ctx, userID, fileSize, uploadMetadata, domain.UploadTypeModulCreate, nil)
}

func (uc *tusModulUsecase) InitiateModulUpdateUpload(ctx context.Context, modulID string, userID string, fileSize int64, uploadMetadata string) (*dto.TusModulUploadResponse, error) {
	modul, err := uc.modulRepo.GetByID(ctx, modulID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Modul")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.InitiateModulUpdateUpload: %w", err))
	}

	if modul.UserID != userID {
		return nil, apperrors.NewForbiddenError("Anda tidak memiliki akses ke modul ini")
	}

	return uc.initiateUpload(ctx, userID, fileSize, uploadMetadata, domain.UploadTypeModulUpdate, &modulID)
}

func (uc *tusModulUsecase) initiateUpload(ctx context.Context, userID string, fileSize int64, uploadMetadata string, uploadType string, modulID *string) (*dto.TusModulUploadResponse, error) {
	slotCheck, err := uc.CheckModulUploadSlot(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !slotCheck.Available {
		return nil, apperrors.NewConflictError(fmt.Sprintf("antrian penuh: %s", slotCheck.Message))
	}

	if err = uc.validateModulFileSize(fileSize); err != nil { //nolint:gocritic // sloppyReassign conflicts with govet shadow
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
		ID:         uploadID,
		UserID:     userID,
		ModulID:    modulID,
		UploadType: uploadType,
		UploadURL:  uploadURL,
		UploadMetadata: domain.TusModulUploadMetadata{
			Judul:     metadata.Judul,
			Deskripsi: metadata.Deskripsi,
		},
		FileSize:      fileSize,
		CurrentOffset: 0,
		Status:        domain.UploadStatusPending,
		Progress:      0,
		ExpiresAt:     expiresAt,
	}

	if err := uc.tusModulUploadRepo.Create(ctx, tusUpload); err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.initiateUpload: create record: %w", err))
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
		_ = uc.tusModulUploadRepo.Delete(ctx, uploadID)
		return nil, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.initiateUpload: init storage: %w", err))
	}

	return &dto.TusModulUploadResponse{
		UploadID:  uploadID,
		UploadURL: uploadURL,
		Offset:    0,
		Length:    fileSize,
	}, nil
}

func (uc *tusModulUsecase) HandleModulChunk(ctx context.Context, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(ctx, uploadID, userID, offset, chunk, nil)
}

func (uc *tusModulUsecase) HandleModulUpdateChunk(ctx context.Context, modulID string, uploadID string, userID string, offset int64, chunk io.Reader) (int64, error) {
	return uc.handleChunk(ctx, uploadID, userID, offset, chunk, &modulID)
}

func (uc *tusModulUsecase) handleChunk(ctx context.Context, uploadID string, userID string, offset int64, chunk io.Reader, modulID *string) (int64, error) {
	tusUpload, err := uc.getOwnedUpload(ctx, uploadID, userID, modulID)
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
		if err = uc.tusModulUploadRepo.UpdateStatus(ctx, uploadID, domain.UploadStatusUploading); err != nil {
			return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.handleChunk: update status: %w", err))
		}
		tusUpload.Status = domain.UploadStatusUploading
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return tusUpload.CurrentOffset, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.handleChunk: write chunk: %w", err))
	}

	progress := float64(newOffset) / float64(tusUpload.FileSize) * 100
	if err = uc.tusModulUploadRepo.UpdateOffset(ctx, uploadID, newOffset, progress); err != nil {
		return newOffset, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.handleChunk: update offset: %w", err))
	}

	tusUpload.CurrentOffset = newOffset
	tusUpload.Progress = progress
	if newOffset >= tusUpload.FileSize {
		if err = uc.completeUpload(ctx, tusUpload, userID); err != nil { //nolint:gocritic // sloppyReassign conflicts with govet shadow
			return newOffset, err
		}
	}

	return newOffset, nil
}

func (uc *tusModulUsecase) completeUpload(ctx context.Context, tusUpload *domain.TusModulUpload, userID string) error {
	dirPath, randomDir, err := uc.fileManager.CreateModulUploadDirectory(userID)
	if err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.completeUpload: create dir: %w", err))
	}

	fileName := fmt.Sprintf("%s.dat", tusUpload.ID)
	finalPath := filepath.Join(dirPath, fileName)
	if err = uc.tusManager.FinalizeUpload(tusUpload.ID, finalPath); err != nil {
		_ = uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.completeUpload: finalize: %w", err))
	}

	var modulID string
	switch tusUpload.UploadType {
	case domain.UploadTypeModulCreate:
		modulID, err = uc.completeModulCreate(ctx, tusUpload, userID, fileName, finalPath, randomDir)
		if err != nil {
			return err
		}
	case domain.UploadTypeModulUpdate:
		modulID, err = uc.completeModulUpdate(ctx, tusUpload, userID, fileName, finalPath, randomDir)
		if err != nil {
			return err
		}
	default:
		return apperrors.NewValidationError("tipe upload tidak didukung", nil)
	}

	if err := uc.tusModulUploadRepo.Complete(ctx, tusUpload.ID, modulID, finalPath); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.completeUpload: complete record: %w", err))
	}

	return nil
}

func (uc *tusModulUsecase) completeModulCreate(ctx context.Context, tusUpload *domain.TusModulUpload, userID string, fileName string, finalPath string, randomDir string) (string, error) {
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

	if err := uc.modulRepo.Create(ctx, modul); err != nil {
		_ = storage.DeleteFile(finalPath)
		_ = uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return "", apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.completeModulCreate: %w", err))
	}

	return modul.ID, nil
}

func (uc *tusModulUsecase) completeModulUpdate(ctx context.Context, tusUpload *domain.TusModulUpload, userID string, fileName string, finalPath string, randomDir string) (string, error) {
	if tusUpload.ModulID == nil {
		return "", apperrors.NewValidationError("modul_id tidak ditemukan dalam upload record", nil)
	}

	modul, err := uc.modulRepo.GetByID(ctx, *tusUpload.ModulID)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.completeModulUpdate: get modul: %w", err))
	}

	oldFilePath := modul.FilePath
	modul.Judul = tusUpload.UploadMetadata.Judul
	modul.Deskripsi = tusUpload.UploadMetadata.Deskripsi
	modul.FilePath = finalPath
	modul.FileName = fileName
	modul.FileSize = tusUpload.FileSize
	modul.MimeType = detectMimeType(finalPath)

	if err := uc.modulRepo.Update(ctx, modul); err != nil {
		_ = storage.DeleteFile(finalPath)
		_ = uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return "", apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.completeModulUpdate: %w", err))
	}

	if oldFilePath != "" {
		if err := storage.DeleteFile(oldFilePath); err != nil {
			// Old file deletion after successful update is critical but non-blocking
			zlog.Warn().Err(err).Str("file", oldFilePath).Msg("TusModulUsecase.completeModulUpdate: failed to delete old file")
		}
	}
	return modul.ID, nil
}

func (uc *tusModulUsecase) GetModulUploadInfo(ctx context.Context, uploadID string, userID string) (*dto.TusModulUploadInfoResponse, error) {
	return uc.getUploadInfo(ctx, uploadID, userID, nil)
}

func (uc *tusModulUsecase) GetModulUpdateUploadInfo(ctx context.Context, modulID string, uploadID string, userID string) (*dto.TusModulUploadInfoResponse, error) {
	return uc.getUploadInfo(ctx, uploadID, userID, &modulID)
}

func (uc *tusModulUsecase) getUploadInfo(ctx context.Context, uploadID string, userID string, modulID *string) (*dto.TusModulUploadInfoResponse, error) {
	tusUpload, err := uc.getOwnedUpload(ctx, uploadID, userID, modulID)
	if err != nil {
		return nil, err
	}

	response := &dto.TusModulUploadInfoResponse{
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

func (uc *tusModulUsecase) GetModulUploadStatus(ctx context.Context, uploadID string, userID string) (offset int64, length int64, err error) {
	return uc.getUploadStatus(ctx, uploadID, userID, nil)
}

func (uc *tusModulUsecase) GetModulUpdateUploadStatus(ctx context.Context, modulID string, uploadID string, userID string) (offset int64, length int64, err error) {
	return uc.getUploadStatus(ctx, uploadID, userID, &modulID)
}

func (uc *tusModulUsecase) getUploadStatus(ctx context.Context, uploadID string, userID string, modulID *string) (offset int64, length int64, err error) {
	tusUpload, err := uc.getOwnedUpload(ctx, uploadID, userID, modulID)
	if err != nil {
		return 0, 0, err
	}

	return tusUpload.CurrentOffset, tusUpload.FileSize, nil
}

func (uc *tusModulUsecase) CancelModulUpload(ctx context.Context, uploadID string, userID string) error {
	return uc.cancelUpload(ctx, uploadID, userID, nil)
}

func (uc *tusModulUsecase) CancelModulUpdateUpload(ctx context.Context, modulID string, uploadID string, userID string) error {
	return uc.cancelUpload(ctx, uploadID, userID, &modulID)
}

func (uc *tusModulUsecase) cancelUpload(ctx context.Context, uploadID string, userID string, modulID *string) error {
	tusUpload, err := uc.getOwnedUpload(ctx, uploadID, userID, modulID)
	if err != nil {
		return err
	}

	if tusUpload.Status == domain.UploadStatusCompleted {
		return apperrors.NewTusCompletedError()
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.cancelUpload: cancel storage: %w", err))
	}

	if err := uc.tusModulUploadRepo.UpdateStatus(ctx, uploadID, domain.UploadStatusCancelled); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.cancelUpload: update status: %w", err))
	}

	return nil
}

func (uc *tusModulUsecase) getOwnedUpload(ctx context.Context, uploadID string, userID string, modulID *string) (*domain.TusModulUpload, error) {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(ctx, uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Upload")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("TusModulUsecase.getOwnedUpload: %w", err))
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

func (uc *tusModulUsecase) parseModulMetadata(metadataHeader string) (*dto.TusModulUploadInitRequest, error) {
	if metadataHeader == "" {
		return nil, apperrors.NewValidationError("metadata wajib diisi", nil)
	}

	metadataMap := upload.ParseTusMetadata(metadataHeader)
	judul, ok := metadataMap["judul"]
	if !ok || judul == "" {
		return nil, apperrors.NewValidationError("judul wajib diisi", nil)
	}

	if len(judul) < 3 || len(judul) > 255 {
		return nil, apperrors.NewValidationError("judul harus antara 3-255 karakter", nil)
	}

	return &dto.TusModulUploadInitRequest{
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
