package usecase

import (
	"context"
	"errors"
	"invento-service/config"
	"invento-service/internal/domain"
	"invento-service/internal/dto"
	"invento-service/internal/storage"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

// newAdminTestUsecase creates a userUsecase with MockAuthService for admin tests.
func newAdminTestUsecase(mockUserRepo *MockUserRepository, mockRoleRepo *MockRoleRepository, mockAuthService *MockAuthService) UserUsecase {
	mockProjectRepo := new(MockProjectRepository)
	mockModulRepo := new(MockModulRepository)
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
	}
	pathResolver := storage.NewPathResolver(cfg)
	return NewUserUsecase(mockUserRepo, mockRoleRepo, mockProjectRepo, mockModulRepo, mockAuthService, nil, pathResolver, cfg, zerolog.Nop())
}

// =============================================================================
// AdminCreateUser Tests
// =============================================================================

func TestAdminCreateUser_Success(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	roleID := uint(3)
	role := &domain.Role{ID: roleID, NamaRole: "Mahasiswa"}
	password := "secure123"

	req := dto.CreateUserRequest{
		Email:    "student1@student.polije.ac.id",
		Name:     "Budi Santoso",
		Password: &password,
		RoleID:   int(roleID),
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockAuthService.On("AdminCreateUser", mock.Anything, req.Email, password).Return("supabase-uid-123", nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(nil)

	result, err := uc.AdminCreateUser(context.Background(), req)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "supabase-uid-123", result.ID)
	assert.Equal(t, req.Email, result.Email)
	assert.Equal(t, req.Name, result.Name)
	assert.Equal(t, "Mahasiswa", result.RoleName)
	assert.True(t, result.IsActive)
	// Password was provided, so GeneratedPassword should be nil
	assert.Nil(t, result.GeneratedPassword)

	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestAdminCreateUser_InvalidRole(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	req := dto.CreateUserRequest{
		Email:  "test@example.com",
		Name:   "Test User",
		RoleID: 999,
	}

	mockRoleRepo.On("GetByID", mock.Anything, uint(999)).Return(nil, gorm.ErrRecordNotFound)

	result, err := uc.AdminCreateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "Role tidak ditemukan")

	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertNotCalled(t, "AdminCreateUser")
	mockUserRepo.AssertNotCalled(t, "Create")
}

func TestAdminCreateUser_MahasiswaDomainValidation(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	roleID := uint(3)
	role := &domain.Role{ID: roleID, NamaRole: "Mahasiswa"}

	req := dto.CreateUserRequest{
		Email:  "student@gmail.com",
		Name:   "Test Student",
		RoleID: int(roleID),
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)

	result, err := uc.AdminCreateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "@student.polije.ac.id")

	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertNotCalled(t, "AdminCreateUser")
	mockUserRepo.AssertNotCalled(t, "Create")
}

func TestAdminCreateUser_AuthServiceFailure(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	roleID := uint(2)
	role := &domain.Role{ID: roleID, NamaRole: "Dosen"}

	req := dto.CreateUserRequest{
		Email:  "dosen@polije.ac.id",
		Name:   "Dr. Ahmad",
		RoleID: int(roleID),
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockAuthService.On("AdminCreateUser", mock.Anything, req.Email, mock.AnythingOfType("string")).Return("", errors.New("supabase error: email already exists"))

	result, err := uc.AdminCreateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "supabase error")

	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertNotCalled(t, "Create")
}

func TestAdminCreateUser_DBFailure_Rollback(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	roleID := uint(2)
	role := &domain.Role{ID: roleID, NamaRole: "Dosen"}

	req := dto.CreateUserRequest{
		Email:  "dosen2@polije.ac.id",
		Name:   "Dr. Siti",
		RoleID: int(roleID),
	}

	mockRoleRepo.On("GetByID", mock.Anything, roleID).Return(role, nil)
	mockAuthService.On("AdminCreateUser", mock.Anything, req.Email, mock.AnythingOfType("string")).Return("supabase-uid-456", nil)
	mockUserRepo.On("Create", mock.Anything, mock.AnythingOfType("*domain.User")).Return(errors.New("db constraint violation"))
	// Rollback: DeleteUser should be called when DB create fails
	mockAuthService.On("DeleteUser", mock.Anything, "supabase-uid-456").Return(nil)

	result, err := uc.AdminCreateUser(context.Background(), req)

	assert.Error(t, err)
	assert.Nil(t, result)

	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	// Verify rollback: DeleteUser was called
	mockAuthService.AssertCalled(t, "DeleteUser", mock.Anything, "supabase-uid-456")
}

// =============================================================================
// BulkImportUsers Tests
// =============================================================================

// createTestExcelFile creates an Excel file for import testing with the "Data Import" sheet.
func createTestExcelFile(t *testing.T, headers []string, rows [][]interface{}) *excelize.File {
	t.Helper()
	f := excelize.NewFile()

	// Create "Data Import" sheet (required by ExcelHelper.ParseImportFile)
	_, err := f.NewSheet("Data Import")
	assert.NoError(t, err)
	err = f.DeleteSheet("Sheet1")
	assert.NoError(t, err)

	// Set headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue("Data Import", cell, header)
	}

	// Set data rows
	for rowIdx, row := range rows {
		for colIdx, val := range row {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			f.SetCellValue("Data Import", cell, val)
		}
	}

	return f
}

func TestBulkImportUsers_Success(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	headers := []string{"Email", "Nama", "Password", "Jenis Kelamin", "Role"}
	rows := [][]interface{}{
		{"user1@student.polije.ac.id", "User Satu", "pass1234", "Laki-laki", "Mahasiswa"},
		{"user2@student.polije.ac.id", "User Dua", "pass5678", "Perempuan", "Mahasiswa"},
	}
	file := createTestExcelFile(t, headers, rows)

	req := dto.ImportUsersRequest{DefaultRoleID: 3}

	roleID := uint(3)
	role := &domain.Role{ID: roleID, NamaRole: "Mahasiswa"}

	// FindByEmails to check existing users
	mockUserRepo.On("FindByEmails", mock.Anything, mock.AnythingOfType("[]string")).Return([]domain.User{}, nil)

	// Role lookup for each row (by name)
	mockRoleRepo.On("GetByName", mock.Anything, "Mahasiswa").Return(role, nil)

	// Auth + DB for user 1
	mockAuthService.On("AdminCreateUser", mock.Anything, "user1@student.polije.ac.id", "pass1234").Return("uid-1", nil)
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "user1@student.polije.ac.id"
	})).Return(nil)

	// Auth + DB for user 2
	mockAuthService.On("AdminCreateUser", mock.Anything, "user2@student.polije.ac.id", "pass5678").Return("uid-2", nil)
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "user2@student.polije.ac.id"
	})).Return(nil)

	report, err := uc.BulkImportUsers(context.Background(), file, req)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2, report.TotalBaris)
	assert.Equal(t, 2, report.Berhasil)
	assert.Equal(t, 0, report.Dilewati)

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestBulkImportUsers_PartialFailure(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	headers := []string{"Email", "Nama", "Password", "Jenis Kelamin", "Role"}
	rows := [][]interface{}{
		{"user1@student.polije.ac.id", "User Satu", "pass1234", "Laki-laki", "Mahasiswa"},
		{"user2@student.polije.ac.id", "User Dua", "pass5678", "Perempuan", "Mahasiswa"},
		{"user3@student.polije.ac.id", "User Tiga", "pass9012", "Laki-laki", "Mahasiswa"},
	}
	file := createTestExcelFile(t, headers, rows)

	req := dto.ImportUsersRequest{DefaultRoleID: 3}

	roleID := uint(3)
	role := &domain.Role{ID: roleID, NamaRole: "Mahasiswa"}

	mockUserRepo.On("FindByEmails", mock.Anything, mock.AnythingOfType("[]string")).Return([]domain.User{}, nil)
	mockRoleRepo.On("GetByName", mock.Anything, "Mahasiswa").Return(role, nil)

	// User 1 succeeds
	mockAuthService.On("AdminCreateUser", mock.Anything, "user1@student.polije.ac.id", "pass1234").Return("uid-1", nil)
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "user1@student.polije.ac.id"
	})).Return(nil)

	// User 2 fails auth
	mockAuthService.On("AdminCreateUser", mock.Anything, "user2@student.polije.ac.id", "pass5678").Return("", errors.New("auth error"))

	// User 3 succeeds
	mockAuthService.On("AdminCreateUser", mock.Anything, "user3@student.polije.ac.id", "pass9012").Return("uid-3", nil)
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "user3@student.polije.ac.id"
	})).Return(nil)

	report, err := uc.BulkImportUsers(context.Background(), file, req)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 3, report.TotalBaris)
	assert.Equal(t, 2, report.Berhasil)
	assert.Equal(t, 1, report.Dilewati)

	// Check the failed row detail
	var failedRow *dto.ImportReportRow
	for i := range report.Detail {
		if report.Detail[i].Status == "dilewati" {
			failedRow = &report.Detail[i]
			break
		}
	}
	assert.NotNil(t, failedRow)
	assert.Equal(t, "user2@student.polije.ac.id", failedRow.Email)
	assert.Contains(t, failedRow.Alasan, "Gagal membuat akun")

	mockUserRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestBulkImportUsers_DuplicateEmails(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	headers := []string{"Email", "Nama", "Password", "Jenis Kelamin", "Role"}
	rows := [][]interface{}{
		{"same@student.polije.ac.id", "User Satu", "pass1234", "Laki-laki", "Mahasiswa"},
		{"same@student.polije.ac.id", "User Dua", "pass5678", "Perempuan", "Mahasiswa"},
	}
	file := createTestExcelFile(t, headers, rows)

	req := dto.ImportUsersRequest{DefaultRoleID: 3}

	roleID := uint(3)
	role := &domain.Role{ID: roleID, NamaRole: "Mahasiswa"}

	mockUserRepo.On("FindByEmails", mock.Anything, mock.AnythingOfType("[]string")).Return([]domain.User{}, nil)
	mockRoleRepo.On("GetByName", mock.Anything, "Mahasiswa").Return(role, nil)

	// Only the first occurrence should be created
	mockAuthService.On("AdminCreateUser", mock.Anything, "same@student.polije.ac.id", "pass1234").Return("uid-1", nil)
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "same@student.polije.ac.id"
	})).Return(nil)

	report, err := uc.BulkImportUsers(context.Background(), file, req)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2, report.TotalBaris)
	assert.Equal(t, 1, report.Berhasil)
	assert.Equal(t, 1, report.Dilewati)

	// Check the duplicate row
	var dupRow *dto.ImportReportRow
	for i := range report.Detail {
		if report.Detail[i].Status == "dilewati" {
			dupRow = &report.Detail[i]
			break
		}
	}
	assert.NotNil(t, dupRow)
	assert.Contains(t, dupRow.Alasan, "duplikat")

	mockUserRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestBulkImportUsers_ExistingUser(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	headers := []string{"Email", "Nama", "Password", "Jenis Kelamin", "Role"}
	rows := [][]interface{}{
		{"existing@student.polije.ac.id", "Existing User", "pass1234", "Laki-laki", "Mahasiswa"},
		{"new@student.polije.ac.id", "New User", "pass5678", "Perempuan", "Mahasiswa"},
	}
	file := createTestExcelFile(t, headers, rows)

	req := dto.ImportUsersRequest{DefaultRoleID: 3}

	roleID := uint(3)
	role := &domain.Role{ID: roleID, NamaRole: "Mahasiswa"}

	// Return existing user for the first email
	mockUserRepo.On("FindByEmails", mock.Anything, mock.AnythingOfType("[]string")).Return([]domain.User{
		{ID: "existing-uid", Email: "existing@student.polije.ac.id", Name: "Existing User"},
	}, nil)

	mockRoleRepo.On("GetByName", mock.Anything, "Mahasiswa").Return(role, nil)

	// Only the new user should be created
	mockAuthService.On("AdminCreateUser", mock.Anything, "new@student.polije.ac.id", "pass5678").Return("uid-new", nil)
	mockUserRepo.On("Create", mock.Anything, mock.MatchedBy(func(u *domain.User) bool {
		return u.Email == "new@student.polije.ac.id"
	})).Return(nil)

	report, err := uc.BulkImportUsers(context.Background(), file, req)

	assert.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, 2, report.TotalBaris)
	assert.Equal(t, 1, report.Berhasil)
	assert.Equal(t, 1, report.Dilewati)

	// Check the skipped row
	var skippedRow *dto.ImportReportRow
	for i := range report.Detail {
		if report.Detail[i].Status == "dilewati" {
			skippedRow = &report.Detail[i]
			break
		}
	}
	assert.NotNil(t, skippedRow)
	assert.Equal(t, "existing@student.polije.ac.id", skippedRow.Email)
	assert.Contains(t, skippedRow.Alasan, "sudah terdaftar")

	mockUserRepo.AssertExpectations(t)
	mockAuthService.AssertExpectations(t)
}

func TestBulkImportUsers_InvalidExcelFormat(t *testing.T) {
	t.Parallel()
	mockUserRepo := new(MockUserRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockAuthService := new(MockAuthService)

	uc := newAdminTestUsecase(mockUserRepo, mockRoleRepo, mockAuthService)

	// Create a file WITHOUT the "Data Import" sheet — only default "Sheet1"
	file := excelize.NewFile()

	req := dto.ImportUsersRequest{DefaultRoleID: 3}

	report, err := uc.BulkImportUsers(context.Background(), file, req)

	assert.Error(t, err)
	assert.Nil(t, report)
	assert.Contains(t, err.Error(), "Data Import")

	mockAuthService.AssertNotCalled(t, "AdminCreateUser")
	mockUserRepo.AssertNotCalled(t, "Create")
}
