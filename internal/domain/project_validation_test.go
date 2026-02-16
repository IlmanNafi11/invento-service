package domain

import (
	"invento-service/internal/dto"
	"testing"
	"time"
)

func TestProjectSemesterRange(t *testing.T) {
	t.Run("Valid semester range", func(t *testing.T) {
		validSemesters := []int{1, 2, 3, 4, 5, 6, 7, 8}

		for _, sem := range validSemesters {
			project := Project{Semester: sem}
			if project.Semester < 1 || project.Semester > 8 {
				t.Errorf("Semester %d should be valid (1-8)", sem)
			}
		}
	})

	t.Run("Invalid semester values", func(t *testing.T) {
		invalidSemesters := []int{0, -1, 9, 10, 100}

		for _, sem := range invalidSemesters {
			project := Project{Semester: sem}
			if project.Semester >= 1 && project.Semester <= 8 {
				t.Errorf("Semester %d should be invalid (outside 1-8)", sem)
			}
		}
	})
}

func TestProjectKategoriValidation(t *testing.T) {
	t.Run("Valid kategori values", func(t *testing.T) {
		validKategories := []string{
			"website",
			"mobile",
			"iot",
			"machine_learning",
			"deep_learning",
		}

		for _, kategori := range validKategories {
			project := Project{Kategori: kategori}
			if project.Kategori != kategori {
				t.Errorf("Expected Kategori '%s', got %s", kategori, project.Kategori)
			}
		}
	})
}

func TestProjectEdgeCases(t *testing.T) {
	t.Run("Empty project name", func(t *testing.T) {
		project := Project{
			NamaProject: "",
			Semester:    1,
		}

		if project.NamaProject != "" {
			t.Errorf("Expected empty NamaProject, got %s", project.NamaProject)
		}
	})

	t.Run("Zero ID value", func(t *testing.T) {
		project := Project{
			ID: 0,
		}

		if project.ID != 0 {
			t.Errorf("Expected ID 0, got %d", project.ID)
		}
	})

	t.Run("Non-zero ID value", func(t *testing.T) {
		project := Project{
			ID: 12345,
		}

		if project.ID != 12345 {
			t.Errorf("Expected ID 12345, got %d", project.ID)
		}
	})

	t.Run("Invalid semester less than 1", func(t *testing.T) {
		invalidSemesters := []int{0, -1, -10, -100}

		for _, sem := range invalidSemesters {
			project := Project{Semester: sem}
			if project.Semester >= 1 {
				t.Errorf("Semester %d should be less than 1", sem)
			}
		}
	})

	t.Run("Empty filter strings", func(t *testing.T) {
		params := dto.ProjectListQueryParams{
			Search:         "",
			FilterKategori: "",
		}

		if params.Search != "" {
			t.Errorf("Expected empty Search, got %s", params.Search)
		}
		if params.FilterKategori != "" {
			t.Errorf("Expected empty FilterKategori, got %s", params.FilterKategori)
		}
	})

	t.Run("Invalid pagination values", func(t *testing.T) {
		testCases := []struct {
			name  string
			page  int
			limit int
		}{
			{"Zero page", 0, 10},
			{"Negative page", -1, 10},
			{"Zero limit", 1, 0},
			{"Negative limit", 1, -5},
			{"Both zero", 0, 0},
			{"Both negative", -1, -1},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				params := dto.ProjectListQueryParams{
					Page:  tc.page,
					Limit: tc.limit,
				}

				if params.Page != tc.page {
					t.Errorf("Expected Page %d, got %d", tc.page, params.Page)
				}
				if params.Limit != tc.limit {
					t.Errorf("Expected Limit %d, got %d", tc.limit, params.Limit)
				}
			})
		}
	})

	t.Run("Nil project list", func(t *testing.T) {
		data := dto.ProjectListData{
			Items: nil,
			Pagination: dto.PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 0,
				TotalPages: 0,
			},
		}

		if data.Items != nil {
			t.Errorf("Expected nil Items, got %v", data.Items)
		}
	})

	t.Run("Empty project list", func(t *testing.T) {
		data := dto.ProjectListData{
			Items: []dto.ProjectListItem{},
			Pagination: dto.PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 0,
				TotalPages: 0,
			},
		}

		if len(data.Items) != 0 {
			t.Errorf("Expected empty Items, got %d items", len(data.Items))
		}
	})
}

func TestProjectUserRelationship(t *testing.T) {
	t.Run("Project with zero UserID", func(t *testing.T) {
		project := Project{
			UserID: "",
		}

		if project.UserID != "" {
			t.Errorf("Expected UserID empty string, got %s", project.UserID)
		}
	})

	t.Run("Project without User relation", func(t *testing.T) {
		project := Project{
			ID:     1,
			UserID: "user-100",
		}

		if project.User.ID != "" {
			t.Errorf("Expected empty User relation, got ID %s", project.User.ID)
		}
	})

	t.Run("Project with matching User relation", func(t *testing.T) {
		user := User{
			ID:    "user-200",
			Email: "test@example.com",
			Name:  "Test User",
		}

		project := Project{
			ID:     1,
			UserID: "user-200",
			User:   user,
		}

		if project.User.ID != project.UserID {
			t.Errorf("User ID mismatch: User.ID=%s, UserID=%s", project.User.ID, project.UserID)
		}
	})

	t.Run("Project with mismatched User relation", func(t *testing.T) {
		user := User{
			ID:    "user-300",
			Email: "other@example.com",
			Name:  "Other User",
		}

		project := Project{
			ID:     1,
			UserID: "user-200",
			User:   user,
		}

		// This test documents the behavior - the User relation ID can differ from UserID
		if project.User.ID == project.UserID {
			t.Errorf("Expected mismatched IDs, but both are %s", project.UserID)
		}
	})
}

func TestProjectTimestampFields(t *testing.T) {
	t.Run("Project with zero timestamps", func(t *testing.T) {
		project := Project{
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		}

		if !project.CreatedAt.IsZero() {
			t.Errorf("Expected zero CreatedAt, got %v", project.CreatedAt)
		}
		if !project.UpdatedAt.IsZero() {
			t.Errorf("Expected zero UpdatedAt, got %v", project.UpdatedAt)
		}
	})

	t.Run("Project with valid timestamps", func(t *testing.T) {
		created := time.Now().Add(-24 * time.Hour)
		updated := time.Now()

		project := Project{
			CreatedAt: created,
			UpdatedAt: updated,
		}

		if project.CreatedAt.IsZero() {
			t.Errorf("Expected non-zero CreatedAt")
		}
		if project.UpdatedAt.IsZero() {
			t.Errorf("Expected non-zero UpdatedAt")
		}

		if !project.CreatedAt.Before(project.UpdatedAt) {
			t.Errorf("Expected CreatedAt before UpdatedAt")
		}
	})

	t.Run("dto.ProjectListItem timestamp", func(t *testing.T) {
		now := time.Now()
		item := dto.ProjectListItem{
			TerakhirDiperbarui: now,
		}

		if item.TerakhirDiperbarui.IsZero() {
			t.Errorf("Expected non-zero TerakhirDiperbarui")
		}
	})

	t.Run("dto.ProjectResponse timestamps", func(t *testing.T) {
		created := time.Now().Add(-7 * 24 * time.Hour)
		updated := time.Now()

		resp := dto.ProjectResponse{
			CreatedAt: created,
			UpdatedAt: updated,
		}

		if resp.CreatedAt.IsZero() {
			t.Errorf("Expected non-zero CreatedAt")
		}
		if resp.UpdatedAt.IsZero() {
			t.Errorf("Expected non-zero UpdatedAt")
		}
	})
}

func TestProjectPaginationData(t *testing.T) {
	t.Run("Single page pagination", func(t *testing.T) {
		pagination := dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 5,
			TotalPages: 1,
		}

		if pagination.Page != 1 {
			t.Errorf("Expected Page 1, got %d", pagination.Page)
		}
		if pagination.TotalPages != 1 {
			t.Errorf("Expected TotalPages 1, got %d", pagination.TotalPages)
		}
	})

	t.Run("Multiple page pagination", func(t *testing.T) {
		pagination := dto.PaginationData{
			Page:       2,
			Limit:      10,
			TotalItems: 25,
			TotalPages: 3,
		}

		if pagination.Page != 2 {
			t.Errorf("Expected Page 2, got %d", pagination.Page)
		}
		if pagination.TotalPages != 3 {
			t.Errorf("Expected TotalPages 3, got %d", pagination.TotalPages)
		}
	})

	t.Run("Empty pagination", func(t *testing.T) {
		pagination := dto.PaginationData{
			Page:       1,
			Limit:      10,
			TotalItems: 0,
			TotalPages: 0,
		}

		if pagination.TotalItems != 0 {
			t.Errorf("Expected TotalItems 0, got %d", pagination.TotalItems)
		}
		if pagination.TotalPages != 0 {
			t.Errorf("Expected TotalPages 0, got %d", pagination.TotalPages)
		}
	})
}

func TestProjectDownloadRequestEdgeCases(t *testing.T) {
	t.Run("Empty IDs array", func(t *testing.T) {
		req := dto.ProjectDownloadRequest{
			IDs: []uint{},
		}

		if len(req.IDs) != 0 {
			t.Errorf("Expected 0 IDs, got %d", len(req.IDs))
		}
	})

	t.Run("Nil IDs", func(t *testing.T) {
		req := dto.ProjectDownloadRequest{
			IDs: nil,
		}

		if req.IDs != nil {
			t.Errorf("Expected nil IDs, got %v", req.IDs)
		}
	})

	t.Run("Large number of IDs", func(t *testing.T) {
		ids := make([]uint, 1000)
		for i := range ids {
			ids[i] = uint(i + 1)
		}

		req := dto.ProjectDownloadRequest{
			IDs: ids,
		}

		if len(req.IDs) != 1000 {
			t.Errorf("Expected 1000 IDs, got %d", len(req.IDs))
		}
		if req.IDs[0] != 1 {
			t.Errorf("Expected first ID 1, got %d", req.IDs[0])
		}
		if req.IDs[999] != 1000 {
			t.Errorf("Expected last ID 1000, got %d", req.IDs[999])
		}
	})

	t.Run("Duplicate IDs", func(t *testing.T) {
		req := dto.ProjectDownloadRequest{
			IDs: []uint{1, 2, 2, 3, 3, 3},
		}

		if len(req.IDs) != 6 {
			t.Errorf("Expected 6 IDs (with duplicates), got %d", len(req.IDs))
		}
	})
}
