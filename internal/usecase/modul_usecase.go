package usecase

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"invento-service/internal/dto"
	"invento-service/internal/storage"
	"invento-service/internal/usecase/repo"

	apperrors "invento-service/internal/errors"

	zlog "github.com/rs/zerolog/log"
)

type ModulUsecase interface {
	GetList(ctx context.Context, userID, search, filterType, filterStatus string, page, limit int) (*dto.ModulListData, error)
	GetByID(ctx context.Context, modulID, userID string) (*dto.ModulResponse, error)
	UpdateMetadata(ctx context.Context, modulID, userID string, req dto.UpdateModulRequest) error
	Delete(ctx context.Context, modulID, userID string) error
	Download(ctx context.Context, userID string, modulIDs []string) (string, error)
}

type modulUsecase struct {
	modulRepo repo.ModulRepository
}

func NewModulUsecase(modulRepo repo.ModulRepository) ModulUsecase {
	return &modulUsecase{
		modulRepo: modulRepo,
	}
}

func (uc *modulUsecase) GetList(ctx context.Context, userID, search, filterType, filterStatus string, page, limit int) (*dto.ModulListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	moduls, total, err := uc.modulRepo.GetByUserID(ctx, userID, search, filterType, filterStatus, page, limit)
	if err != nil {
		return nil, apperrors.NewInternalError(fmt.Errorf("ModulUsecase.GetList: %w", err))
	}

	totalPages := (total + limit - 1) / limit

	return &dto.ModulListData{
		Items: moduls,
		Pagination: dto.PaginationData{
			Page:       page,
			Limit:      limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (uc *modulUsecase) GetByID(ctx context.Context, modulID, userID string) (*dto.ModulResponse, error) {
	modul, err := uc.modulRepo.GetByID(ctx, modulID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Modul")
		}
		return nil, apperrors.NewInternalError(fmt.Errorf("ModulUsecase.GetByID: %w", err))
	}

	if modul.UserID != userID {
		return nil, apperrors.NewForbiddenError("Tidak memiliki akses ke modul ini")
	}

	return &dto.ModulResponse{
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

func (uc *modulUsecase) Delete(ctx context.Context, modulID, userID string) error {
	modul, err := uc.modulRepo.GetByID(ctx, modulID)
	if err != nil {
		if errors.Is(err, apperrors.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Modul")
		}
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Delete: %w", err))
	}

	if modul.UserID != userID {
		return apperrors.NewForbiddenError("Tidak memiliki akses ke modul ini")
	}

	if err := uc.modulRepo.Delete(ctx, modulID); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Delete: %w", err))
	}

	if modul.FilePath != "" {
		if err := storage.DeleteFile(modul.FilePath); err != nil {
			// File deletion after DB delete is critical but non-blocking;
			// log the error so it can be investigated.
			zlog.Warn().Err(err).Str("file", modul.FilePath).Msg("ModulUsecase.Delete: failed to delete modul file")
		}
	}

	return nil
}

func (uc *modulUsecase) Download(ctx context.Context, userID string, modulIDs []string) (string, error) {
	if len(modulIDs) == 0 {
		return "", apperrors.NewValidationError("ID modul tidak boleh kosong", nil)
	}

	moduls, err := uc.modulRepo.GetByIDs(ctx, modulIDs, userID)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	if len(moduls) == 0 {
		return "", apperrors.NewNotFoundError("Modul")
	}

	if len(moduls) == 1 {
		if _, statErr := os.Stat(moduls[0].FilePath); os.IsNotExist(statErr) {
			return "", apperrors.NewNotFoundError("File modul")
		}
		return moduls[0].FilePath, nil
	}

	var filePaths []string
	for _, modul := range moduls {
		filePaths = append(filePaths, modul.FilePath)
	}

	tempDir := "./uploads/temp"
	if err = os.MkdirAll(tempDir, 0o755); err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	identifier, err := storage.GenerateUniqueIdentifier(8)
	if err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	zipFileName := fmt.Sprintf("moduls_%s.zip", identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := storage.CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", apperrors.NewInternalError(fmt.Errorf("ModulUsecase.Download: %w", err))
	}

	return zipFilePath, nil
}

func (uc *modulUsecase) UpdateMetadata(ctx context.Context, modulID, userID string, req dto.UpdateModulRequest) error {
	modul, err := uc.modulRepo.GetByID(ctx, modulID)
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

	if err := uc.modulRepo.UpdateMetadata(ctx, modul); err != nil {
		return apperrors.NewInternalError(fmt.Errorf("ModulUsecase.UpdateMetadata: %w", err))
	}

	return nil
}
