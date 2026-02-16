package storage

import (
	"errors"
	"fmt"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"mime/multipart"
)

type UserHelper struct {
	pathResolver *PathResolver
	config       *config.Config
}

func NewUserHelper(pathResolver *PathResolver, cfg *config.Config) *UserHelper {
	return &UserHelper{
		pathResolver: pathResolver,
		config:       cfg,
	}
}

func (uh *UserHelper) BuildProfileData(user *domain.User, jumlahProject, jumlahModul int) *dto.ProfileData {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	var fotoProfilPath *string
	if user.FotoProfil != nil && *user.FotoProfil != "" {
		fotoProfilPath = uh.pathResolver.ConvertToAPIPath(user.FotoProfil)
	}

	return &dto.ProfileData{
		Name:          user.Name,
		Email:         user.Email,
		JenisKelamin:  user.JenisKelamin,
		FotoProfil:    fotoProfilPath,
		Role:          roleName,
		CreatedAt:     user.CreatedAt,
		JumlahProject: jumlahProject,
		JumlahModul:   jumlahModul,
	}
}

func (uh *UserHelper) AggregateUserPermissions(permissions [][]string) []dto.UserPermissionItem {
	resourceMap := make(map[string][]string)

	for _, perm := range permissions {
		if len(perm) >= 3 {
			resource := perm[1]
			action := perm[2]
			resourceMap[resource] = append(resourceMap[resource], action)
		}
	}

	var result []dto.UserPermissionItem
	for resource, actions := range resourceMap {
		result = append(result, dto.UserPermissionItem{
			Resource: resource,
			Actions:  actions,
		})
	}

	return result
}

func (uh *UserHelper) SaveProfilePhoto(fotoProfil *multipart.FileHeader, userID string, currentPhotoPath *string) (*string, error) {
	if fotoProfil == nil {
		return nil, nil
	}

	if err := ValidateImageFile(fotoProfil); err != nil {
		return nil, err
	}

	if currentPhotoPath != nil && *currentPhotoPath != "" {
		if err := DeleteFile(*currentPhotoPath); err != nil {
			return nil, errors.New("gagal menghapus foto profil lama")
		}
	}

	profilDir := uh.pathResolver.GetProfilDirectory(userID)
	if err := uh.pathResolver.EnsureDirectoryExists(profilDir); err != nil {
		return nil, errors.New("gagal membuat direktori profil")
	}

	ext := GetFileExtension(fotoProfil.Filename)
	filename := fmt.Sprintf("profil%s", ext)
	destPath := uh.pathResolver.GetProfilFilePath(userID, filename)

	if err := SaveUploadedFile(fotoProfil, destPath); err != nil {
		return nil, errors.New("gagal menyimpan foto profil")
	}

	return &destPath, nil
}
