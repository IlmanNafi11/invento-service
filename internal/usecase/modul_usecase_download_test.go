package usecase

import (
	"context"
	"errors"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	apperrors "invento-service/internal/errors"
)

func TestModulUsecase_GetList_WithFilters(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	expectedModuls := []dto.ModulListItem{
		{
			ID:                 "550e8400-e29b-41d4-a716-446655440001",
			Judul:              "Test Modul",
			Deskripsi:          "Test Deskripsi",
			FileName:           "test.pdf",
			MimeType:           "application/pdf",
			FileSize:           1572864,
			Status:             "completed",
			TerakhirDiperbarui: time.Now(),
		},
	}

	mockModulRepo.On("GetByUserID", mock.Anything, "user-1", "test", "application/pdf", "completed", 1, 10).
		Return(expectedModuls, 1, nil)

	result, err := modulUC.GetList(context.Background(), "user-1", "test", "application/pdf", "completed", 1, 10)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Items, 1)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_Success tests successful modul retrieval
func TestModulUsecase_GetByID_Success(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "user-1"

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    userID,
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(expectedModul, nil)

	result, err := modulUC.GetByID(context.Background(), modulID, userID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, modulID, result.ID)
	assert.Equal(t, "Test Modul", result.Judul)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_NotFound tests modul retrieval when not found
func TestModulUsecase_GetByID_NotFound(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440999"
	userID := "user-1"

	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(nil, apperrors.ErrRecordNotFound)

	result, err := modulUC.GetByID(context.Background(), modulID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_GetByID_Unauthorized tests access denial for different user
func TestModulUsecase_GetByID_Unauthorized(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440000"
	userID := "user-2"

	expectedModul := &domain.Modul{
		ID:        modulID,
		UserID:    "user-1",
		Judul:     "Test Modul",
		Deskripsi: "Test Deskripsi",
		FileName:  "test.pdf",
		MimeType:  "application/pdf",
		FileSize:  1572864,
		FilePath:  "/uploads/test.pdf",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(expectedModul, nil)

	result, err := modulUC.GetByID(context.Background(), modulID, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_Download_SingleFile tests single file download
func TestModulUsecase_Download_SingleFile(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	modulIDs := []string{"550e8400-e29b-41d4-a716-446655440000"}

	tempFile, err := os.CreateTemp(t.TempDir(), "modul-*.pdf")
	require.NoError(t, err)
	require.NoError(t, tempFile.Close())

	expectedModuls := []domain.Modul{
		{
			ID:        "550e8400-e29b-41d4-a716-446655440000",
			UserID:    userID,
			Judul:     "Test Modul.pdf",
			Deskripsi: "Test Deskripsi",
			FileName:  "test.pdf",
			FilePath:  tempFile.Name(),
		},
	}

	mockModulRepo.On("GetByIDs", mock.Anything, modulIDs, userID).Return(expectedModuls, nil)

	result, err := modulUC.Download(context.Background(), userID, modulIDs)

	assert.NoError(t, err)
	assert.Equal(t, tempFile.Name(), result)

	mockModulRepo.AssertExpectations(t)
}

// TestModulUsecase_Download_EmptyIDs tests empty modul IDs
func TestModulUsecase_Download_EmptyIDs(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	modulIDs := []string{}

	result, err := modulUC.Download(context.Background(), userID, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrValidation, appErr.Code)
}

// TestModulUsecase_Download_NotFound tests download when moduls not found
func TestModulUsecase_Download_NotFound(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	userID := "user-1"
	modulIDs := []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002"}

	mockModulRepo.On("GetByIDs", mock.Anything, modulIDs, userID).Return([]domain.Modul{}, nil)

	result, err := modulUC.Download(context.Background(), userID, modulIDs)

	assert.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	assert.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_UpdateMetadata_NotFound(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440999"
	userID := "user-1"
	req := dto.UpdateModulRequest{Judul: "Baru"}

	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(nil, apperrors.ErrRecordNotFound).Once()

	err := modulUC.UpdateMetadata(context.Background(), modulID, userID, req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_UpdateMetadata_Unauthorized(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440001"
	req := dto.UpdateModulRequest{Judul: "Baru"}

	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(&domain.Modul{ID: modulID, UserID: "owner"}, nil).Once()

	err := modulUC.UpdateMetadata(context.Background(), modulID, "user-1", req)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockModulRepo.AssertNotCalled(t, "UpdateMetadata", mock.Anything)
	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_Delete_NotFound(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440999"
	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(nil, apperrors.ErrRecordNotFound).Once()

	err := modulUC.Delete(context.Background(), modulID, "user-1")
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrNotFound, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_Delete_Unauthorized(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440001"
	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(&domain.Modul{ID: modulID, UserID: "owner"}, nil).Once()

	err := modulUC.Delete(context.Background(), modulID, "user-1")
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrForbidden, appErr.Code)

	mockModulRepo.AssertNotCalled(t, "Delete", modulID)
	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_Download_RepositoryError(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulIDs := []string{"550e8400-e29b-41d4-a716-446655440001"}
	mockModulRepo.On("GetByIDs", mock.Anything, modulIDs, "user-1").Return(nil, assert.AnError).Once()

	result, err := modulUC.Download(context.Background(), "user-1", modulIDs)
	require.Error(t, err)
	assert.Empty(t, result)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_GetList_InvalidPagination(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	mockModulRepo.On("GetByUserID", mock.Anything, "user-1", "", "", "", 1, 10).
		Return([]dto.ModulListItem{}, 0, nil).Once()

	result, err := modulUC.GetList(context.Background(), "user-1", "", "", "", 0, 0)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.Equal(t, 1, result.Pagination.Page)
	assert.Equal(t, 10, result.Pagination.Limit)

	mockModulRepo.AssertExpectations(t)
}

func TestModulUsecase_GetByID_RepositoryError(t *testing.T) {
	t.Parallel()
	mockModulRepo := new(MockModulRepository)
	modulUC := NewModulUsecase(mockModulRepo)

	modulID := "550e8400-e29b-41d4-a716-446655440001"
	mockModulRepo.On("GetByID", mock.Anything, modulID).Return(nil, assert.AnError).Once()

	result, err := modulUC.GetByID(context.Background(), modulID, "user-1")
	require.Error(t, err)
	assert.Nil(t, result)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, apperrors.ErrInternal, appErr.Code)

	mockModulRepo.AssertExpectations(t)
}
