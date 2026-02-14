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

func (r *userRepository) GetByIDs(userIDs []string) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.Where("id IN ? AND is_active = ?", userIDs, true).Preload("Role").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *userRepository) buildUserListQuery(search, filterRole string) *gorm.DB {
	query := r.db.Table("user_profiles").
		Joins("LEFT JOIN roles ON roles.id = user_profiles.role_id").
		Where("user_profiles.is_active = ?", true)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("user_profiles.email ILIKE ?", searchPattern)
	}

	if filterRole != "" {
		query = query.Where("roles.nama_role = ?", filterRole)
	}

	return query
}

func (r *userRepository) GetAll(search, filterRole string, page, limit int) ([]domain.UserListItem, int, error) {
	var userListItems []domain.UserListItem
	var total int64

	baseQuery := r.buildUserListQuery(search, filterRole)

	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := baseQuery.
		Select("user_profiles.id, user_profiles.email, user_profiles.created_at as dibuat_pada, COALESCE(roles.nama_role, '') as role").
		Offset(offset).Limit(limit).Order("user_profiles.created_at DESC").
		Scan(&userListItems).Error; err != nil {
		return nil, 0, err
	}

	return userListItems, int(total), nil
}

func (r *userRepository) GetProfileWithCounts(userID string) (*domain.User, int, int, error) {
	var user domain.User
	err := r.db.Where("id = ? AND is_active = ?", userID, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, 0, 0, err
	}

	var projectCount int64
	r.db.Table("projects").Where("user_id = ?", userID).Count(&projectCount)

	var modulCount int64
	r.db.Table("moduls").Where("user_id = ?", userID).Count(&modulCount)

	return &user, int(projectCount), int(modulCount), nil
}

func (r *userRepository) GetUserFiles(userID string, search string, page, limit int) ([]domain.UserFileItem, int, error) {
	var items []domain.UserFileItem
	var total int64

	countQuery := r.db.Raw(`
		SELECT COUNT(*) FROM (
			SELECT p.id FROM projects p WHERE p.user_id = ? AND (? = '' OR p.nama_project ILIKE '%' || ? || '%')
			UNION ALL
			SELECT m.id FROM moduls m WHERE m.user_id = ? AND (? = '' OR m.file_name ILIKE '%' || ? || '%')
		) combined
	`, userID, search, search, userID, search, search)

	if err := countQuery.Scan(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	if err := r.db.Raw(`
		SELECT id, nama_file, kategori, download_url FROM (
			SELECT
				CAST(p.id AS TEXT) as id,
				p.nama_project as nama_file,
				'Project' as kategori,
				p.path_file as download_url,
				p.updated_at
			FROM projects p
			WHERE p.user_id = ?
				AND (? = '' OR p.nama_project ILIKE '%' || ? || '%')
			UNION ALL
			SELECT
				CAST(m.id AS TEXT) as id,
				m.file_name as nama_file,
				'Modul' as kategori,
				m.file_path as download_url,
				m.updated_at
			FROM moduls m
			WHERE m.user_id = ?
				AND (? = '' OR m.file_name ILIKE '%' || ? || '%')
		) combined
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, userID, search, search, userID, search, search, limit, offset).Scan(&items).Error; err != nil {
		return nil, 0, err
	}

	return items, int(total), nil
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

func (r *userRepository) GetByRoleID(roleID uint) ([]domain.UserListItem, error) {
	var userListItems []domain.UserListItem

	err := r.db.Table("user_profiles").
		Select("user_profiles.id, user_profiles.email, user_profiles.created_at as dibuat_pada, COALESCE(roles.nama_role, '') as role").
		Joins("LEFT JOIN roles ON roles.id = user_profiles.role_id").
		Where("user_profiles.is_active = ? AND user_profiles.role_id = ?", true, roleID).
		Order("user_profiles.created_at DESC").
		Scan(&userListItems).Error

	if err != nil {
		return nil, err
	}

	return userListItems, nil
}

func (r *userRepository) BulkUpdateRole(userIDs []string, roleID uint) error {
	return r.db.Model(&domain.User{}).
		Where("id IN ? AND is_active = ?", userIDs, true).
		Update("role_id", roleID).Error
}
