package domain

import (
	"testing"
	"time"

	"invento-service/internal/dto"
)

func TestProjectStruct(t *testing.T) {
	t.Parallel()
	t.Run("Project struct initialization", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		project := Project{
			ID:          1,
			UserID:      "user-100",
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
		if project.UserID != "user-100" {
			t.Errorf("Expected UserID 'user-100', got %s", project.UserID)
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
		t.Parallel()
		user := User{
			ID:    "user-100",
			Email: "user@example.com",
			Name:  "Test User",
		}
		project := Project{
			ID:          1,
			UserID:      "user-100",
			NamaProject: "Mobile App",
			Kategori:    "mobile",
			Semester:    5,
			Ukuran:      "8.2 MB",
			PathFile:    "/uploads/projects/mobile.zip",
			User:        user,
		}

		if project.User.ID != project.UserID {
			t.Errorf("User ID mismatch: %s vs %s", project.User.ID, project.UserID)
		}
	})
}

func TestProjectCreateRequest(t *testing.T) {
	t.Parallel()
	t.Run("dto.CreateProjectRequest with valid data", func(t *testing.T) {
		t.Parallel()
		req := dto.CreateProjectRequest{
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

	t.Run("dto.CreateProjectRequest with minimum semester", func(t *testing.T) {
		t.Parallel()
		req := dto.CreateProjectRequest{
			NamaProject: "IoT Dashboard",
			Semester:    1,
		}

		if req.Semester != 1 {
			t.Errorf("Expected Semester 1, got %d", req.Semester)
		}
	})

	t.Run("dto.CreateProjectRequest with maximum semester", func(t *testing.T) {
		t.Parallel()
		req := dto.CreateProjectRequest{
			NamaProject: "Deep Learning Model",
			Semester:    8,
		}

		if req.Semester != 8 {
			t.Errorf("Expected Semester 8, got %d", req.Semester)
		}
	})
}

func TestProjectUpdateRequest(t *testing.T) {
	t.Parallel()
	t.Run("dto.UpdateProjectRequest with all fields", func(t *testing.T) {
		t.Parallel()
		req := dto.UpdateProjectRequest{
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

	t.Run("dto.UpdateProjectRequest with partial data", func(t *testing.T) {
		t.Parallel()
		req := dto.UpdateProjectRequest{
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

	t.Run("dto.UpdateProjectRequest with valid kategori values", func(t *testing.T) {
		t.Parallel()
		validKategories := []string{"website", "mobile", "iot", "machine_learning", "deep_learning"}

		for _, kategori := range validKategories {
			req := dto.UpdateProjectRequest{
				Kategori: kategori,
			}

			if req.Kategori != kategori {
				t.Errorf("Expected Kategori '%s', got %s", kategori, req.Kategori)
			}
		}
	})
}

func TestProjectListQueryParams(t *testing.T) {
	t.Parallel()
	t.Run("dto.ProjectListQueryParams with all fields", func(t *testing.T) {
		t.Parallel()
		params := dto.ProjectListQueryParams{
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

	t.Run("dto.ProjectListQueryParams with default values", func(t *testing.T) {
		t.Parallel()
		params := dto.ProjectListQueryParams{}

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
	t.Parallel()
	t.Run("dto.ProjectListItem struct", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		item := dto.ProjectListItem{
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
	t.Parallel()
	t.Run("dto.ProjectListData with pagination", func(t *testing.T) {
		t.Parallel()
		items := []dto.ProjectListItem{
			{ID: 1, NamaProject: "Project A", Kategori: "website", Semester: 1},
			{ID: 2, NamaProject: "Project B", Kategori: "mobile", Semester: 2},
			{ID: 3, NamaProject: "Project C", Kategori: "iot", Semester: 3},
		}

		data := dto.ProjectListData{
			Items: items,
			Pagination: dto.PaginationData{
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

	t.Run("dto.ProjectListData empty", func(t *testing.T) {
		t.Parallel()
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
			t.Errorf("Expected 0 items, got %d", len(data.Items))
		}
	})
}

func TestProjectResponse(t *testing.T) {
	t.Parallel()
	t.Run("dto.ProjectResponse struct", func(t *testing.T) {
		t.Parallel()
		now := time.Now()
		resp := dto.ProjectResponse{
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

func TestProjectDownloadRequest(t *testing.T) {
	t.Parallel()
	t.Run("dto.ProjectDownloadRequest with multiple IDs", func(t *testing.T) {
		t.Parallel()
		req := dto.ProjectDownloadRequest{
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

	t.Run("dto.ProjectDownloadRequest with single ID", func(t *testing.T) {
		t.Parallel()
		req := dto.ProjectDownloadRequest{
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

func TestProjectUpdateMetadataRequest(t *testing.T) {
	t.Parallel()
	t.Run("dto.UpdateProjectRequest struct", func(t *testing.T) {
		t.Parallel()
		req := dto.UpdateProjectRequest{
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
