package upload_test

import (
	"context"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Mock repository for TusCleanup tests
type MockTusUploadRepository struct {
	uploads     map[string]domain.TusUpload
	expired     []domain.TusUpload
	active      []domain.TusUpload
	updateError bool
	deleteError bool
	getError    bool
}

func (m *MockTusUploadRepository) GetExpiredUploads(ctx context.Context, before time.Time) ([]domain.TusUpload, error) {
	if m.getError {
		return nil, assert.AnError
	}
	return m.expired, nil
}

func (m *MockTusUploadRepository) GetAbandonedUploads(ctx context.Context, timeout time.Duration) ([]domain.TusUpload, error) {
	if m.getError {
		return nil, assert.AnError
	}
	threshold := time.Now().Add(-timeout)
	result := make([]domain.TusUpload, 0)
	for _, upload := range m.active {
		if upload.UpdatedAt.Before(threshold) {
			result = append(result, upload)
		}
	}
	return result, nil
}

func (m *MockTusUploadRepository) UpdateStatus(ctx context.Context, id string, status string) error {
	if m.updateError {
		return assert.AnError
	}
	if upload, ok := m.uploads[id]; ok {
		upload.Status = status
		m.uploads[id] = upload
	}
	return nil
}

func (m *MockTusUploadRepository) Delete(ctx context.Context, id string) error {
	if m.deleteError {
		return assert.AnError
	}
	delete(m.uploads, id)
	return nil
}

func setupTestConfig() *config.Config {
	return &config.Config{
		App: config.AppConfig{
			Env: "test",
		},
		Upload: config.UploadConfig{
			PathProduction:       "/tmp/uploads",
			PathDevelopment:      "/tmp/uploads",
			TempPathProduction:   "/tmp/uploads/temp",
			TempPathDevelopment:  "/tmp/uploads/temp",
			MaxSize:              100 * 1024 * 1024, // 100MB
			MaxConcurrentProject: 3,
			TusVersion:           "1.0.0",
		},
	}
}

func setupTestTusStore(t *testing.T) (*upload.TusStore, string) {
	cfg := setupTestConfig()
	tempDir := t.TempDir()
	cfg.Upload.PathDevelopment = tempDir
	cfg.Upload.TempPathDevelopment = filepath.Join(tempDir, "temp")

	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)

	return store, tempDir
}
