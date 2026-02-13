package helper

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"strconv"

	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
)

type TusManager struct {
	store       *TusStore
	queue       *TusQueue
	fileManager *FileManager
	config      *config.Config
}

func NewTusManager(store *TusStore, queue *TusQueue, fileManager *FileManager, config *config.Config) *TusManager {
	return &TusManager{
		store:       store,
		queue:       queue,
		fileManager: fileManager,
		config:      config,
	}
}

func (tm *TusManager) CheckUploadSlot() *domain.TusUploadSlotResponse {
	hasActiveUpload := tm.queue.HasActiveUpload()
	queueLength := tm.queue.GetQueueLength()
	canAccept := tm.queue.CanAcceptUpload()

	response := &domain.TusUploadSlotResponse{
		Available:     canAccept,
		QueueLength:   queueLength,
		ActiveUpload:  hasActiveUpload,
		MaxConcurrent: tm.config.Upload.MaxConcurrentProject,
	}

	if canAccept {
		response.Message = "Slot upload tersedia"
	} else {
		response.Message = fmt.Sprintf("Slot upload tidak tersedia, ada %d upload dalam antrian", queueLength)
	}

	return response
}

func (tm *TusManager) ResetUploadQueue() error {
	activeUploadIDs := tm.queue.GetActiveUploads()

	for _, activeUploadID := range activeUploadIDs {
		if err := tm.store.Terminate(activeUploadID); err != nil {
			log.Printf("Warning: gagal menghapus upload aktif %s: %v", activeUploadID, err)
		}
	}

	tm.queue.Clear()

	return nil
}

func (tm *TusManager) ValidateProjectMetadata(metadata map[string]string) error {
	if namaProject, ok := metadata["nama_project"]; !ok || namaProject == "" {
		return apperrors.NewValidationError("nama_project wajib diisi", nil)
	}

	if len(metadata["nama_project"]) < 3 || len(metadata["nama_project"]) > 255 {
		return apperrors.NewValidationError("nama_project harus antara 3-255 karakter", nil)
	}

	if kategori, ok := metadata["kategori"]; ok && kategori != "" {
		validKategori := []string{"website", "mobile", "iot", "machine_learning", "deep_learning"}
		isValid := false
		for _, valid := range validKategori {
			if kategori == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return apperrors.NewValidationError("kategori tidak valid", nil)
		}
	} else {
		metadata["kategori"] = "website"
	}

	if semesterStr, ok := metadata["semester"]; ok && semesterStr != "" {
		semester, err := strconv.Atoi(semesterStr)
		if err != nil || semester < 1 || semester > 8 {
			return apperrors.NewValidationError("semester harus antara 1-8", nil)
		}
		metadata["semester"] = semesterStr
	}

	return nil
}

func (tm *TusManager) ValidateModulMetadata(metadata map[string]string) error {
	if namaFile, ok := metadata["nama_file"]; !ok || namaFile == "" {
		return apperrors.NewValidationError("nama_file wajib diisi", nil)
	}

	if len(metadata["nama_file"]) < 3 || len(metadata["nama_file"]) > 255 {
		return apperrors.NewValidationError("nama_file harus antara 3-255 karakter", nil)
	}

	if tipe, ok := metadata["tipe"]; ok && tipe != "" {
		validTipe := []string{"docx", "xlsx", "pdf", "pptx"}
		isValid := false
		for _, valid := range validTipe {
			if tipe == valid {
				isValid = true
				break
			}
		}
		if !isValid {
			return apperrors.NewValidationError("tipe file tidak valid, harus salah satu dari: docx, xlsx, pdf, pptx", nil)
		}
	} else {
		return apperrors.NewValidationError("tipe wajib diisi", nil)
	}

	if semesterStr, ok := metadata["semester"]; ok && semesterStr != "" {
		semester, err := strconv.Atoi(semesterStr)
		if err != nil || semester < 1 || semester > 8 {
			return apperrors.NewValidationError("semester harus antara 1-8", nil)
		}
		metadata["semester"] = semesterStr
	}

	return nil
}

func (tm *TusManager) InitiateUpload(uploadID string, fileSize int64, metadata map[string]string) error {
	if fileSize > tm.config.Upload.MaxSize {
		return apperrors.NewPayloadTooLargeError(fmt.Sprintf("ukuran file melebihi batas maksimal %d MB", tm.config.Upload.MaxSize/(1024*1024)))
	}

	if fileSize <= 0 {
		return apperrors.NewValidationError("ukuran file tidak valid", nil)
	}

	info := TusFileInfo{
		ID:       uploadID,
		Size:     fileSize,
		Offset:   0,
		Metadata: metadata,
	}

	return tm.store.NewUpload(info)
}

func (tm *TusManager) HandleChunk(uploadID string, offset int64, chunk io.Reader) (int64, error) {
	return tm.store.WriteChunk(uploadID, offset, chunk)
}

func (tm *TusManager) GetUploadStatus(uploadID string) (int64, int64, error) {
	info, err := tm.store.GetInfo(uploadID)
	if err != nil {
		return 0, 0, err
	}

	return info.Offset, info.Size, nil
}

func (tm *TusManager) GetUploadInfo(uploadID string) (TusFileInfo, error) {
	return tm.store.GetInfo(uploadID)
}

func (tm *TusManager) CancelUpload(uploadID string) error {
	return tm.store.Terminate(uploadID)
}

func (tm *TusManager) FinalizeUpload(uploadID string, finalPath string) error {
	return tm.store.FinalizeUpload(uploadID, finalPath)
}

func (tm *TusManager) IsUploadComplete(uploadID string) (bool, error) {
	return tm.store.IsComplete(uploadID)
}

func (tm *TusManager) GetUploadProgress(uploadID string) (float64, error) {
	return tm.store.GetProgress(uploadID)
}

func (tm *TusManager) AddToQueue(uploadID string) {
	tm.queue.Add(uploadID)
}

func (tm *TusManager) RemoveFromQueue(uploadID string) error {
	return tm.queue.Remove(uploadID)
}

func (tm *TusManager) CanAcceptUpload() bool {
	return tm.queue.CanAcceptUpload()
}

func (tm *TusManager) IsActiveUpload(uploadID string) bool {
	return tm.queue.IsActiveUpload(uploadID)
}

func (tm *TusManager) ReadChunkFromBody(body []byte, expectedSize int64) (io.Reader, error) {
	if int64(len(body)) != expectedSize {
		return nil, apperrors.NewValidationError(fmt.Sprintf("ukuran body tidak cocok dengan Content-Length: expected %d, actual %d", expectedSize, len(body)), nil)
	}

	if len(body) == 0 {
		return nil, apperrors.NewValidationError("body kosong", nil)
	}

	return bytes.NewReader(body), nil
}

func (tm *TusManager) GetDefaultTusHeaders() map[string]string {
	return map[string]string{
		"Tus-Resumable": tm.config.Upload.TusVersion,
	}
}

func (tm *TusManager) ValidateFileSize(fileSize int64, maxSize int64) error {
	if fileSize <= 0 {
		return apperrors.NewValidationError("ukuran file tidak valid", nil)
	}

	if fileSize > maxSize {
		maxSizeMB := maxSize / (1024 * 1024)
		return apperrors.NewPayloadTooLargeError(fmt.Sprintf("ukuran file melebihi batas maksimal %d MB", maxSizeMB))
	}

	return nil
}

func (tm *TusManager) ValidateTusVersion(version string) error {
	if version != tm.config.Upload.TusVersion {
		return apperrors.NewTusVersionError(tm.config.Upload.TusVersion)
	}

	return nil
}

func (tm *TusManager) ValidateOffset(uploadID string, clientOffset int64) (int64, error) {
	serverOffset, err := tm.store.GetOffset(uploadID)
	if err != nil {
		return 0, err
	}

	if clientOffset != serverOffset {
		return serverOffset, apperrors.NewTusOffsetError(serverOffset, clientOffset)
	}

	return serverOffset, nil
}

func (tm *TusManager) ValidateContentType(contentType string) error {
	if contentType != "application/offset+octet-stream" {
		return apperrors.NewValidationError("Content-Type harus application/offset+octet-stream", nil)
	}

	return nil
}
