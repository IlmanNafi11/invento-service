package repo

import (
	"fiber-boiler-plate/internal/domain"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("email = ? AND is_active = ?", email, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(id string) (*domain.User, error) {
	var user domain.User
	err := r.db.Where("id = ? AND is_active = ?", id, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	var userListItems []domain.UserListItem
	var total int64

	countQuery := r.db.Table("user_profiles").
		Joins("LEFT JOIN roles ON roles.id = user_profiles.role_id").
		Where("user_profiles.is_active = ?", true)

	if search != "" {
		searchPattern := "%" + search + "%"
		countQuery = countQuery.Where("user_profiles.email LIKE ?", searchPattern)
	}

	if filterRole != "" {
		countQuery = countQuery.Where("roles.nama_role = ?", filterRole)
	}

	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	dataQuery := r.db.Table("user_profiles").
		Select("user_profiles.id, user_profiles.email, user_profiles.created_at as dibuat_pada, COALESCE(roles.nama_role, '') as role").
		Joins("LEFT JOIN roles ON roles.id = user_profiles.role_id").
		Where("user_profiles.is_active = ?", true)

	if search != "" {
		searchPattern := "%" + search + "%"
		dataQuery = dataQuery.Where("user_profiles.email LIKE ?", searchPattern)
	}

	if filterRole != "" {
		dataQuery = dataQuery.Where("roles.nama_role = ?", filterRole)
	}

	offset := (page - 1) * limit
	if err := dataQuery.Offset(offset).Limit(limit).Order("user_profiles.created_at DESC").Scan(&userListItems).Error; err != nil {
		return nil, 0, err
	}

	return userListItems, int(total), nil
}

func (r *userRepository) UpdateRole(userID string, roleID *int) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *userRepository) UpdateProfile(userID string, name string, jenisKelamin *string, fotoProfil *string) error {
	updates := map[string]interface{}{
		"name": name,
	}

	if jenisKelamin != nil {
		updates["jenis_kelamin"] = jenisKelamin
	}

	if fotoProfil != nil {
		updates["foto_profil"] = fotoProfil
	}

	return r.db.Model(&domain.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *userRepository) Delete(userID string) error {
	return r.db.Model(&domain.User{}).Where("id = ?", userID).Update("is_active", false).Error
}
