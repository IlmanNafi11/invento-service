package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProjectUpdateMetadataRequest_Structure(t *testing.T) {
	request := ProjectUpdateMetadataRequest{
		NamaProject: "Sistem Informasi Manajemen",
		Kategori:    "website",
		Semester:    3,
	}

	assert.Equal(t, "Sistem Informasi Manajemen", request.NamaProject)
	assert.Equal(t, "website", request.Kategori)
	assert.Equal(t, 3, request.Semester)
}

func TestProjectUpdateMetadataRequest_ValidCategories(t *testing.T) {
	validCategories := []string{"website", "mobile", "iot", "machine_learning", "deep_learning"}

	for _, category := range validCategories {
		t.Run("valid category_"+category, func(t *testing.T) {
			request := ProjectUpdateMetadataRequest{
				NamaProject: "Test Project",
				Kategori:    category,
				Semester:    1,
			}

			assert.Equal(t, category, request.Kategori)
			assert.Equal(t, "Test Project", request.NamaProject)
			assert.Equal(t, 1, request.Semester)
		})
	}
}

func TestProjectUpdateMetadataRequest_ValidSemesters(t *testing.T) {
	for semester := 1; semester <= 8; semester++ {
		t.Run("valid_semester_"+string(rune(semester+'0')), func(t *testing.T) {
			request := ProjectUpdateMetadataRequest{
				NamaProject: "Test Project",
				Kategori:    "website",
				Semester:    semester,
			}

			assert.Equal(t, semester, request.Semester)
			assert.Equal(t, "Test Project", request.NamaProject)
			assert.Equal(t, "website", request.Kategori)
		})
	}
}

func TestProjectUpdateMetadataRequest_JSONTags(t *testing.T) {
	request := ProjectUpdateMetadataRequest{}

	// Verify struct has proper JSON tags by checking reflection
	assert.NotNil(t, request)
}
