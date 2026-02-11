package repo

import (
	"sync"
	"testing"
	"time"

	"fiber-boiler-plate/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	require.NoError(t, db.AutoMigrate(&domain.TusUpload{}))

	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.SetMaxOpenConns(1)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return db
}

func newTusUpload(id, userID, status string, expiresAt time.Time) domain.TusUpload {
	return domain.TusUpload{
		ID:         id,
		UserID:     userID,
		UploadType: domain.UploadTypeProjectCreate,
		UploadURL:  "https://example.com/upload/" + id,
		UploadMetadata: domain.TusUploadInitRequest{
			NamaProject: "Project " + id,
			Kategori:    "website",
			Semester:    1,
		},
		FileSize:      1024,
		CurrentOffset: 0,
		Status:        status,
		Progress:      0,
		ExpiresAt:     expiresAt,
	}
}

func TestTusUploadRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-1", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))

	require.NoError(t, repository.Create(&upload))

	var saved domain.TusUpload
	require.NoError(t, db.First(&saved, "id = ?", "test-upload-1").Error)
	assert.Equal(t, "test-upload-1", saved.ID)
	assert.Equal(t, "user-1", saved.UserID)
}

func TestTusUploadRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-2", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	found, err := repository.GetByID("test-upload-2")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "test-upload-2", found.ID)

	notFound, err := repository.GetByID("missing-upload")
	require.Error(t, err)
	assert.Nil(t, notFound)
}

func TestTusUploadRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	now := time.Now()
	upload1 := newTusUpload("test-upload-3", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	upload2 := newTusUpload("test-upload-4", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	upload3 := newTusUpload("test-upload-5", "user-2", domain.UploadStatusPending, now.Add(time.Hour))
	require.NoError(t, db.Create(&upload1).Error)
	require.NoError(t, db.Create(&upload2).Error)
	require.NoError(t, db.Create(&upload3).Error)

	uploads, err := repository.GetByUserID("user-1")
	require.NoError(t, err)
	assert.Len(t, uploads, 2)

	empty, err := repository.GetByUserID("unknown-user")
	require.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestTusUploadRepository_UpdateOffset(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-6", "user-1", domain.UploadStatusUploading, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.UpdateOffset("test-upload-6", 512, 50.0))

	updated, err := repository.GetByID("test-upload-6")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(512), updated.CurrentOffset)
	assert.Equal(t, 50.0, updated.Progress)
}

func TestTusUploadRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-7", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.UpdateStatus("test-upload-7", domain.UploadStatusUploading))
	require.NoError(t, repository.UpdateStatus("test-upload-7", domain.UploadStatusCompleted))

	updated, err := repository.GetByID("test-upload-7")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, domain.UploadStatusCompleted, updated.Status)
}

func TestTusUploadRepository_GetExpired(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	now := time.Now()
	expiredPending := newTusUpload("test-upload-8", "user-1", domain.UploadStatusPending, now.Add(-10*time.Minute))
	expiredUploading := newTusUpload("test-upload-9", "user-1", domain.UploadStatusUploading, now.Add(-20*time.Minute))
	nonExpired := newTusUpload("test-upload-10", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	expiredCompleted := newTusUpload("test-upload-11", "user-1", domain.UploadStatusCompleted, now.Add(-time.Hour))
	expiredCancelled := newTusUpload("test-upload-12", "user-1", domain.UploadStatusCancelled, now.Add(-time.Hour))
	expiredAlreadyExpired := newTusUpload("test-upload-13", "user-1", domain.UploadStatusExpired, now.Add(-time.Hour))

	require.NoError(t, db.Create(&expiredPending).Error)
	require.NoError(t, db.Create(&expiredUploading).Error)
	require.NoError(t, db.Create(&nonExpired).Error)
	require.NoError(t, db.Create(&expiredCompleted).Error)
	require.NoError(t, db.Create(&expiredCancelled).Error)
	require.NoError(t, db.Create(&expiredAlreadyExpired).Error)

	uploads, err := repository.GetExpired(now)
	require.NoError(t, err)

	ids := make([]string, 0, len(uploads))
	for _, item := range uploads {
		ids = append(ids, item.ID)
	}

	assert.ElementsMatch(t, []string{"test-upload-8", "test-upload-9"}, ids)
}

func TestTusUploadRepository_GetByUserIDAndStatus(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	now := time.Now()
	upload1 := newTusUpload("test-upload-14", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	upload2 := newTusUpload("test-upload-15", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	upload3 := newTusUpload("test-upload-16", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	upload4 := newTusUpload("test-upload-17", "user-2", domain.UploadStatusUploading, now.Add(time.Hour))
	require.NoError(t, db.Create(&upload1).Error)
	require.NoError(t, db.Create(&upload2).Error)
	require.NoError(t, db.Create(&upload3).Error)
	require.NoError(t, db.Create(&upload4).Error)

	uploads, err := repository.GetByUserIDAndStatus("user-1", domain.UploadStatusUploading)
	require.NoError(t, err)
	assert.Len(t, uploads, 2)
	for _, item := range uploads {
		assert.Equal(t, "user-1", item.UserID)
		assert.Equal(t, domain.UploadStatusUploading, item.Status)
	}

	empty, err := repository.GetByUserIDAndStatus("user-1", domain.UploadStatusFailed)
	require.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestTusUploadRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-18", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.Delete("test-upload-18"))

	_, err := repository.GetByID("test-upload-18")
	require.Error(t, err)
}

func TestTusUploadRepository_ListActive(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	now := time.Now()
	queued := newTusUpload("test-upload-19", "user-1", domain.UploadStatusQueued, now.Add(time.Hour))
	uploading := newTusUpload("test-upload-20", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	pending := newTusUpload("test-upload-21", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	completed := newTusUpload("test-upload-22", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour))
	require.NoError(t, db.Create(&queued).Error)
	require.NoError(t, db.Create(&uploading).Error)
	require.NoError(t, db.Create(&pending).Error)
	require.NoError(t, db.Create(&completed).Error)

	uploads, err := repository.ListActive()
	require.NoError(t, err)

	ids := make([]string, 0, len(uploads))
	for _, item := range uploads {
		ids = append(ids, item.ID)
	}

	assert.ElementsMatch(t, []string{"test-upload-19", "test-upload-20"}, ids)
}

func TestTusUploadRepository_UpdateOffsetOnly(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-23", "user-1", domain.UploadStatusUploading, time.Now().Add(time.Hour))
	upload.Progress = 32.5
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.UpdateOffsetOnly("test-upload-23", 777))

	updated, err := repository.GetByID("test-upload-23")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(777), updated.CurrentOffset)
	assert.Equal(t, 32.5, updated.Progress)
}

func TestTusUploadRepository_UpdateUpload(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-24", "user-1", domain.UploadStatusUploading, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	completedAt := time.Now()
	upload.CurrentOffset = 1024
	upload.Progress = 100.0
	upload.Status = domain.UploadStatusCompleted
	upload.FilePath = "/tmp/result.bin"
	upload.CompletedAt = &completedAt

	require.NoError(t, repository.UpdateUpload(&upload))

	updated, err := repository.GetByID("test-upload-24")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(1024), updated.CurrentOffset)
	assert.Equal(t, 100.0, updated.Progress)
	assert.Equal(t, domain.UploadStatusCompleted, updated.Status)
	assert.Equal(t, "/tmp/result.bin", updated.FilePath)
	require.NotNil(t, updated.CompletedAt)
}

func TestTusUploadRepository_CountActiveByUserID(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	now := time.Now()
	records := []domain.TusUpload{
		newTusUpload("test-upload-25", "user-1", domain.UploadStatusPending, now.Add(time.Hour)),
		newTusUpload("test-upload-26", "user-1", domain.UploadStatusUploading, now.Add(time.Hour)),
		newTusUpload("test-upload-27", "user-1", domain.UploadStatusQueued, now.Add(time.Hour)),
		newTusUpload("test-upload-28", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour)),
		newTusUpload("test-upload-29", "user-2", domain.UploadStatusPending, now.Add(time.Hour)),
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	count, err := repository.CountActiveByUserID("user-1")
	require.NoError(t, err)
	assert.Equal(t, int64(2), count)

	emptyCount, err := repository.CountActiveByUserID("user-unknown")
	require.NoError(t, err)
	assert.Equal(t, int64(0), emptyCount)
}

func TestTusUploadRepository_GetActiveUploadIDs(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	now := time.Now()
	records := []domain.TusUpload{
		newTusUpload("test-upload-30", "user-1", domain.UploadStatusPending, now.Add(time.Hour)),
		newTusUpload("test-upload-31", "user-1", domain.UploadStatusUploading, now.Add(time.Hour)),
		newTusUpload("test-upload-32", "user-1", domain.UploadStatusQueued, now.Add(time.Hour)),
		newTusUpload("test-upload-33", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour)),
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	ids, err := repository.GetActiveUploadIDs()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"test-upload-30", "test-upload-31"}, ids)
}

func TestTusUploadRepository_ConcurrentGetByID(t *testing.T) {
	db := setupTestDB(t)
	repository := NewTusUploadRepository(db)

	upload := newTusUpload("test-upload-34", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	var wg sync.WaitGroup
	errs := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := repository.GetByID("test-upload-34")
			errs <- err
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		assert.NoError(t, err)
	}
}
