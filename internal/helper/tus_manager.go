package helper

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"

	"fiber-boiler-plate/config"
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

type UploadSlotResponse struct {
	Available     bool   `json:"available"`
	Message       string `json:"message"`
	QueueLength   int    `json:"queue_length"`
	ActiveUpload  bool   `json:"active_upload"`
	MaxConcurrent int    `json:"max_concurrent"`
}

func (tm *TusManager) CheckUploadSlot() *UploadSlotResponse {
	hasActiveUpload := tm.queue.HasActiveUpload()
	queueLength := tm.queue.GetQueueLength()
	canAccept := tm.queue.CanAcceptUpload()

	response := &UploadSlotResponse{
		Available:     canAccept,
		QueueLength:   queueLength,
		ActiveUpload:  hasActiveUpload,
		MaxConcurrent: tm.config.Upload.MaxConcurrent,
	}

	if canAccept {
		response.Message = "Slot upload tersedia"
	} else {
		response.Message = fmt.Sprintf("Slot upload tidak tersedia, ada %d upload dalam antrian", queueLength)
	}

	return response
}

func (tm *TusManager) ResetUploadQueue() error {
	activeUploadID := tm.queue.GetActiveUpload()
	
	if activeUploadID != "" {
		if err := tm.store.Terminate(activeUploadID); err != nil {
			log.Printf("Warning: gagal menghapus upload aktif: %v", err)
		}
	}
	
	tm.queue.Clear()
	
	return nil
}

func (tm *TusManager) ParseMetadata(metadataHeader string) (map[string]string, error) {
	if metadataHeader == "" {
		return make(map[string]string), nil
	}

	metadataMap := make(map[string]string)
	pairs := strings.Split(metadataHeader, ",")

	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), " ", 2)
		if len(parts) != 2 {
			continue
		}

		key := parts[0]
		valueB64 := parts[1]

		value, err := base64.StdEncoding.DecodeString(valueB64)
		if err != nil {
			return nil, fmt.Errorf("error decoding metadata key %s: %w", key, err)
		}

		metadataMap[key] = string(value)
	}

	return metadataMap, nil
}

func (tm *TusManager) ValidateMetadata(metadata map[string]string) error {
	if namaProject, ok := metadata["nama_project"]; !ok || namaProject == "" {
		return errors.New("nama_project wajib diisi")
	}

	if len(metadata["nama_project"]) < 3 || len(metadata["nama_project"]) > 255 {
		return errors.New("nama_project harus antara 3-255 karakter")
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
			return errors.New("kategori tidak valid")
		}
	} else {
		metadata["kategori"] = "website"
	}

	if semesterStr, ok := metadata["semester"]; ok && semesterStr != "" {
		semester, err := strconv.Atoi(semesterStr)
		if err != nil || semester < 1 || semester > 8 {
			return errors.New("semester harus antara 1-8")
		}
		metadata["semester"] = semesterStr
	}

	return nil
}

func (tm *TusManager) InitiateUpload(uploadID string, fileSize int64, metadata map[string]string) error {
	if fileSize > tm.config.Upload.MaxSize {
		return fmt.Errorf("ukuran file melebihi batas maksimal %d MB", tm.config.Upload.MaxSize/(1024*1024))
	}

	if fileSize <= 0 {
		return errors.New("ukuran file tidak valid")
	}

	return tm.store.InitiateUpload(uploadID, fileSize)
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
		return nil, fmt.Errorf("ukuran body tidak cocok dengan Content-Length: expected %d, actual %d", expectedSize, len(body))
	}

	if len(body) == 0 {
		return nil, errors.New("body kosong")
	}

	return bytes.NewReader(body), nil
}

func (tm *TusManager) ExtractUserIDFromMetadata(metadata map[string]string) (uint, error) {
	userIDStr, ok := metadata["user_id"]
	if !ok {
		return 0, errors.New("user_id tidak ditemukan dalam metadata")
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("user_id tidak valid: %w", err)
	}

	return uint(userID), nil
}

func (tm *TusManager) GetDefaultTusHeaders() map[string]string {
	return map[string]string{
		"Tus-Resumable": tm.config.Upload.TusVersion,
	}
}

func (tm *TusManager) RespondWithTusHeaders(c interface{}, statusCode int, headers map[string]string) {
	type Context interface {
		Set(key string, value string)
		Status(statusCode int) error
		SendStatus(statusCode int) error
	}

	if ctx, ok := c.(Context); ok {
		for key, value := range headers {
			ctx.Set(key, value)
		}
		ctx.SendStatus(statusCode)
	}
}
