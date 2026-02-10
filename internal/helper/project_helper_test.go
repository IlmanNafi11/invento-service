package helper_test

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/helper"
	"fmt"
	"mime/multipart"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewProjectHelper(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
			PathProduction:  "/var/uploads",
		},
	}

	projectHelper := helper.NewProjectHelper(cfg)

	assert.NotNil(t, projectHelper)
}

func TestProjectHelper_GenerateProjectIdentifier(t *testing.T) {
	cfg := &config.Config{}
	projectHelper := helper.NewProjectHelper(cfg)

	t.Run("Generate unique identifiers", func(t *testing.T) {
		identifiers := make(map[string]bool)

		// Generate 100 identifiers and check for uniqueness
		for i := 0; i < 100; i++ {
			id, err := projectHelper.GenerateProjectIdentifier()
			assert.NoError(t, err)
			assert.Len(t, id, 10)
			assert.False(t, identifiers[id], "Identifier should be unique: "+id)
			identifiers[id] = true
		}
	})

	t.Run("Identifier format", func(t *testing.T) {
		id, err := projectHelper.GenerateProjectIdentifier()
		assert.NoError(t, err)
		assert.Len(t, id, 10)

		// Should only contain lowercase letters and numbers
		for _, char := range id {
			assert.True(t, (char >= 'a' && char <= 'z') || (char >= '0' && char <= '9'))
		}
	})
}

func TestProjectHelper_BuildProjectPath(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
			PathProduction:  "/var/uploads",
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	tests := []struct {
		name       string
		userID     uint
		identifier string
		filename   string
		wantPrefix string
	}{
		{
			name:       "Development path",
			userID:     123,
			identifier: "abc123def4",
			filename:   "project.zip",
			wantPrefix: "/tmp/uploads",
		},
		{
			name:       "Different user ID",
			userID:     456,
			identifier: "xyz789ghi0",
			filename:   "test.zip",
			wantPrefix: "/tmp/uploads",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := projectHelper.BuildProjectPath(tt.userID, tt.identifier, tt.filename)

			assert.Contains(t, result, tt.wantPrefix)
			assert.Contains(t, result, "projects")
			assert.Contains(t, result, fmt.Sprintf("%d", tt.userID))
			assert.Contains(t, result, tt.identifier)
			assert.Contains(t, result, tt.filename)
		})
	}
}

func TestProjectHelper_BuildProjectDirectory(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	userID := uint(123)
	identifier := "testidentifier"
	result := projectHelper.BuildProjectDirectory(userID, identifier)

	assert.Contains(t, result, "projects")
	assert.Contains(t, result, fmt.Sprintf("%d", userID))
	assert.Contains(t, result, identifier)
}

func TestProjectHelper_ValidateProjectFile(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxSize: 10 * 1024 * 1024, // 10MB
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	t.Run("Valid ZIP file", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "project.zip",
			Size:     1024 * 1024, // 1MB
		}

		err := projectHelper.ValidateProjectFile(fileHeader)
		assert.NoError(t, err)
	})

	t.Run("Invalid extension - not ZIP", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "project.pdf",
			Size:     1024,
		}

		err := projectHelper.ValidateProjectFile(fileHeader)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "zip")
	})

	t.Run("File too large", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "large.zip",
			Size:     15 * 1024 * 1024, // 15MB
		}

		err := projectHelper.ValidateProjectFile(fileHeader)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "MB")
	})

	t.Run("Uppercase extension", func(t *testing.T) {
		fileHeader := &multipart.FileHeader{
			Filename: "PROJECT.ZIP",
			Size:     1024,
		}

		err := projectHelper.ValidateProjectFile(fileHeader)
		assert.NoError(t, err) // GetFileExtension converts to lowercase
	})
}

func TestProjectHelper_ValidateProjectFileSize(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxSize: 50 * 1024 * 1024, // 50MB
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	tests := []struct {
		name      string
		fileSize  int64
		wantErr   bool
		errMsg    string
	}{
		{
			name:     "Valid file size",
			fileSize: 5 * 1024 * 1024, // 5MB
			wantErr:  false,
		},
		{
			name:     "Zero file size",
			fileSize: 0,
			wantErr:  true,
			errMsg:   "ukuran file tidak valid",
		},
		{
			name:     "Negative file size",
			fileSize: -1,
			wantErr:  true,
			errMsg:   "ukuran file tidak valid",
		},
		{
			name:     "File too large",
			fileSize: 100 * 1024 * 1024, // 100MB
			wantErr:  true,
			errMsg:   "melebihi batas maksimal",
		},
		{
			name:     "Exactly at limit",
			fileSize: 50 * 1024 * 1024, // 50MB
			wantErr:  false,
		},
		{
			name:     "One byte over limit",
			fileSize: 50*1024*1024 + 1,
			wantErr:  true,
			errMsg:   "melebihi batas maksimal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := projectHelper.ValidateProjectFileSize(tt.fileSize)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateProjectZipExtension(t *testing.T) {
	tests := []struct {
		name      string
		filename  string
		wantErr   bool
	}{
		{
			name:     "Valid lowercase ZIP",
			filename: "project.zip",
			wantErr:  false,
		},
		{
			name:     "Valid uppercase ZIP",
			filename: "PROJECT.ZIP",
			wantErr:  false,
		},
		{
			name:     "Mixed case ZIP",
			filename: "Project.Zip",
			wantErr:  false,
		},
		{
			name:     "PDF file",
			filename: "document.pdf",
			wantErr:  true,
		},
		{
			name:     "No extension",
			filename: "project",
			wantErr:  true,
		},
		{
			name:     "Different extension",
			filename: "project.rar",
			wantErr:  true,
		},
		{
			name:     "ZIP with path",
			filename: "/path/to/project.zip",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := helper.ValidateProjectZipExtension(tt.filename)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "zip")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProjectHelper_ProductionVsDevelopment(t *testing.T) {
	t.Run("Development environment", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "development",
			},
			Upload: config.UploadConfig{
				PathDevelopment: "/dev/uploads",
				PathProduction:  "/prod/uploads",
			},
		}
		projectHelper := helper.NewProjectHelper(cfg)

		path := projectHelper.BuildProjectPath(1, "abc", "test.zip")
		assert.Contains(t, path, "/dev/uploads")
		assert.NotContains(t, path, "/prod/uploads")
	})

	t.Run("Production environment", func(t *testing.T) {
		cfg := &config.Config{
			App: config.AppConfig{
				Env: "production",
			},
			Upload: config.UploadConfig{
				PathDevelopment: "/dev/uploads",
				PathProduction:  "/prod/uploads",
			},
		}
		projectHelper := helper.NewProjectHelper(cfg)

		path := projectHelper.BuildProjectPath(1, "abc", "test.zip")
		assert.Contains(t, path, "/prod/uploads")
		assert.NotContains(t, path, "/dev/uploads")
	})
}

func TestProjectHelper_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	tempDir := t.TempDir()

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment: tempDir,
			MaxSize:         10 * 1024 * 1024, // 10MB
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	t.Run("Complete project workflow", func(t *testing.T) {
		// Generate identifier
		identifier, err := projectHelper.GenerateProjectIdentifier()
		assert.NoError(t, err)
		assert.Len(t, identifier, 10)

		// Build directory path
		userID := uint(123)
		projectDir := projectHelper.BuildProjectDirectory(userID, identifier)

		// Create directory
		err = os.MkdirAll(projectDir, 0755)
		assert.NoError(t, err)
		assert.DirExists(t, projectDir)

		// Build file path
		filename := "myproject.zip"
		projectPath := projectHelper.BuildProjectPath(userID, identifier, filename)

		assert.Contains(t, projectPath, fmt.Sprintf("%d", userID))
		assert.Contains(t, projectPath, identifier)
		assert.Contains(t, projectPath, filename)

		// Create a test zip file
		err = os.WriteFile(projectPath, []byte("test zip content"), 0644)
		assert.NoError(t, err)
		assert.FileExists(t, projectPath)

		// Validate file size
		fileInfo, _ := os.Stat(projectPath)
		err = projectHelper.ValidateProjectFileSize(fileInfo.Size())
		assert.NoError(t, err)
	})

	t.Run("Multiple project directories", func(t *testing.T) {
		userID := uint(456)

		// Create multiple project directories
		for i := 0; i < 5; i++ {
			identifier, err := projectHelper.GenerateProjectIdentifier()
			assert.NoError(t, err)

			projectDir := projectHelper.BuildProjectDirectory(userID, identifier)
			err = os.MkdirAll(projectDir, 0755)
			assert.NoError(t, err)

			// Verify directory exists and is unique
			assert.DirExists(t, projectDir)
		}
	})
}

func TestProjectHelper_EdgeCases(t *testing.T) {
	cfg := &config.Config{
		Upload: config.UploadConfig{
			PathDevelopment: "/tmp/uploads",
			MaxSize:         100 * 1024 * 1024, // 100MB
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	t.Run("Special characters in filename", func(t *testing.T) {
		filename := "project (2024) [v1.0].zip"
		userID := uint(1)
		identifier := "abc123"

		path := projectHelper.BuildProjectPath(userID, identifier, filename)

		assert.Contains(t, path, filename)
		assert.Contains(t, path, filepath.Join("projects", "1", identifier))
	})

	t.Run("Very long filename", func(t *testing.T) {
		longFilename := "this_is_a_very_long_project_name_with_many_characters_and_numbers_123456789.zip"
		userID := uint(999)
		identifier := "xyz789"

		path := projectHelper.BuildProjectPath(userID, identifier, longFilename)

		assert.Contains(t, path, longFilename)
	})

	t.Run("User ID zero", func(t *testing.T) {
		userID := uint(0)
		identifier := "test123"
		filename := "project.zip"

		path := projectHelper.BuildProjectPath(userID, identifier, filename)

		assert.Contains(t, path, "0")
	})

	t.Run("Maximum file size validation", func(t *testing.T) {
		maxSize := int64(100 * 1024 * 1024) // 100MB

		err := projectHelper.ValidateProjectFileSize(maxSize)
		assert.NoError(t, err, "Should accept file exactly at max size")

		err = projectHelper.ValidateProjectFileSize(maxSize + 1)
		assert.Error(t, err, "Should reject file over max size")
	})
}

func TestProjectHelper_FileOperations(t *testing.T) {
	tempDir := t.TempDir()

	cfg := &config.Config{
		App: config.AppConfig{
			Env: "development",
		},
		Upload: config.UploadConfig{
			PathDevelopment: tempDir,
			MaxSize:         5 * 1024 * 1024, // 5MB
		},
	}
	projectHelper := helper.NewProjectHelper(cfg)

	t.Run("Create and validate project file", func(t *testing.T) {
		userID := uint(123)
		identifier, err := projectHelper.GenerateProjectIdentifier()
		require.NoError(t, err)

		projectDir := projectHelper.BuildProjectDirectory(userID, identifier)
		err = os.MkdirAll(projectDir, 0755)
		require.NoError(t, err)

		// Create a small zip file
		zipPath := filepath.Join(projectDir, "testproject.zip")
		err = os.WriteFile(zipPath, []byte("PK\x03\x04"), 0644) // ZIP header
		require.NoError(t, err)

		// Create file header for validation
		fileInfo, _ := os.Stat(zipPath)
		fileHeader := &multipart.FileHeader{
			Filename: "testproject.zip",
			Size:     fileInfo.Size(),
		}

		err = projectHelper.ValidateProjectFile(fileHeader)
		assert.NoError(t, err)
	})
}
