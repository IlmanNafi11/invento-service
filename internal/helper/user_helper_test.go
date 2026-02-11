package helper_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"io"
	"mime/multipart"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserHelper(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
			PathProduction:  "/var/uploads",
		},
	}
	pathResolver := helper.NewPathResolver(cfg)

	userHelper := helper.NewUserHelper(pathResolver, cfg)

	assert.NotNil(t, userHelper)
	assert.NotNil(t, userHelper)
}

func TestUserHelper_BuildProfileData(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
		},
	}
	pathResolver := helper.NewPathResolver(cfg)
	userHelper := helper.NewUserHelper(pathResolver, cfg)

	tests := []struct {
		name          string
		user          *domain.User
		jumlahProject int
		jumlahModul   int
		expectedName  string
		expectedRole  string
	}{
		{
			name: "Complete user with role and photo",
			user: &domain.User{
				Name:         "John Doe",
				Email:        "john@example.com",
				JenisKelamin: stringPtr("L"),
				FotoProfil:   stringPtr("/uploads/profile.jpg"),
				Role: &domain.Role{
					NamaRole: "Admin",
				},
				CreatedAt: time.Now(),
			},
			jumlahProject: 5,
			jumlahModul:   10,
			expectedName:  "John Doe",
			expectedRole:  "Admin",
		},
		{
			name: "User without role",
			user: &domain.User{
				Name:         "Jane Doe",
				Email:        "jane@example.com",
				JenisKelamin: stringPtr("P"),
				FotoProfil:   nil,
				Role:         nil,
				CreatedAt:    time.Now(),
			},
			jumlahProject: 2,
			jumlahModul:   3,
			expectedName:  "Jane Doe",
			expectedRole:  "",
		},
		{
			name: "User with empty photo path",
			user: &domain.User{
				Name:         "Bob Smith",
				Email:        "bob@example.com",
				JenisKelamin: stringPtr("L"),
				FotoProfil:   stringPtr(""),
				Role: &domain.Role{
					NamaRole: "User",
				},
				CreatedAt: time.Now(),
			},
			jumlahProject: 0,
			jumlahModul:   1,
			expectedName:  "Bob Smith",
			expectedRole:  "User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userHelper.BuildProfileData(tt.user, tt.jumlahProject, tt.jumlahModul)

			assert.Equal(t, tt.expectedName, result.Name)
			assert.Equal(t, tt.user.Email, result.Email)
			assert.Equal(t, tt.user.JenisKelamin, result.JenisKelamin)
			assert.Equal(t, tt.expectedRole, result.Role)
			assert.Equal(t, tt.jumlahProject, result.JumlahProject)
			assert.Equal(t, tt.jumlahModul, result.JumlahModul)
			assert.WithinDuration(t, tt.user.CreatedAt, result.CreatedAt, time.Second)
		})
	}
}

func TestUserHelper_AggregateUserPermissions(t *testing.T) {
	cfg := &config.Config{}
	pathResolver := helper.NewPathResolver(cfg)
	userHelper := helper.NewUserHelper(pathResolver, cfg)

	tests := []struct {
		name               string
		permissions        [][]string
		expectedResource   string
		expectedActionsLen int
	}{
		{
			name: "Single permission",
			permissions: [][]string{
				{"role", "projects", "create"},
			},
			expectedResource:   "projects",
			expectedActionsLen: 1,
		},
		{
			name: "Multiple permissions for same resource",
			permissions: [][]string{
				{"role", "projects", "create"},
				{"role", "projects", "read"},
				{"role", "projects", "update"},
				{"role", "projects", "delete"},
			},
			expectedResource:   "projects",
			expectedActionsLen: 4,
		},
		{
			name: "Multiple resources",
			permissions: [][]string{
				{"role", "projects", "create"},
				{"role", "moduls", "read"},
				{"role", "users", "update"},
			},
			expectedResource:   "",
			expectedActionsLen: 0,
		},
		{
			name: "Invalid permission format",
			permissions: [][]string{
				{"role", "projects"},
				{"only_one"},
			},
			expectedResource:   "",
			expectedActionsLen: 0,
		},
		{
			name:               "Empty permissions",
			permissions:        [][]string{},
			expectedResource:   "",
			expectedActionsLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := userHelper.AggregateUserPermissions(tt.permissions)

			if tt.expectedResource != "" {
				found := false
				for _, item := range result {
					if item.Resource == tt.expectedResource {
						assert.Equal(t, tt.expectedActionsLen, len(item.Actions))
						found = true
						break
					}
				}
				assert.True(t, found, "Expected resource not found")
			}
		})
	}
}

func TestUserHelper_AggregateUserPermissions_Deduplication(t *testing.T) {
	cfg := &config.Config{}
	pathResolver := helper.NewPathResolver(cfg)
	userHelper := helper.NewUserHelper(pathResolver, cfg)

	permissions := [][]string{
		{"role", "projects", "create"},
		{"role", "projects", "read"},
		{"role", "projects", "create"}, // Duplicate
		{"role", "projects", "read"},   // Duplicate
	}

	result := userHelper.AggregateUserPermissions(permissions)

	assert.Len(t, result, 1)
	assert.Equal(t, "projects", result[0].Resource)
	// Actions may contain duplicates, test that they're aggregated
	assert.GreaterOrEqual(t, len(result[0].Actions), 2)
}

func TestUserHelper_SaveProfilePhoto(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: tempDir,
		},
	}
	pathResolver := helper.NewPathResolver(cfg)
	userHelper := helper.NewUserHelper(pathResolver, cfg)

	t.Run("Nil fotoProfil returns nil", func(t *testing.T) {
		result, err := userHelper.SaveProfilePhoto(nil, "1", nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Non-fileheader type returns nil", func(t *testing.T) {
		result, err := userHelper.SaveProfilePhoto("string", "1", nil)
		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Valid image file saves successfully", func(t *testing.T) {
		// Create a pipe for multipart writer
		pr, pw := io.Pipe()
		writer := multipart.NewWriter(pw)

		// Write multipart form in goroutine
		go func() {
			defer pw.Close()
			part, err := writer.CreateFormFile("file", "profil.jpg")
			require.NoError(t, err)
			// Write valid JPEG content
			imageContent := []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01} // JPEG header + JFIF
			part.Write(imageContent)
			writer.Close()
		}()

		// Read from pipe to create reader
		reader := multipart.NewReader(pr, writer.Boundary())
		form, err := reader.ReadForm(1 << 20) // 1MB max memory
		require.NoError(t, err)
		defer pr.Close()

		fileHeader := form.File["file"][0]

		userID := "123"

		result, err := userHelper.SaveProfilePhoto(fileHeader, userID, nil)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Contains(t, *result, "profil")
	})

	t.Run("Invalid file extension returns error", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "test.txt",
			Size:     100,
		}

		result, err := userHelper.SaveProfilePhoto(fileHeader, "1", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "format foto profil")
	})

	t.Run("File too large returns error", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "large.jpg",
			Size:     3 * 1024 * 1024, // 3MB
		}

		result, err := userHelper.SaveProfilePhoto(fileHeader, "1", nil)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "2MB")
	})
}

func TestUserHelper_DeleteProfilePhoto(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: tempDir,
		},
	}
	pathResolver := helper.NewPathResolver(cfg)
	userHelper := helper.NewUserHelper(pathResolver, cfg)

	t.Run("Nil photo path returns no error", func(t *testing.T) {
		err := userHelper.DeleteProfilePhoto(nil)
		assert.NoError(t, err)
	})

	t.Run("Empty photo path returns no error", func(t *testing.T) {
		emptyPath := ""
		err := userHelper.DeleteProfilePhoto(&emptyPath)
		assert.NoError(t, err)
	})

	t.Run("Non-existent file returns no error", func(t *testing.T) {
		nonExistentPath := "/tmp/non/existent/file.jpg"
		err := userHelper.DeleteProfilePhoto(&nonExistentPath)
		assert.NoError(t, err)
	})

	t.Run("Existing file is deleted successfully", func(t *testing.T) {
		// Create a temporary file
		tempFile, err := os.CreateTemp(tempDir, "test_*.jpg")
		require.NoError(t, err)
		tempFile.Close()

		filePath := tempFile.Name()
		err = userHelper.DeleteProfilePhoto(&filePath)

		assert.NoError(t, err)
		_, err = os.Stat(filePath)
		assert.True(t, os.IsNotExist(err))
	})
}

func TestUserHelper_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: tempDir,
		},
	}
	pathResolver := helper.NewPathResolver(cfg)
	userHelper := helper.NewUserHelper(pathResolver, cfg)

	t.Run("Build complete user profile", func(t *testing.T) {
		user := &domain.User{
			Name:         "Test User",
			Email:        "test@example.com",
			JenisKelamin: stringPtr("L"),
			FotoProfil:   stringPtr("/uploads/test.jpg"),
			Role: &domain.Role{
				NamaRole: "Mahasiswa",
			},
			CreatedAt: time.Now(),
		}

		profileData := userHelper.BuildProfileData(user, 3, 7)

		assert.Equal(t, "Test User", profileData.Name)
		assert.Equal(t, "Mahasiswa", profileData.Role)
		assert.Equal(t, 3, profileData.JumlahProject)
		assert.Equal(t, 7, profileData.JumlahModul)
	})
}

// Helper function
func stringPtr(s string) *string {
	return &s
}
