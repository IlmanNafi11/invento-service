package usecase

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/usecase/repo"
	"math"

	"gorm.io/gorm"
)

type UserUsecase interface {
	GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error)
	UpdateUserRole(userID uint, roleName string) error
	DeleteUser(userID uint) error
	GetUserFiles(userID uint, params domain.UserFilesQueryParams) (*domain.UserFilesData, error)
	GetProfile(userID uint) (*domain.ProfileData, error)
	GetUserPermissions(userID uint) ([]domain.UserPermissionItem, error)
}

type userUsecase struct {
	userRepo           repo.UserRepository
	roleRepo           repo.RoleRepository
	rolePermissionRepo repo.RolePermissionRepository
}

func NewUserUsecase(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	rolePermissionRepo repo.RolePermissionRepository,
) UserUsecase {
	return &userUsecase{
		userRepo:           userRepo,
		roleRepo:           roleRepo,
		rolePermissionRepo: rolePermissionRepo,
	}
}

func (uc *userUsecase) GetUserList(params domain.UserListQueryParams) (*domain.UserListData, error) {
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	users, total, err := uc.userRepo.GetAll(params.Search, params.FilterRole, params.Page, params.Limit)
	if err != nil {
		return nil, errors.New("gagal mengambil daftar user")
	}

	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	return &domain.UserListData{
		Items: users,
		Pagination: domain.PaginationData{
			Page:       params.Page,
			Limit:      params.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
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

	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 10
	}
	if params.Limit > 100 {
		params.Limit = 100
	}

	return &domain.UserFilesData{
		Items: []domain.UserFileItem{},
		Pagination: domain.PaginationData{
			Page:       params.Page,
			Limit:      params.Limit,
			TotalItems: 0,
			TotalPages: 0,
		},
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

	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	return &domain.ProfileData{
		Email: user.Email,
		Role:  roleName,
	}, nil
}

func (uc *userUsecase) GetUserPermissions(userID uint) ([]domain.UserPermissionItem, error) {
	user, err := uc.userRepo.GetByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user tidak ditemukan")
		}
		return nil, errors.New("gagal mengambil data user")
	}

	if user.RoleID == nil {
		return []domain.UserPermissionItem{}, nil
	}

	permissions, err := uc.rolePermissionRepo.GetPermissionsForRole(*user.RoleID)
	if err != nil {
		return nil, errors.New("gagal mengambil permissions user")
	}

	resourceMap := make(map[string][]string)
	for _, perm := range permissions {
		resourceMap[perm.Resource] = append(resourceMap[perm.Resource], perm.Action)
	}

	var result []domain.UserPermissionItem
	for resource, actions := range resourceMap {
		result = append(result, domain.UserPermissionItem{
			Resource: resource,
			Actions:  actions,
		})
	}

	return result, nil
}
