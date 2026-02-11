package usecase

import (
	"encoding/base64"
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"io"
	"path/filepath"
	"strconv"
	"strings"
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
	InitiateModulUpdateUpload(modulID uint, userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error)
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
		return nil, fmt.Errorf("gagal mengecek slot upload: %v", err)
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
		return nil, errors.New("metadata wajib diisi")
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

	namaFile, ok := metadataMap["nama_file"]
	if !ok || namaFile == "" {
		return nil, errors.New("nama_file wajib diisi")
	}

	if len(namaFile) < 3 || len(namaFile) > 255 {
		return nil, errors.New("nama_file harus antara 3-255 karakter")
	}

	tipe, ok := metadataMap["tipe"]
	if !ok || tipe == "" {
		return nil, errors.New("tipe file wajib diisi")
	}

	validTipe := []string{"docx", "xlsx", "pdf", "pptx"}
	isValid := false
	for _, valid := range validTipe {
		if tipe == valid {
			isValid = true
			break
		}
	}
	if !isValid {
		return nil, errors.New("tipe file harus salah satu dari: docx, xlsx, pdf, pptx")
	}

	semesterStr, ok := metadataMap["semester"]
	if !ok || semesterStr == "" {
		return nil, errors.New("semester wajib diisi")
	}

	semester, err := strconv.Atoi(semesterStr)
	if err != nil || semester < 1 || semester > 8 {
		return nil, errors.New("semester harus berupa angka antara 1-8")
	}

	return &domain.TusModulUploadInitRequest{
		NamaFile: namaFile,
		Tipe:     tipe,
		Semester: semester,
	}, nil
}

func (uc *tusModulUsecase) validateModulFileSize(fileSize int64) error {
	if fileSize <= 0 {
		return errors.New("ukuran file tidak valid")
	}

	maxSize := uc.config.Upload.MaxSizeModul
	if fileSize > maxSize {
		return fmt.Errorf("ukuran file melebihi batas maksimal %d MB", maxSize/(1024*1024))
	}

	return nil
}

func (uc *tusModulUsecase) InitiateModulUpload(userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	slotCheck, err := uc.CheckModulUploadSlot(userID)
	if err != nil {
		return nil, err
	}

	if !slotCheck.Available {
		return nil, fmt.Errorf("antrian penuh: %s", slotCheck.Message)
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
	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Minute)

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
		return nil, fmt.Errorf("gagal membuat upload record: %v", err)
	}

	metadataMap := make(map[string]string)
	metadataMap["nama_file"] = metadata.NamaFile
	metadataMap["tipe"] = metadata.Tipe
	metadataMap["semester"] = strconv.Itoa(metadata.Semester)
	metadataMap["user_id"] = userID

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		uc.tusModulUploadRepo.Delete(uploadID)
		return nil, fmt.Errorf("gagal inisiasi TUS upload: %v", err)
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
			return 0, errors.New("upload tidak ditemukan")
		}
		return 0, fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	if tusUpload.UserID != userID {
		return 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if tusUpload.Status != domain.ModulUploadStatusPending && tusUpload.Status != domain.ModulUploadStatusUploading {
		return 0, fmt.Errorf("upload tidak aktif (status: %s)", tusUpload.Status)
	}

	if offset != tusUpload.CurrentOffset {
		return tusUpload.CurrentOffset, fmt.Errorf("offset tidak valid, expected %d got %d", tusUpload.CurrentOffset, offset)
	}

	if tusUpload.Status == domain.ModulUploadStatusPending {
		uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.ModulUploadStatusUploading)
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return tusUpload.CurrentOffset, fmt.Errorf("gagal menulis chunk: %v", err)
	}

	progress := float64(newOffset) / float64(tusUpload.FileSize) * 100

	if err := uc.tusModulUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, fmt.Errorf("gagal update offset: %v", err)
	}

	if newOffset >= tusUpload.FileSize {
		if err := uc.completeModulUpload(uploadID, userID); err != nil {
			return newOffset, fmt.Errorf("gagal menyelesaikan upload: %v", err)
		}
	}

	return newOffset, nil
}

func (uc *tusModulUsecase) completeModulUpload(uploadID string, userID string) error {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		return fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	dirPath, randomDir, err := uc.fileManager.CreateModulUploadDirectory(userID)
	if err != nil {
		return fmt.Errorf("gagal membuat direktori modul: %v", err)
	}

	fileName := fmt.Sprintf("%s.%s", tusUpload.UploadMetadata.NamaFile, tusUpload.UploadMetadata.Tipe)
	finalPath := filepath.Join(dirPath, fileName)

	if err := uc.tusManager.FinalizeUpload(uploadID, finalPath); err != nil {
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return fmt.Errorf("gagal finalisasi upload: %v", err)
	}

	fileSize := helper.FormatFileSize(tusUpload.FileSize)

	modul := &domain.Modul{
		UserID:   userID,
		NamaFile: tusUpload.UploadMetadata.NamaFile,
		Tipe:     tusUpload.UploadMetadata.Tipe,
		Ukuran:   fileSize,
		Semester: tusUpload.UploadMetadata.Semester,
		PathFile: finalPath,
	}

	if err := uc.modulRepo.Create(modul); err != nil {
		helper.DeleteFile(finalPath)
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return fmt.Errorf("gagal menyimpan data modul: %v", err)
	}

	if err := uc.tusModulUploadRepo.Complete(uploadID, modul.ID, finalPath); err != nil {
		return fmt.Errorf("gagal update status upload: %v", err)
	}

	return nil
}

func (uc *tusModulUsecase) GetModulUploadInfo(uploadID string, userID string) (*domain.TusModulUploadInfoResponse, error) {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("upload tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	if tusUpload.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke upload ini")
	}

	response := &domain.TusModulUploadInfoResponse{
		UploadID:  tusUpload.ID,
		NamaFile:  tusUpload.UploadMetadata.NamaFile,
		Tipe:      tusUpload.UploadMetadata.Tipe,
		Semester:  tusUpload.UploadMetadata.Semester,
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
			return 0, 0, errors.New("upload tidak ditemukan")
		}
		return 0, 0, fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	if tusUpload.UserID != userID {
		return 0, 0, errors.New("tidak memiliki akses ke upload ini")
	}

	return tusUpload.CurrentOffset, tusUpload.FileSize, nil
}

func (uc *tusModulUsecase) CancelModulUpload(uploadID string, userID string) error {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("upload tidak ditemukan")
		}
		return fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	if tusUpload.UserID != userID {
		return errors.New("tidak memiliki akses ke upload ini")
	}

	if tusUpload.Status == domain.ModulUploadStatusCompleted {
		return errors.New("upload sudah selesai dan tidak bisa dibatalkan")
	}

	if err := uc.tusManager.CancelUpload(uploadID); err != nil {
		return fmt.Errorf("gagal membatalkan upload: %v", err)
	}

	if err := uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.ModulUploadStatusCancelled); err != nil {
		return fmt.Errorf("gagal update status: %v", err)
	}

	return nil
}

func (uc *tusModulUsecase) InitiateModulUpdateUpload(modulID uint, userID string, fileSize int64, uploadMetadata string) (*domain.TusModulUploadResponse, error) {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("modul tidak ditemukan")
		}
		return nil, fmt.Errorf("gagal mengambil data modul: %v", err)
	}

	if modul.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke modul ini")
	}

	slotCheck, err := uc.CheckModulUploadSlot(userID)
	if err != nil {
		return nil, err
	}

	if !slotCheck.Available {
		return nil, fmt.Errorf("antrian penuh: %s", slotCheck.Message)
	}

	if err := uc.validateModulFileSize(fileSize); err != nil {
		return nil, err
	}

	metadata, err := uc.parseModulMetadata(uploadMetadata)
	if err != nil {
		return nil, err
	}

	uploadID := uuid.New().String()
	uploadURL := fmt.Sprintf("/modul/%d/update/%s", modulID, uploadID)
	expiresAt := time.Now().Add(time.Duration(uc.config.Upload.IdleTimeout) * time.Minute)

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
		return nil, fmt.Errorf("gagal membuat upload record: %v", err)
	}

	metadataMap := make(map[string]string)
	metadataMap["nama_file"] = metadata.NamaFile
	metadataMap["tipe"] = metadata.Tipe
	metadataMap["semester"] = strconv.Itoa(metadata.Semester)
	metadataMap["user_id"] = userID
	metadataMap["modul_id"] = strconv.FormatUint(uint64(modulID), 10)

	if err := uc.tusManager.InitiateUpload(uploadID, fileSize, metadataMap); err != nil {
		uc.tusModulUploadRepo.Delete(uploadID)
		return nil, fmt.Errorf("gagal inisiasi TUS upload: %v", err)
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
			return 0, errors.New("upload tidak ditemukan")
		}
		return 0, fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	if tusUpload.UserID != userID {
		return 0, errors.New("tidak memiliki akses ke upload ini")
	}

	if tusUpload.Status != domain.ModulUploadStatusPending && tusUpload.Status != domain.ModulUploadStatusUploading {
		return 0, fmt.Errorf("upload tidak aktif (status: %s)", tusUpload.Status)
	}

	if offset != tusUpload.CurrentOffset {
		return tusUpload.CurrentOffset, fmt.Errorf("offset tidak valid, expected %d got %d", tusUpload.CurrentOffset, offset)
	}

	if tusUpload.Status == domain.ModulUploadStatusPending {
		uc.tusModulUploadRepo.UpdateStatus(uploadID, domain.ModulUploadStatusUploading)
	}

	newOffset, err := uc.tusManager.HandleChunk(uploadID, offset, chunk)
	if err != nil {
		return tusUpload.CurrentOffset, fmt.Errorf("gagal menulis chunk: %v", err)
	}

	progress := float64(newOffset) / float64(tusUpload.FileSize) * 100

	if err := uc.tusModulUploadRepo.UpdateOffset(uploadID, newOffset, progress); err != nil {
		return newOffset, fmt.Errorf("gagal update offset: %v", err)
	}

	if newOffset >= tusUpload.FileSize {
		if err := uc.completeModulUpdate(uploadID, userID); err != nil {
			return newOffset, fmt.Errorf("gagal menyelesaikan upload: %v", err)
		}
	}

	return newOffset, nil
}

func (uc *tusModulUsecase) completeModulUpdate(uploadID string, userID string) error {
	tusUpload, err := uc.tusModulUploadRepo.GetByID(uploadID)
	if err != nil {
		return fmt.Errorf("gagal mengambil data upload: %v", err)
	}

	if tusUpload.ModulID == nil {
		return errors.New("modul_id tidak ditemukan dalam upload record")
	}

	modul, err := uc.modulRepo.GetByID(*tusUpload.ModulID)
	if err != nil {
		return fmt.Errorf("gagal mengambil data modul: %v", err)
	}

	oldFilePath := modul.PathFile

	dirPath, randomDir, err := uc.fileManager.CreateModulUploadDirectory(userID)
	if err != nil {
		return fmt.Errorf("gagal membuat direktori modul: %v", err)
	}

	fileName := fmt.Sprintf("%s.%s", tusUpload.UploadMetadata.NamaFile, tusUpload.UploadMetadata.Tipe)
	finalPath := filepath.Join(dirPath, fileName)

	if err := uc.tusManager.FinalizeUpload(uploadID, finalPath); err != nil {
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return fmt.Errorf("gagal finalisasi upload: %v", err)
	}

	modul.NamaFile = tusUpload.UploadMetadata.NamaFile
	modul.Tipe = tusUpload.UploadMetadata.Tipe
	modul.Ukuran = helper.FormatFileSize(tusUpload.FileSize)
	modul.Semester = tusUpload.UploadMetadata.Semester
	modul.PathFile = finalPath

	if err := uc.modulRepo.Update(modul); err != nil {
		helper.DeleteFile(finalPath)
		uc.fileManager.DeleteModulDirectory(userID, randomDir)
		return fmt.Errorf("gagal update data modul: %v", err)
	}

	if err := helper.DeleteFile(oldFilePath); err != nil {
		return nil
	}

	if err := uc.tusModulUploadRepo.Complete(uploadID, modul.ID, finalPath); err != nil {
		return fmt.Errorf("gagal update status upload: %v", err)
	}

	return nil
}
