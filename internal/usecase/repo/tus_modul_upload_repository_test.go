package repo

import (
	"testing"
	"time"

	"invento-service/internal/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTusModulUpload(id, userID, status string, expiresAt time.Time) domain.TusModulUpload {
	return domain.TusModulUpload{
		ID:         id,
		UserID:     userID,
		UploadType: domain.UploadTypeModulCreate,
		UploadURL:  "https://example.com/upload/" + id,
		UploadMetadata: domain.TusModulUploadInitRequest{
			Judul:     "Modul " + id,
			Deskripsi: "Deskripsi modul",
		},
		FileSize:      2048,
		CurrentOffset: 0,
		Status:        status,
		Progress:      0,
		ExpiresAt:     expiresAt,
	}
}

func TestTusModulUploadRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))

	repository := NewTusModulUploadRepository(db)
	upload := newTusModulUpload("test-modul-upload-1", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))

	require.NoError(t, repository.Create(&upload))

	var saved domain.TusModulUpload
	require.NoError(t, db.First(&saved, "id = ?", "test-modul-upload-1").Error)
	assert.Equal(t, "test-modul-upload-1", saved.ID)
}

func TestTusModulUploadRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	upload := newTusModulUpload("test-modul-upload-2", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	found, err := repository.GetByID("test-modul-upload-2")
	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, "test-modul-upload-2", found.ID)

	notFound, err := repository.GetByID("missing-modul-upload")
	require.Error(t, err)
	assert.Nil(t, notFound)
}

func TestTusModulUploadRepository_GetByUserID(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	now := time.Now()
	upload1 := newTusModulUpload("test-modul-upload-3", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	upload2 := newTusModulUpload("test-modul-upload-4", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	upload3 := newTusModulUpload("test-modul-upload-5", "user-2", domain.UploadStatusPending, now.Add(time.Hour))
	require.NoError(t, db.Create(&upload1).Error)
	require.NoError(t, db.Create(&upload2).Error)
	require.NoError(t, db.Create(&upload3).Error)

	uploads, err := repository.GetByUserID("user-1")
	require.NoError(t, err)
	assert.Len(t, uploads, 2)

	empty, err := repository.GetByUserID("missing-user")
	require.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestTusModulUploadRepository_UpdateOffset(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	upload := newTusModulUpload("test-modul-upload-6", "user-1", domain.UploadStatusUploading, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.UpdateOffset("test-modul-upload-6", 1024, 50.0))

	updated, err := repository.GetByID("test-modul-upload-6")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, int64(1024), updated.CurrentOffset)
	assert.Equal(t, 50.0, updated.Progress)
}

func TestTusModulUploadRepository_UpdateStatus(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	upload := newTusModulUpload("test-modul-upload-7", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.UpdateStatus("test-modul-upload-7", domain.UploadStatusUploading))
	require.NoError(t, repository.UpdateStatus("test-modul-upload-7", domain.UploadStatusCompleted))

	updated, err := repository.GetByID("test-modul-upload-7")
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, domain.UploadStatusCompleted, updated.Status)
}

func TestTusModulUploadRepository_Complete(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	upload := newTusModulUpload("test-modul-upload-8", "user-1", domain.UploadStatusUploading, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	modulID := "550e8400-e29b-41d4-a716-446655440088"
	require.NoError(t, repository.Complete("test-modul-upload-8", modulID, "/tmp/modul-88.pdf"))

	updated, err := repository.GetByID("test-modul-upload-8")
	require.NoError(t, err)
	require.NotNil(t, updated)
	require.NotNil(t, updated.ModulID)
	assert.Equal(t, modulID, *updated.ModulID)
	assert.Equal(t, "/tmp/modul-88.pdf", updated.FilePath)
	assert.Equal(t, domain.UploadStatusCompleted, updated.Status)
	assert.Equal(t, 100.0, updated.Progress)
	require.NotNil(t, updated.CompletedAt)
}

func TestTusModulUploadRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	upload := newTusModulUpload("test-modul-upload-9", "user-1", domain.UploadStatusPending, time.Now().Add(time.Hour))
	require.NoError(t, db.Create(&upload).Error)

	require.NoError(t, repository.Delete("test-modul-upload-9"))

	_, err := repository.GetByID("test-modul-upload-9")
	require.Error(t, err)
}

func TestTusModulUploadRepository_GetExpiredUploads(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	now := time.Now()
	expiredPending := newTusModulUpload("test-modul-upload-10", "user-1", domain.UploadStatusPending, now.Add(-time.Hour))
	expiredUploading := newTusModulUpload("test-modul-upload-11", "user-1", domain.UploadStatusUploading, now.Add(-time.Hour))
	nonExpired := newTusModulUpload("test-modul-upload-12", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	expiredCompleted := newTusModulUpload("test-modul-upload-13", "user-1", domain.UploadStatusCompleted, now.Add(-time.Hour))
	expiredCancelled := newTusModulUpload("test-modul-upload-14", "user-1", domain.UploadStatusCancelled, now.Add(-time.Hour))
	expiredAlreadyExpired := newTusModulUpload("test-modul-upload-15", "user-1", domain.UploadStatusExpired, now.Add(-time.Hour))

	require.NoError(t, db.Create(&expiredPending).Error)
	require.NoError(t, db.Create(&expiredUploading).Error)
	require.NoError(t, db.Create(&nonExpired).Error)
	require.NoError(t, db.Create(&expiredCompleted).Error)
	require.NoError(t, db.Create(&expiredCancelled).Error)
	require.NoError(t, db.Create(&expiredAlreadyExpired).Error)

	uploads, err := repository.GetExpiredUploads(now)
	require.NoError(t, err)

	ids := make([]string, 0, len(uploads))
	for _, item := range uploads {
		ids = append(ids, item.ID)
	}

	assert.ElementsMatch(t, []string{"test-modul-upload-10", "test-modul-upload-11"}, ids)
}

func TestTusModulUploadRepository_GetAbandonedUploads(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	now := time.Now()
	timeout := 30 * time.Minute
	cutoff := now.Add(-timeout)

	abandonedUploading := newTusModulUpload("test-modul-upload-16", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	abandonedUploading.UpdatedAt = cutoff.Add(-time.Minute)
	abandonedPending := newTusModulUpload("test-modul-upload-17", "user-1", domain.UploadStatusPending, now.Add(time.Hour))
	abandonedPending.UpdatedAt = cutoff.Add(-2 * time.Minute)
	recentUploading := newTusModulUpload("test-modul-upload-18", "user-1", domain.UploadStatusUploading, now.Add(time.Hour))
	recentUploading.UpdatedAt = cutoff.Add(time.Minute)
	oldCompleted := newTusModulUpload("test-modul-upload-19", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour))
	oldCompleted.UpdatedAt = cutoff.Add(-time.Minute)

	require.NoError(t, db.Create(&abandonedUploading).Error)
	require.NoError(t, db.Create(&abandonedPending).Error)
	require.NoError(t, db.Create(&recentUploading).Error)
	require.NoError(t, db.Create(&oldCompleted).Error)

	uploads, err := repository.GetAbandonedUploads(timeout)
	require.NoError(t, err)

	ids := make([]string, 0, len(uploads))
	for _, item := range uploads {
		ids = append(ids, item.ID)
	}

	assert.ElementsMatch(t, []string{"test-modul-upload-16", "test-modul-upload-17"}, ids)
}

func TestTusModulUploadRepository_CountActiveByUserID(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	now := time.Now()
	records := []domain.TusModulUpload{
		newTusModulUpload("test-modul-upload-20", "user-1", domain.UploadStatusQueued, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-21", "user-1", domain.UploadStatusPending, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-22", "user-1", domain.UploadStatusUploading, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-23", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-24", "user-2", domain.UploadStatusUploading, now.Add(time.Hour)),
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

func TestTusModulUploadRepository_GetActiveByUserID(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	now := time.Now()
	records := []domain.TusModulUpload{
		newTusModulUpload("test-modul-upload-25", "user-1", domain.UploadStatusQueued, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-26", "user-1", domain.UploadStatusPending, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-27", "user-1", domain.UploadStatusUploading, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-28", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-29", "user-2", domain.UploadStatusUploading, now.Add(time.Hour)),
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	uploads, err := repository.GetActiveByUserID("user-1")
	require.NoError(t, err)
	assert.Len(t, uploads, 2)
	for _, item := range uploads {
		assert.Equal(t, "user-1", item.UserID)
		assert.Contains(t, []string{domain.UploadStatusPending, domain.UploadStatusUploading}, item.Status)
	}

	empty, err := repository.GetActiveByUserID("user-unknown")
	require.NoError(t, err)
	assert.Len(t, empty, 0)
}

func TestTusModulUploadRepository_GetActiveUploadIDs(t *testing.T) {
	db := setupTestDB(t)
	require.NoError(t, db.AutoMigrate(&domain.TusModulUpload{}))
	repository := NewTusModulUploadRepository(db)

	now := time.Now()
	records := []domain.TusModulUpload{
		newTusModulUpload("test-modul-upload-30", "user-1", domain.UploadStatusPending, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-31", "user-1", domain.UploadStatusUploading, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-32", "user-1", domain.UploadStatusQueued, now.Add(time.Hour)),
		newTusModulUpload("test-modul-upload-33", "user-1", domain.UploadStatusCompleted, now.Add(time.Hour)),
	}
	for i := range records {
		require.NoError(t, db.Create(&records[i]).Error)
	}

	ids, err := repository.GetActiveUploadIDs()
	require.NoError(t, err)
	assert.ElementsMatch(t, []string{"test-modul-upload-30", "test-modul-upload-31"}, ids)
}
