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
		NamaFile: "Test Modul",
		UserID:   "user-1",
		Tipe:     "pdf",
		Ukuran:   "small",
		Semester: 1,
		PathFile: "/test/modul",
	}

	repo := NewModulRepository(db)
	err = repo.Create(modul)
	assert.NoError(t, err)
	assert.NotZero(t, modul.ID)
}

// TestModulRepository_GetByID_Success tests successful modul retrieval by ID
func TestModulRepository_GetByID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		NamaFile: "Test Modul",
		UserID:   "user-1",
		Tipe:     "pdf",
		Ukuran:   "small",
		Semester: 1,
		PathFile: "/test/modul",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	result, err := repo.GetByID(modul.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, modul.ID, result.ID)
	assert.Equal(t, "Test Modul", result.NamaFile)
}

// TestModulRepository_GetByIDs_Success tests successful modul retrieval by IDs
func TestModulRepository_GetByIDs_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := "user-1"

	moduls := []domain.Modul{
		{NamaFile: "Modul 1", UserID: userID, Tipe: "pdf", Ukuran: "small", Semester: 1, PathFile: "/test1"},
		{NamaFile: "Modul 2", UserID: userID, Tipe: "video", Ukuran: "medium", Semester: 2, PathFile: "/test2"},
	}

	for i := range moduls {
		err = db.Create(&moduls[i]).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)
	ids := []uint{moduls[0].ID, moduls[1].ID}
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
		{NamaFile: "Modul 1", UserID: userID, Tipe: "pdf", Ukuran: "small", Semester: 1, PathFile: "/test1"},
		{NamaFile: "Modul 2", UserID: "user-2", Tipe: "video", Ukuran: "medium", Semester: 2, PathFile: "/test2"},
	}

	for i := range moduls {
		err = db.Create(&moduls[i]).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)
	ids := []uint{moduls[0].ID, moduls[1].ID}
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
		{NamaFile: "Modul 1", UserID: userID, Tipe: "pdf", Ukuran: "small", Semester: 1, PathFile: "/test1"},
		{NamaFile: "Modul 2", UserID: userID, Tipe: "video", Ukuran: "medium", Semester: 2, PathFile: "/test2"},
		{NamaFile: "Modul 3", UserID: "user-2", Tipe: "pdf", Ukuran: "large", Semester: 1, PathFile: "/test3"},
	}

	for _, modul := range moduls {
		err = db.Create(&modul).Error
		require.NoError(t, err)
	}

	repo := NewModulRepository(db)

	// Test without filters
	result, total, err := repo.GetByUserID(userID, "", "", 0, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)

	// Test with search
	result, total, err = repo.GetByUserID(userID, "Modul 1", "", 0, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with type filter
	result, total, err = repo.GetByUserID(userID, "", "pdf", 0, 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with semester filter
	result, total, err = repo.GetByUserID(userID, "", "", 2, 1, 10)
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
		{NamaFile: "Modul 1", UserID: userID, Tipe: "pdf", Ukuran: "small", Semester: 1, PathFile: "/test1"},
		{NamaFile: "Modul 2", UserID: userID, Tipe: "video", Ukuran: "medium", Semester: 2, PathFile: "/test2"},
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
		NamaFile: "Old Name",
		UserID:   "user-1",
		Tipe:     "pdf",
		Ukuran:   "small",
		Semester: 1,
		PathFile: "/test/modul",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	modul.NamaFile = "New Name"
	modul.Tipe = "video"
	err = repo.Update(modul)
	assert.NoError(t, err)

	// Verify update
	var updatedModul domain.Modul
	err = db.First(&updatedModul, modul.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updatedModul.NamaFile)
	assert.Equal(t, "video", updatedModul.Tipe)
}

// TestModulRepository_Delete_Success tests successful modul deletion
func TestModulRepository_Delete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		NamaFile: "To Delete",
		UserID:   "user-1",
		Tipe:     "pdf",
		Ukuran:   "small",
		Semester: 1,
		PathFile: "/test/modul",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	err = repo.Delete(modul.ID)
	assert.NoError(t, err)

	// Verify deletion
	var deletedModul domain.Modul
	err = db.First(&deletedModul, modul.ID).Error
	assert.Error(t, err)
}

// TestModulRepository_UpdateMetadata_Success tests successful metadata update
func TestModulRepository_UpdateMetadata_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	modul := &domain.Modul{
		NamaFile: "Original Name",
		UserID:   "user-1",
		Tipe:     "pdf",
		Ukuran:   "small",
		Semester: 1,
		PathFile: "/test/modul",
	}
	err = db.Create(modul).Error
	require.NoError(t, err)

	repo := NewModulRepository(db)
	modul.NamaFile = "Updated Name"
	modul.Semester = 2
	err = repo.UpdateMetadata(modul)
	assert.NoError(t, err)

	// Verify metadata update
	var updatedModul domain.Modul
	err = db.First(&updatedModul, modul.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updatedModul.NamaFile)
	assert.Equal(t, 2, updatedModul.Semester)
	// PathFile should remain unchanged
	assert.Equal(t, "/test/modul", updatedModul.PathFile)
}
