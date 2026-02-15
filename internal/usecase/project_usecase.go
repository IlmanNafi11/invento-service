package usecase

import (
	"errors"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/helper"
	"invento-service/internal/usecase/repo"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/gorm"
)

type ProjectUsecase interface {
	GetList(userID string, search string, filterSemester int, filterKategori string, page, limit int) (*domain.ProjectListData, error)
	GetByID(projectID uint, userID string) (*domain.ProjectResponse, error)
	UpdateMetadata(projectID uint, userID string, req domain.ProjectUpdateRequest) error
	Delete(projectID uint, userID string) error
	Download(userID string, projectIDs []uint) (string, error)
}

type projectUsecase struct {
	projectRepo repo.ProjectRepository
	fileManager *helper.FileManager
}

func NewProjectUsecase(projectRepo repo.ProjectRepository, fileManager *helper.FileManager) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
		fileManager: fileManager,
	}
}

func (uc *projectUsecase) GetList(userID string, search string, filterSemester int, filterKategori string, page, limit int) (*domain.ProjectListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	projects, total, err := uc.projectRepo.GetByUserID(userID, search, filterSemester, filterKategori, page, limit)
	if err != nil {
		return nil, newInternalError("gagal mengambil data project", err)
	}

	totalPages := (total + limit - 1) / limit

	return &domain.ProjectListData{
		Items: projects,
		Pagination: domain.PaginationData{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (uc *projectUsecase) GetByID(projectID uint, userID string) (*domain.ProjectResponse, error) {
	project, err := uc.getOwnedProject(projectID, userID)
	if err != nil {
		return nil, err
	}

	return &domain.ProjectResponse{
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

func (uc *projectUsecase) UpdateMetadata(projectID uint, userID string, req domain.ProjectUpdateRequest) error {
	project, err := uc.getOwnedProject(projectID, userID)
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

	if err := uc.projectRepo.Update(project); err != nil {
		return newInternalError("gagal mengupdate project", err)
	}

	return nil
}

func (uc *projectUsecase) Delete(projectID uint, userID string) error {
	project, err := uc.getOwnedProject(projectID, userID)
	if err != nil {
		return err
	}

	if err := uc.projectRepo.Delete(projectID); err != nil {
		return newInternalError("gagal menghapus project", err)
	}

	if project.PathFile != "" {
		if err := helper.DeleteFile(project.PathFile); err != nil {
			log.Printf("WARNING: gagal menghapus file project %s: %v", project.PathFile, err)
		}
	}

	return nil
}

func (uc *projectUsecase) Download(userID string, projectIDs []uint) (string, error) {
	if len(projectIDs) == 0 {
		return "", apperrors.NewValidationError("id project tidak boleh kosong", nil)
	}

	if len(projectIDs) == 1 {
		project, err := uc.getOwnedProject(projectIDs[0], userID)
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

	projects, err := uc.projectRepo.GetByIDs(projectIDs, userID)
	if err != nil {
		return "", newInternalError("gagal mengambil data project", err)
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
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", newInternalError("gagal membuat direktori temp", err)
	}

	identifier, err := helper.GenerateUniqueIdentifier(8)
	if err != nil {
		return "", newInternalError("gagal generate identifier", err)
	}

	zipFileName := fmt.Sprintf("projects_%s.zip", identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := helper.CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", newInternalError("gagal membuat file zip", err)
	}

	go func(zipPath string) {
		time.Sleep(5 * time.Minute)
		if err := os.Remove(zipPath); err != nil && !errors.Is(err, os.ErrNotExist) {
			log.Printf("WARNING: gagal menghapus file zip sementara %s: %v", zipPath, err)
		}
	}(zipFilePath)

	return zipFilePath, nil
}

func (uc *projectUsecase) getOwnedProject(projectID uint, userID string) (*domain.Project, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("project")
		}

		return nil, newInternalError("gagal mengambil data project", err)
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
