package usecase

import (
	"context"
	"errors"
	"fmt"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/storage"
	"invento-service/internal/usecase/repo"
	"os"
	"path/filepath"
	"strings"
	"time"

	apperrors "invento-service/internal/errors"

	zlog "github.com/rs/zerolog/log"
)

type ProjectUsecase interface {
	GetList(ctx context.Context, userID, search string, filterSemester int, filterKategori string, page, limit int) (*dto.ProjectListData, error)
	GetByID(ctx context.Context, projectID uint, userID string) (*dto.ProjectResponse, error)
	UpdateMetadata(ctx context.Context, projectID uint, userID string, req dto.UpdateProjectRequest) error
	Delete(ctx context.Context, projectID uint, userID string) error
	Download(ctx context.Context, userID string, projectIDs []uint) (string, error)
}

type projectUsecase struct {
	projectRepo repo.ProjectRepository
	fileManager *storage.FileManager
}

func NewProjectUsecase(projectRepo repo.ProjectRepository, fileManager *storage.FileManager) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
		fileManager: fileManager,
	}
}

func (uc *projectUsecase) GetList(ctx context.Context, userID, search string, filterSemester int, filterKategori string, page, limit int) (*dto.ProjectListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	projects, total, err := uc.projectRepo.GetByUserID(ctx, userID, search, filterSemester, filterKategori, page, limit)
	if err != nil {
		return nil, newInternalError("gagal mengambil data project", fmt.Errorf("ProjectUsecase.GetList: %w", err))
	}

	totalPages := (total + limit - 1) / limit

	return &dto.ProjectListData{
		Items: projects,
		Pagination: dto.PaginationData{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (uc *projectUsecase) GetByID(ctx context.Context, projectID uint, userID string) (*dto.ProjectResponse, error) {
	project, err := uc.getOwnedProject(ctx, projectID, userID)
	if err != nil {
		return nil, err
	}

	return &dto.ProjectResponse{
		ID:          project.ID,
		NamaProject: project.NamaProject,
		Kategori:    project.Kategori,
		Semester:    project.Semester,
		Ukuran:      project.Ukuran,
		PathFile:    project.PathFile,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}, nil
}

func (uc *projectUsecase) UpdateMetadata(ctx context.Context, projectID uint, userID string, req dto.UpdateProjectRequest) error {
	project, err := uc.getOwnedProject(ctx, projectID, userID)
	if err != nil {
		return err
	}

	if req.NamaProject != "" {
		project.NamaProject = req.NamaProject
	}
	if req.Kategori != "" {
		project.Kategori = req.Kategori
	}
	if req.Semester > 0 {
		project.Semester = req.Semester
	}

	if err := uc.projectRepo.Update(ctx, project); err != nil {
		return newInternalError("gagal mengupdate project", fmt.Errorf("ProjectUsecase.UpdateMetadata: %w", err))
	}

	return nil
}

func (uc *projectUsecase) Delete(ctx context.Context, projectID uint, userID string) error {
	project, err := uc.getOwnedProject(ctx, projectID, userID)
	if err != nil {
		return err
	}

	if err := uc.projectRepo.Delete(ctx, projectID); err != nil {
		return newInternalError("gagal menghapus project", fmt.Errorf("ProjectUsecase.Delete: %w", err))
	}

	if project.PathFile != "" {
		if err := storage.DeleteFile(project.PathFile); err != nil {
			// File deletion after DB delete is critical but non-blocking;
			// log the error so it can be investigated.
			zlog.Warn().Err(err).Str("file", project.PathFile).Msg("ProjectUsecase.Delete: failed to delete project file")
		}
	}

	return nil
}

func (uc *projectUsecase) Download(ctx context.Context, userID string, projectIDs []uint) (string, error) {
	if len(projectIDs) == 0 {
		return "", apperrors.NewValidationError("id project tidak boleh kosong", nil)
	}

	if len(projectIDs) == 1 {
		project, err := uc.getOwnedProject(ctx, projectIDs[0], userID)
		if err != nil {
			return "", err
		}

		if project.PathFile == "" {
			return "", apperrors.NewNotFoundError("file project")
		}

		cleanPath := filepath.Clean(project.PathFile)
		if strings.Contains(cleanPath, "..") {
			return "", apperrors.NewValidationError("path file tidak valid", nil)
		}

		return cleanPath, nil
	}

	projects, err := uc.projectRepo.GetByIDs(ctx, projectIDs, userID)
	if err != nil {
		return "", newInternalError("gagal mengambil data project", fmt.Errorf("ProjectUsecase.Download: %w", err))
	}

	if len(projects) == 0 {
		return "", apperrors.NewNotFoundError("project")
	}

	var filePaths []string
	for _, project := range projects {
		if project.PathFile == "" {
			return "", apperrors.NewNotFoundError("file project")
		}

		cleanPath := filepath.Clean(project.PathFile)
		if strings.Contains(cleanPath, "..") {
			return "", apperrors.NewValidationError("path file tidak valid", nil)
		}

		filePaths = append(filePaths, cleanPath)
	}

	tempDir := "./uploads/temp"
	if err = os.MkdirAll(tempDir, 0o755); err != nil {
		return "", newInternalError("gagal membuat direktori temp", fmt.Errorf("ProjectUsecase.Download: %w", err))
	}

	identifier, err := storage.GenerateUniqueIdentifier(8)
	if err != nil {
		return "", newInternalError("gagal generate identifier", fmt.Errorf("ProjectUsecase.Download: %w", err))
	}

	zipFileName := fmt.Sprintf("projects_%s.zip", identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := storage.CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", newInternalError("gagal membuat file zip", fmt.Errorf("ProjectUsecase.Download: %w", err))
	}

	go func(zipPath string) {
		time.Sleep(5 * time.Minute)
		if err := os.Remove(zipPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			zlog.Warn().Err(err).Str("file", zipPath).Msg("failed to delete temporary zip file")
		}
	}(zipFilePath)

	return zipFilePath, nil
}

func (uc *projectUsecase) getOwnedProject(ctx context.Context, projectID uint, userID string) (*domain.Project, error) {
	project, err := uc.projectRepo.GetByID(ctx, projectID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("project")
		}

		return nil, newInternalError("gagal mengambil data project", fmt.Errorf("ProjectUsecase.getOwnedProject: %w", err))
	}

	if project.UserID != userID {
		return nil, apperrors.NewForbiddenError("anda tidak memiliki akses ke project ini")
	}

	return project, nil
}

func newInternalError(message string, err error) *apperrors.AppError {
	appErr := apperrors.NewInternalError(err)
	appErr.Message = message
	return appErr
}
