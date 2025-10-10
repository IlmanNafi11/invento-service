package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

type ModulUsecase interface {
	Create(userID uint, userEmail string, userRole string, files []*multipart.FileHeader, namaFiles []string) (*domain.ModulCreateResponse, error)
	GetList(userID uint, search string, filterType string, page, limit int) (*domain.ModulListData, error)
	GetByID(modulID, userID uint) (*domain.ModulResponse, error)
	Update(modulID, userID uint, userEmail string, userRole string, namaFile string, file *multipart.FileHeader) (*domain.ModulResponse, error)
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

func (uc *modulUsecase) Create(userID uint, userEmail string, userRole string, files []*multipart.FileHeader, namaFiles []string) (*domain.ModulCreateResponse, error) {
	if len(files) == 0 {
		return nil, errors.New("file wajib diupload")
	}

	if len(files) != len(namaFiles) {
		return nil, errors.New("jumlah file dan nama file harus sama")
	}

	var modulResponses []domain.ModulResponse

	for i, fileHeader := range files {
		if err := helper.ValidateModulFile(fileHeader); err != nil {
			return nil, err
		}

		tipe := helper.GetFileType(fileHeader.Filename)
		ukuran := helper.GetFileSize(fileHeader)

		modulDir, err := helper.CreateModulDirectory(userEmail, userRole, tipe)
		if err != nil {
			return nil, errors.New("gagal membuat direktori modul")
		}

		filename := fileHeader.Filename
		destPath := filepath.Join(modulDir, filename)

		if err := helper.SaveUploadedFile(fileHeader, destPath); err != nil {
			return nil, errors.New("gagal menyimpan file")
		}

		modul := &domain.Modul{
			UserID:   userID,
			NamaFile: namaFiles[i],
			Tipe:     tipe,
			Ukuran:   ukuran,
			PathFile: destPath,
		}

		if err := uc.modulRepo.Create(modul); err != nil {
			helper.DeleteFile(destPath)
			return nil, errors.New("gagal menyimpan data modul")
		}

		modulResponses = append(modulResponses, domain.ModulResponse{
			ID:        modul.ID,
			NamaFile:  modul.NamaFile,
			Tipe:      modul.Tipe,
			Ukuran:    modul.Ukuran,
			PathFile:  modul.PathFile,
			CreatedAt: modul.CreatedAt,
			UpdatedAt: modul.UpdatedAt,
		})
	}

	return &domain.ModulCreateResponse{
		Items: modulResponses,
	}, nil
}

func (uc *modulUsecase) GetList(userID uint, search string, filterType string, page, limit int) (*domain.ModulListData, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	moduls, total, err := uc.modulRepo.GetByUserID(userID, search, filterType, page, limit)
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
		PathFile:  modul.PathFile,
		CreatedAt: modul.CreatedAt,
		UpdatedAt: modul.UpdatedAt,
	}, nil
}

func (uc *modulUsecase) Update(modulID, userID uint, userEmail string, userRole string, namaFile string, file *multipart.FileHeader) (*domain.ModulResponse, error) {
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

	if namaFile != "" {
		modul.NamaFile = namaFile
	}

	if file != nil {
		if err := helper.ValidateModulFile(file); err != nil {
			return nil, err
		}

		oldPath := modul.PathFile

		tipe := helper.GetFileType(file.Filename)
		ukuran := helper.GetFileSize(file)

		modulDir, err := helper.CreateModulDirectory(userEmail, userRole, tipe)
		if err != nil {
			return nil, errors.New("gagal membuat direktori modul")
		}

		filename := file.Filename
		destPath := filepath.Join(modulDir, filename)

		if err := helper.SaveUploadedFile(file, destPath); err != nil {
			return nil, errors.New("gagal menyimpan file")
		}

		helper.DeleteFile(oldPath)

		modul.Tipe = tipe
		modul.Ukuran = ukuran
		modul.PathFile = destPath
	}

	if err := uc.modulRepo.Update(modul); err != nil {
		return nil, errors.New("gagal memperbarui data modul")
	}

	return &domain.ModulResponse{
		ID:        modul.ID,
		NamaFile:  modul.NamaFile,
		Tipe:      modul.Tipe,
		Ukuran:    modul.Ukuran,
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
