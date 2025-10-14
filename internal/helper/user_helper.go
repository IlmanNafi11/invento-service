package helper

import (
	"errors"
	"fiber-boiler-plate/internal/domain"
	"mime/multipart"
	"path/filepath"
)

type UserHelper struct{}

func NewUserHelper() *UserHelper {
	return &UserHelper{}
}

func (uh *UserHelper) BuildProfileData(user *domain.User, jumlahProject, jumlahModul int) *domain.ProfileData {
	roleName := ""
	if user.Role != nil {
		roleName = user.Role.NamaRole
	}

	return &domain.ProfileData{
		Name:          user.Name,
		Email:         user.Email,
		JenisKelamin:  user.JenisKelamin,
		FotoProfil:    user.FotoProfil,
		Role:          roleName,
		CreatedAt:     user.CreatedAt,
		JumlahProject: jumlahProject,
		JumlahModul:   jumlahModul,
	}
}

func (uh *UserHelper) AggregateUserPermissions(permissions [][]string) []domain.UserPermissionItem {
	resourceMap := make(map[string][]string)

	for _, perm := range permissions {
		if len(perm) >= 3 {
			resource := perm[1]
			action := perm[2]
			resourceMap[resource] = append(resourceMap[resource], action)
		}
	}

	var result []domain.UserPermissionItem
	for resource, actions := range resourceMap {
		result = append(result, domain.UserPermissionItem{
			Resource: resource,
			Actions:  actions,
		})
	}

	return result
}

func (uh *UserHelper) SaveProfilePhoto(fotoProfil interface{}, currentPhotoPath *string, userEmail, userRole string) (*string, error) {
	if fotoProfil == nil {
		return nil, nil
	}

	fileHeader, ok := fotoProfil.(*multipart.FileHeader)
	if !ok {
		return nil, nil
	}

	if err := ValidateImageFile(fileHeader); err != nil {
		return nil, err
	}

	var profilDir string
	if currentPhotoPath != nil && *currentPhotoPath != "" {
		profilDir = filepath.Dir(*currentPhotoPath)
		if err := DeleteFile(*currentPhotoPath); err != nil {
			return nil, errors.New("gagal menghapus foto profil lama")
		}
	} else {
		var err error
		profilDir, err = CreateProfilDirectory(userEmail, userRole)
		if err != nil {
			return nil, errors.New("gagal membuat direktori profil")
		}
	}

	ext := GetFileExtension(fileHeader.Filename)
	filename := "profil" + ext
	destPath := filepath.Join(profilDir, filename)

	if err := SaveUploadedFile(fileHeader, destPath); err != nil {
		return nil, errors.New("gagal menyimpan foto profil")
	}

	return &destPath, nil
}

func (uh *UserHelper) DeleteProfilePhoto(photoPath *string) error {
	if photoPath != nil && *photoPath != "" {
		return DeleteFile(*photoPath)
	}
	return nil
}

func (uh *UserHelper) NormalizeJenisKelamin(jenisKelamin string) *string {
	if jenisKelamin == "" {
		return nil
	}
	return &jenisKelamin
}
