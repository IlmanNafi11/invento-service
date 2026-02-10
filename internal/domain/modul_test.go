package domain

import (
	"testing"
	"time"
)

func TestModulStruct(t *testing.T) {
	t.Run("Modul struct initialization", func(t *testing.T) {
		now := time.Now()
		modul := Modul{
			ID:       1,
			UserID:   100,
			NamaFile: "test_module.pdf",
			Tipe:     "pdf",
			Ukuran:   "2.5 MB",
			Semester: 3,
			PathFile: "/uploads/modules/test_module.pdf",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if modul.ID != 1 {
			t.Errorf("Expected ID 1, got %d", modul.ID)
		}
		if modul.UserID != 100 {
			t.Errorf("Expected UserID 100, got %d", modul.UserID)
		}
		if modul.NamaFile != "test_module.pdf" {
			t.Errorf("Expected NamaFile 'test_module.pdf', got %s", modul.NamaFile)
		}
		if modul.Tipe != "pdf" {
			t.Errorf("Expected Tipe 'pdf', got %s", modul.Tipe)
		}
		if modul.Semester != 3 {
			t.Errorf("Expected Semester 3, got %d", modul.Semester)
		}
	})

	t.Run("Modul with User relation", func(t *testing.T) {
		user := User{
			ID:    100,
			Email: "user@example.com",
			Name:  "Test User",
		}
		modul := Modul{
			ID:       1,
			UserID:   100,
			NamaFile: "test.pdf",
			Tipe:     "pdf",
			Ukuran:   "1 MB",
			Semester: 1,
			PathFile: "/uploads/test.pdf",
			User:     user,
		}

		if modul.User.ID != modul.UserID {
			t.Errorf("User ID mismatch: %d vs %d", modul.User.ID, modul.UserID)
		}
	})
}

func TestModulRequestStructs(t *testing.T) {
	t.Run("ModulCreateRequest", func(t *testing.T) {
		req := ModulCreateRequest{
			NamaFile: "test_module.docx",
		}

		if req.NamaFile != "test_module.docx" {
			t.Errorf("Expected NamaFile 'test_module.docx', got %s", req.NamaFile)
		}
	})

	t.Run("ModulUpdateRequest", func(t *testing.T) {
		req := ModulUpdateRequest{
			NamaFile: "updated_module.pdf",
			Semester: 5,
		}

		if req.NamaFile != "updated_module.pdf" {
			t.Errorf("Expected NamaFile 'updated_module.pdf', got %s", req.NamaFile)
		}
		if req.Semester != 5 {
			t.Errorf("Expected Semester 5, got %d", req.Semester)
		}
	})

	t.Run("ModulUpdateRequest with partial data", func(t *testing.T) {
		req := ModulUpdateRequest{
			Semester: 4,
		}

		if req.NamaFile != "" {
			t.Errorf("Expected empty NamaFile, got %s", req.NamaFile)
		}
		if req.Semester != 4 {
			t.Errorf("Expected Semester 4, got %d", req.Semester)
		}
	})
}

func TestModulListQueryParams(t *testing.T) {
	t.Run("ModulListQueryParams with all fields", func(t *testing.T) {
		params := ModulListQueryParams{
			Search:         "test",
			FilterType:     "pdf",
			FilterSemester: 3,
			Page:           1,
			Limit:          20,
		}

		if params.Search != "test" {
			t.Errorf("Expected Search 'test', got %s", params.Search)
		}
		if params.FilterType != "pdf" {
			t.Errorf("Expected FilterType 'pdf', got %s", params.FilterType)
		}
		if params.FilterSemester != 3 {
			t.Errorf("Expected FilterSemester 3, got %d", params.FilterSemester)
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
			ID:                 1,
			NamaFile:           "test.pdf",
			Tipe:               "pdf",
			Ukuran:             "1.5 MB",
			Semester:           2,
			PathFile:           "/uploads/test.pdf",
			TerakhirDiperbarui: now,
		}

		if item.ID != 1 {
			t.Errorf("Expected ID 1, got %d", item.ID)
		}
		if item.NamaFile != "test.pdf" {
			t.Errorf("Expected NamaFile 'test.pdf', got %s", item.NamaFile)
		}
	})

	t.Run("ModulListData with pagination", func(t *testing.T) {
		items := []ModulListItem{
			{ID: 1, NamaFile: "test1.pdf", Tipe: "pdf", Semester: 1},
			{ID: 2, NamaFile: "test2.pdf", Tipe: "pdf", Semester: 2},
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
			ID:        1,
			NamaFile:  "test.pdf",
			Tipe:      "pdf",
			Ukuran:    "1 MB",
			Semester:  1,
			PathFile:  "/uploads/test.pdf",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if resp.NamaFile != "test.pdf" {
			t.Errorf("Expected NamaFile 'test.pdf', got %s", resp.NamaFile)
		}
	})

	t.Run("ModulCreateResponse with items", func(t *testing.T) {
		items := []ModulResponse{
			{ID: 1, NamaFile: "test1.pdf", Tipe: "pdf"},
			{ID: 2, NamaFile: "test2.pdf", Tipe: "docx"},
		}
		resp := ModulCreateResponse{
			Items: items,
		}

		if len(resp.Items) != 2 {
			t.Errorf("Expected 2 items, got %d", len(resp.Items))
		}
	})
}

func TestModulDownloadRequest(t *testing.T) {
	t.Run("ModulDownloadRequest with multiple IDs", func(t *testing.T) {
		req := ModulDownloadRequest{
			IDs: []uint{1, 2, 3, 4, 5},
		}

		if len(req.IDs) != 5 {
			t.Errorf("Expected 5 IDs, got %d", len(req.IDs))
		}
		if req.IDs[0] != 1 {
			t.Errorf("Expected first ID 1, got %d", req.IDs[0])
		}
	})

	t.Run("ModulDownloadRequest with single ID", func(t *testing.T) {
		req := ModulDownloadRequest{
			IDs: []uint{100},
		}

		if len(req.IDs) != 1 {
			t.Errorf("Expected 1 ID, got %d", len(req.IDs))
		}
	})
}

func TestModulSemesterRange(t *testing.T) {
	validSemesters := []int{1, 2, 3, 4, 5, 6, 7, 8}

	for _, sem := range validSemesters {
		modul := Modul{Semester: sem}
		if modul.Semester < 1 || modul.Semester > 8 {
			t.Errorf("Semester %d should be valid (1-8)", sem)
		}
	}

	invalidSemesters := []int{0, -1, 9, 10}
	for _, sem := range invalidSemesters {
		modul := Modul{Semester: sem}
		if modul.Semester >= 1 && modul.Semester <= 8 {
			t.Errorf("Semester %d should be invalid (outside 1-8)", sem)
		}
	}
}
