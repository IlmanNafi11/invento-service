package domain

import (
	"testing"
	"time"
)

func TestUserListQueryParams(t *testing.T) {
	t.Run("UserListQueryParams with all fields", func(t *testing.T) {
		params := UserListQueryParams{
			Search:     "john",
			FilterRole: "admin",
			Page:       1,
			Limit:      20,
		}

		if params.Search != "john" {
			t.Errorf("Expected Search 'john', got %s", params.Search)
		}
		if params.FilterRole != "admin" {
			t.Errorf("Expected FilterRole 'admin', got %s", params.FilterRole)
		}
		if params.Page != 1 {
			t.Errorf("Expected Page 1, got %d", params.Page)
		}
		if params.Limit != 20 {
			t.Errorf("Expected Limit 20, got %d", params.Limit)
		}
	})

	t.Run("UserListQueryParams with default values", func(t *testing.T) {
		params := UserListQueryParams{}

		if params.Search != "" {
			t.Errorf("Expected empty Search, got %s", params.Search)
		}
		if params.FilterRole != "" {
			t.Errorf("Expected empty FilterRole, got %s", params.FilterRole)
		}
		if params.Page != 0 {
			t.Errorf("Expected Page 0, got %d", params.Page)
		}
	})
}

func TestUserListItem(t *testing.T) {
	t.Run("UserListItem struct", func(t *testing.T) {
		now := time.Now()
		item := UserListItem{
			ID:         1,
			Email:      "user@example.com",
			Role:       "Admin",
			DibuatPada: now,
		}

		if item.ID != 1 {
			t.Errorf("Expected ID 1, got %d", item.ID)
		}
		if item.Email != "user@example.com" {
			t.Errorf("Expected Email 'user@example.com', got %s", item.Email)
		}
		if item.Role != "Admin" {
			t.Errorf("Expected Role 'Admin', got %s", item.Role)
		}
	})
}

func TestUserListData(t *testing.T) {
	t.Run("UserListData with pagination", func(t *testing.T) {
		items := []UserListItem{
			{ID: 1, Email: "user1@example.com", Role: "Admin"},
			{ID: 2, Email: "user2@example.com", Role: "User"},
		}

		data := UserListData{
			Items: items,
			Pagination: PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		if len(data.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(data.Items))
		}
		if data.Pagination.TotalItems != 2 {
			t.Errorf("Expected TotalItems 2, got %d", data.Pagination.TotalItems)
		}
	})

	t.Run("UserListData empty", func(t *testing.T) {
		data := UserListData{
			Items: []UserListItem{},
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

func TestUpdateUserRoleRequest(t *testing.T) {
	t.Run("UpdateUserRoleRequest struct", func(t *testing.T) {
		req := UpdateUserRoleRequest{
			Role: "editor",
		}

		if req.Role != "editor" {
			t.Errorf("Expected Role 'editor', got %s", req.Role)
		}
	})

	t.Run("UpdateUserRoleRequest with different roles", func(t *testing.T) {
		roles := []string{"admin", "editor", "viewer", "moderator"}

		for _, role := range roles {
			req := UpdateUserRoleRequest{Role: role}
			if req.Role != role {
				t.Errorf("Expected Role '%s', got %s", role, req.Role)
			}
		}
	})
}

func TestProfileData(t *testing.T) {
	t.Run("ProfileData with all fields", func(t *testing.T) {
		now := time.Now()
		jenisKelamin := "Laki-laki"
		fotoProfil := "https://example.com/photo.jpg"

		profile := ProfileData{
			Name:          "John Doe",
			Email:         "john@example.com",
			JenisKelamin:  &jenisKelamin,
			FotoProfil:    &fotoProfil,
			Role:          "Admin",
			CreatedAt:     now,
			JumlahProject: 10,
			JumlahModul:   25,
		}

		if profile.Name != "John Doe" {
			t.Errorf("Expected Name 'John Doe', got %s", profile.Name)
		}
		if profile.Email != "john@example.com" {
			t.Errorf("Expected Email 'john@example.com', got %s", profile.Email)
		}
		if profile.JenisKelamin == nil || *profile.JenisKelamin != "Laki-laki" {
			t.Error("Expected JenisKelamin to be 'Laki-laki'")
		}
		if profile.FotoProfil == nil || *profile.FotoProfil != "https://example.com/photo.jpg" {
			t.Error("Expected FotoProfil to be set")
		}
		if profile.Role != "Admin" {
			t.Errorf("Expected Role 'Admin', got %s", profile.Role)
		}
		if profile.JumlahProject != 10 {
			t.Errorf("Expected JumlahProject 10, got %d", profile.JumlahProject)
		}
		if profile.JumlahModul != 25 {
			t.Errorf("Expected JumlahModul 25, got %d", profile.JumlahModul)
		}
	})

	t.Run("ProfileData without optional fields", func(t *testing.T) {
		now := time.Now()
		profile := ProfileData{
			Name:          "Jane Doe",
			Email:         "jane@example.com",
			JenisKelamin:  nil,
			FotoProfil:    nil,
			Role:          "User",
			CreatedAt:     now,
			JumlahProject: 5,
			JumlahModul:   10,
		}

		if profile.JenisKelamin != nil {
			t.Error("Expected JenisKelamin to be nil")
		}
		if profile.FotoProfil != nil {
			t.Error("Expected FotoProfil to be nil")
		}
	})
}

func TestUpdateProfileRequest(t *testing.T) {
	t.Run("UpdateProfileRequest with all fields", func(t *testing.T) {
		req := UpdateProfileRequest{
			Name:         "Updated Name",
			JenisKelamin: "Laki-laki",
		}

		if req.Name != "Updated Name" {
			t.Errorf("Expected Name 'Updated Name', got %s", req.Name)
		}
		if req.JenisKelamin != "Laki-laki" {
			t.Errorf("Expected JenisKelamin 'Laki-laki', got %s", req.JenisKelamin)
		}
	})

	t.Run("UpdateProfileRequest with different jenis kelamin", func(t *testing.T) {
		validValues := []string{"Laki-laki", "Perempuan"}

		for _, jk := range validValues {
			req := UpdateProfileRequest{
				Name:         "Test User",
				JenisKelamin: jk,
			}

			if req.JenisKelamin != jk {
				t.Errorf("Expected JenisKelamin '%s', got %s", jk, req.JenisKelamin)
			}
		}
	})

	t.Run("UpdateProfileRequest with only name", func(t *testing.T) {
		req := UpdateProfileRequest{
			Name:         "Only Name",
			JenisKelamin: "",
		}

		if req.Name != "Only Name" {
			t.Errorf("Expected Name 'Only Name', got %s", req.Name)
		}
		if req.JenisKelamin != "" {
			t.Errorf("Expected empty JenisKelamin, got %s", req.JenisKelamin)
		}
	})
}

func TestUserFileItem(t *testing.T) {
	t.Run("UserFileItem struct", func(t *testing.T) {
		item := UserFileItem{
			ID:          1,
			NamaFile:    "project_report.pdf",
			Kategori:    "project",
			DownloadURL: "https://example.com/download/project_report.pdf",
		}

		if item.ID != 1 {
			t.Errorf("Expected ID 1, got %d", item.ID)
		}
		if item.NamaFile != "project_report.pdf" {
			t.Errorf("Expected NamaFile 'project_report.pdf', got %s", item.NamaFile)
		}
		if item.Kategori != "project" {
			t.Errorf("Expected Kategori 'project', got %s", item.Kategori)
		}
		if item.DownloadURL != "https://example.com/download/project_report.pdf" {
			t.Errorf("Expected DownloadURL 'https://example.com/download/project_report.pdf', got %s", item.DownloadURL)
		}
	})

	t.Run("UserFileItem with different kategori", func(t *testing.T) {
		kategories := []string{"project", "modul"}

		for _, kategori := range kategories {
			item := UserFileItem{
				ID:          1,
				NamaFile:    "test.pdf",
				Kategori:    kategori,
				DownloadURL: "https://example.com/download/test.pdf",
			}

			if item.Kategori != kategori {
				t.Errorf("Expected Kategori '%s', got %s", kategori, item.Kategori)
			}
		}
	})
}

func TestUserFilesQueryParams(t *testing.T) {
	t.Run("UserFilesQueryParams with all fields", func(t *testing.T) {
		params := UserFilesQueryParams{
			Search: "report",
			Page:   2,
			Limit:  15,
		}

		if params.Search != "report" {
			t.Errorf("Expected Search 'report', got %s", params.Search)
		}
		if params.Page != 2 {
			t.Errorf("Expected Page 2, got %d", params.Page)
		}
		if params.Limit != 15 {
			t.Errorf("Expected Limit 15, got %d", params.Limit)
		}
	})

	t.Run("UserFilesQueryParams with default values", func(t *testing.T) {
		params := UserFilesQueryParams{}

		if params.Search != "" {
			t.Errorf("Expected empty Search, got %s", params.Search)
		}
		if params.Page != 0 {
			t.Errorf("Expected Page 0, got %d", params.Page)
		}
	})
}

func TestUserFilesData(t *testing.T) {
	t.Run("UserFilesData with pagination", func(t *testing.T) {
		items := []UserFileItem{
			{ID: 1, NamaFile: "project1.pdf", Kategori: "project", DownloadURL: "url1"},
			{ID: 2, NamaFile: "modul1.pdf", Kategori: "modul", DownloadURL: "url2"},
		}

		data := UserFilesData{
			Items: items,
			Pagination: PaginationData{
				Page:       1,
				Limit:      10,
				TotalItems: 2,
				TotalPages: 1,
			},
		}

		if len(data.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(data.Items))
		}
		if data.Pagination.TotalItems != 2 {
			t.Errorf("Expected TotalItems 2, got %d", data.Pagination.TotalItems)
		}
	})
}

func TestUserPermissionItem(t *testing.T) {
	t.Run("UserPermissionItem struct", func(t *testing.T) {
		item := UserPermissionItem{
			Resource: "projects",
			Actions:  []string{"create", "read", "update", "delete"},
		}

		if item.Resource != "projects" {
			t.Errorf("Expected Resource 'projects', got %s", item.Resource)
		}
		if len(item.Actions) != 4 {
			t.Errorf("Expected 4 actions, got %d", len(item.Actions))
		}
	})

	t.Run("UserPermissionItem with different resources", func(t *testing.T) {
		resources := []string{"projects", "users", "roles", "moduls"}

		for _, resource := range resources {
			item := UserPermissionItem{
				Resource: resource,
				Actions:  []string{"read"},
			}

			if item.Resource != resource {
				t.Errorf("Expected Resource '%s', got %s", resource, item.Resource)
			}
		}
	})
}

func TestDownloadUserFilesRequest(t *testing.T) {
	t.Run("DownloadUserFilesRequest with project IDs", func(t *testing.T) {
		req := DownloadUserFilesRequest{
			ProjectIDs: []uint{1, 2, 3},
			ModulIDs:   []uint{4, 5},
		}

		if len(req.ProjectIDs) != 3 {
			t.Errorf("Expected 3 ProjectIDs, got %d", len(req.ProjectIDs))
		}
		if len(req.ModulIDs) != 2 {
			t.Errorf("Expected 2 ModulIDs, got %d", len(req.ModulIDs))
		}
	})

	t.Run("DownloadUserFilesRequest with only project IDs", func(t *testing.T) {
		req := DownloadUserFilesRequest{
			ProjectIDs: []uint{1, 2, 3, 4, 5},
			ModulIDs:   []uint{},
		}

		if len(req.ProjectIDs) != 5 {
			t.Errorf("Expected 5 ProjectIDs, got %d", len(req.ProjectIDs))
		}
		if len(req.ModulIDs) != 0 {
			t.Errorf("Expected 0 ModulIDs, got %d", len(req.ModulIDs))
		}
	})

	t.Run("DownloadUserFilesRequest with only modul IDs", func(t *testing.T) {
		req := DownloadUserFilesRequest{
			ProjectIDs: []uint{},
			ModulIDs:   []uint{10, 20, 30},
		}

		if len(req.ProjectIDs) != 0 {
			t.Errorf("Expected 0 ProjectIDs, got %d", len(req.ProjectIDs))
		}
		if len(req.ModulIDs) != 3 {
			t.Errorf("Expected 3 ModulIDs, got %d", len(req.ModulIDs))
		}
	})

	t.Run("DownloadUserFilesRequest empty", func(t *testing.T) {
		req := DownloadUserFilesRequest{
			ProjectIDs: []uint{},
			ModulIDs:   []uint{},
		}

		if len(req.ProjectIDs) != 0 {
			t.Errorf("Expected 0 ProjectIDs, got %d", len(req.ProjectIDs))
		}
		if len(req.ModulIDs) != 0 {
			t.Errorf("Expected 0 ModulIDs, got %d", len(req.ModulIDs))
		}
	})
}
