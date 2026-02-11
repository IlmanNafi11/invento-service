package usecase

import (
	"errors"
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

	"gorm.io/gorm"
)

type UserUsecase interface {
	GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error)
	UpdateUserRole(userID uint, roleName string) error
	DeleteUser(userID uint) error
	GetUserFiles(userID uint, params domain.UserFilesQueryParams) (*domain.UserFilesData, error)
	GetProfile(userID uint) (*domain.ProfileData, error)
	UpdateProfile(userID uint, req domain.UpdateProfileRequest, fotoProfil interface{}) (*domain.ProfileData, error)
	GetUserPermissions(userID uint) ([]domain.UserPermissionItem, error)
	DownloadUserFiles(ownerUserID uint, projectIDs, modulIDs []uint) (string, error)
}

type userUsecase struct {
	userRepo       repo.UserRepository
	roleRepo       repo.RoleRepository
	projectRepo    repo.ProjectRepository
	modulRepo      repo.ModulRepository
	casbinEnforcer *helper.CasbinEnforcer
	userHelper     *helper.UserHelper
	downloadHelper *helper.DownloadHelper
	pathResolver   *helper.PathResolver
	config         *config.Config
	db             *gorm.DB
}

func NewUserUsecase(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	casbinEnforcer *helper.CasbinEnforcer,
	pathResolver *helper.PathResolver,
	cfg *config.Config,
	db *gorm.DB,
) UserUsecase {
	return &userUsecase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		projectRepo:    projectRepo,
		modulRepo:      modulRepo,
		casbinEnforcer: casbinEnforcer,
		userHelper:     helper.NewUserHelper(pathResolver, cfg),
		downloadHelper: helper.NewDownloadHelper(pathResolver),
		pathResolver:   pathResolver,
		config:         cfg,
		db:             db,
	}
}

func (uc *userUsecase) GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error) {
	normalizedParams := helper.NormalizePaginationParams(params.Page, params.Limit)
	params.Page = normalizedParams.Page
	params.Limit = normalizedParams.Limit

	users, total, err := uc.userRepo.GetAll(params.Search, params.FilterRole, params.Page, params.Limit)
	if err != nil {
		return nil, errors.New("gagal mengambil daftar user")
	}

	pagination := helper.CalculatePagination(params.Page, params.Limit, total)

	return &domain.UserListData{
		Items:      users,
		Pagination: pagination,
	}, nil
}

func (uc *userUsecase) UpdateUserRole(userID uint, roleName string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user tidak ditemukan")
		}
		return errors.New("gagal mengambil data user")
	}

	role, err := uc.roleRepo.GetByName(roleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("role tidak ditemukan")
		}
		return errors.New("gagal mengambil data role")
	}

	if user.RoleID != nil && *user.RoleID == role.ID {
		return nil
	}

	if err := uc.userRepo.UpdateRole(userID, &role.ID); err != nil {
		return errors.New("gagal memperbarui role user")
	}

	return nil
}

func (uc *userUsecase) DeleteUser(userID uint) error {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user tidak ditemukan")
		}
		return errors.New("gagal mengambil data user")
	}

	if err := uc.userRepo.Delete(userID); err != nil {
		return errors.New("gagal menghapus user")
	}

	return nil
}

func (uc *userUsecase) GetUserFiles(userID uint, params domain.UserFilesQueryParams) (*domain.UserFilesData, error) {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	normalizedParams := helper.NormalizePaginationParams(params.Page, params.Limit)
	params.Page = normalizedParams.Page
	params.Limit = normalizedParams.Limit

	var allFiles []domain.UserFileItem
	var projectFiles []struct {
		ID          uint
		NamaFile    string
		Kategori    string
		DownloadURL string
	}

	search := params.Search
	if err := uc.db.Raw(`
		SELECT 
			p.id,
			p.nama_project as nama_file,
			'Project' as kategori,
			p.path_file as download_url
		FROM projects p
		WHERE p.user_id = ?
			AND (? = '' OR p.nama_project LIKE CONCAT('%', ?, '%'))
		ORDER BY p.updated_at DESC
	`, userID, search, search).Scan(&projectFiles).Error; err != nil {
		return nil, errors.New("gagal mengambil data project")
	}

	for _, pf := range projectFiles {
		normalizedURL := pf.DownloadURL
		if normalizedPath := uc.pathResolver.ConvertToAPIPath(&pf.DownloadURL); normalizedPath != nil {
			normalizedURL = *normalizedPath
		}
		allFiles = append(allFiles, domain.UserFileItem{
			ID:          pf.ID,
			NamaFile:    pf.NamaFile,
			Kategori:    pf.Kategori,
			DownloadURL: normalizedURL,
		})
	}

	var modulFiles []struct {
		ID          uint
		NamaFile    string
		Kategori    string
		DownloadURL string
	}

	if err := uc.db.Raw(`
		SELECT 
			m.id,
			m.nama_file,
			'Modul' as kategori,
			m.path_file as download_url
		FROM moduls m
		WHERE m.user_id = ?
			AND (? = '' OR m.nama_file LIKE CONCAT('%', ?, '%'))
		ORDER BY m.updated_at DESC
	`, userID, search, search).Scan(&modulFiles).Error; err != nil {
		return nil, errors.New("gagal mengambil data modul")
	}

	for _, mf := range modulFiles {
		normalizedURL := mf.DownloadURL
		if normalizedPath := uc.pathResolver.ConvertToAPIPath(&mf.DownloadURL); normalizedPath != nil {
			normalizedURL = *normalizedPath
		}
		allFiles = append(allFiles, domain.UserFileItem{
			ID:          mf.ID,
			NamaFile:    mf.NamaFile,
			Kategori:    mf.Kategori,
			DownloadURL: normalizedURL,
		})
	}

	total := len(allFiles)
	offset := (params.Page - 1) * params.Limit
	end := offset + params.Limit

	if offset > total {
		offset = total
	}
	if end > total {
		end = total
	}

	paginatedFiles := allFiles[offset:end]
	pagination := helper.CalculatePagination(params.Page, params.Limit, total)

	return &domain.UserFilesData{
		Items:      paginatedFiles,
		Pagination: pagination,
	}, nil
}

func (uc *userUsecase) GetProfile(userID uint) (*domain.ProfileData, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	jumlahProject, err := uc.projectRepo.CountByUserID(userID)
	if err != nil {
		jumlahProject = 0
	}

	jumlahModul, err := uc.modulRepo.CountByUserID(userID)
	if err != nil {
		jumlahModul = 0
	}

	return uc.userHelper.BuildProfileData(user, jumlahProject, jumlahModul), nil
}

func (uc *userUsecase) UpdateProfile(userID uint, req domain.UpdateProfileRequest, fotoProfil interface{}) (*domain.ProfileData, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	var jenisKelaminPtr *string
	if req.JenisKelamin != "" {
		jenisKelaminPtr = &req.JenisKelamin
	}

	fotoProfilPath, err := uc.userHelper.SaveProfilePhoto(fotoProfil, userID, user.FotoProfil)
	if err != nil {
		return nil, err
	}

	if err := uc.userRepo.UpdateProfile(userID, req.Name, jenisKelaminPtr, fotoProfilPath); err != nil {
		return nil, errors.New("gagal memperbarui profil")
	}

	return uc.GetProfile(userID)
}

func (uc *userUsecase) GetUserPermissions(userID uint) ([]domain.UserPermissionItem, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	if user.Role == nil {
		return []domain.UserPermissionItem{}, nil
	}

	// Handle nil casbinEnforcer gracefully
	if uc.casbinEnforcer == nil {
		return []domain.UserPermissionItem{}, nil
	}

	permissions, err := uc.casbinEnforcer.GetPermissionsForRole(user.Role.NamaRole)
	if err != nil {
		return nil, errors.New("gagal mengambil permissions user")
	}

	return uc.userHelper.AggregateUserPermissions(permissions), nil
}

func (uc *userUsecase) DownloadUserFiles(ownerUserID uint, projectIDs, modulIDs []uint) (string, error) {
	if err := uc.downloadHelper.ValidateDownloadRequest(projectIDs, modulIDs); err != nil {
		return "", err
	}

	_, err := uc.userRepo.GetByID(ownerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", errors.New("user tidak ditemukan")
		}
		return "", errors.New("gagal mengambil data user")
	}

	projects, err := uc.projectRepo.GetByIDsForUser(projectIDs, ownerUserID)
	if err != nil {
		return "", errors.New("gagal mengambil data project")
	}

	moduls, err := uc.modulRepo.GetByIDsForUser(modulIDs, ownerUserID)
	if err != nil {
		return "", errors.New("gagal mengambil data modul")
	}

	if len(projects)+len(moduls) == 0 {
		return "", errors.New("file tidak ditemukan")
	}

	filePaths, _, err := uc.downloadHelper.PrepareFilesForDownload(projects, moduls)
	if err != nil {
		return "", err
	}

	zipPath, err := uc.downloadHelper.CreateDownloadZip(filePaths, ownerUserID)
	if err != nil {
		return "", err
	}

	return zipPath, nil
}
