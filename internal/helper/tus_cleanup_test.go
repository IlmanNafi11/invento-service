package helper

import (
	"path/filepath"
	"testing"
	"time"

	"invento-service/config"
	"invento-service/internal/domain"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockTusUploadRepository struct{ mock.Mock }

func (m *mockTusUploadRepository) GetExpiredUploads(before time.Time) ([]domain.TusUpload, error) {
	args := m.Called(before)
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *mockTusUploadRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusUpload, error) {
	args := m.Called(timeout)
	return args.Get(0).([]domain.TusUpload), args.Error(1)
}

func (m *mockTusUploadRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *mockTusUploadRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockTusModulUploadCleanupRepository struct{ mock.Mock }

func (m *mockTusModulUploadCleanupRepository) GetExpiredUploads(before time.Time) ([]domain.TusModulUpload, error) {
	args := m.Called(before)
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *mockTusModulUploadCleanupRepository) GetAbandonedUploads(timeout time.Duration) ([]domain.TusModulUpload, error) {
	args := m.Called(timeout)
	return args.Get(0).([]domain.TusModulUpload), args.Error(1)
}

func (m *mockTusModulUploadCleanupRepository) UpdateStatus(id string, status string) error {
	args := m.Called(id, status)
	return args.Error(0)
}

func (m *mockTusModulUploadCleanupRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func newTestTusCleanup(t *testing.T) (*TusCleanup, *TusStore, *mockTusUploadRepository, *mockTusModulUploadCleanupRepository) {
	t.Helper()

	base := t.TempDir()
	cfg := &config.Config{
		App: config.AppConfig{Env: "development"},
		Upload: config.UploadConfig{
			PathDevelopment:     filepath.Join(base, "uploads"),
			TempPathDevelopment: filepath.Join(base, "temp"),
		},
	}

	store := NewTusStore(NewPathResolver(cfg), 1024*1024)
	projectRepo := &mockTusUploadRepository{}
	modulRepo := &mockTusModulUploadCleanupRepository{}

	cleanup := NewTusCleanup(projectRepo, modulRepo, store, store, 1, 1, zerolog.Nop())
	return cleanup, store, projectRepo, modulRepo
}

func TestTusCleanup_NewTusCleanup_InitializesConfig(t *testing.T) {
	cleanup, _, _, _ := newTestTusCleanup(t)

	require.NotNil(t, cleanup)
	assert.Equal(t, 1*time.Second, cleanup.cleanupInterval)
	assert.Equal(t, 1*time.Second, cleanup.idleTimeout)
	assert.Equal(t, 30*time.Minute, cleanup.lockTTL)
	assert.False(t, cleanup.isRunning)
	assert.NotNil(t, cleanup.stopChan)
}

func TestTusCleanup_StartStop_ManagesGoroutineLifecycle(t *testing.T) {
	cleanup, _, projectRepo, modulRepo := newTestTusCleanup(t)

	projectRepo.On("GetExpiredUploads", mock.Anything).Return([]domain.TusUpload{}, nil)
	projectRepo.On("GetAbandonedUploads", mock.Anything).Return([]domain.TusUpload{}, nil)
	modulRepo.On("GetExpiredUploads", mock.Anything).Return([]domain.TusModulUpload{}, nil)
	modulRepo.On("GetAbandonedUploads", mock.Anything).Return([]domain.TusModulUpload{}, nil)

	cleanup.Start()
	assert.True(t, cleanup.isRunning)

	cleanup.Start()
	assert.True(t, cleanup.isRunning)

	cleanup.Stop()
	assert.False(t, cleanup.isRunning)

	cleanup.Stop()
	assert.False(t, cleanup.isRunning)
}

func TestTusCleanup_CleanupExpiredProjects_UpdatesExpiredStatuses(t *testing.T) {
	cleanup, _, projectRepo, _ := newTestTusCleanup(t)

	expired := []domain.TusUpload{{ID: "p1"}, {ID: "p2"}}
	projectRepo.On("GetExpiredUploads", mock.Anything).Return(expired, nil).Once()
	projectRepo.On("UpdateStatus", "p1", domain.UploadStatusExpired).Return(nil).Once()
	projectRepo.On("UpdateStatus", "p2", domain.UploadStatusExpired).Return(nil).Once()

	err := cleanup.CleanupExpiredProjects()
	require.NoError(t, err)
	projectRepo.AssertExpectations(t)
}

func TestTusCleanup_CleanupAbandonedProjects_UpdatesFailedForIdleUploads(t *testing.T) {
	cleanup, _, projectRepo, _ := newTestTusCleanup(t)

	projectRepo.On("GetAbandonedUploads", cleanup.idleTimeout).Return([]domain.TusUpload{{ID: "old"}}, nil).Once()
	projectRepo.On("UpdateStatus", "old", domain.UploadStatusFailed).Return(nil).Once()

	err := cleanup.CleanupAbandonedProjects()
	require.NoError(t, err)
	projectRepo.AssertExpectations(t)
	projectRepo.AssertNotCalled(t, "UpdateStatus", "recent", domain.UploadStatusFailed)
}

func TestTusCleanup_CleanupExpiredModuls_UpdatesExpiredStatuses(t *testing.T) {
	cleanup, _, _, modulRepo := newTestTusCleanup(t)

	expired := []domain.TusModulUpload{{ID: "m1"}, {ID: "m2"}}
	modulRepo.On("GetExpiredUploads", mock.Anything).Return(expired, nil).Once()
	modulRepo.On("UpdateStatus", "m1", domain.UploadStatusExpired).Return(nil).Once()
	modulRepo.On("UpdateStatus", "m2", domain.UploadStatusExpired).Return(nil).Once()

	err := cleanup.CleanupExpiredModuls()
	require.NoError(t, err)
	modulRepo.AssertExpectations(t)
}

func TestTusCleanup_CleanupAbandonedModuls_UpdatesFailedStatuses(t *testing.T) {
	cleanup, _, _, modulRepo := newTestTusCleanup(t)

	abandoned := []domain.TusModulUpload{{ID: "m1"}}
	modulRepo.On("GetAbandonedUploads", cleanup.idleTimeout).Return(abandoned, nil).Once()
	modulRepo.On("UpdateStatus", "m1", domain.UploadStatusFailed).Return(nil).Once()

	err := cleanup.CleanupAbandonedModuls()
	require.NoError(t, err)
	modulRepo.AssertExpectations(t)
}

func TestTusCleanup_CleanupUpload_TerminatesStoreAndDeletesRecord(t *testing.T) {
	cleanup, store, projectRepo, _ := newTestTusCleanup(t)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "p-clean", Size: 10, Metadata: map[string]string{}}))

	projectRepo.On("Delete", "p-clean").Return(nil).Once()

	err := cleanup.CleanupUpload("p-clean")
	require.NoError(t, err)

	_, err = store.GetInfo("p-clean")
	assert.Error(t, err)
	projectRepo.AssertExpectations(t)
}

func TestTusCleanup_CleanupModulUpload_TerminatesStoreAndDeletesRecord(t *testing.T) {
	cleanup, store, _, modulRepo := newTestTusCleanup(t)
	require.NoError(t, store.NewUpload(TusFileInfo{ID: "m-clean", Size: 10, Metadata: map[string]string{}}))

	modulRepo.On("Delete", "m-clean").Return(nil).Once()

	err := cleanup.CleanupModulUpload("m-clean")
	require.NoError(t, err)

	_, err = store.GetInfo("m-clean")
	assert.Error(t, err)
	modulRepo.AssertExpectations(t)
}

func TestTusCleanup_CleanupModulUpload_NilModulRepoNoError(t *testing.T) {
	cleanup, _, _, _ := newTestTusCleanup(t)
	cleanup.modulRepo = nil

	err := cleanup.CleanupModulUpload("any")
	assert.NoError(t, err)
}

func TestTusCleanup_PerformCleanup_ExecutesCycleAndCleansStaleLocks(t *testing.T) {
	cleanup, store, projectRepo, modulRepo := newTestTusCleanup(t)

	projectRepo.On("GetExpiredUploads", mock.Anything).Return([]domain.TusUpload{}, nil).Once()
	projectRepo.On("GetAbandonedUploads", cleanup.idleTimeout).Return([]domain.TusUpload{}, nil).Once()
	modulRepo.On("GetExpiredUploads", mock.Anything).Return([]domain.TusModulUpload{}, nil).Once()
	modulRepo.On("GetAbandonedUploads", cleanup.idleTimeout).Return([]domain.TusModulUpload{}, nil).Once()

	_ = store.getLock("stale")
	store.locksMutex.Lock()
	store.locks["stale"].lastAccess = time.Now().Add(-2 * time.Hour)
	store.locksMutex.Unlock()

	cleanup.performCleanup()

	assert.Equal(t, 0, store.GetLockCount())
	projectRepo.AssertExpectations(t)
	modulRepo.AssertExpectations(t)
}

func TestTusCleanup_CleanupExpiredModuls_NilRepoReturnsNil(t *testing.T) {
	cleanup, _, _, _ := newTestTusCleanup(t)
	cleanup.modulRepo = nil

	err := cleanup.CleanupExpiredModuls()
	assert.NoError(t, err)
}

func TestTusCleanup_CleanupAbandonedModuls_NilRepoReturnsNil(t *testing.T) {
	cleanup, _, _, _ := newTestTusCleanup(t)
	cleanup.modulRepo = nil

	err := cleanup.CleanupAbandonedModuls()
	assert.NoError(t, err)
}
