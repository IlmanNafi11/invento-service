package storage

import (
	"errors"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"invento-service/config"
)

type ProjectHelper struct {
	config *config.Config
}

func NewProjectHelper(cfg *config.Config) *ProjectHelper {
	return &ProjectHelper{
		config: cfg,
	}
}

func (ph *ProjectHelper) GenerateProjectIdentifier() (string, error) {
	return GenerateRandomString(10)
}

func (ph *ProjectHelper) BuildProjectPath(userID uint, identifier string, filename string) string {
	basePath := ph.getBasePath()
	return filepath.Join(basePath, "projects", fmt.Sprintf("%d", userID), identifier, filename)
}

func (ph *ProjectHelper) BuildProjectDirectory(userID uint, identifier string) string {
	basePath := ph.getBasePath()
	return filepath.Join(basePath, "projects", fmt.Sprintf("%d", userID), identifier)
}

func (ph *ProjectHelper) getBasePath() string {
	if ph.config.App.Env == config.EnvProduction {
		return ph.config.Upload.PathProduction
	}
	return ph.config.Upload.PathDevelopment
}

func (ph *ProjectHelper) ValidateProjectFile(fileHeader *multipart.FileHeader) error {
	ext := GetFileExtension(fileHeader.Filename)
	if ext != ".zip" {
		return errors.New("file project harus berupa zip")
	}

	maxSize := ph.config.Upload.MaxSize
	if fileHeader.Size > maxSize {
		maxSizeMB := maxSize / (1024 * 1024)
		return fmt.Errorf("ukuran file project tidak boleh lebih dari %d MB", maxSizeMB)
	}

	return nil
}

func (ph *ProjectHelper) ValidateProjectFileSize(fileSize int64) error {
	if fileSize <= 0 {
		return errors.New("ukuran file tidak valid")
	}

	maxSize := ph.config.Upload.MaxSize
	if fileSize > maxSize {
		maxSizeMB := maxSize / (1024 * 1024)
		return fmt.Errorf("ukuran file melebihi batas maksimal %d MB", maxSizeMB)
	}

	return nil
}

func ValidateProjectZipExtension(filename string) error {
	ext := GetFileExtension(filename)
	if ext != ".zip" {
		return errors.New("file project harus berupa zip")
	}
	return nil
}
