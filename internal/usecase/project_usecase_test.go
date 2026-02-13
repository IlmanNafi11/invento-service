package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	apperrors "fiber-boiler-plate/internal/errors"
	"fiber-boiler-plate/internal/helper"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestProjectUsecase_CreateProject_Success(t *testing.T) {
	// This test would require file upload handling which is complex
	// For now, we'll test the GetByID and List operations
}

// TestGetProjectByID_Success tests successful project retrieval by ID
func TestProjectUsecase_GetProjectByID_Success(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(1)

	project := &domain.Project{
		ID:          projectID,
		UserID:      userID,
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/uploads/test.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)

	result, err := projectUC.GetByID(projectID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, projectID, result.ID)
	assert.Equal(t, "Test Project", result.NamaProject)
	assert.Equal(t, "website", result.Kategori)

	mockProjectRepo.AssertExpectations(t)
}

// TestGetProjectByID_NotFound tests project retrieval when project doesn't exist
func TestProjectUsecase_GetProjectByID_NotFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(999)

	mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	result, err := projectUC.GetByID(projectID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestGetProjectByID_AccessDenied tests project retrieval when user doesn't have access
func TestProjectUsecase_GetProjectByID_AccessDenied(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(2)

	project := &domain.Project{
		ID:          projectID,
		UserID:      "user-999", // Different user
		NamaProject: "Other User Project",
		Kategori:    "mobile",
		Semester:    2,
		Ukuran:      "medium",
		PathFile:    "/uploads/other.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)

	result, err := projectUC.GetByID(projectID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestListProjects_Success tests successful project list retrieval
func TestProjectUsecase_ListProjects_Success(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	search := "test"
	filterSemester := 1
	filterKategori := "website"
	page := 1
	limit := 10

	projects := []domain.ProjectListItem{
		{
			ID:                 1,
			NamaProject:        "Test Project 1",
			Kategori:           "website",
			Semester:           1,
			Ukuran:             "small",
			PathFile:           "/uploads/test1.zip",
			TerakhirDiperbarui: time.Now(),
		},
		{
			ID:                 2,
			NamaProject:        "Test Project 2",
			Kategori:           "website",
			Semester:           1,
			Ukuran:             "medium",
			PathFile:           "/uploads/test2.zip",
			TerakhirDiperbarui: time.Now(),
		},
	}

	total := 2

	mockProjectRepo.On("GetByUserID", userID, search, filterSemester, filterKategori, page, limit).Return(projects, total, nil)

	result, err := projectUC.GetList(userID, search, filterSemester, filterKategori, page, limit)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 2)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 2, result.Pagination.TotalItems)
	assert.Equal(t, 1, result.Pagination.TotalPages)

	mockProjectRepo.AssertExpectations(t)
}

// TestListProjects_Empty tests project list retrieval with no results
func TestProjectUsecase_ListProjects_Empty(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	search := "nonexistent"
	filterSemester := 0
	filterKategori := ""
	page := 1
	limit := 10

	projects := []domain.ProjectListItem{}
	total := 0

	mockProjectRepo.On("GetByUserID", userID, search, filterSemester, filterKategori, page, limit).Return(projects, total, nil)

	result, err := projectUC.GetList(userID, search, filterSemester, filterKategori, page, limit)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 0)
	assert.Equal(t, 0, result.Pagination.TotalItems)

	mockProjectRepo.AssertExpectations(t)
}

// TestListProjects_Pagination tests pagination in project list
func TestProjectUsecase_ListProjects_Pagination(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	search := ""
	filterSemester := 0
	filterKategori := ""
	page := 2
	limit := 10

	projects := []domain.ProjectListItem{
		{
			ID:                 11,
			NamaProject:        "Project 11",
			Kategori:           "mobile",
			Semester:           2,
			Ukuran:             "large",
			PathFile:           "/uploads/project11.zip",
			TerakhirDiperbarui: time.Now(),
		},
	}

	total := 15

	mockProjectRepo.On("GetByUserID", userID, search, filterSemester, filterKategori, page, limit).Return(projects, total, nil)

	result, err := projectUC.GetList(userID, search, filterSemester, filterKategori, page, limit)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)
	assert.Equal(t, 2, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 15, result.Pagination.TotalItems)
	assert.Equal(t, 2, result.Pagination.TotalPages)

	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_ListProjects_RepoError(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	search := "test"
	filterSemester := 1
	filterKategori := "website"
	page := 1
	limit := 10

	mockProjectRepo.On("GetByUserID", userID, search, filterSemester, filterKategori, page, limit).Return(nil, 0, assert.AnError)

	result, err := projectUC.GetList(userID, search, filterSemester, filterKategori, page, limit)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_ListProjects_EmptyWithDefaultPagination(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	search := ""
	filterSemester := 0
	filterKategori := ""

	mockProjectRepo.On("GetByUserID", userID, search, filterSemester, filterKategori, 1, 10).Return([]domain.ProjectListItem{}, 0, nil)

	result, err := projectUC.GetList(userID, search, filterSemester, filterKategori, 0, 0)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 0)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)
	assert.Equal(t, 0, result.Pagination.TotalItems)
	assert.Equal(t, 0, result.Pagination.TotalPages)

	mockProjectRepo.AssertExpectations(t)
}

// TestUpdateProject_Success tests successful project update
func TestProjectUsecase_UpdateProject_Success(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(1)

	project := &domain.Project{
		ID:          projectID,
		UserID:      userID,
		NamaProject: "Old Project Name",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/uploads/test.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	req := domain.ProjectUpdateRequest{
		NamaProject: "Updated Project Name",
		Kategori:    "mobile",
		Semester:    2,
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	mockProjectRepo.On("Update", mock.AnythingOfType("*domain.Project")).Return(nil)

	err := projectUC.UpdateMetadata(projectID, userID, req)

	assert.NoError(t, err)

	mockProjectRepo.AssertExpectations(t)
}

// TestUpdateProject_NotFound tests project update when project doesn't exist
func TestProjectUsecase_UpdateProject_NotFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(999)

	req := domain.ProjectUpdateRequest{
		NamaProject: "Updated Name",
	}

	mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	err := projectUC.UpdateMetadata(projectID, userID, req)

	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestUpdateProject_AccessDenied tests project update when user doesn't have access
func TestProjectUsecase_UpdateProject_AccessDenied(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(2)

	project := &domain.Project{
		ID:          projectID,
		UserID:      "user-999", // Different user
		NamaProject: "Other User Project",
		Kategori:    "mobile",
		Semester:    2,
		Ukuran:      "medium",
		PathFile:    "/uploads/other.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	req := domain.ProjectUpdateRequest{
		NamaProject: "Updated Name",
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)

	err := projectUC.UpdateMetadata(projectID, userID, req)

	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestDeleteProject_Success tests successful project deletion
func TestProjectUsecase_DeleteProject_Success(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(1)

	project := &domain.Project{
		ID:          projectID,
		UserID:      userID,
		NamaProject: "Test Project",
		Kategori:    "website",
		Semester:    1,
		Ukuran:      "small",
		PathFile:    "/uploads/test.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)
	mockProjectRepo.On("Delete", projectID).Return(nil)

	err := projectUC.Delete(projectID, userID)

	assert.NoError(t, err)

	mockProjectRepo.AssertExpectations(t)
}

// TestDeleteProject_NotFound tests project deletion when project doesn't exist
func TestProjectUsecase_DeleteProject_NotFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(999)

	mockProjectRepo.On("GetByID", projectID).Return(nil, gorm.ErrRecordNotFound)

	err := projectUC.Delete(projectID, userID)

	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestDeleteProject_AccessDenied tests project deletion when user doesn't have access
func TestProjectUsecase_DeleteProject_AccessDenied(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)
	projectUC := NewProjectUsecase(mockProjectRepo, nil)

	userID := "user-1"
	projectID := uint(2)

	project := &domain.Project{
		ID:          projectID,
		UserID:      "user-999", // Different user
		NamaProject: "Other User Project",
		Kategori:    "mobile",
		Semester:    2,
		Ukuran:      "medium",
		PathFile:    "/uploads/other.zip",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockProjectRepo.On("GetByID", projectID).Return(project, nil)

	err := projectUC.Delete(projectID, userID)

	assert.Error(t, err)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestProjectUsecase_Download_SingleFile tests single project download
func TestProjectUsecase_Download_SingleFile(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	project := &domain.Project{
		ID:       1,
		UserID:   userID,
		PathFile: "/uploads/project1.pdf",
	}

	mockProjectRepo.On("GetByID", uint(1)).Return(project, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.NoError(t, err)
	assert.Equal(t, "/uploads/project1.pdf", result)

	mockProjectRepo.AssertExpectations(t)
}

// TestProjectUsecase_Download_EmptyIDs tests empty project IDs
func TestProjectUsecase_Download_EmptyIDs(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{}

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrValidation, appErr.Code)
}

// TestProjectUsecase_Download_NotFound tests project not found
func TestProjectUsecase_Download_NotFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1, 2}

	mockProjectRepo.On("GetByIDs", projectIDs, userID).Return([]domain.Project{}, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

// TestProjectUsecase_Download_Error tests error during download
func TestProjectUsecase_Download_Error(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1, 2}

	mockProjectRepo.On("GetByIDs", projectIDs, userID).Return(nil, assert.AnError)

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_SingleFile_GetOwnedProjectError(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	mockProjectRepo.On("GetByID", uint(1)).Return(nil, assert.AnError)

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_SingleFile_Forbidden(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	project := &domain.Project{
		ID:       1,
		UserID:   "user-2",
		PathFile: "/uploads/project1.pdf",
	}

	mockProjectRepo.On("GetByID", uint(1)).Return(project, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_SingleFile_PathTraversal(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	project := &domain.Project{
		ID:       1,
		UserID:   userID,
		PathFile: "../uploads/project1.pdf",
	}

	mockProjectRepo.On("GetByID", uint(1)).Return(project, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrValidation, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_MultipleFiles_Success(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "project1.txt")
	file2 := filepath.Join(tempDir, "project2.txt")
	require.NoError(t, os.WriteFile(file1, []byte("file-1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("file-2"), 0644))

	userID := "user-1"
	projectIDs := []uint{1, 2}
	projects := []domain.Project{
		{ID: 1, UserID: userID, PathFile: file1},
		{ID: 2, UserID: userID, PathFile: file2},
	}

	mockProjectRepo.On("GetByIDs", projectIDs, userID).Return(projects, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.FileExists(t, result)

	require.NoError(t, os.Remove(result))
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_MultipleFiles_PartialFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "project1.txt")
	require.NoError(t, os.WriteFile(file1, []byte("file-1"), 0644))

	userID := "user-1"
	projectIDs := []uint{1, 2, 3}
	projects := []domain.Project{
		{ID: 1, UserID: userID, PathFile: file1},
	}

	mockProjectRepo.On("GetByIDs", projectIDs, userID).Return(projects, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.FileExists(t, result)

	require.NoError(t, os.Remove(result))
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_MultipleFiles_NonexistentFile(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := helper.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1, 2}
	projects := []domain.Project{
		{ID: 1, UserID: userID, PathFile: "/tmp/this-file-should-not-exist-project1.txt"},
		{ID: 2, UserID: userID, PathFile: "/tmp/this-file-should-not-exist-project2.txt"},
	}

	mockProjectRepo.On("GetByIDs", projectIDs, userID).Return(projects, nil)

	result, err := projectUC.Download(userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}
