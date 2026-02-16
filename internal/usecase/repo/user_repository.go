package repo

import (
	"context"
	"invento-service/internal/domain"
	"invento-service/internal/dto"

	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("email = ? AND is_active = ?", email, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*domain.User, error) {
	var user domain.User
	err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByIDs(ctx context.Context, userIDs []string) ([]*domain.User, error) {
	var users []*domain.User
	err := r.db.WithContext(ctx).Where("id IN ? AND is_active = ?", userIDs, true).Preload("Role").Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) buildUserListQuery(ctx context.Context, search, filterRole string) *gorm.DB {
	query := r.db.WithContext(ctx).Table("user_profiles").
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

func (r *userRepository) GetAll(ctx context.Context, search, filterRole string, page, limit int) ([]dto.UserListItem, int, error) {
	var userListItems []dto.UserListItem
	var total int64

	baseQuery := r.buildUserListQuery(ctx, search, filterRole)

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

func (r *userRepository) GetProfileWithCounts(ctx context.Context, userID string) (userResult *domain.User, projectTotal, modulTotal int, err error) {
	var user domain.User

	err = r.db.WithContext(ctx).Where("id = ? AND is_active = ?", userID, true).Preload("Role").First(&user).Error
	if err != nil {
		return nil, 0, 0, err
	}

	var projectCount int64
	r.db.WithContext(ctx).Table("projects").Where("user_id = ?", userID).Count(&projectCount)

	var modulCount int64
	r.db.WithContext(ctx).Table("moduls").Where("user_id = ?", userID).Count(&modulCount)

	return &user, int(projectCount), int(modulCount), nil
}

func (r *userRepository) GetUserFiles(ctx context.Context, userID, search string, page, limit int) ([]dto.UserFileItem, int, error) {
	var items []dto.UserFileItem
	var total int64

	countQuery := r.db.WithContext(ctx).Raw(`
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
	if err := r.db.WithContext(ctx).Raw(`
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

func (r *userRepository) UpdateRole(ctx context.Context, userID string, roleID *int) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}

func (r *userRepository) UpdateProfile(ctx context.Context, userID, name string, jenisKelamin, fotoProfil *string) error {
	updates := map[string]interface{}{
		"name": name,
	}

	if jenisKelamin != nil {
		updates["jenis_kelamin"] = jenisKelamin
	}

	if fotoProfil != nil {
		updates["foto_profil"] = fotoProfil
	}

	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Updates(updates).Error
}

func (r *userRepository) Delete(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).Where("id = ?", userID).Update("is_active", false).Error
}

func (r *userRepository) GetByRoleID(ctx context.Context, roleID uint) ([]dto.UserListItem, error) {
	var userListItems []dto.UserListItem

	err := r.db.WithContext(ctx).Table("user_profiles").
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

func (r *userRepository) BulkUpdateRole(ctx context.Context, userIDs []string, roleID uint) error {
	return r.db.WithContext(ctx).Model(&domain.User{}).
		Where("id IN ? AND is_active = ?", userIDs, true).
		Update("role_id", roleID).Error
}
