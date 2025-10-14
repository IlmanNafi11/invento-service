package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type ModulUsecase interface {
	GetList(userID uint, search string, filterType string, filterSemester int, page, limit int) (*domain.ModulListData, error)
	GetByID(modulID, userID uint) (*domain.ModulResponse, error)
	UpdateMetadata(modulID, userID uint, req domain.ModulUpdateRequest) error
	Delete(modulID, userID uint) error
	Download(userID uint, modulIDs []uint) (string, error)
}

type modulUsecase struct {
	modulRepo repo.ModulRepository
}

func NewModulUsecase(modulRepo repo.ModulRepository) ModulUsecase {
	return &modulUsecase{
		modulRepo: modulRepo,
	}
}



func (uc *modulUsecase) GetList(userID uint, search string, filterType string, filterSemester int, page, limit int) (*domain.ModulListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	moduls, total, err := uc.modulRepo.GetByUserID(userID, search, filterType, filterSemester, page, limit)
	if err != nil {
		return nil, errors.New("gagal mengambil data modul")
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

func (uc *modulUsecase) GetByID(modulID, userID uint) (*domain.ModulResponse, error) {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("modul tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data modul")
	}

	if modul.UserID != userID {
		return nil, errors.New("tidak memiliki akses ke modul ini")
	}

	return &domain.ModulResponse{
		ID:        modul.ID,
		NamaFile:  modul.NamaFile,
		Tipe:      modul.Tipe,
		Ukuran:    modul.Ukuran,
		Semester:  modul.Semester,
		PathFile:  modul.PathFile,
		CreatedAt: modul.CreatedAt,
		UpdatedAt: modul.UpdatedAt,
	}, nil
}



func (uc *modulUsecase) Delete(modulID, userID uint) error {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("modul tidak ditemukan")
		}
		return errors.New("gagal mengambil data modul")
	}

	if modul.UserID != userID {
		return errors.New("tidak memiliki akses ke modul ini")
	}

	helper.DeleteFile(modul.PathFile)

	if err := uc.modulRepo.Delete(modulID); err != nil {
		return errors.New("gagal menghapus modul")
	}

	return nil
}

func (uc *modulUsecase) Download(userID uint, modulIDs []uint) (string, error) {
	if len(modulIDs) == 0 {
		return "", errors.New("id modul tidak boleh kosong")
	}

	moduls, err := uc.modulRepo.GetByIDs(modulIDs, userID)
	if err != nil {
		return "", errors.New("gagal mengambil data modul")
	}

	if len(moduls) == 0 {
		return "", errors.New("modul tidak ditemukan")
	}

	if len(moduls) == 1 {
		return moduls[0].PathFile, nil
	}

	var filePaths []string
	for _, modul := range moduls {
		filePaths = append(filePaths, modul.PathFile)
	}

	tempDir := "./uploads/temp"
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", errors.New("gagal membuat direktori temp")
	}

	identifier, err := helper.GenerateUniqueIdentifier(8)
	if err != nil {
		return "", errors.New("gagal generate identifier")
	}

	zipFileName := fmt.Sprintf("moduls_%s.zip", identifier)
	zipFilePath := filepath.Join(tempDir, zipFileName)

	if err := helper.CreateZipArchive(filePaths, zipFilePath); err != nil {
		return "", errors.New("gagal membuat file zip")
	}

	return zipFilePath, nil
}

func (uc *modulUsecase) UpdateMetadata(modulID, userID uint, req domain.ModulUpdateRequest) error {
	modul, err := uc.modulRepo.GetByID(modulID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("modul tidak ditemukan")
		}
		return errors.New("gagal mengambil data modul")
	}

	if modul.UserID != userID {
		return errors.New("tidak memiliki akses ke modul ini")
	}

	if req.NamaFile != "" {
		modul.NamaFile = req.NamaFile
	}
	if req.Semester > 0 {
		modul.Semester = req.Semester
	}

	if err := uc.modulRepo.Update(modul); err != nil {
		return errors.New("gagal update metadata modul")
	}

	return nil
}
