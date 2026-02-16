package repo_test

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/usecase/repo"
	"testing"

	apperrors "invento-service/internal/errors"
	testhelper "invento-service/internal/testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestModulRepository_Create_Success tests successful modul creation
func TestModulRepository_Create_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	err = modulRepo.Create(ctx, modul)
	assert.NoError(t, err)
	assert.NotEmpty(t, modul.ID)
}

// TestModulRepository_GetByID_Success tests successful modul retrieval by ID
func TestModulRepository_GetByID_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	result, err := modulRepo.GetByID(ctx, modul.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, modul.ID, result.ID)
	assert.Equal(t, "Test Modul", result.Judul)
}

// TestModulRepository_GetByIDs_Success tests successful modul retrieval by IDs
func TestModulRepository_GetByIDs_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	ids := []string{moduls[0].ID, moduls[1].ID}
	result, err := modulRepo.GetByIDs(ctx, ids, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// TestModulRepository_GetByUserID_Success tests successful modul retrieval by user ID
func TestModulRepository_GetByUserID_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()

	// Test without filters
	result, total, err := modulRepo.GetByUserID(ctx, userID, "", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)

	// Test with search
	result, total, err = modulRepo.GetByUserID(ctx, userID, "Modul 1", "", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with type filter
	result, total, err = modulRepo.GetByUserID(ctx, userID, "", "application/pdf", "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with status filter
	result, total, err = modulRepo.GetByUserID(ctx, userID, "", "", "pending", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
}

// TestModulRepository_CountByUserID_Success tests successful count
func TestModulRepository_CountByUserID_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	count, err := modulRepo.CountByUserID(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestModulRepository_Update_Success tests successful modul update
func TestModulRepository_Update_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	modul.Judul = "New Name"
	modul.Status = "pending"
	err = modulRepo.Update(ctx, modul)
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
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	err = modulRepo.Delete(ctx, modul.ID)
	assert.NoError(t, err)

	// Verify deletion
	var deletedModul domain.Modul
	err = db.First(&deletedModul, "id = ?", modul.ID).Error
	assert.Error(t, err)
}

// TestModulRepository_UpdateMetadata_Success tests successful metadata update
func TestModulRepository_UpdateMetadata_Success(t *testing.T) {
	t.Parallel()
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

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	modul.Judul = "Updated Name"
	modul.Deskripsi = "Updated Deskripsi"
	err = modulRepo.UpdateMetadata(ctx, modul)
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

func TestModulRepository_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	result, err := modulRepo.GetByID(ctx, "nonexistent-id")
	assert.ErrorIs(t, err, apperrors.ErrRecordNotFound)
	assert.Nil(t, result)
}

func TestModulRepository_CountByUserID_NoModuls(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	count, err := modulRepo.CountByUserID(ctx, "nonexistent-user")
	assert.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestModulRepository_Delete_Nonexistent(t *testing.T) {
	t.Parallel()
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modulRepo := repo.NewModulRepository(db, zerolog.Nop())
	ctx := context.Background()
	err = modulRepo.Delete(ctx, "nonexistent-id")
	assert.NoError(t, err)
}
