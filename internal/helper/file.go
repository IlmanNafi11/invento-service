package helper

import (
	"archive/zip"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
)

func GenerateUniqueIdentifier(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func GetFileExtension(filename string) string {
	return strings.ToLower(filepath.Ext(filename))
}

func ValidateZipFile(fileHeader *multipart.FileHeader) error {
	ext := GetFileExtension(fileHeader.Filename)
	if ext != ".zip" {
		return errors.New("file harus berupa zip")
	}
	return nil
}

func GetFileSize(fileHeader *multipart.FileHeader) string {
	size := fileHeader.Size
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)

	switch {
	case size >= GB:
		return fmt.Sprintf("%.2fGB", float64(size)/float64(GB))
	case size >= MB:
		return fmt.Sprintf("%.2fMB", float64(size)/float64(MB))
	case size >= KB:
		return fmt.Sprintf("%.2fKB", float64(size)/float64(KB))
	default:
		return fmt.Sprintf("%dB", size)
	}
}

func DetectProjectCategory(filename string) string {
	lowerFilename := strings.ToLower(filename)

	if strings.Contains(lowerFilename, "website") || strings.Contains(lowerFilename, "web") {
		return "website"
	} else if strings.Contains(lowerFilename, "mobile") || strings.Contains(lowerFilename, "android") || strings.Contains(lowerFilename, "ios") {
		return "mobile"
	} else if strings.Contains(lowerFilename, "iot") || strings.Contains(lowerFilename, "arduino") || strings.Contains(lowerFilename, "sensor") {
		return "iot"
	} else if strings.Contains(lowerFilename, "machine") || strings.Contains(lowerFilename, "ml") {
		return "machine_learning"
	} else if strings.Contains(lowerFilename, "deep") || strings.Contains(lowerFilename, "dl") || strings.Contains(lowerFilename, "neural") {
		return "deep_learning"
	}

	return "website"
}

func CreateUserDirectory(email string, role string) (string, error) {
	emailParts := strings.Split(email, "@")
	if len(emailParts) != 2 {
		return "", errors.New("format email tidak valid")
	}

	username := emailParts[0]

	identifier, err := GenerateUniqueIdentifier(1)
	if err != nil {
		return "", err
	}

	dirName := fmt.Sprintf("%s-%s-%s", username, role, identifier)

	baseDir := "./uploads/project"
	userDir := filepath.Join(baseDir, dirName)

	if err := os.MkdirAll(userDir, 0755); err != nil {
		return "", err
	}

	return userDir, nil
}

func SaveUploadedFile(fileHeader *multipart.FileHeader, destPath string) error {
	file, err := fileHeader.Open()
	if err != nil {
		return err
	}
	defer file.Close()

	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return err
	}

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.ReadFrom(file)
	return err
}

func DeleteFile(path string) error {
	if _, err := os.Stat(path); err == nil {
		return os.Remove(path)
	}
	return nil
}

func CreateZipArchive(filePaths []string, outputPath string) error {
	zipFile, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	for _, filePath := range filePaths {
		if err := addFileToZip(zipWriter, filePath); err != nil {
			return err
		}
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
