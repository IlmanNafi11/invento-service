package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTusUploadRepository_Create_Success tests successful TUS upload creation
func TestTusUploadRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-id-123",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		UploadURL:  "https://example.com/upload/test-upload-id-123",
		FileSize:   1024000,
		Status:     domain.UploadStatusQueued,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	repo := NewTusUploadRepository(db)
	err = repo.Create(upload)
	assert.NoError(t, err)
	assert.NotEmpty(t, upload.ID)
	assert.Equal(t, "test-upload-id-123", upload.ID)
}

// TestTusUploadRepository_Create_WithProject_Success tests successful TUS upload creation with project
func TestTusUploadRepository_Create_WithProject_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)
	project, err := testhelper.CreateTestProject(db, "Test Project", user.ID)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-id-456",
		UserID:     user.ID,
		ProjectID:  &project.ID,
		UploadType: domain.UploadTypeProjectUpdate,
		UploadURL:  "https://example.com/upload/test-upload-id-456",
		FileSize:   2048000,
		Status:     domain.UploadStatusPending,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	repo := NewTusUploadRepository(db)
	err = repo.Create(upload)
	assert.NoError(t, err)
	assert.NotNil(t, upload.ProjectID)
	assert.Equal(t, project.ID, *upload.ProjectID)
}

// TestTusUploadRepository_GetByID_Success tests successful upload retrieval by ID
func TestTusUploadRepository_GetByID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-id-789",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		UploadURL:  "https://example.com/upload/test-upload-id-789",
		FileSize:   512000,
		Status:     domain.UploadStatusUploading,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	result, err := repo.GetByID("test-upload-id-789")
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-upload-id-789", result.ID)
	assert.Equal(t, user.ID, result.UserID)
}

// TestTusUploadRepository_GetByID_NotFound tests upload retrieval with non-existent ID
func TestTusUploadRepository_GetByID_NotFound(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	repo := NewTusUploadRepository(db)
	result, err := repo.GetByID("non-existent-id")
	assert.Error(t, err)
	assert.Nil(t, result)
}

// TestTusUploadRepository_GetByUserID_Success tests successful uploads retrieval by user ID
func TestTusUploadRepository_GetByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	uploads := []domain.TusUpload{
		{ID: "upload-1", UserID: user.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 1000, Status: domain.UploadStatusCompleted, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "upload-2", UserID: user.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 2000, Status: domain.UploadStatusUploading, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "upload-3", UserID: user.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 3000, Status: domain.UploadStatusQueued, ExpiresAt: time.Now().Add(24 * time.Hour)},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	// Create uploads for another user
	otherUser, err := testhelper.CreateTestUser(db, "other@example.com", "Other User", 1)
	require.NoError(t, err)
	otherUpload := domain.TusUpload{ID: "upload-other", UserID: otherUser.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 4000, Status: domain.UploadStatusPending, ExpiresAt: time.Now().Add(24 * time.Hour)}
	err = db.Create(&otherUpload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	result, err := repo.GetByUserID(user.ID)
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	// Verify ordering (DESC by created_at)
	assert.Equal(t, "upload-3", result[0].ID)
	assert.Equal(t, "upload-2", result[1].ID)
	assert.Equal(t, "upload-1", result[2].ID)
}

// TestTusUploadRepository_UpdateOffset_Success tests successful offset and progress update
func TestTusUploadRepository_UpdateOffset_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:            "test-upload-offset",
		UserID:        user.ID,
		UploadType:    domain.UploadTypeProjectCreate,
		FileSize:      1000000,
		CurrentOffset: 0,
		Progress:      0,
		Status:        domain.UploadStatusUploading,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	newOffset := int64(500000)
	newProgress := 50.0

	err = repo.UpdateOffset("test-upload-offset", newOffset, newProgress)
	assert.NoError(t, err)

	// Verify update
	var updatedUpload domain.TusUpload
	err = db.First(&updatedUpload, "id = ?", "test-upload-offset").Error
	assert.NoError(t, err)
	assert.Equal(t, newOffset, updatedUpload.CurrentOffset)
	assert.Equal(t, newProgress, updatedUpload.Progress)
}

// TestTusUploadRepository_UpdateOffsetOnly_Success tests successful offset-only update
func TestTusUploadRepository_UpdateOffsetOnly_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:            "test-upload-offset-only",
		UserID:        user.ID,
		UploadType:    domain.UploadTypeProjectCreate,
		FileSize:      1000000,
		CurrentOffset: 0,
		Progress:      25.0,
		Status:        domain.UploadStatusUploading,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	newOffset := int64(250000)

	err = repo.UpdateOffsetOnly("test-upload-offset-only", newOffset)
	assert.NoError(t, err)

	// Verify offset updated but progress unchanged
	var updatedUpload domain.TusUpload
	err = db.First(&updatedUpload, "id = ?", "test-upload-offset-only").Error
	assert.NoError(t, err)
	assert.Equal(t, newOffset, updatedUpload.CurrentOffset)
	assert.Equal(t, 25.0, updatedUpload.Progress) // Progress should remain unchanged
}

// TestTusUploadRepository_UpdateUpload_Success tests successful full upload update
func TestTusUploadRepository_UpdateUpload_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-update",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000000,
		Status:     domain.UploadStatusUploading,
		FilePath:   "/tmp/old-path",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	upload.Status = domain.UploadStatusCompleted
	upload.FilePath = "/var/uploads/new-path"
	upload.Progress = 100.0
	completedAt := time.Now()
	upload.CompletedAt = &completedAt

	err = repo.UpdateUpload(upload)
	assert.NoError(t, err)

	// Verify update
	var updatedUpload domain.TusUpload
	err = db.First(&updatedUpload, "id = ?", "test-upload-update").Error
	assert.NoError(t, err)
	assert.Equal(t, domain.UploadStatusCompleted, updatedUpload.Status)
	assert.Equal(t, "/var/uploads/new-path", updatedUpload.FilePath)
	assert.Equal(t, 100.0, updatedUpload.Progress)
	assert.NotNil(t, updatedUpload.CompletedAt)
}

// TestTusUploadRepository_UpdateStatus_Success tests successful status update
func TestTusUploadRepository_UpdateStatus_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-status",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000000,
		Status:     domain.UploadStatusPending,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	err = repo.UpdateStatus("test-upload-status", domain.UploadStatusUploading)
	assert.NoError(t, err)

	// Verify status update
	var updatedUpload domain.TusUpload
	err = db.First(&updatedUpload, "id = ?", "test-upload-status").Error
	assert.NoError(t, err)
	assert.Equal(t, domain.UploadStatusUploading, updatedUpload.Status)
}

// TestTusUploadRepository_UpdateStatus_ToCancelled tests cancellation
func TestTusUploadRepository_UpdateStatus_ToCancelled(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-cancel",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000000,
		Status:     domain.UploadStatusUploading,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	err = repo.UpdateStatus("test-upload-cancel", domain.UploadStatusCancelled)
	assert.NoError(t, err)

	var updatedUpload domain.TusUpload
	err = db.First(&updatedUpload, "id = ?", "test-upload-cancel").Error
	assert.NoError(t, err)
	assert.Equal(t, domain.UploadStatusCancelled, updatedUpload.Status)
}

// TestTusUploadRepository_UpdateStatus_ToFailed tests failure status
func TestTusUploadRepository_UpdateStatus_ToFailed(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-fail",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000000,
		Status:     domain.UploadStatusUploading,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	err = repo.UpdateStatus("test-upload-fail", domain.UploadStatusFailed)
	assert.NoError(t, err)

	var updatedUpload domain.TusUpload
	err = db.First(&updatedUpload, "id = ?", "test-upload-fail").Error
	assert.NoError(t, err)
	assert.Equal(t, domain.UploadStatusFailed, updatedUpload.Status)
}

// TestTusUploadRepository_GetExpired_Success tests retrieving expired uploads
func TestTusUploadRepository_GetExpired_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	// Create expired uploads (not in final states)
	expiredUpload1 := &domain.TusUpload{
		ID:         "expired-1",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000,
		Status:     domain.UploadStatusPending,
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	expiredUpload2 := &domain.TusUpload{
		ID:         "expired-2",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectUpdate,
		FileSize:   2000,
		Status:     domain.UploadStatusUploading,
		ExpiresAt:  time.Now().Add(-2 * time.Hour),
	}

	// Create non-expired upload
	activeUpload := &domain.TusUpload{
		ID:         "active-1",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   3000,
		Status:     domain.UploadStatusPending,
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}

	// Create completed upload (should not be returned even if expired)
	completedUpload := &domain.TusUpload{
		ID:         "completed-1",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   4000,
		Status:     domain.UploadStatusCompleted,
		ExpiresAt:  time.Now().Add(-3 * time.Hour),
	}

	// Create cancelled upload (should not be returned)
	cancelledUpload := &domain.TusUpload{
		ID:         "cancelled-1",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   5000,
		Status:     domain.UploadStatusCancelled,
		ExpiresAt:  time.Now().Add(-4 * time.Hour),
	}

	for _, upload := range []*domain.TusUpload{expiredUpload1, expiredUpload2, activeUpload, completedUpload, cancelledUpload} {
		err = db.Create(upload).Error
		require.NoError(t, err)
	}

	repo := NewTusUploadRepository(db)
	result, err := repo.GetExpired(time.Now())
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify only expired non-final state uploads are returned
	ids := make([]string, len(result))
	for i, upload := range result {
		ids[i] = upload.ID
	}
	assert.Contains(t, ids, "expired-1")
	assert.Contains(t, ids, "expired-2")
	assert.NotContains(t, ids, "active-1")
	assert.NotContains(t, ids, "completed-1")
	assert.NotContains(t, ids, "cancelled-1")
}

// TestTusUploadRepository_GetByUserIDAndStatus_Success tests retrieval by user and status
func TestTusUploadRepository_GetByUserIDAndStatus_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	// Create uploads with different statuses
	uploads := []domain.TusUpload{
		{ID: "upload-uploading-1", UserID: user.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 1000, Status: domain.UploadStatusUploading, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "upload-uploading-2", UserID: user.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 2000, Status: domain.UploadStatusUploading, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "upload-completed-1", UserID: user.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 3000, Status: domain.UploadStatusCompleted, ExpiresAt: time.Now().Add(24 * time.Hour)},
	}

	for i := range uploads {
		err = db.Create(&uploads[i]).Error
		require.NoError(t, err)
	}

	repo := NewTusUploadRepository(db)
	result, err := repo.GetByUserIDAndStatus(user.ID, domain.UploadStatusUploading)
	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Verify all returned uploads have the correct status
	for _, upload := range result {
		assert.Equal(t, domain.UploadStatusUploading, upload.Status)
		assert.Equal(t, user.ID, upload.UserID)
	}
}

// TestTusUploadRepository_GetByUserIDAndStatus_Empty tests retrieval with no matching uploads
func TestTusUploadRepository_GetByUserIDAndStatus_Empty(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	// Create upload with different status
	upload := &domain.TusUpload{
		ID:         "upload-1",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000,
		Status:     domain.UploadStatusCompleted,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	result, err := repo.GetByUserIDAndStatus(user.ID, domain.UploadStatusUploading)
	assert.NoError(t, err)
	assert.Len(t, result, 0)
}

// TestTusUploadRepository_Delete_Success tests successful upload deletion
func TestTusUploadRepository_Delete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	upload := &domain.TusUpload{
		ID:         "test-upload-delete",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000,
		Status:     domain.UploadStatusCompleted,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)
	err = repo.Delete("test-upload-delete")
	assert.NoError(t, err)

	// Verify deletion
	var deletedUpload domain.TusUpload
	err = db.First(&deletedUpload, "id = ?", "test-upload-delete").Error
	assert.Error(t, err)
}

// TestTusUploadRepository_ListActive_Success tests listing all active uploads
func TestTusUploadRepository_ListActive_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user1, err := testhelper.CreateTestUser(db, "user1@example.com", "User 1", 1)
	require.NoError(t, err)
	user2, err := testhelper.CreateTestUser(db, "user2@example.com", "User 2", 1)
	require.NoError(t, err)

	// Create active uploads (queued and uploading)
	activeUploads := []domain.TusUpload{
		{ID: "active-1", UserID: user1.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 1000, Status: domain.UploadStatusQueued, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "active-2", UserID: user1.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 2000, Status: domain.UploadStatusUploading, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "active-3", UserID: user2.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 3000, Status: domain.UploadStatusQueued, ExpiresAt: time.Now().Add(24 * time.Hour)},
	}

	// Create inactive uploads
	inactiveUploads := []domain.TusUpload{
		{ID: "inactive-1", UserID: user1.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 4000, Status: domain.UploadStatusCompleted, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "inactive-2", UserID: user2.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 5000, Status: domain.UploadStatusFailed, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "inactive-3", UserID: user1.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 6000, Status: domain.UploadStatusCancelled, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "inactive-4", UserID: user2.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 7000, Status: domain.UploadStatusExpired, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "inactive-5", UserID: user1.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 8000, Status: domain.UploadStatusPending, ExpiresAt: time.Now().Add(24 * time.Hour)},
	}

	for _, upload := range activeUploads {
		err = db.Create(&upload).Error
		require.NoError(t, err)
	}
	for _, upload := range inactiveUploads {
		err = db.Create(&upload).Error
		require.NoError(t, err)
	}

	repo := NewTusUploadRepository(db)
	result, err := repo.ListActive()
	assert.NoError(t, err)
	assert.Len(t, result, 3)

	// Verify all returned uploads are in active states
	for _, upload := range result {
		assert.True(t, upload.Status == domain.UploadStatusQueued || upload.Status == domain.UploadStatusUploading)
	}
}

// TestTusUploadRepository_ProgressTracking_Success tests upload progress tracking from 0 to 100
func TestTusUploadRepository_ProgressTracking_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	const totalSize = int64(1000000)
	upload := &domain.TusUpload{
		ID:            "progress-test",
		UserID:        user.ID,
		UploadType:    domain.UploadTypeProjectCreate,
		FileSize:      totalSize,
		CurrentOffset: 0,
		Progress:      0,
		Status:        domain.UploadStatusPending,
		ExpiresAt:     time.Now().Add(24 * time.Hour),
	}
	err = db.Create(upload).Error
	require.NoError(t, err)

	repo := NewTusUploadRepository(db)

	// Simulate upload progress
	progressSteps := []struct {
		offset   int64
		progress float64
	}{
		{250000, 25.0},
		{500000, 50.0},
		{750000, 75.0},
		{1000000, 100.0},
	}

	for _, step := range progressSteps {
		err = repo.UpdateOffset("progress-test", step.offset, step.progress)
		assert.NoError(t, err)

		var updated domain.TusUpload
		err = db.First(&updated, "id = ?", "progress-test").Error
		assert.NoError(t, err)
		assert.Equal(t, step.offset, updated.CurrentOffset)
		assert.Equal(t, step.progress, updated.Progress)
	}

	// Mark as completed
	err = repo.UpdateStatus("progress-test", domain.UploadStatusCompleted)
	assert.NoError(t, err)

	var final domain.TusUpload
	err = db.First(&final, "id = ?", "progress-test").Error
	assert.NoError(t, err)
	assert.Equal(t, domain.UploadStatusCompleted, final.Status)
	assert.Equal(t, float64(100), final.Progress)
	assert.Equal(t, totalSize, final.CurrentOffset)
}

// TestTusUploadRepository_UploadExpiration_Cleanup tests cleanup of expired uploads
func TestTusUploadRepository_UploadExpiration_Cleanup(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	// Create expired uploads in various states
	expiredPending := &domain.TusUpload{
		ID:         "expired-pending",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   1000,
		Status:     domain.UploadStatusPending,
		ExpiresAt:  time.Now().Add(-1 * time.Hour),
	}
	expiredUploading := &domain.TusUpload{
		ID:         "expired-uploading",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectUpdate,
		FileSize:   2000,
		Status:     domain.UploadStatusUploading,
		ExpiresAt:  time.Now().Add(-2 * time.Hour),
	}
	expiredQueued := &domain.TusUpload{
		ID:         "expired-queued",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   3000,
		Status:     domain.UploadStatusQueued,
		ExpiresAt:  time.Now().Add(-3 * time.Hour),
	}

	// Create valid uploads
	validPending := &domain.TusUpload{
		ID:         "valid-pending",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   4000,
		Status:     domain.UploadStatusPending,
		ExpiresAt:  time.Now().Add(1 * time.Hour),
	}
	validUploading := &domain.TusUpload{
		ID:         "valid-uploading",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectUpdate,
		FileSize:   5000,
		Status:     domain.UploadStatusUploading,
		ExpiresAt:  time.Now().Add(2 * time.Hour),
	}

	// Create expired uploads in final states (should not be returned)
	expiredCompleted := &domain.TusUpload{
		ID:         "expired-completed",
		UserID:     user.ID,
		UploadType: domain.UploadTypeProjectCreate,
		FileSize:   6000,
		Status:     domain.UploadStatusCompleted,
		ExpiresAt:  time.Now().Add(-5 * time.Hour),
	}

	for _, upload := range []*domain.TusUpload{expiredPending, expiredUploading, expiredQueued, validPending, validUploading, expiredCompleted} {
		err = db.Create(upload).Error
		require.NoError(t, err)
	}

	repo := NewTusUploadRepository(db)
	expired, err := repo.GetExpired(time.Now())
	assert.NoError(t, err)
	assert.Len(t, expired, 3)

	// Mark expired uploads as expired status
	for _, upload := range expired {
		err = repo.UpdateStatus(upload.ID, domain.UploadStatusExpired)
		assert.NoError(t, err)
	}

	// Verify updates - no uploads should now be returned as expired
	stillExpired, err := repo.GetExpired(time.Now())
	assert.NoError(t, err)
	assert.Len(t, stillExpired, 0)
}

// TestTusUploadRepository_MultiUser_Isolation tests that uploads are properly isolated by user
func TestTusUploadRepository_MultiUser_Isolation(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user1, err := testhelper.CreateTestUser(db, "user1@example.com", "User 1", 1)
	require.NoError(t, err)
	user2, err := testhelper.CreateTestUser(db, "user2@example.com", "User 2", 1)
	require.NoError(t, err)

	// Create uploads for user1
	user1Uploads := []domain.TusUpload{
		{ID: "user1-1", UserID: user1.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 1000, Status: domain.UploadStatusUploading, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "user1-2", UserID: user1.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 2000, Status: domain.UploadStatusCompleted, ExpiresAt: time.Now().Add(24 * time.Hour)},
	}

	// Create uploads for user2
	user2Uploads := []domain.TusUpload{
		{ID: "user2-1", UserID: user2.ID, UploadType: domain.UploadTypeProjectCreate, FileSize: 3000, Status: domain.UploadStatusUploading, ExpiresAt: time.Now().Add(24 * time.Hour)},
		{ID: "user2-2", UserID: user2.ID, UploadType: domain.UploadTypeProjectUpdate, FileSize: 4000, Status: domain.UploadStatusPending, ExpiresAt: time.Now().Add(24 * time.Hour)},
	}

	for _, upload := range user1Uploads {
		err = db.Create(&upload).Error
		require.NoError(t, err)
	}
	for _, upload := range user2Uploads {
		err = db.Create(&upload).Error
		require.NoError(t, err)
	}

	repo := NewTusUploadRepository(db)

	// Verify user isolation
	user1Results, err := repo.GetByUserID(user1.ID)
	assert.NoError(t, err)
	assert.Len(t, user1Results, 2)
	for _, upload := range user1Results {
		assert.Equal(t, user1.ID, upload.UserID)
	}

	user2Results, err := repo.GetByUserID(user2.ID)
	assert.NoError(t, err)
	assert.Len(t, user2Results, 2)
	for _, upload := range user2Results {
		assert.Equal(t, user2.ID, upload.UserID)
	}

	// Verify status-based retrieval also respects user boundary
	user1Uploading, err := repo.GetByUserIDAndStatus(user1.ID, domain.UploadStatusUploading)
	assert.NoError(t, err)
	assert.Len(t, user1Uploading, 1)
	assert.Equal(t, "user1-1", user1Uploading[0].ID)

	user2Uploading, err := repo.GetByUserIDAndStatus(user2.ID, domain.UploadStatusUploading)
	assert.NoError(t, err)
	assert.Len(t, user2Uploading, 1)
	assert.Equal(t, "user2-1", user2Uploading[0].ID)
}

// TestTusUploadRepository_StatusTransitions tests valid status transitions
func TestTusUploadRepository_StatusTransitions(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	transitions := []struct {
		fromStatus string
		toStatus   string
	}{
		{domain.UploadStatusQueued, domain.UploadStatusPending},
		{domain.UploadStatusPending, domain.UploadStatusUploading},
		{domain.UploadStatusUploading, domain.UploadStatusCompleted},
		{domain.UploadStatusPending, domain.UploadStatusCancelled},
		{domain.UploadStatusUploading, domain.UploadStatusFailed},
	}

	for i, transition := range transitions {
		upload := &domain.TusUpload{
			ID:         "transition-test-" + string(rune('a'+i)),
			UserID:     user.ID,
			UploadType: domain.UploadTypeProjectCreate,
			FileSize:   1000,
			Status:     transition.fromStatus,
			ExpiresAt:  time.Now().Add(24 * time.Hour),
		}
		err = db.Create(upload).Error
		require.NoError(t, err)

		repo := NewTusUploadRepository(db)
		err = repo.UpdateStatus(upload.ID, transition.toStatus)
		assert.NoError(t, err)

		var updated domain.TusUpload
		err = db.First(&updated, "id = ?", upload.ID).Error
		assert.NoError(t, err)
		assert.Equal(t, transition.toStatus, updated.Status)
	}
}

// TestTusUploadRepository_Create_WithMetadata tests creation with upload metadata
func TestTusUploadRepository_Create_WithMetadata(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	user, err := testhelper.CreateTestUser(db, "test@example.com", "Test User", 1)
	require.NoError(t, err)

	metadata := domain.TusUploadInitRequest{
		NamaProject: "Test Project for Upload",
		Kategori:    "website",
		Semester:    1,
	}

	upload := &domain.TusUpload{
		ID:             "metadata-test",
		UserID:         user.ID,
		UploadType:     domain.UploadTypeProjectCreate,
		UploadURL:      "https://example.com/upload/metadata-test",
		UploadMetadata: metadata,
		FileSize:       1024000,
		Status:         domain.UploadStatusQueued,
		ExpiresAt:      time.Now().Add(24 * time.Hour),
	}

	repo := NewTusUploadRepository(db)
	err = repo.Create(upload)
	assert.NoError(t, err)

	// Retrieve and verify metadata
	var retrieved domain.TusUpload
	err = db.First(&retrieved, "id = ?", "metadata-test").Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Project for Upload", retrieved.UploadMetadata.NamaProject)
	assert.Equal(t, "website", retrieved.UploadMetadata.Kategori)
	assert.Equal(t, 1, retrieved.UploadMetadata.Semester)
}
