package domain

import (
	"testing"
	"time"
)

func TestModulStruct(t *testing.T) {
	t.Run("Modul struct initialization", func(t *testing.T) {
		now := time.Now()
		modul := Modul{
			ID:        "550e8400-e29b-41d4-a716-446655440000",
			UserID:    "user-100",
			Judul:     "Test Module",
			Deskripsi: "Test Description",
			FilePath:  "/uploads/modules/test_module.pdf",
			FileName:  "test_module.pdf",
			FileSize:  2621440,
			MimeType:  "application/pdf",
			Status:    "completed",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if modul.ID != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("Expected ID '550e8400-e29b-41d4-a716-446655440000', got %s", modul.ID)
		}
		if modul.UserID != "user-100" {
			t.Errorf("Expected UserID 'user-100', got %s", modul.UserID)
		}
		if modul.Judul != "Test Module" {
			t.Errorf("Expected Judul 'Test Module', got %s", modul.Judul)
		}
		if modul.MimeType != "application/pdf" {
			t.Errorf("Expected MimeType 'application/pdf', got %s", modul.MimeType)
		}
	})

	t.Run("Modul with User relation", func(t *testing.T) {
		user := User{
			ID:    "user-100",
			Email: "user@example.com",
			Name:  "Test User",
		}
		modul := Modul{
			ID:        "550e8400-e29b-41d4-a716-446655440001",
			UserID:    "user-100",
			Judul:     "Test PDF",
			Deskripsi: "Description",
			FilePath:  "/uploads/test.pdf",
			FileName:  "test.pdf",
			FileSize:  1048576,
			MimeType:  "application/pdf",
			Status:    "completed",
			User:      user,
		}

		if modul.User.ID != modul.UserID {
			t.Errorf("User ID mismatch: %s vs %s", modul.User.ID, modul.UserID)
		}
	})
}

func TestModulRequestStructs(t *testing.T) {
	t.Run("ModulUpdateRequest", func(t *testing.T) {
		req := ModulUpdateRequest{
			Judul:     "updated_module.pdf",
			Deskripsi: "Updated description",
		}

		if req.Judul != "updated_module.pdf" {
			t.Errorf("Expected Judul 'updated_module.pdf', got %s", req.Judul)
		}
		if req.Deskripsi != "Updated description" {
			t.Errorf("Expected Deskripsi 'Updated description', got %s", req.Deskripsi)
		}
	})

	t.Run("ModulUpdateRequest with partial data", func(t *testing.T) {
		req := ModulUpdateRequest{
			Deskripsi: "Only description updated",
		}

		if req.Judul != "" {
			t.Errorf("Expected empty Judul, got %s", req.Judul)
		}
		if req.Deskripsi != "Only description updated" {
			t.Errorf("Expected Deskripsi 'Only description updated', got %s", req.Deskripsi)
		}
	})
}

func TestModulListQueryParams(t *testing.T) {
	t.Run("ModulListQueryParams with all fields", func(t *testing.T) {
		params := ModulListQueryParams{
			Search:       "test",
			FilterType:   "application/pdf",
			FilterStatus: "completed",
			Page:         1,
			Limit:        20,
		}

		if params.Search != "test" {
			t.Errorf("Expected Search 'test', got %s", params.Search)
		}
		if params.FilterType != "application/pdf" {
			t.Errorf("Expected FilterType 'application/pdf', got %s", params.FilterType)
		}
		if params.FilterStatus != "completed" {
			t.Errorf("Expected FilterStatus 'completed', got %s", params.FilterStatus)
		}
		if params.Page != 1 {
			t.Errorf("Expected Page 1, got %d", params.Page)
		}
		if params.Limit != 20 {
			t.Errorf("Expected Limit 20, got %d", params.Limit)
		}
	})

	t.Run("ModulListQueryParams with default values", func(t *testing.T) {
		params := ModulListQueryParams{}

		if params.Search != "" {
			t.Errorf("Expected empty Search, got %s", params.Search)
		}
		if params.Page != 0 {
			t.Errorf("Expected Page 0, got %d", params.Page)
		}
	})
}

func TestModulResponseStructs(t *testing.T) {
	t.Run("ModulListItem", func(t *testing.T) {
		now := time.Now()
		item := ModulListItem{
			ID:                 "550e8400-e29b-41d4-a716-446655440002",
			Judul:              "Test Module",
			Deskripsi:          "Test Description",
			FileName:           "test.pdf",
			MimeType:           "application/pdf",
			FileSize:           1572864,
			Status:             "completed",
			TerakhirDiperbarui: now,
		}

		if item.ID != "550e8400-e29b-41d4-a716-446655440002" {
			t.Errorf("Expected ID '550e8400-e29b-41d4-a716-446655440002', got %s", item.ID)
		}
		if item.Judul != "Test Module" {
			t.Errorf("Expected Judul 'Test Module', got %s", item.Judul)
		}
	})

	t.Run("ModulListData with pagination", func(t *testing.T) {
		items := []ModulListItem{
			{ID: "550e8400-e29b-41d4-a716-446655440003", Judul: "test1.pdf", MimeType: "application/pdf", Status: "completed"},
			{ID: "550e8400-e29b-41d4-a716-446655440004", Judul: "test2.pdf", MimeType: "application/pdf", Status: "completed"},
		}
		data := ModulListData{
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

	t.Run("ModulResponse", func(t *testing.T) {
		now := time.Now()
		resp := ModulResponse{
			ID:        "550e8400-e29b-41d4-a716-446655440005",
			Judul:     "test.pdf",
			Deskripsi: "Test description",
			FileName:  "test.pdf",
			MimeType:  "application/pdf",
			FileSize:  1048576,
			Status:    "completed",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if resp.Judul != "test.pdf" {
			t.Errorf("Expected Judul 'test.pdf', got %s", resp.Judul)
		}
	})

}

func TestModulDownloadRequest(t *testing.T) {
	t.Run("ModulDownloadRequest with multiple IDs", func(t *testing.T) {
		req := ModulDownloadRequest{
			IDs: []string{"550e8400-e29b-41d4-a716-446655440001", "550e8400-e29b-41d4-a716-446655440002", "550e8400-e29b-41d4-a716-446655440003"},
		}

		if len(req.IDs) != 3 {
			t.Errorf("Expected 3 IDs, got %d", len(req.IDs))
		}
		if req.IDs[0] != "550e8400-e29b-41d4-a716-446655440001" {
			t.Errorf("Expected first ID '550e8400-e29b-41d4-a716-446655440001', got %s", req.IDs[0])
		}
	})

	t.Run("ModulDownloadRequest with single ID", func(t *testing.T) {
		req := ModulDownloadRequest{
			IDs: []string{"550e8400-e29b-41d4-a716-446655440100"},
		}

		if len(req.IDs) != 1 {
			t.Errorf("Expected 1 ID, got %d", len(req.IDs))
		}
	})
}

func TestModulStatus(t *testing.T) {
	validStatuses := []string{"pending", "completed", "failed"}

	for _, status := range validStatuses {
		modul := Modul{Status: status}
		if modul.Status == "" {
			t.Errorf("Status %s should be valid", status)
		}
	}
}
