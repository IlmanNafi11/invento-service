package usecase

import (
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/helper"
	"invento-service/internal/usecase/repo"
	"mime/multipart"
	"strconv"

	"gorm.io/gorm"
)

type UserUsecase interface {
	GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error)
	UpdateUserRole(userID string, roleName string) error
	DeleteUser(userID string) error
	GetUserFiles(userID string, params domain.UserFilesQueryParams) (*domain.UserFilesData, error)
	GetProfile(userID string) (*domain.ProfileData, error)
	UpdateProfile(userID string, req domain.UpdateProfileRequest, fotoProfil *multipart.FileHeader) (*domain.ProfileData, error)
	GetUserPermissions(userID string) ([]domain.UserPermissionItem, error)
	DownloadUserFiles(ownerUserID string, projectIDs, modulIDs []string) (string, error)
	GetUsersForRole(roleID uint) ([]domain.UserListItem, error)
	BulkAssignRole(userIDs []string, roleID uint) error
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
}

func NewUserUsecase(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	casbinEnforcer *helper.CasbinEnforcer,
	pathResolver *helper.PathResolver,
	cfg *config.Config,
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
	}
}

func (uc *userUsecase) GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error) {
	normalizedParams := helper.NormalizePaginationParams(params.Page, params.Limit)
	params.Page = normalizedParams.Page
	params.Limit = normalizedParams.Limit

	users, total, err := uc.userRepo.GetAll(params.Search, params.FilterRole, params.Page, params.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	pagination := helper.CalculatePagination(params.Page, params.Limit, total)

	return &domain.UserListData{
		Items:      users,
		Pagination: pagination,
	}, nil
}

func (uc *userUsecase) UpdateUserRole(userID string, roleName string) error {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("User")
		}
		return apperrors.NewInternalError(err)
	}

	role, err := uc.roleRepo.GetByName(roleName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Role")
		}
		return apperrors.NewInternalError(err)
	}

	roleID := int(role.ID)
	if user.RoleID != nil && *user.RoleID == roleID {
		return nil
	}

	if uc.casbinEnforcer != nil && user.Role != nil {
		if err := uc.casbinEnforcer.RemoveRoleForUser(userID, user.Role.NamaRole); err != nil {
			return apperrors.NewInternalError(err)
		}
	}

	if err := uc.userRepo.UpdateRole(userID, &roleID); err != nil {
		return apperrors.NewInternalError(err)
	}

	if uc.casbinEnforcer != nil {
		if err := uc.casbinEnforcer.AddRoleForUser(userID, roleName); err != nil {
			return apperrors.NewInternalError(err)
		}

		if err := uc.casbinEnforcer.SavePolicy(); err != nil {
			return apperrors.NewInternalError(err)
		}
	}

	return nil
}

func (uc *userUsecase) DeleteUser(userID string) error {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("User")
		}
		return apperrors.NewInternalError(err)
	}

	if err := uc.userRepo.Delete(userID); err != nil {
		return apperrors.NewInternalError(err)
	}

	return nil
}

func (uc *userUsecase) GetUserFiles(userID string, params domain.UserFilesQueryParams) (*domain.UserFilesData, error) {
	_, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
	}

	normalizedParams := helper.NormalizePaginationParams(params.Page, params.Limit)
	params.Page = normalizedParams.Page
	params.Limit = normalizedParams.Limit

	items, total, err := uc.userRepo.GetUserFiles(userID, params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	for i := range items {
		if normalizedPath := uc.pathResolver.ConvertToAPIPath(&items[i].DownloadURL); normalizedPath != nil {
			items[i].DownloadURL = *normalizedPath
		}
	}

	pagination := helper.CalculatePagination(params.Page, params.Limit, total)

	return &domain.UserFilesData{
		Items:      items,
		Pagination: pagination,
	}, nil
}

func (uc *userUsecase) GetProfile(userID string) (*domain.ProfileData, error) {
	user, jumlahProject, jumlahModul, err := uc.userRepo.GetProfileWithCounts(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
	}

	return uc.userHelper.BuildProfileData(user, jumlahProject, jumlahModul), nil
}

func (uc *userUsecase) UpdateProfile(userID string, req domain.UpdateProfileRequest, fotoProfil *multipart.FileHeader) (*domain.ProfileData, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
	}

	var jenisKelaminPtr *string
	if req.JenisKelamin != "" {
		jenisKelaminPtr = &req.JenisKelamin
	}

	fotoProfilPath, err := uc.userHelper.SaveProfilePhoto(fotoProfil, userID, user.FotoProfil)
	if err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	if err := uc.userRepo.UpdateProfile(userID, req.Name, jenisKelaminPtr, fotoProfilPath); err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	user.Name = req.Name
	if jenisKelaminPtr != nil {
		user.JenisKelamin = jenisKelaminPtr
	}
	if fotoProfilPath != nil {
		user.FotoProfil = fotoProfilPath
	}

	jumlahProject, _ := uc.projectRepo.CountByUserID(userID)
	jumlahModul, _ := uc.modulRepo.CountByUserID(userID)

	return uc.userHelper.BuildProfileData(user, jumlahProject, jumlahModul), nil
}

func (uc *userUsecase) GetUserPermissions(userID string) ([]domain.UserPermissionItem, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
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
		return nil, apperrors.NewInternalError(err)
	}

	return uc.userHelper.AggregateUserPermissions(permissions), nil
}

func (uc *userUsecase) DownloadUserFiles(ownerUserID string, projectIDs, modulIDs []string) (string, error) {
	if err := uc.downloadHelper.ValidateDownloadRequest(projectIDs, modulIDs); err != nil {
		return "", err
	}

	_, err := uc.userRepo.GetByID(ownerUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", apperrors.NewNotFoundError("User")
		}
		return "", apperrors.NewInternalError(err)
	}

	// Convert string IDs to uint for repository calls
	projectIDsUint := make([]uint, 0, len(projectIDs))
	for _, idStr := range projectIDs {
		id, err := strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return "", apperrors.NewValidationError("Format project ID tidak valid", err)
		}
		projectIDsUint = append(projectIDsUint, uint(id))
	}

	projects, err := uc.projectRepo.GetByIDs(projectIDsUint, ownerUserID)
	if err != nil {
		return "", apperrors.NewInternalError(err)
	}

	moduls, err := uc.modulRepo.GetByIDs(modulIDs, ownerUserID)
	if err != nil {
		return "", apperrors.NewInternalError(err)
	}

	if len(projects)+len(moduls) == 0 {
		return "", apperrors.NewNotFoundError("File")
	}

	filePaths, _, err := uc.downloadHelper.PrepareFilesForDownload(projects, moduls)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return "", appErr
		}
		return "", apperrors.NewInternalError(err)
	}

	zipPath, err := uc.downloadHelper.CreateDownloadZip(filePaths, ownerUserID)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return "", appErr
		}
		return "", apperrors.NewInternalError(err)
	}

	return zipPath, nil
}

func (uc *userUsecase) GetUsersForRole(roleID uint) ([]domain.UserListItem, error) {
	_, err := uc.roleRepo.GetByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	users, err := uc.userRepo.GetByRoleID(roleID)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return users, nil
}

func (uc *userUsecase) BulkAssignRole(userIDs []string, roleID uint) error {
	role, err := uc.roleRepo.GetByID(roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Role")
		}
		return apperrors.NewInternalError(err)
	}

	users, err := uc.userRepo.GetByIDs(userIDs)
	if err != nil {
		return apperrors.NewInternalError(err)
	}

	for _, user := range users {
		if uc.casbinEnforcer != nil && user.Role != nil {
			if err := uc.casbinEnforcer.RemoveRoleForUser(user.ID, user.Role.NamaRole); err != nil {
				return apperrors.NewInternalError(err)
			}
		}

		if uc.casbinEnforcer != nil {
			if err := uc.casbinEnforcer.AddRoleForUser(user.ID, role.NamaRole); err != nil {
				return apperrors.NewInternalError(err)
			}
		}
	}

	if err := uc.userRepo.BulkUpdateRole(userIDs, roleID); err != nil {
		return apperrors.NewInternalError(err)
	}

	if uc.casbinEnforcer != nil {
		if err := uc.casbinEnforcer.SavePolicy(); err != nil {
			return apperrors.NewInternalError(err)
		}
	}

	return nil
}
