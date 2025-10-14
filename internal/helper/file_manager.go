package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fiber-boiler-plate/config"
	"fmt"
	"os"
	"path/filepath"
)

type FileManager struct {
	config *config.Config
}

func NewFileManager(cfg *config.Config) *FileManager {
	return &FileManager{
		config: cfg,
	}
}

func (fm *FileManager) GenerateRandomDirectory() (string, error) {
	bytes := make([]byte, 5)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (fm *FileManager) GetUserUploadPath(userID uint) (string, error) {
	var basePath string

	if fm.config.App.Env == "production" {
		basePath = fm.config.Upload.PathProduction
	} else {
		basePath = fm.config.Upload.PathDevelopment
	}

	userDirPath := filepath.Join(basePath, fmt.Sprintf("%d", userID))

	if err := os.MkdirAll(userDirPath, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori user: %w", err)
	}

	return userDirPath, nil
}

func (fm *FileManager) CreateProjectUploadDirectory(userID uint) (string, string, error) {
	randomDir, err := fm.GenerateRandomDirectory()
	if err != nil {
		return "", "", fmt.Errorf("gagal generate random directory: %w", err)
	}

	userUploadPath, err := fm.GetUserUploadPath(userID)
	if err != nil {
		return "", "", err
	}

	projectDirPath := filepath.Join(userUploadPath, randomDir)

	if err := os.MkdirAll(projectDirPath, 0755); err != nil {
		return "", "", fmt.Errorf("gagal membuat direktori project: %w", err)
	}

	return projectDirPath, randomDir, nil
}

func (fm *FileManager) GetProjectFilePath(userID uint, randomDir, filename string) string {
	var basePath string

	if fm.config.App.Env == "production" {
		basePath = fm.config.Upload.PathProduction
	} else {
		basePath = fm.config.Upload.PathDevelopment
	}

	return filepath.Join(basePath, fmt.Sprintf("%d", userID), randomDir, filename)
}

func (fm *FileManager) DeleteUserDirectory(userID uint) error {
	var basePath string

	if fm.config.App.Env == "production" {
		basePath = fm.config.Upload.PathProduction
	} else {
		basePath = fm.config.Upload.PathDevelopment
	}

	userDirPath := filepath.Join(basePath, fmt.Sprintf("%d", userID))

	return os.RemoveAll(userDirPath)
}

func (fm *FileManager) DeleteProjectDirectory(userID uint, randomDir string) error {
	var basePath string

	if fm.config.App.Env == "production" {
		basePath = fm.config.Upload.PathProduction
	} else {
		basePath = fm.config.Upload.PathDevelopment
	}

	projectDirPath := filepath.Join(basePath, fmt.Sprintf("%d", userID), randomDir)

	return os.RemoveAll(projectDirPath)
}

func (fm *FileManager) GetUploadFilePath(uploadID string) string {
	var basePath string

	if fm.config.App.Env == "production" {
		basePath = fm.config.Upload.TempPathProduction
	} else {
		basePath = fm.config.Upload.TempPathDevelopment
	}

	return filepath.Join(basePath, "uploads", uploadID)
}

func (fm *FileManager) GetModulBasePath() string {
	if fm.config.App.Env == "production" {
		return fm.config.Upload.PathProduction
	}
	return fm.config.Upload.PathDevelopment
}

func (fm *FileManager) GetUserModulPath(userID uint) (string, error) {
	basePath := fm.GetModulBasePath()
	userModulPath := filepath.Join(basePath, "moduls", fmt.Sprintf("%d", userID))

	if err := os.MkdirAll(userModulPath, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori modul user: %w", err)
	}

	return userModulPath, nil
}

func (fm *FileManager) CreateModulUploadDirectory(userID uint) (string, string, error) {
	randomDir, err := fm.GenerateRandomDirectory()
	if err != nil {
		return "", "", fmt.Errorf("gagal generate random directory: %w", err)
	}

	userModulPath, err := fm.GetUserModulPath(userID)
	if err != nil {
		return "", "", err
	}

	modulDirPath := filepath.Join(userModulPath, randomDir)

	if err := os.MkdirAll(modulDirPath, 0755); err != nil {
		return "", "", fmt.Errorf("gagal membuat direktori modul: %w", err)
	}

	return modulDirPath, randomDir, nil
}

func (fm *FileManager) GetModulFilePath(userID uint, randomDir, filename string) string {
	basePath := fm.GetModulBasePath()
	return filepath.Join(basePath, "moduls", fmt.Sprintf("%d", userID), randomDir, filename)
}

func (fm *FileManager) DeleteModulDirectory(userID uint, randomDir string) error {
	basePath := fm.GetModulBasePath()
	modulDirPath := filepath.Join(basePath, "moduls", fmt.Sprintf("%d", userID), randomDir)
	return os.RemoveAll(modulDirPath)
}

func (fm *FileManager) DeleteUserModulDirectory(userID uint) error {
	basePath := fm.GetModulBasePath()
	userModulPath := filepath.Join(basePath, "moduls", fmt.Sprintf("%d", userID))
	return os.RemoveAll(userModulPath)
}
