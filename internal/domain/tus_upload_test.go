package domain

import (
	"invento-service/internal/dto"
	"testing"
	"time"
)

func TestTusUploadStruct(t *testing.T) {
	t.Run("TusUpload struct initialization", func(t *testing.T) {
		now := time.Now()
		projectID := uint(100)

		upload := TusUpload{
			ID:         "upload-123",
			UserID: "user-1",
			ProjectID:  &projectID,
			UploadType: UploadTypeProjectCreate,
			UploadURL:  "https://example.com/upload/upload-123",
			FileSize:   1024000,
			FilePath:   "/uploads/project.zip",
			Status:     UploadStatusPending,
			Progress:   0,
			CreatedAt:  now,
			UpdatedAt:  now,
			ExpiresAt:  now.Add(24 * time.Hour),
		}

		if upload.ID != "upload-123" {
			t.Errorf("Expected ID 'upload-123', got %s", upload.ID)
		}
		if upload.UserID != "user-1" {
			t.Errorf("Expected UserID 'user-1', got %s", upload.UserID)
		}
		if upload.ProjectID == nil || *upload.ProjectID != 100 {
			t.Errorf("Expected ProjectID 100, got %v", upload.ProjectID)
		}
		if upload.UploadType != UploadTypeProjectCreate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeProjectCreate, upload.UploadType)
		}
		if upload.Status != UploadStatusPending {
			t.Errorf("Expected Status '%s', got %s", UploadStatusPending, upload.Status)
		}
	})

	t.Run("TusUpload without ProjectID", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-456",
			UserID: "user-2",
			ProjectID:  nil,
			UploadType: UploadTypeProjectUpdate,
			FileSize:   512000,
			Status:     UploadStatusQueued,
		}

		if upload.ProjectID != nil {
			t.Errorf("Expected nil ProjectID, got %v", upload.ProjectID)
		}
		if upload.UploadType != UploadTypeProjectUpdate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeProjectUpdate, upload.UploadType)
		}
	})

	t.Run("TusUpload with completed status", func(t *testing.T) {
		now := time.Now()
		completedAt := now

		upload := TusUpload{
			ID:          "upload-789",
			UserID: "user-3",
			UploadType:  UploadTypeProjectCreate,
			FileSize:    2048000,
			Status:      UploadStatusCompleted,
			Progress:    100.0,
			CompletedAt: &completedAt,
		}

		if upload.Status != UploadStatusCompleted {
			t.Errorf("Expected Status '%s', got %s", UploadStatusCompleted, upload.Status)
		}
		if upload.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0, got %f", upload.Progress)
		}
		if upload.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set, got nil")
		}
	})
}

func TestUploadStatusConstants(t *testing.T) {
	t.Run("All upload status constants are defined", func(t *testing.T) {
		statuses := []string{
			UploadStatusQueued,
			UploadStatusPending,
			UploadStatusUploading,
			UploadStatusCompleted,
			UploadStatusCancelled,
			UploadStatusFailed,
			UploadStatusExpired,
		}

		expectedStatuses := []string{
			"queued",
			"pending",
			"uploading",
			"completed",
			"cancelled",
			"failed",
			"expired",
		}

		for i, status := range statuses {
			if status != expectedStatuses[i] {
				t.Errorf("Expected status '%s', got '%s'", expectedStatuses[i], status)
			}
		}
	})
}

func TestUploadTypeConstants(t *testing.T) {
	t.Run("All upload type constants are defined", func(t *testing.T) {
		types := []string{
			UploadTypeProjectCreate,
			UploadTypeProjectUpdate,
		}

		expectedTypes := []string{
			"project_create",
			"project_update",
		}

		for i, uploadType := range types {
			if uploadType != expectedTypes[i] {
				t.Errorf("Expected upload type '%s', got '%s'", expectedTypes[i], uploadType)
			}
		}
	})
}

func TestTusUploadInitRequest(t *testing.T) {
	t.Run("TusUploadMetadata with valid data", func(t *testing.T) {
		req := TusUploadMetadata{
			NamaProject: "My Awesome Project",
			Kategori:    "website",
			Semester:    3,
		}

		if req.NamaProject != "My Awesome Project" {
			t.Errorf("Expected NamaProject 'My Awesome Project', got %s", req.NamaProject)
		}
		if req.Kategori != "website" {
			t.Errorf("Expected Kategori 'website', got %s", req.Kategori)
		}
		if req.Semester != 3 {
			t.Errorf("Expected Semester 3, got %d", req.Semester)
		}
	})

	t.Run("TusUploadMetadata with different kategori", func(t *testing.T) {
		validKategories := []string{"website", "mobile", "iot", "machine_learning", "deep_learning"}

		for _, kategori := range validKategories {
			req := TusUploadMetadata{
				NamaProject: "Test Project",
				Kategori:    kategori,
				Semester:    1,
			}

			if req.Kategori != kategori {
				t.Errorf("Expected Kategori '%s', got %s", kategori, req.Kategori)
			}
		}
	})
}

func TestTusUploadResponse(t *testing.T) {
	t.Run("dto.TusUploadResponse struct", func(t *testing.T) {
		resp := dto.TusUploadResponse{
			UploadID:  "upload-123",
			UploadURL: "https://example.com/tus/upload-123",
			Offset:    0,
			Length:    1024000,
		}

		if resp.UploadID != "upload-123" {
			t.Errorf("Expected UploadID 'upload-123', got %s", resp.UploadID)
		}
		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 1024000 {
			t.Errorf("Expected Length 1024000, got %d", resp.Length)
		}
	})

	t.Run("dto.TusUploadResponse with progress", func(t *testing.T) {
		resp := dto.TusUploadResponse{
			UploadID:  "upload-456",
			UploadURL: "https://example.com/tus/upload-456",
			Offset:    512000,
			Length:    1024000,
		}

		if resp.Offset != 512000 {
			t.Errorf("Expected Offset 512000, got %d", resp.Offset)
		}
		expectedProgress := float64(resp.Offset) / float64(resp.Length) * 100
		if expectedProgress != 50.0 {
			t.Errorf("Expected progress 50.0, got %f", expectedProgress)
		}
	})
}

func TestTusUploadInfoResponse(t *testing.T) {
	t.Run("dto.TusUploadInfoResponse struct", func(t *testing.T) {
		now := time.Now()
		projectID := uint(100)

		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-123",
			ProjectID:   projectID,
			NamaProject: "Test Project",
			Kategori:    "mobile",
			Semester:    5,
			Status:      UploadStatusUploading,
			Progress:    65.5,
			Offset:      655000,
			Length:      1000000,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if resp.UploadID != "upload-123" {
			t.Errorf("Expected UploadID 'upload-123', got %s", resp.UploadID)
		}
		if resp.ProjectID != 100 {
			t.Errorf("Expected ProjectID 100, got %d", resp.ProjectID)
		}
		if resp.Status != UploadStatusUploading {
			t.Errorf("Expected Status '%s', got %s", UploadStatusUploading, resp.Status)
		}
		if resp.Progress != 65.5 {
			t.Errorf("Expected Progress 65.5, got %f", resp.Progress)
		}
	})

	t.Run("dto.TusUploadInfoResponse without ProjectID", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-456",
			NamaProject: "New Project",
			Kategori:    "iot",
			Semester:    2,
			Status:      UploadStatusPending,
			Progress:    0,
		}

		if resp.ProjectID != 0 {
			t.Errorf("Expected ProjectID 0, got %d", resp.ProjectID)
		}
	})
}

func TestTusUploadSlotResponse(t *testing.T) {
	t.Run("dto.TusUploadSlotResponse available", func(t *testing.T) {
		resp := dto.TusUploadSlotResponse{
			Available:     true,
			Message:       "Upload slot available",
			QueueLength:   2,
			ActiveUpload:  false,
			MaxConcurrent: 5,
		}

		if !resp.Available {
			t.Error("Expected Available to be true")
		}
		if resp.Message != "Upload slot available" {
			t.Errorf("Expected Message 'Upload slot available', got %s", resp.Message)
		}
		if resp.QueueLength != 2 {
			t.Errorf("Expected QueueLength 2, got %d", resp.QueueLength)
		}
		if resp.ActiveUpload {
			t.Error("Expected ActiveUpload to be false")
		}
		if resp.MaxConcurrent != 5 {
			t.Errorf("Expected MaxConcurrent 5, got %d", resp.MaxConcurrent)
		}
	})

	t.Run("dto.TusUploadSlotResponse not available", func(t *testing.T) {
		resp := dto.TusUploadSlotResponse{
			Available:     false,
			Message:       "No upload slots available",
			QueueLength:   10,
			ActiveUpload:  true,
			MaxConcurrent: 5,
		}

		if resp.Available {
			t.Error("Expected Available to be false")
		}
		if resp.ActiveUpload != true {
			t.Error("Expected ActiveUpload to be true")
		}
	})
}

func TestTusUploadProgressCalculation(t *testing.T) {
	t.Run("Calculate progress percentage", func(t *testing.T) {
		testCases := []struct {
			offset      int64
			length      int64
			expectedPct float64
		}{
			{0, 1000, 0.0},
			{500, 1000, 50.0},
			{1000, 1000, 100.0},
			{750, 1000, 75.0},
		}

		for _, tc := range testCases {
			progress := float64(tc.offset) / float64(tc.length) * 100
			if progress != tc.expectedPct {
				t.Errorf("Expected progress %f, got %f", tc.expectedPct, progress)
			}
		}
	})
}

func TestTusUploadStatusTransitions(t *testing.T) {
	t.Run("Valid status transitions", func(t *testing.T) {
		transitions := map[string][]string{
			UploadStatusQueued:    {UploadStatusPending},
			UploadStatusPending:   {UploadStatusUploading, UploadStatusCancelled, UploadStatusExpired},
			UploadStatusUploading: {UploadStatusCompleted, UploadStatusFailed, UploadStatusCancelled},
		}

		for from, validTo := range transitions {
			for _, to := range validTo {
				found := false
				for _, v := range validTo {
					if to == v {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Transition from %s to %s should be valid", from, to)
				}
			}
		}
	})
}

func TestTusUploadMetadata(t *testing.T) {
	t.Run("TusUpload with metadata", func(t *testing.T) {
		now := time.Now()
		metadata := TusUploadMetadata{
			NamaProject: "Test Project",
			Kategori:    "machine_learning",
			Semester:    7,
		}

		upload := TusUpload{
			ID:             "upload-metadata-1",
			UserID: "user-1",
			UploadType:     UploadTypeProjectCreate,
			UploadMetadata: metadata,
			FileSize:       5000000,
			Status:         UploadStatusPending,
			CreatedAt:      now,
			UpdatedAt:      now,
			ExpiresAt:      now.Add(24 * time.Hour),
		}

		if upload.UploadMetadata.NamaProject != "Test Project" {
			t.Errorf("Expected NamaProject 'Test Project', got %s", upload.UploadMetadata.NamaProject)
		}
		if upload.UploadMetadata.Kategori != "machine_learning" {
			t.Errorf("Expected Kategori 'machine_learning', got %s", upload.UploadMetadata.Kategori)
		}
		if upload.UploadMetadata.Semester != 7 {
			t.Errorf("Expected Semester 7, got %d", upload.UploadMetadata.Semester)
		}
	})
}

func TestTusUploadInitRequest_EdgeCases(t *testing.T) {
	t.Run("Minimum valid NamaProject length", func(t *testing.T) {
		req := TusUploadMetadata{
			NamaProject: "ABC", // min=3
			Kategori:    "website",
			Semester:    1,
		}

		if len(req.NamaProject) != 3 {
			t.Errorf("Expected NamaProject length 3, got %d", len(req.NamaProject))
		}
	})

	t.Run("Maximum valid NamaProject length", func(t *testing.T) {
		// Create a 255 character string
		longName := string(make([]byte, 255))
		for i := range longName {
			longName = longName[:i] + "A" + longName[i+1:]
		}

		req := TusUploadMetadata{
			NamaProject: longName, // max=255
			Kategori:    "mobile",
			Semester:    1,
		}

		if len(req.NamaProject) != 255 {
			t.Errorf("Expected NamaProject length 255, got %d", len(req.NamaProject))
		}
	})

	t.Run("Minimum valid Semester", func(t *testing.T) {
		req := TusUploadMetadata{
			NamaProject: "Test Project",
			Kategori:    "iot",
			Semester:    1, // min=1
		}

		if req.Semester != 1 {
			t.Errorf("Expected Semester 1, got %d", req.Semester)
		}
	})

	t.Run("Maximum valid Semester", func(t *testing.T) {
		req := TusUploadMetadata{
			NamaProject: "Test Project",
			Kategori:    "deep_learning",
			Semester:    8, // max=8
		}

		if req.Semester != 8 {
			t.Errorf("Expected Semester 8, got %d", req.Semester)
		}
	})

	t.Run("All valid kategori values", func(t *testing.T) {
		validCategories := []string{
			"website",
			"mobile",
			"iot",
			"machine_learning",
			"deep_learning",
		}

		for _, kategori := range validCategories {
			req := TusUploadMetadata{
				NamaProject: "Test Project",
				Kategori:    kategori,
				Semester:    1,
			}

			if req.Kategori != kategori {
				t.Errorf("Expected Kategori '%s', got '%s'", kategori, req.Kategori)
			}
		}
	})
}

func TestTusUploadResponse_EdgeCases(t *testing.T) {
	t.Run("Zero offset and length", func(t *testing.T) {
		resp := dto.TusUploadResponse{
			UploadID:  "upload-zero",
			UploadURL: "https://example.com/tus/upload-zero",
			Offset:    0,
			Length:    0,
		}

		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 0 {
			t.Errorf("Expected Length 0, got %d", resp.Length)
		}
	})

	t.Run("Empty upload ID", func(t *testing.T) {
		resp := dto.TusUploadResponse{
			UploadID:  "",
			UploadURL: "https://example.com/tus/",
			Offset:    0,
			Length:    1000,
		}

		if resp.UploadID != "" {
			t.Errorf("Expected empty UploadID, got '%s'", resp.UploadID)
		}
	})

	t.Run("Empty upload URL", func(t *testing.T) {
		resp := dto.TusUploadResponse{
			UploadID:  "upload-123",
			UploadURL: "",
			Offset:    0,
			Length:    1000,
		}

		if resp.UploadURL != "" {
			t.Errorf("Expected empty UploadURL, got '%s'", resp.UploadURL)
		}
	})

	t.Run("Large file size", func(t *testing.T) {
		resp := dto.TusUploadResponse{
			UploadID:  "upload-large",
			UploadURL: "https://example.com/tus/upload-large",
			Offset:    0,
			Length:    1024 * 1024 * 1024 * 5, // 5GB
		}

		expectedLength := int64(1024 * 1024 * 1024 * 5)
		if resp.Length != expectedLength {
			t.Errorf("Expected Length %d, got %d", expectedLength, resp.Length)
		}
	})
}

func TestTusUploadInfoResponse_EdgeCases(t *testing.T) {
	t.Run("Zero ProjectID", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-zero-project",
			ProjectID:   0,
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    1,
			Status:      UploadStatusPending,
			Progress:    0,
		}

		if resp.ProjectID != 0 {
			t.Errorf("Expected ProjectID 0, got %d", resp.ProjectID)
		}
	})

	t.Run("Invalid progress - negative value", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-negative-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "mobile",
			Semester:    2,
			Status:      UploadStatusUploading,
			Progress:    -10.5,
		}

		if resp.Progress != -10.5 {
			t.Errorf("Expected Progress -10.5, got %f", resp.Progress)
		}
	})

	t.Run("Invalid progress - exceeds 100", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-excess-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "iot",
			Semester:    3,
			Status:      UploadStatusUploading,
			Progress:    150.0,
		}

		if resp.Progress != 150.0 {
			t.Errorf("Expected Progress 150.0, got %f", resp.Progress)
		}
	})

	t.Run("Zero progress", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-zero-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "machine_learning",
			Semester:    4,
			Status:      UploadStatusPending,
			Progress:    0,
		}

		if resp.Progress != 0 {
			t.Errorf("Expected Progress 0, got %f", resp.Progress)
		}
	})

	t.Run("Exact 100 percent progress", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-complete-progress",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "deep_learning",
			Semester:    5,
			Status:      UploadStatusCompleted,
			Progress:    100.0,
		}

		if resp.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0, got %f", resp.Progress)
		}
	})

	t.Run("Zero offset and length", func(t *testing.T) {
		resp := dto.TusUploadInfoResponse{
			UploadID:    "upload-zero-offset",
			ProjectID:   1,
			NamaProject: "Test Project",
			Kategori:    "website",
			Semester:    1,
			Status:      UploadStatusPending,
			Progress:    0,
			Offset:      0,
			Length:      0,
		}

		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 0 {
			t.Errorf("Expected Length 0, got %d", resp.Length)
		}
	})
}

func TestTusUploadSlotResponse_EdgeCases(t *testing.T) {
	t.Run("Zero queue length", func(t *testing.T) {
		resp := dto.TusUploadSlotResponse{
			Available:     true,
			Message:       "No uploads in queue",
			QueueLength:   0,
			ActiveUpload:  false,
			MaxConcurrent: 5,
		}

		if resp.QueueLength != 0 {
			t.Errorf("Expected QueueLength 0, got %d", resp.QueueLength)
		}
	})

	t.Run("Zero max concurrent", func(t *testing.T) {
		resp := dto.TusUploadSlotResponse{
			Available:     false,
			Message:       "Uploads disabled",
			QueueLength:   0,
			ActiveUpload:  false,
			MaxConcurrent: 0,
		}

		if resp.MaxConcurrent != 0 {
			t.Errorf("Expected MaxConcurrent 0, got %d", resp.MaxConcurrent)
		}
	})

	t.Run("Empty message", func(t *testing.T) {
		resp := dto.TusUploadSlotResponse{
			Available:     true,
			Message:       "",
			QueueLength:   1,
			ActiveUpload:  true,
			MaxConcurrent: 5,
		}

		if resp.Message != "" {
			t.Errorf("Expected empty Message, got '%s'", resp.Message)
		}
	})

	t.Run("Large queue length", func(t *testing.T) {
		resp := dto.TusUploadSlotResponse{
			Available:     false,
			Message:       "Very long queue",
			QueueLength:   1000,
			ActiveUpload:  true,
			MaxConcurrent: 5,
		}

		if resp.QueueLength != 1000 {
			t.Errorf("Expected QueueLength 1000, got %d", resp.QueueLength)
		}
	})
}

func TestTusUpload_EdgeCases(t *testing.T) {
	t.Run("Nil ProjectID", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-no-project",
			UserID: "user-1",
			ProjectID:  nil,
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.ProjectID != nil {
			t.Errorf("Expected nil ProjectID, got %v", upload.ProjectID)
		}
	})

	t.Run("Zero file size", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-zero-size",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   0,
			Status:     UploadStatusPending,
		}

		if upload.FileSize != 0 {
			t.Errorf("Expected FileSize 0, got %d", upload.FileSize)
		}
	})

	t.Run("Negative file size", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-negative-size",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   -1,
			Status:     UploadStatusPending,
		}

		if upload.FileSize != -1 {
			t.Errorf("Expected FileSize -1, got %d", upload.FileSize)
		}
	})

	t.Run("Empty upload ID", func(t *testing.T) {
		upload := TusUpload{
			ID:         "",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.ID != "" {
			t.Errorf("Expected empty ID, got '%s'", upload.ID)
		}
	})

	t.Run("Zero UserID", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-zero-user",
			UserID:     "",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.UserID != "" {
			t.Errorf("Expected UserID empty, got %s", upload.UserID)
		}
	})

	t.Run("Empty upload URL", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-empty-url",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			UploadURL:  "",
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.UploadURL != "" {
			t.Errorf("Expected empty UploadURL, got '%s'", upload.UploadURL)
		}
	})

	t.Run("Empty file path", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-empty-path",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			FilePath:   "",
			Status:     UploadStatusPending,
		}

		if upload.FilePath != "" {
			t.Errorf("Expected empty FilePath, got '%s'", upload.FilePath)
		}
	})

	t.Run("Zero current offset", func(t *testing.T) {
		upload := TusUpload{
			ID:            "upload-zero-offset",
			UserID: "user-1",
			UploadType:    UploadTypeProjectCreate,
			FileSize:      1024,
			CurrentOffset: 0,
			Status:        UploadStatusPending,
		}

		if upload.CurrentOffset != 0 {
			t.Errorf("Expected CurrentOffset 0, got %d", upload.CurrentOffset)
		}
	})

	t.Run("Negative current offset", func(t *testing.T) {
		upload := TusUpload{
			ID:            "upload-negative-offset",
			UserID: "user-1",
			UploadType:    UploadTypeProjectCreate,
			FileSize:      1024,
			CurrentOffset: -1,
			Status:        UploadStatusPending,
		}

		if upload.CurrentOffset != -1 {
			t.Errorf("Expected CurrentOffset -1, got %d", upload.CurrentOffset)
		}
	})

	t.Run("Zero progress", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-zero-progress",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusPending,
			Progress:   0,
		}

		if upload.Progress != 0 {
			t.Errorf("Expected Progress 0, got %f", upload.Progress)
		}
	})

	t.Run("Negative progress", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-negative-progress",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusUploading,
			Progress:   -5.5,
		}

		if upload.Progress != -5.5 {
			t.Errorf("Expected Progress -5.5, got %f", upload.Progress)
		}
	})

	t.Run("Progress exceeds 100", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-excess-progress",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     UploadStatusUploading,
			Progress:   120.5,
		}

		if upload.Progress != 120.5 {
			t.Errorf("Expected Progress 120.5, got %f", upload.Progress)
		}
	})

	t.Run("Nil CompletedAt", func(t *testing.T) {
		upload := TusUpload{
			ID:          "upload-no-completed",
			UserID: "user-1",
			UploadType:  UploadTypeProjectCreate,
			FileSize:    1024,
			Status:      UploadStatusPending,
			CompletedAt: nil,
		}

		if upload.CompletedAt != nil {
			t.Errorf("Expected nil CompletedAt, got %v", upload.CompletedAt)
		}
	})

	t.Run("Empty upload type", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-empty-type",
			UserID: "user-1",
			UploadType: "",
			FileSize:   1024,
			Status:     UploadStatusPending,
		}

		if upload.UploadType != "" {
			t.Errorf("Expected empty UploadType, got '%s'", upload.UploadType)
		}
	})

	t.Run("Empty status", func(t *testing.T) {
		upload := TusUpload{
			ID:         "upload-empty-status",
			UserID: "user-1",
			UploadType: UploadTypeProjectCreate,
			FileSize:   1024,
			Status:     "",
		}

		if upload.Status != "" {
			t.Errorf("Expected empty Status, got '%s'", upload.Status)
		}
	})
}

func TestTusUpload_Timestamps(t *testing.T) {
	t.Run("Valid timestamps", func(t *testing.T) {
		now := time.Now()
		past := now.Add(-1 * time.Hour)
		future := now.Add(24 * time.Hour)

		upload := TusUpload{
			ID:        "upload-timestamps",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			CreatedAt: past,
			UpdatedAt: now,
			ExpiresAt: future,
		}

		if upload.CreatedAt.IsZero() {
			t.Error("Expected non-zero CreatedAt")
		}
		if upload.UpdatedAt.IsZero() {
			t.Error("Expected non-zero UpdatedAt")
		}
		if upload.ExpiresAt.IsZero() {
			t.Error("Expected non-zero ExpiresAt")
		}
	})

	t.Run("Zero CreatedAt", func(t *testing.T) {
		upload := TusUpload{
			ID:        "upload-zero-created",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			CreatedAt: time.Time{},
		}

		if !upload.CreatedAt.IsZero() {
			t.Error("Expected zero CreatedAt")
		}
	})

	t.Run("Zero UpdatedAt", func(t *testing.T) {
		upload := TusUpload{
			ID:        "upload-zero-updated",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			UpdatedAt: time.Time{},
		}

		if !upload.UpdatedAt.IsZero() {
			t.Error("Expected zero UpdatedAt")
		}
	})

	t.Run("Zero ExpiresAt", func(t *testing.T) {
		upload := TusUpload{
			ID:        "upload-zero-expires",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			ExpiresAt: time.Time{},
		}

		if !upload.ExpiresAt.IsZero() {
			t.Error("Expected zero ExpiresAt")
		}
	})

	t.Run("Expired timestamp", func(t *testing.T) {
		past := time.Now().Add(-1 * time.Hour)

		upload := TusUpload{
			ID:        "upload-expired",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusExpired,
			ExpiresAt: past,
		}

		if upload.ExpiresAt.After(time.Now()) {
			t.Error("Expected ExpiresAt to be in the past")
		}
	})

	t.Run("Future expiration", func(t *testing.T) {
		future := time.Now().Add(24 * time.Hour)

		upload := TusUpload{
			ID:        "upload-future-expiration",
			UserID: "user-1",
			FileSize:  1024,
			Status:    UploadStatusPending,
			ExpiresAt: future,
		}

		if upload.ExpiresAt.Before(time.Now()) {
			t.Error("Expected ExpiresAt to be in the future")
		}
	})
}
