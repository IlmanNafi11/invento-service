package helper

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DownloadHelper struct {
	pathResolver *PathResolver
}

func NewDownloadHelper(pathResolver *PathResolver) *DownloadHelper {
	return &DownloadHelper{
		pathResolver: pathResolver,
	}
}

func (dh *DownloadHelper) ValidateDownloadRequest(projectIDs, modulIDs []string) error {
	if len(projectIDs) == 0 && len(modulIDs) == 0 {
		return errors.New("project IDs atau modul IDs harus diisi minimal salah satu")
	}
	return nil
}

func (dh *DownloadHelper) resolvePath(relativePath string) string {
	if filepath.IsAbs(relativePath) {
		return relativePath
	}

	if dh.pathResolver == nil {
		logger := NewLogger()
		logger.Error("resolvePath - pathResolver is nil!")
		return relativePath
	}

	basePath := dh.pathResolver.GetBasePath()
	cleanBasePath := strings.TrimRight(basePath, "/")

	pathToJoin := relativePath
	if strings.HasPrefix(relativePath, "uploads/") {
		pathToJoin = strings.TrimPrefix(relativePath, "uploads/")
	}
	if strings.HasPrefix(relativePath, "temp/") {
		pathToJoin = strings.TrimPrefix(relativePath, "temp/")
	}

	resolved := filepath.Join(cleanBasePath, pathToJoin)

	absPath, err := filepath.Abs(resolved)
	if err != nil {
		logger := NewLogger()
		logger.Warn("resolvePath - Failed to get absolute path: %v, using relative path", err)
		return resolved
	}

	return absPath
}

func (dh *DownloadHelper) PrepareFilesForDownload(projects []domain.Project, moduls []domain.Modul) ([]string, []string, error) {
	var filePaths []string
	var notFoundFiles []string

	for _, project := range projects {
		resolvedPath := dh.resolvePath(project.PathFile)

		if _, err := os.Stat(resolvedPath); err == nil {
			filePaths = append(filePaths, resolvedPath)
		} else {
			notFoundFiles = append(notFoundFiles, fmt.Sprintf("Project ID %d: %s", project.ID, project.PathFile))
		}
	}

	for _, modul := range moduls {
		resolvedPath := dh.resolvePath(modul.FilePath)

		if _, err := os.Stat(resolvedPath); err == nil {
			filePaths = append(filePaths, resolvedPath)
		} else {
			notFoundFiles = append(notFoundFiles, fmt.Sprintf("Modul ID %s: %s", modul.ID, modul.FilePath))
		}
	}

	if len(filePaths) == 0 {
		return nil, notFoundFiles, errors.New("semua file tidak ditemukan di server")
	}

	return filePaths, notFoundFiles, nil
}

func (dh *DownloadHelper) CreateDownloadZip(filePaths []string, userID string) (string, error) {
	if len(filePaths) == 0 {
		return "", errors.New("tidak ada file untuk didownload")
	}

	if len(filePaths) == 1 {
		return filePaths[0], nil
	}

	basePath := dh.pathResolver.GetBasePath()
	tempDir := filepath.Join(basePath, "temp")

	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", errors.New("gagal membuat direktori temp")
	}

	identifier, err := GenerateUniqueIdentifier(8)
	if err != nil {
		return "", errors.New("gagal generate identifier")
	}

	zipFileName := fmt.Sprintf("user_%s_files_%s.zip", userID, identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", errors.New("gagal membuat file zip")
	}

	return zipFilePath, nil
}

func (dh *DownloadHelper) GetFilesByIDs(projectIDs []uint, modulIDs []string, projects []domain.Project, moduls []domain.Modul) ([]domain.Project, []domain.Modul) {
	var selectedProjects []domain.Project
	var selectedModuls []domain.Modul

	projectMap := make(map[uint]domain.Project)
	for _, p := range projects {
		projectMap[p.ID] = p
	}

	modulMap := make(map[string]domain.Modul)
	for _, m := range moduls {
		modulMap[m.ID] = m
	}

	for _, id := range projectIDs {
		if project, exists := projectMap[id]; exists {
			selectedProjects = append(selectedProjects, project)
		}
	}

	for _, id := range modulIDs {
		if modul, exists := modulMap[id]; exists {
			selectedModuls = append(selectedModuls, modul)
		}
	}

	return selectedProjects, selectedModuls
}
