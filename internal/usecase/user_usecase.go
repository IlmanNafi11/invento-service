package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/dto"
	apperrors "invento-service/internal/errors"
	"invento-service/internal/httputil"
	"invento-service/internal/rbac"
	"invento-service/internal/storage"
	"invento-service/internal/usecase/repo"
	"mime/multipart"
	"strconv"

	"gorm.io/gorm"
)

type UserUsecase interface {
	GetUserList(ctx context.Context, params dto.UserListQueryParams) (*dto.UserListData, error)
	UpdateUserRole(ctx context.Context, userID string, roleName string) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserFiles(ctx context.Context, userID string, params dto.UserFilesQueryParams) (*dto.UserFilesData, error)
	GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest, fotoProfil *multipart.FileHeader) (*dto.ProfileData, error)
	GetUserPermissions(ctx context.Context, userID string) ([]dto.UserPermissionItem, error)
	DownloadUserFiles(ctx context.Context, ownerUserID string, projectIDs, modulIDs []string) (string, error)
	GetUsersForRole(ctx context.Context, roleID uint) ([]dto.UserListItem, error)
	BulkAssignRole(ctx context.Context, userIDs []string, roleID uint) error
}

type userUsecase struct {
	userRepo       repo.UserRepository
	roleRepo       repo.RoleRepository
	projectRepo    repo.ProjectRepository
	modulRepo      repo.ModulRepository
	casbinEnforcer *rbac.CasbinEnforcer
	userHelper     *storage.UserHelper
	downloadHelper *storage.DownloadHelper
	pathResolver   *storage.PathResolver
	config         *config.Config
}

func NewUserUsecase(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	casbinEnforcer *rbac.CasbinEnforcer,
	pathResolver *storage.PathResolver,
	cfg *config.Config,
) UserUsecase {
	return &userUsecase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		projectRepo:    projectRepo,
		modulRepo:      modulRepo,
		casbinEnforcer: casbinEnforcer,
		userHelper:     storage.NewUserHelper(pathResolver, cfg),
		downloadHelper: storage.NewDownloadHelper(pathResolver),
		pathResolver:   pathResolver,
		config:         cfg,
	}
}

func (uc *userUsecase) GetUserList(ctx context.Context, params dto.UserListQueryParams) (*dto.UserListData, error) {
	normalizedParams := httputil.NormalizePaginationParams(params.Page, params.Limit)
	params.Page = normalizedParams.Page
	params.Limit = normalizedParams.Limit

	users, total, err := uc.userRepo.GetAll(ctx, params.Search, params.FilterRole, params.Page, params.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	pagination := httputil.CalculatePagination(params.Page, params.Limit, total)

	return &dto.UserListData{
		Items:      users,
		Pagination: pagination,
	}, nil
}

func (uc *userUsecase) UpdateUserRole(ctx context.Context, userID string, roleName string) error {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("User")
		}
		return apperrors.NewInternalError(err)
	}

	role, err := uc.roleRepo.GetByName(ctx, roleName)
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

	if err := uc.userRepo.UpdateRole(ctx, userID, &roleID); err != nil {
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

func (uc *userUsecase) DeleteUser(ctx context.Context, userID string) error {
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("User")
		}
		return apperrors.NewInternalError(err)
	}

	if err := uc.userRepo.Delete(ctx, userID); err != nil {
		return apperrors.NewInternalError(err)
	}

	return nil
}

func (uc *userUsecase) GetUserFiles(ctx context.Context, userID string, params dto.UserFilesQueryParams) (*dto.UserFilesData, error) {
	_, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
	}

	normalizedParams := httputil.NormalizePaginationParams(params.Page, params.Limit)
	params.Page = normalizedParams.Page
	params.Limit = normalizedParams.Limit

	items, total, err := uc.userRepo.GetUserFiles(ctx, userID, params.Search, params.Page, params.Limit)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	for i := range items {
		if normalizedPath := uc.pathResolver.ConvertToAPIPath(&items[i].DownloadURL); normalizedPath != nil {
			items[i].DownloadURL = *normalizedPath
		}
	}

	pagination := httputil.CalculatePagination(params.Page, params.Limit, total)

	return &dto.UserFilesData{
		Items:      items,
		Pagination: pagination,
	}, nil
}

func (uc *userUsecase) GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error) {
	user, jumlahProject, jumlahModul, err := uc.userRepo.GetProfileWithCounts(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
	}

	return uc.userHelper.BuildProfileData(user, jumlahProject, jumlahModul), nil
}

func (uc *userUsecase) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest, fotoProfil *multipart.FileHeader) (*dto.ProfileData, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
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

	if err := uc.userRepo.UpdateProfile(ctx, userID, req.Name, jenisKelaminPtr, fotoProfilPath); err != nil {
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

func (uc *userUsecase) GetUserPermissions(ctx context.Context, userID string) ([]dto.UserPermissionItem, error) {
	user, err := uc.userRepo.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("User")
		}
		return nil, apperrors.NewInternalError(err)
	}

	if user.Role == nil {
		return []dto.UserPermissionItem{}, nil
	}

	// Handle nil casbinEnforcer gracefully
	if uc.casbinEnforcer == nil {
		return []dto.UserPermissionItem{}, nil
	}

	permissions, err := uc.casbinEnforcer.GetPermissionsForRole(user.Role.NamaRole)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return uc.userHelper.AggregateUserPermissions(permissions), nil
}

func (uc *userUsecase) DownloadUserFiles(ctx context.Context, ownerUserID string, projectIDs, modulIDs []string) (string, error) {
	if err := uc.downloadHelper.ValidateDownloadRequest(projectIDs, modulIDs); err != nil {
		return "", err
	}

	_, err := uc.userRepo.GetByID(ctx, ownerUserID)
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

func (uc *userUsecase) GetUsersForRole(ctx context.Context, roleID uint) ([]dto.UserListItem, error) {
	_, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	users, err := uc.userRepo.GetByRoleID(ctx, roleID)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}

	return users, nil
}

func (uc *userUsecase) BulkAssignRole(ctx context.Context, userIDs []string, roleID uint) error {
	role, err := uc.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return apperrors.NewNotFoundError("Role")
		}
		return apperrors.NewInternalError(err)
	}

	users, err := uc.userRepo.GetByIDs(ctx, userIDs)
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

	if err := uc.userRepo.BulkUpdateRole(ctx, userIDs, roleID); err != nil {
		return apperrors.NewInternalError(err)
	}

	if uc.casbinEnforcer != nil {
		if err := uc.casbinEnforcer.SavePolicy(); err != nil {
			return apperrors.NewInternalError(err)
		}
	}

	return nil
}
