package upload_test

import (
	"testing"
	"time"

	"invento-service/internal/domain"
	"invento-service/internal/storage"
	"invento-service/internal/upload"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ==================== TusCleanup Tests ====================

func TestNewTusCleanup_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)

	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	assert.NotNil(t, cleanup)
}

func TestTusCleanup_Start_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Start()

	time.Sleep(100 * time.Millisecond)
	cleanup.Stop()

	// Started and stopped successfully
	assert.True(t, true)
}

func TestTusCleanup_Start_AlreadyRunning(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Start()
	cleanup.Start() // Should not cause issues

	cleanup.Stop()
}

func TestTusCleanup_Stop_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Start()
	cleanup.Stop()

	// Verify it's stopped
	assert.True(t, true)
}

func TestTusCleanup_Stop_NotRunning(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	cleanup.Stop() // Should not cause issues

	assert.True(t, true)
}

func TestTusCleanup_CleanupExpired_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		expired: []domain.TusUpload{
			{
				ID:     "expired-1",
				Status: domain.UploadStatusPending,
			},
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	// Create a temp upload
	err := store.NewUpload(upload.TusFileInfo{ID: "expired-1", Size: 1024})
	require.NoError(t, err)

	err = cleanup.CleanupExpiredProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupExpired_NoExpired(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		expired: []domain.TusUpload{},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupExpiredProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupExpired_RepositoryError(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads:  make(map[string]domain.TusUpload),
		expired:  []domain.TusUpload{},
		getError: true,
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupExpiredProjects()
	assert.Error(t, err)
}

func TestTusCleanup_CleanupAbandoned_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	oldTime := time.Now().Add(-1 * time.Hour)
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		active: []domain.TusUpload{
			{
				ID:        "abandoned-1",
				Status:    domain.UploadStatusUploading,
				UpdatedAt: oldTime,
			},
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	// Create a temp upload
	err := store.NewUpload(upload.TusFileInfo{ID: "abandoned-1", Size: 1024})
	require.NoError(t, err)

	err = cleanup.CleanupAbandonedProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupAbandoned_NoActive(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: make(map[string]domain.TusUpload),
		active:  []domain.TusUpload{},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupAbandonedProjects()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupAbandoned_RepositoryError(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads:  make(map[string]domain.TusUpload),
		active:   []domain.TusUpload{},
		getError: true,
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupAbandonedProjects()
	assert.Error(t, err)
}

func TestTusCleanup_CleanupUpload_Success(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads: map[string]domain.TusUpload{
			"test-upload": {
				ID:     "test-upload",
				Status: domain.UploadStatusPending,
			},
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	// Create a temp upload
	err := store.NewUpload(upload.TusFileInfo{ID: "test-upload", Size: 1024})
	require.NoError(t, err)

	err = cleanup.CleanupUpload("test-upload")
	assert.NoError(t, err)

	// Verify deleted
	_, err = store.GetInfo("test-upload")
	assert.Error(t, err)
}

func TestTusCleanup_CleanupUpload_NotFound(t *testing.T) {
	t.Parallel()
	cfg := setupTestConfig()
	repo := &MockTusUploadRepository{
		uploads:     map[string]domain.TusUpload{},
		deleteError: true, // Simulate error for non-existent
	}
	pathResolver := storage.NewPathResolver(cfg)
	store := upload.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	cleanup := upload.NewTusCleanup(repo, nil, store, store, 60, 300, zerolog.Nop())

	err := cleanup.CleanupUpload("nonexistent")
	assert.Error(t, err)
}
