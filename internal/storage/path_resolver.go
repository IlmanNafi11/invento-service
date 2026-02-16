package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"invento-service/config"
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
	if pr.env == config.EnvProduction {
		return pr.pathProduction
	}
	return pr.pathDevelopment
}

func (pr *PathResolver) GetTempPath() string {
	if pr.env == config.EnvProduction {
		return pr.tempPathProduction
	}
	return pr.tempPathDevelopment
}

func (pr *PathResolver) GetProjectPath(userID string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "projects", userID)
}

func (pr *PathResolver) GetProjectFilePath(userID string, identifier string, filename string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "projects", userID, identifier, filename)
}

func (pr *PathResolver) GetProjectDirectory(userID string, identifier string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "projects", userID, identifier)
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
	if err := os.MkdirAll(path, 0o755); err != nil {
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

func (pr *PathResolver) GetProfilPath(userID string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "profil", userID)
}

func (pr *PathResolver) GetProfilFilePath(userID string, filename string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "profil", userID, filename)
}

func (pr *PathResolver) GetProfilDirectory(userID string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "profil", userID)
}

func (pr *PathResolver) GetModulPath(userID uint, identifier string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "moduls", fmt.Sprintf("%d", userID), identifier)
}

func (pr *PathResolver) GetModulFilePath(userID uint, identifier string, filename string) string {
	basePath := pr.GetBasePath()
	return filepath.Join(basePath, "moduls", fmt.Sprintf("%d", userID), identifier, filename)
}

func (pr *PathResolver) ConvertToAPIPath(absolutePath *string) *string {
	if absolutePath == nil {
		return nil
	}

	if *absolutePath == "" {
		return absolutePath
	}

	basePath := pr.GetBasePath()
	path := *absolutePath

	path = filepath.Clean(path)
	basePath = filepath.Clean(basePath)

	if len(path) > len(basePath) && strings.HasPrefix(path, basePath) {
		relativePath := path[len(basePath):]

		relativePath = filepath.ToSlash(relativePath)

		if relativePath != "" && relativePath[0] != '/' {
			relativePath = "/" + relativePath
		}

		apiPath := "/uploads" + relativePath
		result := new(string)
		*result = apiPath
		return result
	}

	return absolutePath
}
