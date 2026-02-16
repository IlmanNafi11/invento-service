package http_test

import (
	"context"
	httpcontroller "invento-service/internal/controller/http"
	"invento-service/internal/dto"
	"mime/multipart"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/mock"
)

// MockUserUsecase is a mock implementation of usecase.UserUsecase
type MockUserUsecase struct {
	mock.Mock
}

func (m *MockUserUsecase) GetUserList(ctx context.Context, params dto.UserListQueryParams) (*dto.UserListData, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserListData), args.Error(1)
}

func (m *MockUserUsecase) UpdateUserRole(ctx context.Context, userID string, roleName string) error {
	args := m.Called(ctx, userID, roleName)
	return args.Error(0)
}

func (m *MockUserUsecase) DeleteUser(ctx context.Context, userID string) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockUserUsecase) GetUserFiles(ctx context.Context, userID string, params dto.UserFilesQueryParams) (*dto.UserFilesData, error) {
	args := m.Called(ctx, userID, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.UserFilesData), args.Error(1)
}

func (m *MockUserUsecase) GetProfile(ctx context.Context, userID string) (*dto.ProfileData, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileData), args.Error(1)
}

func (m *MockUserUsecase) UpdateProfile(ctx context.Context, userID string, req dto.UpdateProfileRequest, fotoProfil *multipart.FileHeader) (*dto.ProfileData, error) {
	args := m.Called(ctx, userID, req, fotoProfil)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileData), args.Error(1)
}

func (m *MockUserUsecase) GetUserPermissions(ctx context.Context, userID string) ([]dto.UserPermissionItem, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.UserPermissionItem), args.Error(1)
}

func (m *MockUserUsecase) DownloadUserFiles(ctx context.Context, ownerUserID string, projectIDs, modulIDs []string) (string, error) {
	args := m.Called(ctx, ownerUserID, projectIDs, modulIDs)
	if args.Get(0) == nil || args.Get(0).(string) == "" {
		return "", args.Error(1)
	}
	return args.String(0), args.Error(1)
}

func (m *MockUserUsecase) GetUsersForRole(ctx context.Context, roleID uint) ([]dto.UserListItem, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.UserListItem), args.Error(1)
}

func (m *MockUserUsecase) BulkAssignRole(ctx context.Context, userIDs []string, roleID uint) error {
	args := m.Called(ctx, userIDs, roleID)
	return args.Error(0)
}

// Helper function to create a test app with authenticated middleware for UserController
func setupTestAppWithAuthForUser(controller *httpcontroller.UserController) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
	})

	app.Use(func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader != "" {
			if userID := c.Get("X-Test-User-ID"); userID != "" {
				c.Locals("user_id", "00000000-0000-0000-0000-000000000001")
				c.Locals("user_email", "test@example.com")
				c.Locals("user_role", "admin")
			}
		}
		return c.Next()
	})

	return app
}

// Helper function
func stringPtr(s string) *string {
	return &s
}

// createTempFile creates a temporary file for testing and returns its name.
// Caller is responsible for cleanup via os.Remove.
func createTempFile(content string) (string, error) {
	tmpFile, err := os.CreateTemp("", "test_download_*.zip")
	if err != nil {
		return "", err
	}
	if _, err = tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		return "", err
	}
	tmpFile.Close()
	return tmpFile.Name(), nil
}
