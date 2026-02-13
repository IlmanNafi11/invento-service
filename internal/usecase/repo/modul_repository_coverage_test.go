package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModulRepository_Create_Success tests successful modul creation
func TestModulRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		UserID:    "user-1",
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
	}

	repo := NewModulRepository(db)
	err = repo.Create(modul)
	assert.NoError(t, err)
	assert.NotEmpty(t, modul.ID)
}

// TestModulRepository_GetByID_Success tests successful modul retrieval by ID
func TestModulRepository_GetByID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		UserID:    "user-1",
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	result, err := repo.GetByID(modul.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, modul.ID, result.ID)
	assert.Equal(t, "Test Modul", result.Judul)
}

// TestModulRepository_GetByIDs_Success tests successful modul retrieval by IDs
func TestModulRepository_GetByIDs_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := "user-1"

	moduls := []domain.Modul{
		{Judul: "Modul 1", Deskripsi: "Deskripsi 1", UserID: userID, FileName: "test1.pdf", FilePath: "/test1", FileSize: 1024, MimeType: "application/pdf", Status: "completed"},
		{Judul: "Modul 2", Deskripsi: "Deskripsi 2", UserID: userID, FileName: "test2.pdf", FilePath: "/test2", FileSize: 2048, MimeType: "application/pdf", Status: "completed"},
	}

	for i := range moduls {
		err = db.Create(&moduls[i]).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)
	ids := []string{moduls[0].ID, moduls[1].ID}
	result, err := repo.GetByIDs(ids, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// TestModulRepository_GetByIDsForUser_Success tests successful modul retrieval for user
func TestModulRepository_GetByIDsForUser_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := "user-1"

	moduls := []domain.Modul{
		{Judul: "Modul 1", Deskripsi: "Deskripsi 1", UserID: userID, FileName: "test1.pdf", FilePath: "/test1", FileSize: 1024, MimeType: "application/pdf", Status: "completed"},
		{Judul: "Modul 2", Deskripsi: "Deskripsi 2", UserID: "user-2", FileName: "test2.pdf", FilePath: "/test2", FileSize: 2048, MimeType: "application/pdf", Status: "completed"},
	}

	for i := range moduls {
		err = db.Create(&moduls[i]).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)
	ids := []string{moduls[0].ID, moduls[1].ID}
	result, err := repo.GetByIDsForUser(ids, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 1) // Only user's modul
}

// TestModulRepository_GetByUserID_Success tests successful modul retrieval by user ID
func TestModulRepository_GetByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := "user-1"

	moduls := []domain.Modul{
		{Judul: "Modul 1", Deskripsi: "Deskripsi 1", UserID: userID, FileName: "test1.pdf", FilePath: "/test1", FileSize: 1024, MimeType: "application/pdf", Status: "completed"},
		{Judul: "Modul 2", Deskripsi: "Deskripsi 2", UserID: userID, FileName: "test2.pdf", FilePath: "/test2", FileSize: 2048, MimeType: "video/mp4", Status: "pending"},
		{Judul: "Modul 3", Deskripsi: "Deskripsi 3", UserID: "user-2", FileName: "test3.pdf", FilePath: "/test3", FileSize: 3072, MimeType: "application/pdf", Status: "completed"},
	}

	for _, modul := range moduls {
		err = db.Create(&modul).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)

	// Test without filters
	result, total, err := repo.GetByUserID(userID, "", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)

	// Test with search
	result, total, err = repo.GetByUserID(userID, "Modul 1", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with type filter
	result, total, err = repo.GetByUserID(userID, "", "application/pdf", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with status filter
	result, total, err = repo.GetByUserID(userID, "", "", "pending", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
}

// TestModulRepository_CountByUserID_Success tests successful count
func TestModulRepository_CountByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := "user-1"

	moduls := []domain.Modul{
		{Judul: "Modul 1", Deskripsi: "Deskripsi 1", UserID: userID, FileName: "test1.pdf", FilePath: "/test1", FileSize: 1024, MimeType: "application/pdf", Status: "completed"},
		{Judul: "Modul 2", Deskripsi: "Deskripsi 2", UserID: userID, FileName: "test2.pdf", FilePath: "/test2", FileSize: 2048, MimeType: "application/pdf", Status: "completed"},
	}

	for _, modul := range moduls {
		err = db.Create(&modul).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)
	count, err := repo.CountByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestModulRepository_Update_Success tests successful modul update
func TestModulRepository_Update_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		Judul:     "Old Name",
		Deskripsi: "Old Deskripsi",
		UserID:    "user-1",
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	modul.Judul = "New Name"
	modul.Status = "pending"
	err = repo.Update(modul)
	assert.NoError(t, err)

	// Verify update
	var updatedModul domain.Modul
	err = db.First(&updatedModul, "id = ?", modul.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updatedModul.Judul)
	assert.Equal(t, "pending", updatedModul.Status)
}

// TestModulRepository_Delete_Success tests successful modul deletion
func TestModulRepository_Delete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		Judul:     "To Delete",
		Deskripsi: "To Delete Deskripsi",
		UserID:    "user-1",
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	err = repo.Delete(modul.ID)
	assert.NoError(t, err)

	// Verify deletion
	var deletedModul domain.Modul
	err = db.First(&deletedModul, "id = ?", modul.ID).Error
	assert.Error(t, err)
}

// TestModulRepository_UpdateMetadata_Success tests successful metadata update
func TestModulRepository_UpdateMetadata_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		Judul:     "Original Name",
		Deskripsi: "Original Deskripsi",
		UserID:    "user-1",
		FileName:  "test.pdf",
		FilePath:  "/test/modul",
		FileSize:  1024,
		MimeType:  "application/pdf",
		Status:    "completed",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	modul.Judul = "Updated Name"
	modul.Deskripsi = "Updated Deskripsi"
	err = repo.UpdateMetadata(modul)
	assert.NoError(t, err)

	// Verify metadata update
	var updatedModul domain.Modul
	err = db.First(&updatedModul, "id = ?", modul.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedModul.Judul)
	assert.Equal(t, "Updated Deskripsi", updatedModul.Deskripsi)
	// FilePath should remain unchanged
	assert.Equal(t, "/test/modul", updatedModul.FilePath)
}
