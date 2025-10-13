package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

type TusStore struct {
	pathResolver *PathResolver
	maxFileSize  int64
	locks        map[string]*sync.RWMutex
	locksMutex   sync.RWMutex
}

type TusFileInfo struct {
	ID       string            `json:"id"`
	Size     int64             `json:"size"`
	Offset   int64             `json:"offset"`
	Metadata map[string]string `json:"metadata"`
}

func NewTusStore(pathResolver *PathResolver, maxFileSize int64) *TusStore {
	return &TusStore{
		pathResolver: pathResolver,
		maxFileSize:  maxFileSize,
		locks:        make(map[string]*sync.RWMutex),
	}
}

func (ts *TusStore) getLock(uploadID string) *sync.RWMutex {
	ts.locksMutex.Lock()
	defer ts.locksMutex.Unlock()

	if lock, exists := ts.locks[uploadID]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	ts.locks[uploadID] = lock
	return lock
}

func (ts *TusStore) removeLock(uploadID string) {
	ts.locksMutex.Lock()
	defer ts.locksMutex.Unlock()

	delete(ts.locks, uploadID)
}

func (ts *TusStore) NewUpload(info TusFileInfo) error {
	if info.Size > ts.maxFileSize {
		return fmt.Errorf("ukuran file melebihi batas maksimal %d bytes", ts.maxFileSize)
	}

	uploadPath := ts.pathResolver.GetUploadPath(info.ID)
	if err := ts.pathResolver.EnsureDirectoryExists(uploadPath); err != nil {
		return err
	}

	filePath := ts.pathResolver.GetUploadFilePath(info.ID)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("gagal membuat file: %w", err)
	}
	defer file.Close()

	if err := file.Truncate(info.Size); err != nil {
		return fmt.Errorf("gagal preallocate file: %w", err)
	}

	if err := ts.saveInfo(info); err != nil {
		os.Remove(filePath)
		return err
	}

	return nil
}

func (ts *TusStore) WriteChunk(uploadID string, offset int64, src io.Reader) (int64, error) {
	lock := ts.getLock(uploadID)
	lock.Lock()
	defer lock.Unlock()

	filePath := ts.pathResolver.GetUploadFilePath(uploadID)
	file, err := os.OpenFile(filePath, os.O_WRONLY, 0644)
	if err != nil {
		return offset, fmt.Errorf("gagal membuka file: %w", err)
	}
	defer file.Close()

	if _, err := file.Seek(offset, 0); err != nil {
		return offset, fmt.Errorf("gagal seek ke offset: %w", err)
	}

	bytesWritten, err := io.Copy(file, src)
	if err != nil {
		return offset + bytesWritten, fmt.Errorf("gagal menulis chunk: %w", err)
	}

	if err := file.Sync(); err != nil {
		return offset + bytesWritten, fmt.Errorf("gagal sync file: %w", err)
	}

	newOffset := offset + bytesWritten
	return newOffset, nil
}

func (ts *TusStore) GetInfo(uploadID string) (TusFileInfo, error) {
	infoPath := ts.pathResolver.GetUploadInfoPath(uploadID)

	data, err := os.ReadFile(infoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return TusFileInfo{}, errors.New("upload tidak ditemukan")
		}
		return TusFileInfo{}, fmt.Errorf("gagal membaca info: %w", err)
	}

	var info TusFileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return TusFileInfo{}, fmt.Errorf("gagal parse info: %w", err)
	}

	return info, nil
}

func (ts *TusStore) InitiateUpload(uploadID string, fileSize int64) error {
	if fileSize > ts.maxFileSize {
		return fmt.Errorf("ukuran file melebihi batas maksimal %d bytes", ts.maxFileSize)
	}

	if fileSize <= 0 {
		return errors.New("ukuran file tidak valid")
	}

	uploadPath := ts.pathResolver.GetUploadPath(uploadID)
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori upload: %w", err)
	}

	filePath := ts.pathResolver.GetUploadFilePath(uploadID)
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("gagal membuat file upload: %w", err)
	}
	defer file.Close()

	if err := file.Truncate(fileSize); err != nil {
		return fmt.Errorf("gagal alokasi file size: %w", err)
	}

	info := TusFileInfo{
		ID:       uploadID,
		Size:     fileSize,
		Offset:   0,
		Metadata: make(map[string]string),
	}

	return ts.saveInfo(info)
}

func (ts *TusStore) FinalizeUpload(uploadID string, finalPath string) error {
	lock := ts.getLock(uploadID)
	lock.Lock()
	defer lock.Unlock()

	tempFilePath := ts.pathResolver.GetUploadFilePath(uploadID)

	if _, err := os.Stat(tempFilePath); err != nil {
		return fmt.Errorf("file temporary tidak ditemukan: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(finalPath), 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori final: %w", err)
	}

	if err := MoveFile(tempFilePath, finalPath); err != nil {
		return fmt.Errorf("gagal memindahkan file: %w", err)
	}

	if err := os.RemoveAll(ts.pathResolver.GetUploadPath(uploadID)); err != nil {
		return fmt.Errorf("gagal membersihkan temporary files: %w", err)
	}

	ts.removeLock(uploadID)
	return nil
}

func (ts *TusStore) saveInfo(info TusFileInfo) error {
	infoPath := ts.pathResolver.GetUploadInfoPath(info.ID)

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("gagal marshal info: %w", err)
	}

	if err := os.WriteFile(infoPath, data, 0644); err != nil {
		return fmt.Errorf("gagal menyimpan info: %w", err)
	}

	return nil
}

func (ts *TusStore) GetFilePath(uploadID string) string {
	return ts.pathResolver.GetUploadFilePath(uploadID)
}

func (ts *TusStore) IsComplete(uploadID string) (bool, error) {
	info, err := ts.GetInfo(uploadID)
	if err != nil {
		return false, err
	}

	return info.Offset >= info.Size, nil
}

func (ts *TusStore) Terminate(uploadID string) error {
	lock := ts.getLock(uploadID)
	lock.Lock()
	defer lock.Unlock()

	uploadPath := ts.pathResolver.GetUploadPath(uploadID)
	if err := os.RemoveAll(uploadPath); err != nil {
		return fmt.Errorf("gagal menghapus upload: %w", err)
	}

	ts.removeLock(uploadID)
	return nil
}

func (ts *TusStore) GetProgress(uploadID string) (float64, error) {
	info, err := ts.GetInfo(uploadID)
	if err != nil {
		return 0, err
	}

	if info.Size == 0 {
		return 0, nil
	}

	progress := (float64(info.Offset) / float64(info.Size)) * 100
	return progress, nil
}
