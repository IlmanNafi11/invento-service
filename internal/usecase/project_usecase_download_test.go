package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/storage"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestProjectUsecase_Download_SingleFile tests single project download
func TestProjectUsecase_Download_SingleFile(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	project := &domain.Project{
		ID:       1,
		UserID:   userID,
		PathFile: "/uploads/project1.pdf",
	}

	mockProjectRepo.On("GetByID", mock.Anything, uint(1)).Return(project, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

	assert.NoError(t, err)
	assert.Equal(t, "/uploads/project1.pdf", result)

	mockProjectRepo.AssertExpectations(t)
}

// TestProjectUsecase_Download_EmptyIDs tests empty project IDs
func TestProjectUsecase_Download_EmptyIDs(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{}

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

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
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1, 2}

	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDs, userID).Return([]domain.Project{}, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

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
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1, 2}

	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDs, userID).Return(nil, assert.AnError)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

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
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	mockProjectRepo.On("GetByID", mock.Anything, uint(1)).Return(nil, assert.AnError)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

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
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	project := &domain.Project{
		ID:       1,
		UserID:   "user-2",
		PathFile: "/uploads/project1.pdf",
	}

	mockProjectRepo.On("GetByID", mock.Anything, uint(1)).Return(project, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

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
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1}

	project := &domain.Project{
		ID:       1,
		UserID:   userID,
		PathFile: "../uploads/project1.pdf",
	}

	mockProjectRepo.On("GetByID", mock.Anything, uint(1)).Return(project, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

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
	fileManager := storage.NewFileManager(cfg)
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

	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDs, userID).Return(projects, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.FileExists(t, result)

	require.NoError(t, os.Remove(result))
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_MultipleFiles_PartialFound(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	tempDir := t.TempDir()
	file1 := filepath.Join(tempDir, "project1.txt")
	require.NoError(t, os.WriteFile(file1, []byte("file-1"), 0644))

	userID := "user-1"
	projectIDs := []uint{1, 2, 3}
	projects := []domain.Project{
		{ID: 1, UserID: userID, PathFile: file1},
	}

	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDs, userID).Return(projects, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)
	assert.FileExists(t, result)

	require.NoError(t, os.Remove(result))
	mockProjectRepo.AssertExpectations(t)
}

func TestProjectUsecase_Download_MultipleFiles_NonexistentFile(t *testing.T) {
	mockProjectRepo := new(MockProjectRepository)

	cfg := &config.Config{}
	fileManager := storage.NewFileManager(cfg)
	projectUC := NewProjectUsecase(mockProjectRepo, fileManager)

	userID := "user-1"
	projectIDs := []uint{1, 2}
	projects := []domain.Project{
		{ID: 1, UserID: userID, PathFile: "/tmp/this-file-should-not-exist-project1.txt"},
		{ID: 2, UserID: userID, PathFile: "/tmp/this-file-should-not-exist-project2.txt"},
	}

	mockProjectRepo.On("GetByIDs", mock.Anything, projectIDs, userID).Return(projects, nil)

	result, err := projectUC.Download(context.Background(), userID, projectIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockProjectRepo.AssertExpectations(t)
}
