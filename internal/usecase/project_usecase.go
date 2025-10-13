package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type ProjectUsecase interface {
	GetList(userID uint, search string, filterSemester int, filterKategori string, page, limit int) (*domain.ProjectListData, error)
	GetByID(projectID, userID uint) (*domain.ProjectResponse, error)
	UpdateMetadata(projectID, userID uint, req domain.ProjectUpdateMetadataRequest) error
	Delete(projectID, userID uint) error
	Download(userID uint, projectIDs []uint) (string, error)
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

func (uc *projectUsecase) GetList(userID uint, search string, filterSemester int, filterKategori string, page, limit int) (*domain.ProjectListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	projects, total, err := uc.projectRepo.GetByUserID(userID, search, filterSemester, filterKategori, page, limit)
	if err != nil {
		return nil, errors.New("gagal mengambil data project")
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

func (uc *projectUsecase) GetByID(projectID, userID uint) (*domain.ProjectResponse, error) {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("project tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data project")
	}

	if project.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke project ini")
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

func (uc *projectUsecase) UpdateMetadata(projectID, userID uint, req domain.ProjectUpdateMetadataRequest) error {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project tidak ditemukan")
		}
		return errors.New("gagal mengambil data project")
	}

	if project.UserID != userID {
		return errors.New("tidak memiliki akses ke project ini")
	}

	project.NamaProject = req.NamaProject
	project.Kategori = req.Kategori
	project.Semester = req.Semester

	if err := uc.projectRepo.Update(project); err != nil {
		return errors.New("gagal update metadata project")
	}

	return nil
}

func (uc *projectUsecase) Delete(projectID, userID uint) error {
	project, err := uc.projectRepo.GetByID(projectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("project tidak ditemukan")
		}
		return errors.New("gagal mengambil data project")
	}

	if project.UserID != userID {
		return errors.New("tidak memiliki akses ke project ini")
	}

	helper.DeleteFile(project.PathFile)

	if err := uc.projectRepo.Delete(projectID); err != nil {
		return errors.New("gagal menghapus project")
	}

	return nil
}

func (uc *projectUsecase) Download(userID uint, projectIDs []uint) (string, error) {
	if len(projectIDs) == 0 {
		return "", errors.New("id project tidak boleh kosong")
	}

	projects, err := uc.projectRepo.GetByIDs(projectIDs, userID)
	if err != nil {
		return "", errors.New("gagal mengambil data project")
	}

	if len(projects) == 0 {
		return "", errors.New("project tidak ditemukan")
	}

	if len(projects) == 1 {
		return projects[0].PathFile, nil
	}

	var filePaths []string
	for _, project := range projects {
		filePaths = append(filePaths, project.PathFile)
	}

	tempDir := "./uploads/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", errors.New("gagal membuat direktori temp")
	}

	identifier, err := helper.GenerateUniqueIdentifier(8)
	if err != nil {
		return "", errors.New("gagal generate identifier")
	}

	zipFileName := fmt.Sprintf("projects_%s.zip", identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := helper.CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", errors.New("gagal membuat file zip")
	}

	return zipFilePath, nil
}
