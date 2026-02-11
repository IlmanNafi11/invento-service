package repo

import (
	"fiber-boiler-plate/internal/domain"
	testhelper "fiber-boiler-plate/internal/testing"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProjectRepository_Create_Success tests successful project creation
func TestProjectRepository_Create_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	project := &domain.Project{
		NamaProject: "Test Project",
		UserID:      1,
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
	}

	repo := NewProjectRepository(db)
	err = repo.Create(project)
	assert.NoError(t, err)
	assert.NotZero(t, project.ID)
}

// TestProjectRepository_GetByID_Success tests successful project retrieval by ID
func TestProjectRepository_GetByID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	project := &domain.Project{
		NamaProject: "Test Project",
		UserID:      1,
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	repo := NewProjectRepository(db)
	result, err := repo.GetByID(project.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, project.ID, result.ID)
	assert.Equal(t, "Test Project", result.NamaProject)
}

// TestProjectRepository_GetByIDs_Success tests successful project retrieval by IDs
func TestProjectRepository_GetByIDs_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	projects := []domain.Project{
		{NamaProject: "Project 1", UserID: userID, Kategori: "website", Semester: 1, Ukuran: "small", PathFile: "/test1"},
		{NamaProject: "Project 2", UserID: userID, Kategori: "mobile", Semester: 2, Ukuran: "medium", PathFile: "/test2"},
	}

	for i := range projects {
		err = db.Create(&projects[i]).Error
		require.NoError(t, err)
	}

	repo := NewProjectRepository(db)
	ids := []uint{projects[0].ID, projects[1].ID}
	result, err := repo.GetByIDs(ids, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
}

// TestProjectRepository_GetByIDsForUser_Success tests successful project retrieval for user
func TestProjectRepository_GetByIDsForUser_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	projects := []domain.Project{
		{NamaProject: "Project 1", UserID: userID, Kategori: "website", Semester: 1, Ukuran: "small", PathFile: "/test1"},
		{NamaProject: "Project 2", UserID: 2, Kategori: "mobile", Semester: 2, Ukuran: "medium", PathFile: "/test2"},
	}

	for i := range projects {
		err = db.Create(&projects[i]).Error
		require.NoError(t, err)
	}

	repo := NewProjectRepository(db)
	ids := []uint{projects[0].ID, projects[1].ID}
	result, err := repo.GetByIDsForUser(ids, userID)
	assert.NoError(t, err)
	assert.Len(t, result, 1) // Only user's project
}

// TestProjectRepository_GetByUserID_Success tests successful project retrieval by user ID
func TestProjectRepository_GetByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	projects := []domain.Project{
		{NamaProject: "Project 1", UserID: userID, Kategori: "website", Semester: 1, Ukuran: "small", PathFile: "/test1"},
		{NamaProject: "Project 2", UserID: userID, Kategori: "mobile", Semester: 2, Ukuran: "medium", PathFile: "/test2"},
		{NamaProject: "Project 3", UserID: 2, Kategori: "desktop", Semester: 1, Ukuran: "large", PathFile: "/test3"},
	}

	for _, proj := range projects {
		err = db.Create(&proj).Error
		require.NoError(t, err)
	}

	repo := NewProjectRepository(db)

	// Test without filters
	result, total, err := repo.GetByUserID(userID, "", 0, "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, 2, total)

	// Test with search
	result, total, err = repo.GetByUserID(userID, "Project 1", 0, "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with semester filter
	result, total, err = repo.GetByUserID(userID, "", 2, "", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)

	// Test with category filter
	result, total, err = repo.GetByUserID(userID, "", 0, "website", 1, 10)
	assert.NoError(t, err)
	assert.Len(t, result, 1)
	assert.Equal(t, 1, total)
}

// TestProjectRepository_CountByUserID_Success tests successful count
func TestProjectRepository_CountByUserID_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	userID := uint(1)

	projects := []domain.Project{
		{NamaProject: "Project 1", UserID: userID, Kategori: "website", Semester: 1, Ukuran: "small", PathFile: "/test1"},
		{NamaProject: "Project 2", UserID: userID, Kategori: "mobile", Semester: 2, Ukuran: "medium", PathFile: "/test2"},
	}

	for _, proj := range projects {
		err = db.Create(&proj).Error
		require.NoError(t, err)
	}

	repo := NewProjectRepository(db)
	count, err := repo.CountByUserID(userID)
	assert.NoError(t, err)
	assert.Equal(t, 2, count)
}

// TestProjectRepository_Update_Success tests successful project update
func TestProjectRepository_Update_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	project := &domain.Project{
		NamaProject: "Old Name",
		UserID:      1,
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	repo := NewProjectRepository(db)
	project.NamaProject = "New Name"
	project.Kategori = "mobile"
	err = repo.Update(project)
	assert.NoError(t, err)

	// Verify update
	var updatedProject domain.Project
	err = db.First(&updatedProject, project.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "New Name", updatedProject.NamaProject)
	assert.Equal(t, "mobile", updatedProject.Kategori)
}

// TestProjectRepository_Delete_Success tests successful project deletion
func TestProjectRepository_Delete_Success(t *testing.T) {
	db, err := testhelper.SetupTestDatabase()
	require.NoError(t, err)
	defer testhelper.TeardownTestDatabase(db)

	project := &domain.Project{
		NamaProject: "To Delete",
		UserID:      1,
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/test/path",
	}
	err = db.Create(project).Error
	require.NoError(t, err)

	repo := NewProjectRepository(db)
	err = repo.Delete(project.ID)
	assert.NoError(t, err)

	// Verify deletion
	var deletedProject domain.Project
	err = db.First(&deletedProject, project.ID).Error
	assert.Error(t, err)
}
