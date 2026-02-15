package helper

import (
	"errors"
	"mime/multipart"
	"path/filepath"

	"invento-service/config"
)

type ModulHelper struct {
	config *config.Config
}

func NewModulHelper(cfg *config.Config) *ModulHelper {
	return &ModulHelper{
		config: cfg,
	}
}

func (mh *ModulHelper) GenerateModulIdentifier() (string, error) {
	return GenerateRandomString(10)
}

func (mh *ModulHelper) BuildModulPath(userID string, identifier string, filename string) string {
	basePath := mh.getBasePath()
	return filepath.Join(basePath, "moduls", userID, identifier, filename)
}

func (mh *ModulHelper) BuildModulDirectory(userID string, identifier string) string {
	basePath := mh.getBasePath()
	return filepath.Join(basePath, "moduls", userID, identifier)
}

func (mh *ModulHelper) getBasePath() string {
	if mh.config.App.Env == "production" {
		return mh.config.Upload.PathProduction
	}
	return mh.config.Upload.PathDevelopment
}

func (mh *ModulHelper) ValidateModulFile(fileHeader *multipart.FileHeader) error {
	ext := GetFileExtension(fileHeader.Filename)
	validExtensions := []string{".docx", ".xlsx", ".pdf", ".pptx"}

	isValid := false
	for _, validExt := range validExtensions {
		if ext == validExt {
			isValid = true
			break
		}
	}

	if !isValid {
		return errors.New("file modul harus berupa docx, xlsx, pdf, atau pptx")
	}

	maxSize := int64(50 * 1024 * 1024)
	if fileHeader.Size > maxSize {
		return errors.New("ukuran file modul tidak boleh lebih dari 50 MB")
	}

	return nil
}

func (mh *ModulHelper) ValidateModulFileSize(fileSize int64) error {
	if fileSize <= 0 {
		return errors.New("ukuran file tidak valid")
	}

	maxSize := int64(50 * 1024 * 1024)
	if fileSize > maxSize {
		return errors.New("ukuran file melebihi batas maksimal 50 MB")
	}

	return nil
}

func ValidateModulFileExtension(tipe string) error {
	validTypes := []string{"docx", "xlsx", "pdf", "pptx"}

	for _, validType := range validTypes {
		if tipe == validType {
			return nil
		}
	}

	return errors.New("tipe file harus salah satu dari: docx, xlsx, pdf, pptx")
}

func GetModulFileExtension(tipe string) string {
	switch tipe {
	case "docx":
		return ".docx"
	case "xlsx":
		return ".xlsx"
	case "pdf":
		return ".pdf"
	case "pptx":
		return ".pptx"
	default:
		return ""
	}
}
