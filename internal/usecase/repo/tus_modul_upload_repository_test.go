package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTusModulUploadRepository_Create_Success tests successful upload creation
func TestTusModulUploadRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-id-1",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadURL:  "https://example.com/upload",
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:      1024000,
		CurrentOffset: 0,
		Status:        domain.ModulUploadStatusQueued,
		Progress:      0,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}

	repo := NewTusModulUploadRepository(db)
	err = repo.Create(upload)
	assert.NoError(t, err)
	assert.NotEmpty(t, upload.ID)
}

// TestTusModulUploadRepository_Create_DuplicateID tests duplicate upload ID error
func TestTusModulUploadRepository_Create_DuplicateID(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload1 := &domain.TusModulUpload{
		ID:         "duplicate-id",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:  1024000,
		Status:    domain.ModulUploadStatusQueued,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	repo := NewTusModulUploadRepository(db)
	err = repo.Create(upload1)
	require.NoError(t, err)

	// Try to create another upload with the same ID
	upload2 := &domain.TusModulUpload{
		ID:         "duplicate-id",
		UserID:     2,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Another Modul.pdf",
			Tipe:     "pdf",
			Semester: 2,
		},
		FileSize:  2048000,
		Status:    domain.ModulUploadStatusQueued,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	err = repo.Create(upload2)
	assert.Error(t, err)
}

// TestTusModulUploadRepository_GetByID_Success tests successful upload retrieval by ID
func TestTusModulUploadRepository_GetByID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-id-2",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadURL:  "https://example.com/upload",
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:  1024000,
		Status:    domain.ModulUploadStatusQueued,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetByID(upload.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, upload.ID, result.ID)
	assert.Equal(t, uint(1), result.UserID)
	assert.Equal(t, "Test Modul.pdf", result.UploadMetadata.NamaFile)
}

// TestTusModulUploadRepository_GetByID_NotFound tests upload not found
func TestTusModulUploadRepository_GetByID_NotFound(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetByID("non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestTusModulUploadRepository_GetByUserID_Success tests successful uploads retrieval by user ID
func TestTusModulUploadRepository_GetByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	uploads := []domain.TusModulUpload{
		{
			ID:         "upload-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Modul 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "upload-2",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Modul 2.pdf",
				Tipe:     "pdf",
				Semester: 2,
			},
			FileSize:  2048000,
			Status:    domain.ModulUploadStatusUploading,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "upload-3",
			UserID:     2,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Other Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusQueued,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// TestTusModulUploadRepository_UpdateOffset_Success tests successful offset update
func TestTusModulUploadRepository_UpdateOffset_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-id-3",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:      1024000,
		CurrentOffset: 0,
		Status:        domain.ModulUploadStatusUploading,
		Progress:      0,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusModulUploadRepository(db)
	err = repo.UpdateOffset(upload.ID, 512000, 50.0)
	assert.NoError(t, err)

	// Verify update
	var updatedUpload domain.TusModulUpload
	err = db.Where("id = ?", upload.ID).First(&updatedUpload).Error
	require.NoError(t, err)
	assert.Equal(t, int64(512000), updatedUpload.CurrentOffset)
	assert.Equal(t, 50.0, updatedUpload.Progress)
}

// TestTusModulUploadRepository_UpdateOffset_NotFound tests offset update for non-existent upload
// Note: GORM doesn't return error when updating non-existent records with Model/Where
func TestTusModulUploadRepository_UpdateOffset_NotFound(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewTusModulUploadRepository(db)
	err = repo.UpdateOffset("non-existent-id", 512000, 50.0)
	// GORM doesn't error on non-existent updates - it just affects 0 rows
	assert.NoError(t, err)
}

// TestTusModulUploadRepository_UpdateStatus_Success tests successful status update
func TestTusModulUploadRepository_UpdateStatus_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-id-4",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:  1024000,
		Status:    domain.ModulUploadStatusQueued,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusModulUploadRepository(db)
	err = repo.UpdateStatus(upload.ID, domain.ModulUploadStatusUploading)
	assert.NoError(t, err)

	// Verify update
	var updatedUpload domain.TusModulUpload
	err = db.Where("id = ?", upload.ID).First(&updatedUpload).Error
	require.NoError(t, err)
	assert.Equal(t, domain.ModulUploadStatusUploading, updatedUpload.Status)
}

// TestTusModulUploadRepository_UpdateStatus_NotFound tests status update for non-existent upload
// Note: GORM doesn't return error when updating non-existent records with Model/Where
func TestTusModulUploadRepository_UpdateStatus_NotFound(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewTusModulUploadRepository(db)
	err = repo.UpdateStatus("non-existent-id", domain.ModulUploadStatusUploading)
	// GORM doesn't error on non-existent updates - it just affects 0 rows
	assert.NoError(t, err)
}

// TestTusModulUploadRepository_Complete_Success tests successful upload completion
func TestTusModulUploadRepository_Complete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-id-5",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:      1024000,
		CurrentOffset: 1024000,
		Status:        domain.ModulUploadStatusUploading,
		Progress:      100.0,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	modulID := uint(123)
	filePath := "/uploads/modul/test-upload-id-5.pdf"

	repo := NewTusModulUploadRepository(db)
	err = repo.Complete(upload.ID, modulID, filePath)
	assert.NoError(t, err)

	// Verify completion
	var completedUpload domain.TusModulUpload
	err = db.Where("id = ?", upload.ID).First(&completedUpload).Error
	require.NoError(t, err)
	assert.NotNil(t, completedUpload.ModulID)
	assert.Equal(t, modulID, *completedUpload.ModulID)
	assert.Equal(t, filePath, completedUpload.FilePath)
	assert.Equal(t, domain.ModulUploadStatusCompleted, completedUpload.Status)
	assert.Equal(t, 100.0, completedUpload.Progress)
	assert.NotNil(t, completedUpload.CompletedAt)
}

// TestTusModulUploadRepository_Complete_NotFound tests completion for non-existent upload
// Note: GORM doesn't return error when updating non-existent records with Model/Where
func TestTusModulUploadRepository_Complete_NotFound(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewTusModulUploadRepository(db)
	err = repo.Complete("non-existent-id", 123, "/uploads/path.pdf")
	// GORM doesn't error on non-existent updates - it just affects 0 rows
	assert.NoError(t, err)
}

// TestTusModulUploadRepository_Delete_Success tests successful upload deletion
func TestTusModulUploadRepository_Delete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-id-6",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:  1024000,
		Status:    domain.ModulUploadStatusQueued,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusModulUploadRepository(db)
	err = repo.Delete(upload.ID)
	assert.NoError(t, err)

	// Verify deletion
	var deletedUpload domain.TusModulUpload
	err = db.Where("id = ?", upload.ID).First(&deletedUpload).Error
	assert.Error(t, err)
}

// TestTusModulUploadRepository_Delete_NotFound tests deletion for non-existent upload
func TestTusModulUploadRepository_Delete_NotFound(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewTusModulUploadRepository(db)
	err = repo.Delete("non-existent-id")
	// GORM doesn't return error when deleting non-existent record
	// if you use Delete with struct (soft delete), it might not error
	assert.NoError(t, err)
}

// TestTusModulUploadRepository_GetExpiredUploads_Success tests successful retrieval of expired uploads
func TestTusModulUploadRepository_GetExpiredUploads_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	now := time.Now()

	uploads := []domain.TusModulUpload{
		{
			ID:         "expired-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Expired Modul 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusQueued,
			ExpiresAt: now.Add(-1 * time.Hour), // Expired
		},
		{
			ID:         "expired-2",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Expired Modul 2.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  2048000,
			Status:    domain.ModulUploadStatusUploading,
			ExpiresAt: now.Add(-2 * time.Hour), // Expired
		},
		{
			ID:         "active-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Active Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusQueued,
			ExpiresAt: now.Add(24 * time.Hour), // Not expired
		},
		{
			ID:         "completed-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Completed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			ExpiresAt: now.Add(-1 * time.Hour), // Expired but completed - should be excluded
		},
		{
			ID:         "cancelled-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Cancelled Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCancelled,
			ExpiresAt: now.Add(-1 * time.Hour), // Expired but cancelled - should be excluded
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetExpiredUploads()
	assert.NoError(t, err)
	assert.Len(t, result, 2) // Only expired-1 and expired-2
}

// TestTusModulUploadRepository_GetAbandonedUploads_Success tests successful retrieval of abandoned uploads
func TestTusModulUploadRepository_GetAbandonedUploads_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	now := time.Now()
	timeout := 30 * time.Minute

	uploads := []domain.TusModulUpload{
		{
			ID:         "abandoned-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Abandoned Modul 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:      1024000,
			CurrentOffset: 512000,
			Status:        domain.ModulUploadStatusUploading,
			Progress:      50.0,
			ExpiresAt:     now.Add(24 * time.Hour),
			UpdatedAt:     now.Add(-1 * time.Hour), // Updated more than timeout ago
		},
		{
			ID:         "abandoned-2",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Abandoned Modul 2.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:      2048000,
			CurrentOffset: 0,
			Status:        domain.ModulUploadStatusPending,
			Progress:      0,
			ExpiresAt:     now.Add(24 * time.Hour),
			UpdatedAt:     now.Add(-45 * time.Minute), // Updated more than timeout ago
		},
		{
			ID:         "active-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Active Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:      512000,
			CurrentOffset: 256000,
			Status:        domain.ModulUploadStatusUploading,
			Progress:      50.0,
			ExpiresAt:     now.Add(24 * time.Hour),
			UpdatedAt:     now.Add(-10 * time.Minute), // Updated recently - should be excluded
		},
		{
			ID:         "completed-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Completed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			ExpiresAt: now.Add(24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour), // Updated long ago but completed - should be excluded
		},
		{
			ID:         "failed-1",
			UserID:     1,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Failed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusFailed,
			ExpiresAt: now.Add(24 * time.Hour),
			UpdatedAt: now.Add(-1 * time.Hour), // Updated long ago but failed - should be excluded
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetAbandonedUploads(timeout)
	assert.NoError(t, err)
	assert.Len(t, result, 2) // Only abandoned-1 and abandoned-2
}

// TestTusModulUploadRepository_CountActiveByUserID_Success tests successful count of active uploads
func TestTusModulUploadRepository_CountActiveByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	uploads := []domain.TusModulUpload{
		{
			ID:         "queued-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Queued Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusQueued,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "pending-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Pending Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  2048000,
			Status:    domain.ModulUploadStatusPending,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "uploading-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Uploading Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusUploading,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "completed-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Completed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "failed-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Failed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusFailed,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	count, err := repo.CountActiveByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 3, count) // queued-1, pending-1, uploading-1
}

// TestTusModulUploadRepository_CountActiveByUserID_NoUploads tests count with no active uploads
func TestTusModulUploadRepository_CountActiveByUserID_NoUploads(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	repo := NewTusModulUploadRepository(db)
	count, err := repo.CountActiveByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestTusModulUploadRepository_GetActiveByUserID_Success tests successful retrieval of active uploads
func TestTusModulUploadRepository_GetActiveByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	uploads := []domain.TusModulUpload{
		{
			ID:         "queued-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Queued Modul 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusQueued,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "pending-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Pending Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  2048000,
			Status:    domain.ModulUploadStatusPending,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "uploading-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Uploading Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusUploading,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "completed-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Completed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetActiveByUserID(userID)
	assert.NoError(t, err)
	assert.Len(t, result, 3) // queued-1, pending-1, uploading-1
}

// TestTusModulUploadRepository_GetActiveByUserID_Ordering tests that active uploads are ordered by created_at ASC
func TestTusModulUploadRepository_GetActiveByUserID_Ordering(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	baseTime := time.Now()

	// Create uploads with specific creation times
	uploads := []domain.TusModulUpload{
		{
			ID:         "upload-3",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Upload 3.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusQueued,
			CreatedAt: baseTime.Add(3 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "upload-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Upload 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  2048000,
			Status:    domain.ModulUploadStatusQueued,
			CreatedAt: baseTime.Add(1 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "upload-2",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Upload 2.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusQueued,
			CreatedAt: baseTime.Add(2 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetActiveByUserID(userID)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	// Verify ascending order
	assert.Equal(t, "upload-1", result[0].ID)
	assert.Equal(t, "upload-2", result[1].ID)
	assert.Equal(t, "upload-3", result[2].ID)
}

// TestTusModulUploadRepository_GetActiveByUserID_NoActiveUploads tests retrieval with no active uploads
func TestTusModulUploadRepository_GetActiveByUserID_NoActiveUploads(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	// Create only completed uploads
	uploads := []domain.TusModulUpload{
		{
			ID:         "completed-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Completed Modul 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "failed-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Failed Modul.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusFailed,
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetActiveByUserID(userID)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestTusModulUploadRepository_GetByUserID_Ordering tests that uploads are ordered by created_at DESC
func TestTusModulUploadRepository_GetByUserID_Ordering(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)
	baseTime := time.Now()

	uploads := []domain.TusModulUpload{
		{
			ID:         "upload-1",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Upload 1.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  1024000,
			Status:    domain.ModulUploadStatusCompleted,
			CreatedAt: baseTime.Add(1 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "upload-2",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Upload 2.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  2048000,
			Status:    domain.ModulUploadStatusCompleted,
			CreatedAt: baseTime.Add(2 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
		{
			ID:         "upload-3",
			UserID:     userID,
			UploadType: domain.ModulUploadTypeCreate,
			UploadMetadata: domain.TusModulUploadInitRequest{
				NamaFile: "Upload 3.pdf",
				Tipe:     "pdf",
				Semester: 1,
			},
			FileSize:  512000,
			Status:    domain.ModulUploadStatusCompleted,
			CreatedAt: baseTime.Add(3 * time.Hour),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusModulUploadRepository(db)
	result, err := repo.GetByUserID(userID)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	// Verify descending order (newest first)
	assert.Equal(t, "upload-3", result[0].ID)
	assert.Equal(t, "upload-2", result[1].ID)
	assert.Equal(t, "upload-1", result[2].ID)
}

// TestTusModulUploadRepository_Complete_SetsProgressAndTimestamp tests that Complete sets progress to 100 and timestamp
func TestTusModulUploadRepository_Complete_SetsProgressAndTimestamp(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	upload := &domain.TusModulUpload{
		ID:         "test-upload-complete",
		UserID:     1,
		UploadType: domain.ModulUploadTypeCreate,
		UploadMetadata: domain.TusModulUploadInitRequest{
			NamaFile: "Test Modul.pdf",
			Tipe:     "pdf",
			Semester: 1,
		},
		FileSize:      1024000,
		CurrentOffset: 1024000,
		Status:        domain.ModulUploadStatusUploading,
		Progress:      99.5,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	modulID := uint(456)
	filePath := "/uploads/modul/test-upload-complete.pdf"

	repo := NewTusModulUploadRepository(db)
	beforeComplete := time.Now()
	err = repo.Complete(upload.ID, modulID, filePath)
	afterComplete := time.Now()

	assert.NoError(t, err)

	// Verify completion
	var completedUpload domain.TusModulUpload
	err = db.Where("id = ?", upload.ID).First(&completedUpload).Error
	require.NoError(t, err)

	assert.Equal(t, 100.0, completedUpload.Progress)
	assert.NotNil(t, completedUpload.CompletedAt)
	assert.True(t, completedUpload.CompletedAt.After(beforeComplete) || completedUpload.CompletedAt.Equal(beforeComplete))
	assert.True(t, completedUpload.CompletedAt.Before(afterComplete) || completedUpload.CompletedAt.Equal(afterComplete))
}
