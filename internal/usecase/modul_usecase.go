package usecase

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/helper"
	"invento-service/internal/usecase/repo"
)

type ModulUsecase interface {
	GetList(userID string, search string, filterType string, filterStatus string, page, limit int) (*domain.ModulListData, error)
	GetByID(modulID string, userID string) (*domain.ModulResponse, error)
	UpdateMetadata(modulID string, userID string, req domain.ModulUpdateRequest) error
	Delete(modulID string, userID string) error
	Download(userID string, modulIDs []string) (string, error)
}

type modulUsecase struct {
	modulRepo repo.ModulRepository
}

func NewModulUsecase(modulRepo repo.ModulRepository) ModulUsecase {
	return &modulUsecase{
		modulRepo: modulRepo,
	}
}

func (uc *modulUsecase) GetList(userID string, search string, filterType string, filterStatus string, page, limit int) (*domain.ModulListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	moduls, total, err := uc.modulRepo.GetByUserID(userID, search, filterType, filterStatus, page, limit)
	if err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("ModulUsecase.GetList: %w", err))
	}

	totalPages := (total + limit - 1) / limit

	return &domain.ModulListData{
		Items: moduls,
		Pagination: domain.PaginationData{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (uc *modulUsecase) GetByID(modulID string, userID string) (*domain.ModulResponse, error) {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Modul")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("ModulUsecase.GetByID: %w", err))
	}

	if modul.UserID != userID {
		return nil, apperrors.NewForbiddenError("Tidak memiliki akses ke modul ini")
	}

	return &domain.ModulResponse{
		ID:        modul.ID,
		Judul:     modul.Judul,
		Deskripsi: modul.Deskripsi,
		FileName:  modul.FileName,
		MimeType:  modul.MimeType,
		FileSize:  modul.FileSize,
		Status:    modul.Status,
		CreatedAt: modul.CreatedAt,
		UpdatedAt: modul.UpdatedAt,
	}, nil
}

func (uc *modulUsecase) Delete(modulID string, userID string) error {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Modul")
		}
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Delete: %w", err))
	}

	if modul.UserID != userID {
		return apperrors.NewForbiddenError("Tidak memiliki akses ke modul ini")
	}

	if err := uc.modulRepo.Delete(modulID); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Delete: %w", err))
	}

	if modul.FilePath != "" {
		if err := helper.DeleteFile(modul.FilePath); err != nil {
			// File deletion after DB delete is critical but non-blocking;
			// log the error so it can be investigated.
			log.Printf("WARNING: ModulUsecase.Delete: gagal menghapus file modul %s: %v", modul.FilePath, err)
		}
	}

	return nil
}

func (uc *modulUsecase) Download(userID string, modulIDs []string) (string, error) {
	if len(modulIDs) == 0 {
		return "", apperrors.NewValidationError("ID modul tidak boleh kosong", nil)
	}

	moduls, err := uc.modulRepo.GetByIDs(modulIDs, userID)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	if len(moduls) == 0 {
		return "", apperrors.NewNotFoundError("Modul")
	}

	if len(moduls) == 1 {
		if _, err := os.Stat(moduls[0].FilePath); os.IsNotExist(err) {
			return "", apperrors.NewNotFoundError("File modul")
		}
		return moduls[0].FilePath, nil
	}

	var filePaths []string
	for _, modul := range moduls {
		filePaths = append(filePaths, modul.FilePath)
	}

	tempDir := "./uploads/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	identifier, err := helper.GenerateUniqueIdentifier(8)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	zipFileName := fmt.Sprintf("moduls_%s.zip", identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := helper.CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	return zipFilePath, nil
}

func (uc *modulUsecase) UpdateMetadata(modulID string, userID string, req domain.ModulUpdateRequest) error {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Modul")
		}
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.UpdateMetadata: %w", err))
	}

	if modul.UserID != userID {
		return apperrors.NewForbiddenError("Tidak memiliki akses ke modul ini")
	}

	if req.Judul != "" {
		modul.Judul = req.Judul
	}
	if req.Deskripsi != "" {
		modul.Deskripsi = req.Deskripsi
	}

	if err := uc.modulRepo.UpdateMetadata(modul); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.UpdateMetadata: %w", err))
	}

	return nil
}
