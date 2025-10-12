package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type ProjectUsecase interface {
	Create(userID uint, userEmail string, userRole string, files []*multipart.FileHeader, namaProjects []string, kategori []string, semesters []int) (*domain.ProjectCreateResponse, error)
	GetList(userID uint, search string, filterSemester int, filterKategori string, page, limit int) (*domain.ProjectListData, error)
	GetByID(projectID, userID uint) (*domain.ProjectResponse, error)
	Update(projectID, userID uint, namaProject string, kategori string, semester int, file *multipart.FileHeader) (*domain.ProjectResponse, error)
	Delete(projectID, userID uint) error
	Download(userID uint, projectIDs []uint) (string, error)
}

type projectUsecase struct {
	projectRepo repo.ProjectRepository
}

func NewProjectUsecase(projectRepo repo.ProjectRepository) ProjectUsecase {
	return &projectUsecase{
		projectRepo: projectRepo,
	}
}

func (uc *projectUsecase) Create(userID uint, userEmail string, userRole string, files []*multipart.FileHeader, namaProjects []string, kategori []string, semesters []int) (*domain.ProjectCreateResponse, error) {
	if len(files) == 0 {
		return nil, errors.New("file wajib diupload")
	}

	if len(files) != len(namaProjects) || len(files) != len(semesters) {
		return nil, errors.New("jumlah file, nama project, dan semester harus sama")
	}

	userDir, err := helper.CreateUserDirectory(userEmail, userRole)
	if err != nil {
		return nil, errors.New("gagal membuat direktori user")
	}

	var projectResponses []domain.ProjectResponse

	for i, fileHeader := range files {
		if err := helper.ValidateZipFile(fileHeader); err != nil {
			return nil, err
		}

		kategoriInput := kategori
		projectKategori := "website"
		if i < len(kategoriInput) && kategoriInput[i] != "" {
			projectKategori = kategoriInput[i]
		}

		ukuran := helper.GetFileSize(fileHeader)

		filename := fileHeader.Filename
		destPath := filepath.Join(userDir, filename)

		if err := helper.SaveUploadedFile(fileHeader, destPath); err != nil {
			return nil, errors.New("gagal menyimpan file")
		}

		project := &domain.Project{
			UserID:      userID,
			NamaProject: namaProjects[i],
			Kategori:    projectKategori,
			Semester:    semesters[i],
			Ukuran:      ukuran,
			PathFile:    destPath,
		}

		if err := uc.projectRepo.Create(project); err != nil {
			helper.DeleteFile(destPath)
			return nil, errors.New("gagal menyimpan data project")
		}

		projectResponses = append(projectResponses, domain.ProjectResponse{
			ID:          project.ID,
			NamaProject: project.NamaProject,
			Kategori:    projectKategori,
			Semester:    project.Semester,
			Ukuran:      project.Ukuran,
			PathFile:    project.PathFile,
			CreatedAt:   project.CreatedAt,
			UpdatedAt:   project.UpdatedAt,
		})
	}

	return &domain.ProjectCreateResponse{
		Items: projectResponses,
	}, nil
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

func (uc *projectUsecase) Update(projectID, userID uint, namaProject string, kategori string, semester int, file *multipart.FileHeader) (*domain.ProjectResponse, error) {
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

	if namaProject != "" {
		project.NamaProject = namaProject
	}

	if kategori != "" {
		project.Kategori = kategori
	}

	if semester > 0 {
		project.Semester = semester
	}

	if file != nil {
		if err := helper.ValidateZipFile(file); err != nil {
			return nil, err
		}

		oldPath := project.PathFile

		ukuran := helper.GetFileSize(file)

		filename := file.Filename
		destPath := filepath.Join(filepath.Dir(oldPath), filename)

		if err := helper.SaveUploadedFile(file, destPath); err != nil {
			return nil, errors.New("gagal menyimpan file")
		}

		helper.DeleteFile(oldPath)

		project.Ukuran = ukuran
		project.PathFile = destPath
	}

	if err := uc.projectRepo.Update(project); err != nil {
		return nil, errors.New("gagal memperbarui data project")
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
