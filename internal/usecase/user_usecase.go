package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/helper"
	"invento-service/internal/httputil"
	"invento-service/internal/rbac"
	"invento-service/internal/storage"
	"invento-service/internal/usecase/repo"
	"mime/multipart"
	"regexp"
	"strconv"
	"strings"

	apperrors "invento-service/internal/errors"

	"github.com/rs/zerolog"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

type UserUsecase interface {
	GetUserList(ctx context.Context, params dto.UserListQueryParams) (*dto.UserListData, error)
	UpdateUserRole(ctx context.Context, userID, roleName string) error
	DeleteUser(ctx context.Context, userID string) error
	GetUserFiles(ctx context.Context, userID string, params dto.UserFilesQueryParams) (*dto.UserFilesData, error)
	GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error)
	UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest, fotoProfil *multipart.FileHeader) (*dto.ProfileData, error)
	GetUserPermissions(ctx context.Context, userID string) ([]dto.UserPermissionItem, error)
	DownloadUserFiles(ctx context.Context, ownerUserID string, projectIDs, modulIDs []string) (string, error)
	GetUsersForRole(ctx context.Context, roleID uint) ([]dto.UserListItem, error)
	BulkAssignRole(ctx context.Context, userIDs []string, roleID uint) error
	AdminCreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.CreateUserResponse, error)
	BulkImportUsers(ctx context.Context, file *excelize.File, req dto.ImportUsersRequest) (*dto.ImportReport, error)
}

type userUsecase struct {
	userRepo       repo.UserRepository
	roleRepo       repo.RoleRepository
	projectRepo    repo.ProjectRepository
	modulRepo      repo.ModulRepository
	authService    domain.AuthService
	casbinEnforcer *rbac.CasbinEnforcer
	userHelper     *storage.UserHelper
	downloadHelper *storage.DownloadHelper
	pathResolver   *storage.PathResolver
	excelHelper    *helper.ExcelHelper
	config         *config.Config
}

func NewUserUsecase(
	userRepo repo.UserRepository,
	roleRepo repo.RoleRepository,
	projectRepo repo.ProjectRepository,
	modulRepo repo.ModulRepository,
	authService domain.AuthService,
	casbinEnforcer *rbac.CasbinEnforcer,
	pathResolver *storage.PathResolver,
	cfg *config.Config,
	logger zerolog.Logger,
) UserUsecase {
	return &userUsecase{
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		projectRepo:    projectRepo,
		modulRepo:      modulRepo,
		authService:    authService,
		casbinEnforcer: casbinEnforcer,
		userHelper:     storage.NewUserHelper(pathResolver, cfg),
		downloadHelper: storage.NewDownloadHelper(pathResolver, logger),
		pathResolver:   pathResolver,
		excelHelper:    helper.NewExcelHelper(),
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

func (uc *userUsecase) UpdateUserRole(ctx context.Context, userID, roleName string) error {
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

	jumlahProject, _ := uc.projectRepo.CountByUserID(context.Background(), userID)
	jumlahModul, _ := uc.modulRepo.CountByUserID(context.Background(), userID)

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
	var id uint64
	for _, idStr := range projectIDs {
		id, err = strconv.ParseUint(idStr, 10, 32)
		if err != nil {
			return "", apperrors.NewValidationError("Format project ID tidak valid", err)
		}
		projectIDsUint = append(projectIDsUint, uint(id))
	}

	projects, err := uc.projectRepo.GetByIDs(ctx, projectIDsUint, ownerUserID)
	if err != nil {
		return "", apperrors.NewInternalError(err)
	}

	moduls, err := uc.modulRepo.GetByIDs(ctx, modulIDs, ownerUserID)
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

const mahasiswaEmailDomain = "@student.polije.ac.id"

func generateRandomPassword() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

type createUserParams struct {
	Email        string
	Name         string
	NIM          string
	JenisKelamin string
	Password     string
	RoleID       uint
	RoleName     string
}

type createUserResult struct {
	User     domain.User
	Password string
}

func (uc *userUsecase) createSingleUser(ctx context.Context, params createUserParams) (*createUserResult, error) {
	password := params.Password
	if password == "" {
		pwd, err := generateRandomPassword()
		if err != nil {
			return nil, apperrors.NewInternalError(err)
		}
		password = pwd
	}

	supabaseUserID, err := uc.authService.AdminCreateUser(ctx, params.Email, password)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, apperrors.NewInternalError(err)
	}

	roleID := int(params.RoleID)
	var jenisKelaminPtr *string
	if params.JenisKelamin != "" {
		jenisKelaminPtr = &params.JenisKelamin
	}

	user := domain.User{
		ID:           supabaseUserID,
		Email:        params.Email,
		Name:         params.Name,
		JenisKelamin: jenisKelaminPtr,
		RoleID:       &roleID,
		IsActive:     true,
	}
	if err := uc.userRepo.SaveOrUpdate(ctx, &user); err != nil {
		_ = uc.authService.DeleteUser(ctx, supabaseUserID)
		return nil, apperrors.NewInternalError(err)
	}

	if uc.casbinEnforcer != nil {
		if err := uc.casbinEnforcer.AddRoleForUser(supabaseUserID, params.RoleName); err != nil {
			_ = uc.userRepo.Delete(ctx, supabaseUserID)
			_ = uc.authService.DeleteUser(ctx, supabaseUserID)
			return nil, apperrors.NewInternalError(err)
		}
	}

	return &createUserResult{
		User:     user,
		Password: password,
	}, nil
}

func (uc *userUsecase) AdminCreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.CreateUserResponse, error) {
	role, err := uc.roleRepo.GetByID(ctx, uint(req.RoleID))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperrors.NewNotFoundError("Role")
		}
		return nil, apperrors.NewInternalError(err)
	}

	if strings.EqualFold(role.NamaRole, "Mahasiswa") && !strings.HasSuffix(strings.ToLower(req.Email), mahasiswaEmailDomain) {
		return nil, apperrors.NewValidationError("Email mahasiswa harus menggunakan domain "+mahasiswaEmailDomain, nil)
	}

	password := ""
	if req.Password != nil && *req.Password != "" {
		password = *req.Password
	}

	result, err := uc.createSingleUser(ctx, createUserParams{
		Email:    req.Email,
		Name:     req.Name,
		Password: password,
		RoleID:   uint(req.RoleID),
		RoleName: role.NamaRole,
	})
	if err != nil {
		return nil, err
	}

	if uc.casbinEnforcer != nil {
		if err := uc.casbinEnforcer.SavePolicy(); err != nil {
			_ = uc.userRepo.Delete(ctx, result.User.ID)
			_ = uc.authService.DeleteUser(ctx, result.User.ID)
			return nil, apperrors.NewInternalError(err)
		}
	}

	var generatedPassword *string
	if req.Password == nil || *req.Password == "" {
		generatedPassword = &result.Password
	}

	return &dto.CreateUserResponse{
		ID:                result.User.ID,
		Email:             req.Email,
		Name:              req.Name,
		RoleID:            req.RoleID,
		RoleName:          role.NamaRole,
		GeneratedPassword: generatedPassword,
		IsActive:          true,
	}, nil
}

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func (uc *userUsecase) BulkImportUsers(ctx context.Context, file *excelize.File, req dto.ImportUsersRequest) (*dto.ImportReport, error) {
	rows, err := uc.excelHelper.ParseImportFile(file)
	if err != nil {
		return nil, apperrors.NewValidationError(err.Error(), err)
	}

	report := &dto.ImportReport{
		TotalBaris: len(rows),
		Detail:     make([]dto.ImportReportRow, 0, len(rows)),
	}

	if len(rows) == 0 {
		return report, nil
	}

	emails := make([]string, 0, len(rows))
	for _, row := range rows {
		if row.Email != "" {
			emails = append(emails, strings.ToLower(row.Email))
		}
	}

	existingUsers, err := uc.userRepo.FindByEmails(ctx, emails)
	if err != nil {
		return nil, apperrors.NewInternalError(err)
	}
	existingEmails := make(map[string]bool, len(existingUsers))
	for _, u := range existingUsers {
		existingEmails[strings.ToLower(u.Email)] = true
	}

	seenEmails := make(map[string]bool)

	for _, row := range rows {
		emailLower := strings.ToLower(row.Email)

		if row.Email == "" || row.Nama == "" {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: "Email dan Nama wajib diisi",
			})
			report.Dilewati++
			continue
		}

		if !emailRegex.MatchString(row.Email) {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: "Format email tidak valid",
			})
			report.Dilewati++
			continue
		}

		if seenEmails[emailLower] {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: "Email duplikat dalam file",
			})
			report.Dilewati++
			continue
		}

		if existingEmails[emailLower] {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: "Email sudah terdaftar",
			})
			report.Dilewati++
			continue
		}

		var role *domain.Role
		var roleID int
		if row.Role != "" {
			role, err = uc.roleRepo.GetByName(ctx, row.Role)
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					report.Detail = append(report.Detail, dto.ImportReportRow{
						Baris:  row.RowNumber,
						Email:  row.Email,
						Nama:   row.Nama,
						Status: "dilewati",
						Alasan: fmt.Sprintf("Role '%s' tidak ditemukan", row.Role),
					})
					report.Dilewati++
					continue
				}
				return nil, apperrors.NewInternalError(err)
			}
			roleID = int(role.ID)
		} else {
			roleID = req.DefaultRoleID
			role, err = uc.roleRepo.GetByID(ctx, uint(roleID))
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					report.Detail = append(report.Detail, dto.ImportReportRow{
						Baris:  row.RowNumber,
						Email:  row.Email,
						Nama:   row.Nama,
						Status: "dilewati",
						Alasan: "Default role tidak ditemukan",
					})
					report.Dilewati++
					continue
				}
				return nil, apperrors.NewInternalError(err)
			}
		}

		if strings.EqualFold(role.NamaRole, "Mahasiswa") && !strings.HasSuffix(emailLower, mahasiswaEmailDomain) {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: "Mahasiswa harus menggunakan email @student.polije.ac.id",
			})
			report.Dilewati++
			continue
		}

		if row.JenisKelamin != "" && row.JenisKelamin != "Laki-laki" && row.JenisKelamin != "Perempuan" {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: "Jenis Kelamin harus 'Laki-laki' atau 'Perempuan'",
			})
			report.Dilewati++
			continue
		}

		result, createErr := uc.createSingleUser(ctx, createUserParams{
			Email:        row.Email,
			Name:         row.Nama,
			JenisKelamin: row.JenisKelamin,
			Password:     row.Password,
			RoleID:       uint(roleID),
			RoleName:     role.NamaRole,
		})
		if createErr != nil {
			report.Detail = append(report.Detail, dto.ImportReportRow{
				Baris:  row.RowNumber,
				Email:  row.Email,
				Nama:   row.Nama,
				Status: "dilewati",
				Alasan: fmt.Sprintf("Gagal membuat akun: %s", createErr.Error()),
			})
			report.Dilewati++
			continue
		}

		var reportPassword string
		if row.Password == "" {
			reportPassword = result.Password
		}

		seenEmails[emailLower] = true
		existingEmails[emailLower] = true

		report.Detail = append(report.Detail, dto.ImportReportRow{
			Baris:    row.RowNumber,
			Email:    row.Email,
			Nama:     row.Nama,
			Status:   "berhasil",
			Password: reportPassword,
		})
		report.Berhasil++
	}

	if uc.casbinEnforcer != nil && report.Berhasil > 0 {
		if err := uc.casbinEnforcer.SavePolicy(); err != nil {
			return nil, apperrors.NewInternalError(err)
		}
	}

	return report, nil
}
