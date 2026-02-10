package domain

import (
	"testing"
	"time"
)

func TestProjectStruct(t *testing.T) {
	t.Run("Project struct initialization", func(t *testing.T) {
		now := time.Now()
		project := Project{
			ID:          1,
			UserID:      100,
			NamaProject: "E-Commerce Platform",
			Kategori:    "website",
			Semester:    3,
			Ukuran:      "15.5 MB",
			PathFile:    "/uploads/projects/ecommerce.zip",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if project.ID != 1 {
			t.Errorf("Expected ID 1, got %d", project.ID)
		}
		if project.UserID != 100 {
			t.Errorf("Expected UserID 100, got %d", project.UserID)
		}
		if project.NamaProject != "E-Commerce Platform" {
			t.Errorf("Expected NamaProject 'E-Commerce Platform', got %s", project.NamaProject)
		}
		if project.Kategori != "website" {
			t.Errorf("Expected Kategori 'website', got %s", project.Kategori)
		}
		if project.Semester != 3 {
			t.Errorf("Expected Semester 3, got %d", project.Semester)
		}
	})

	t.Run("Project with User relation", func(t *testing.T) {
		user := User{
			ID:    100,
			Email: "user@example.com",
			Name:  "Test User",
		}
		project := Project{
			ID:          1,
			UserID:      100,
			NamaProject: "Mobile App",
			Kategori:    "mobile",
			Semester:    5,
			Ukuran:      "8.2 MB",
			PathFile:    "/uploads/projects/mobile.zip",
			User:        user,
		}

		if project.User.ID != project.UserID {
			t.Errorf("User ID mismatch: %d vs %d", project.User.ID, project.UserID)
		}
	})
}

func TestProjectCreateRequest(t *testing.T) {
	t.Run("ProjectCreateRequest with valid data", func(t *testing.T) {
		req := ProjectCreateRequest{
			NamaProject: "AI Chatbot",
			Semester:    7,
		}

		if req.NamaProject != "AI Chatbot" {
			t.Errorf("Expected NamaProject 'AI Chatbot', got %s", req.NamaProject)
		}
		if req.Semester != 7 {
			t.Errorf("Expected Semester 7, got %d", req.Semester)
		}
	})

	t.Run("ProjectCreateRequest with minimum semester", func(t *testing.T) {
		req := ProjectCreateRequest{
			NamaProject: "IoT Dashboard",
			Semester:    1,
		}

		if req.Semester != 1 {
			t.Errorf("Expected Semester 1, got %d", req.Semester)
		}
	})

	t.Run("ProjectCreateRequest with maximum semester", func(t *testing.T) {
		req := ProjectCreateRequest{
			NamaProject: "Deep Learning Model",
			Semester:    8,
		}

		if req.Semester != 8 {
			t.Errorf("Expected Semester 8, got %d", req.Semester)
		}
	})
}

func TestProjectUpdateRequest(t *testing.T) {
	t.Run("ProjectUpdateRequest with all fields", func(t *testing.T) {
		req := ProjectUpdateRequest{
			NamaProject: "Updated Project Name",
			Kategori:    "machine_learning",
			Semester:    6,
		}

		if req.NamaProject != "Updated Project Name" {
			t.Errorf("Expected NamaProject 'Updated Project Name', got %s", req.NamaProject)
		}
		if req.Kategori != "machine_learning" {
			t.Errorf("Expected Kategori 'machine_learning', got %s", req.Kategori)
		}
		if req.Semester != 6 {
			t.Errorf("Expected Semester 6, got %d", req.Semester)
		}
	})

	t.Run("ProjectUpdateRequest with partial data", func(t *testing.T) {
		req := ProjectUpdateRequest{
			Semester: 4,
		}

		if req.NamaProject != "" {
			t.Errorf("Expected empty NamaProject, got %s", req.NamaProject)
		}
		if req.Kategori != "" {
			t.Errorf("Expected empty Kategori, got %s", req.Kategori)
		}
		if req.Semester != 4 {
			t.Errorf("Expected Semester 4, got %d", req.Semester)
		}
	})

	t.Run("ProjectUpdateRequest with valid kategori values", func(t *testing.T) {
		validKategories := []string{"website", "mobile", "iot", "machine_learning", "deep_learning"}

		for _, kategori := range validKategories {
			req := ProjectUpdateRequest{
				Kategori: kategori,
			}

			if req.Kategori != kategori {
				t.Errorf("Expected Kategori '%s', got %s", kategori, req.Kategori)
			}
		}
	})
}

func TestProjectListQueryParams(t *testing.T) {
	t.Run("ProjectListQueryParams with all fields", func(t *testing.T) {
		params := ProjectListQueryParams{
			Search:         "ecommerce",
			FilterSemester: 3,
			FilterKategori: "website",
			Page:           1,
			Limit:          20,
		}

		if params.Search != "ecommerce" {
			t.Errorf("Expected Search 'ecommerce', got %s", params.Search)
		}
		if params.FilterSemester != 3 {
			t.Errorf("Expected FilterSemester 3, got %d", params.FilterSemester)
		}
		if params.FilterKategori != "website" {
			t.Errorf("Expected FilterKategori 'website', got %s", params.FilterKategori)
		}
		if params.Page != 1 {
			t.Errorf("Expected Page 1, got %d", params.Page)
		}
		if params.Limit != 20 {
			t.Errorf("Expected Limit 20, got %d", params.Limit)
		}
	})

	t.Run("ProjectListQueryParams with default values", func(t *testing.T) {
		params := ProjectListQueryParams{}

		if params.Search != "" {
			t.Errorf("Expected empty Search, got %s", params.Search)
		}
		if params.FilterSemester != 0 {
			t.Errorf("Expected FilterSemester 0, got %d", params.FilterSemester)
		}
		if params.Page != 0 {
			t.Errorf("Expected Page 0, got %d", params.Page)
		}
	})
}

func TestProjectListItem(t *testing.T) {
	t.Run("ProjectListItem struct", func(t *testing.T) {
		now := time.Now()
		item := ProjectListItem{
			ID:                 1,
			NamaProject:        "Smart Home System",
			Kategori:           "iot",
			Semester:           4,
			Ukuran:             "12.3 MB",
			PathFile:           "/uploads/smarthome.zip",
			TerakhirDiperbarui: now,
		}

		if item.ID != 1 {
			t.Errorf("Expected ID 1, got %d", item.ID)
		}
		if item.NamaProject != "Smart Home System" {
			t.Errorf("Expected NamaProject 'Smart Home System', got %s", item.NamaProject)
		}
		if item.Kategori != "iot" {
			t.Errorf("Expected Kategori 'iot', got %s", item.Kategori)
		}
		if item.Semester != 4 {
			t.Errorf("Expected Semester 4, got %d", item.Semester)
		}
	})
}

func TestProjectListData(t *testing.T) {
	t.Run("ProjectListData with pagination", func(t *testing.T) {
		items := []ProjectListItem{
			{ID: 1, NamaProject: "Project A", Kategori: "website", Semester: 1},
			{ID: 2, NamaProject: "Project B", Kategori: "mobile", Semester: 2},
			{ID: 3, NamaProject: "Project C", Kategori: "iot", Semester: 3},
		}

		data := ProjectListData{
			Items: items,
			Pagination: PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 3,
				TotalPages: 1,
			},
		}

		if len(data.Items) != 3 {
			t.Errorf("Expected 3 items, got %d", len(data.Items))
		}
		if data.Pagination.TotalItems != 3 {
			t.Errorf("Expected TotalItems 3, got %d", data.Pagination.TotalItems)
		}
	})

	t.Run("ProjectListData empty", func(t *testing.T) {
		data := ProjectListData{
			Items: []ProjectListItem{},
			Pagination: PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 0,
				TotalPages: 0,
			},
		}

		if len(data.Items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(data.Items))
		}
	})
}

func TestProjectResponse(t *testing.T) {
	t.Run("ProjectResponse struct", func(t *testing.T) {
		now := time.Now()
		resp := ProjectResponse{
			ID:          1,
			NamaProject: "Data Analytics Platform",
			Kategori:    "machine_learning",
			Semester:    6,
			Ukuran:      "25.7 MB",
			PathFile:    "/uploads/analytics.zip",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if resp.NamaProject != "Data Analytics Platform" {
			t.Errorf("Expected NamaProject 'Data Analytics Platform', got %s", resp.NamaProject)
		}
		if resp.Kategori != "machine_learning" {
			t.Errorf("Expected Kategori 'machine_learning', got %s", resp.Kategori)
		}
		if resp.Semester != 6 {
			t.Errorf("Expected Semester 6, got %d", resp.Semester)
		}
	})
}

func TestProjectCreateResponse(t *testing.T) {
	t.Run("ProjectCreateResponse with items", func(t *testing.T) {
		items := []ProjectResponse{
			{ID: 1, NamaProject: "Project A", Kategori: "website", Semester: 1},
			{ID: 2, NamaProject: "Project B", Kategori: "mobile", Semester: 2},
		}

		resp := ProjectCreateResponse{
			Items: items,
		}

		if len(resp.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(resp.Items))
		}
	})

	t.Run("ProjectCreateResponse with single item", func(t *testing.T) {
		items := []ProjectResponse{
			{ID: 1, NamaProject: "Single Project", Kategori: "iot", Semester: 5},
		}

		resp := ProjectCreateResponse{
			Items: items,
		}

		if len(resp.Items) != 1 {
			t.Errorf("Expected 1 item, got %d", len(resp.Items))
		}
		if resp.Items[0].NamaProject != "Single Project" {
			t.Errorf("Expected NamaProject 'Single Project', got %s", resp.Items[0].NamaProject)
		}
	})
}

func TestProjectDownloadRequest(t *testing.T) {
	t.Run("ProjectDownloadRequest with multiple IDs", func(t *testing.T) {
		req := ProjectDownloadRequest{
			IDs: []uint{1, 2, 3, 4, 5},
		}

		if len(req.IDs) != 5 {
			t.Errorf("Expected 5 IDs, got %d", len(req.IDs))
		}
		if req.IDs[0] != 1 {
			t.Errorf("Expected first ID 1, got %d", req.IDs[0])
		}
		if req.IDs[4] != 5 {
			t.Errorf("Expected last ID 5, got %d", req.IDs[4])
		}
	})

	t.Run("ProjectDownloadRequest with single ID", func(t *testing.T) {
		req := ProjectDownloadRequest{
			IDs: []uint{100},
		}

		if len(req.IDs) != 1 {
			t.Errorf("Expected 1 ID, got %d", len(req.IDs))
		}
		if req.IDs[0] != 100 {
			t.Errorf("Expected ID 100, got %d", req.IDs[0])
		}
	})
}

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

func TestProjectUpdateMetadataRequest(t *testing.T) {
	t.Run("ProjectUpdateRequest struct", func(t *testing.T) {
		req := ProjectUpdateRequest{
			NamaProject: "Updated Metadata Project",
			Kategori:    "deep_learning",
			Semester:    8,
		}

		if req.NamaProject != "Updated Metadata Project" {
			t.Errorf("Expected NamaProject 'Updated Metadata Project', got %s", req.NamaProject)
		}
		if req.Kategori != "deep_learning" {
			t.Errorf("Expected Kategori 'deep_learning', got %s", req.Kategori)
		}
		if req.Semester != 8 {
			t.Errorf("Expected Semester 8, got %d", req.Semester)
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
		params := ProjectListQueryParams{
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
				params := ProjectListQueryParams{
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
		data := ProjectListData{
			Items: nil,
			Pagination: PaginationData{
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
		data := ProjectListData{
			Items: []ProjectListItem{},
			Pagination: PaginationData{
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
			UserID: 0,
		}

		if project.UserID != 0 {
			t.Errorf("Expected UserID 0, got %d", project.UserID)
		}
	})

	t.Run("Project without User relation", func(t *testing.T) {
		project := Project{
			ID:     1,
			UserID: 100,
		}

		if project.User.ID != 0 {
			t.Errorf("Expected empty User relation, got ID %d", project.User.ID)
		}
	})

	t.Run("Project with matching User relation", func(t *testing.T) {
		user := User{
			ID:    200,
			Email: "test@example.com",
			Name:  "Test User",
		}

		project := Project{
			ID:     1,
			UserID: 200,
			User:   user,
		}

		if project.User.ID != project.UserID {
			t.Errorf("User ID mismatch: User.ID=%d, UserID=%d", project.User.ID, project.UserID)
		}
	})

	t.Run("Project with mismatched User relation", func(t *testing.T) {
		user := User{
			ID:    300,
			Email: "other@example.com",
			Name:  "Other User",
		}

		project := Project{
			ID:     1,
			UserID: 200,
			User:   user,
		}

		// This test documents the behavior - the User relation ID can differ from UserID
		if project.User.ID == project.UserID {
			t.Errorf("Expected mismatched IDs, but both are %d", project.UserID)
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

	t.Run("ProjectListItem timestamp", func(t *testing.T) {
		now := time.Now()
		item := ProjectListItem{
			TerakhirDiperbarui: now,
		}

		if item.TerakhirDiperbarui.IsZero() {
			t.Errorf("Expected non-zero TerakhirDiperbarui")
		}
	})

	t.Run("ProjectResponse timestamps", func(t *testing.T) {
		created := time.Now().Add(-7 * 24 * time.Hour)
		updated := time.Now()

		resp := ProjectResponse{
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
		pagination := PaginationData{
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
		pagination := PaginationData{
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
		pagination := PaginationData{
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

func TestProjectCreateResponseEdgeCases(t *testing.T) {
	t.Run("Empty items array", func(t *testing.T) {
		resp := ProjectCreateResponse{
			Items: []ProjectResponse{},
		}

		if len(resp.Items) != 0 {
			t.Errorf("Expected 0 items, got %d", len(resp.Items))
		}
	})

	t.Run("Nil items", func(t *testing.T) {
		resp := ProjectCreateResponse{
			Items: nil,
		}

		if resp.Items != nil {
			t.Errorf("Expected nil Items, got %v", resp.Items)
		}
	})
}

func TestProjectDownloadRequestEdgeCases(t *testing.T) {
	t.Run("Empty IDs array", func(t *testing.T) {
		req := ProjectDownloadRequest{
			IDs: []uint{},
		}

		if len(req.IDs) != 0 {
			t.Errorf("Expected 0 IDs, got %d", len(req.IDs))
		}
	})

	t.Run("Nil IDs", func(t *testing.T) {
		req := ProjectDownloadRequest{
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

		req := ProjectDownloadRequest{
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
		req := ProjectDownloadRequest{
			IDs: []uint{1, 2, 2, 3, 3, 3},
		}

		if len(req.IDs) != 6 {
			t.Errorf("Expected 6 IDs (with duplicates), got %d", len(req.IDs))
		}
	})
}
