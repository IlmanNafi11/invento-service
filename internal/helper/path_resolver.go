package helper

import (
	"fiber-boiler-plate/config"
	"fmt"
	"os"
	"path/filepath"
)

type PathResolver struct {
	env                 string
	pathProduction      string
	pathDevelopment     string
	tempPathProduction  string
	tempPathDevelopment string
}

func NewPathResolver(cfg *config.Config) *PathResolver {
	return &PathResolver{
		env:                 cfg.App.Env,
		pathProduction:      cfg.Upload.PathProduction,
		pathDevelopment:     cfg.Upload.PathDevelopment,
		tempPathProduction:  cfg.Upload.TempPathProduction,
		tempPathDevelopment: cfg.Upload.TempPathDevelopment,
	}
}

func (pr *PathResolver) GetBasePath() string {
	if pr.env == "production" {
		return pr.pathProduction
	}
	return pr.pathDevelopment
}

func (pr *PathResolver) GetTempPath() string {
	if pr.env == "production" {
		return pr.tempPathProduction
	}
	return pr.tempPathDevelopment
}

func (pr *PathResolver) GetProjectPath(userID uint) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "project")
}

func (pr *PathResolver) GetUploadPath(uploadID string) string {
	tempPath := pr.GetTempPath()
	return filepath.Join(tempPath, "uploads", uploadID)
}

func (pr *PathResolver) GetUploadFilePath(uploadID string) string {
	uploadPath := pr.GetUploadPath(uploadID)
	return filepath.Join(uploadPath, "file.zip")
}

func (pr *PathResolver) GetUploadInfoPath(uploadID string) string {
	uploadPath := pr.GetUploadPath(uploadID)
	return filepath.Join(uploadPath, "info.json")
}

func (pr *PathResolver) EnsureDirectoryExists(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("gagal membuat direktori: %w", err)
	}
	return nil
}

func (pr *PathResolver) DirectoryExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}

func (pr *PathResolver) FileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
