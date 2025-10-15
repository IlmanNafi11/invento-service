package helper

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fmt"
	"os"
	"path/filepath"
)

type DownloadHelper struct{}

func NewDownloadHelper() *DownloadHelper {
	return &DownloadHelper{}
}

func (dh *DownloadHelper) ValidateDownloadRequest(projectIDs, modulIDs []uint) error {
	if len(projectIDs) == 0 && len(modulIDs) == 0 {
		return errors.New("project IDs atau modul IDs harus diisi minimal salah satu")
	}
	return nil
}

func (dh *DownloadHelper) PrepareFilesForDownload(projects []domain.Project, moduls []domain.Modul) ([]string, []string, error) {
	var filePaths []string
	var notFoundFiles []string
	
	for _, project := range projects {
		if _, err := os.Stat(project.PathFile); err == nil {
			filePaths = append(filePaths, project.PathFile)
		} else {
			notFoundFiles = append(notFoundFiles, fmt.Sprintf("Project ID %d: %s", project.ID, project.PathFile))
		}
	}
	
	for _, modul := range moduls {
		if _, err := os.Stat(modul.PathFile); err == nil {
			filePaths = append(filePaths, modul.PathFile)
		} else {
			notFoundFiles = append(notFoundFiles, fmt.Sprintf("Modul ID %d: %s", modul.ID, modul.PathFile))
		}
	}
	
	if len(filePaths) == 0 {
		return nil, notFoundFiles, errors.New("semua file tidak ditemukan di server")
	}
	
	return filePaths, notFoundFiles, nil
}

func (dh *DownloadHelper) CreateDownloadZip(filePaths []string, userID uint) (string, error) {
	if len(filePaths) == 0 {
		return "", errors.New("tidak ada file untuk didownload")
	}
	
	if len(filePaths) == 1 {
		return filePaths[0], nil
	}
	
	tempDir := "./uploads/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", errors.New("gagal membuat direktori temp")
	}
	
	identifier, err := GenerateUniqueIdentifier(8)
	if err != nil {
		return "", errors.New("gagal generate identifier")
	}
	
	zipFileName := fmt.Sprintf("user_%d_files_%s.zip", userID, identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)
	
	if err := CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", errors.New("gagal membuat file zip")
	}
	
	return zipFilePath, nil
}

func (dh *DownloadHelper) GetFilesByIDs(projectIDs, modulIDs []uint, projects []domain.Project, moduls []domain.Modul) ([]domain.Project, []domain.Modul) {
	var selectedProjects []domain.Project
	var selectedModuls []domain.Modul
	
	projectMap := make(map[uint]domain.Project)
	for _, p := range projects {
		projectMap[p.ID] = p
	}
	
	modulMap := make(map[uint]domain.Modul)
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
