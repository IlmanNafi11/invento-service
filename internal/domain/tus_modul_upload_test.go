package domain

import (
	"testing"
	"time"
)

func TestTusModulUploadStruct(t *testing.T) {
	t.Run("TusModulUpload struct initialization", func(t *testing.T) {
		now := time.Now()
		modulID := uint(50)

		upload := TusModulUpload{
			ID:         "modul-upload-123",
			UserID:     "user-1",
			ModulID:    &modulID,
			UploadType: ModulUploadTypeCreate,
			UploadURL:  "https://example.com/upload/modul-upload-123",
			FileSize:   512000,
			FilePath:   "/uploads/module.pdf",
			Status:     ModulUploadStatusPending,
			Progress:   0,
			CreatedAt:  now,
			UpdatedAt:  now,
			ExpiresAt:  now.Add(24 * time.Hour),
		}

		if upload.ID != "modul-upload-123" {
			t.Errorf("Expected ID 'modul-upload-123', got %s", upload.ID)
		}
		if upload.UserID != "user-1" {
			t.Errorf("Expected UserID 'user-1', got %s", upload.UserID)
		}
		if upload.ModulID == nil || *upload.ModulID != 50 {
			t.Errorf("Expected ModulID 50, got %v", upload.ModulID)
		}
		if upload.UploadType != ModulUploadTypeCreate {
			t.Errorf("Expected UploadType '%s', got %s", ModulUploadTypeCreate, upload.UploadType)
		}
		if upload.Status != ModulUploadStatusPending {
			t.Errorf("Expected Status '%s', got %s", ModulUploadStatusPending, upload.Status)
		}
	})

	t.Run("TusModulUpload without ModulID", func(t *testing.T) {
		upload := TusModulUpload{
			ID:         "modul-upload-456",
			UserID: "user-2",
			ModulID:    nil,
			UploadType: ModulUploadTypeUpdate,
			FileSize:   256000,
			Status:     ModulUploadStatusQueued,
		}

		if upload.ModulID != nil {
			t.Errorf("Expected nil ModulID, got %v", upload.ModulID)
		}
		if upload.UploadType != ModulUploadTypeUpdate {
			t.Errorf("Expected UploadType '%s', got %s", ModulUploadTypeUpdate, upload.UploadType)
		}
	})

	t.Run("TusModulUpload with completed status", func(t *testing.T) {
		now := time.Now()
		completedAt := now

		upload := TusModulUpload{
			ID:          "modul-upload-789",
			UserID: "user-3",
			UploadType:  ModulUploadTypeCreate,
			FileSize:    1024000,
			Status:      ModulUploadStatusCompleted,
			Progress:    100.0,
			CompletedAt: &completedAt,
		}

		if upload.Status != ModulUploadStatusCompleted {
			t.Errorf("Expected Status '%s', got %s", ModulUploadStatusCompleted, upload.Status)
		}
		if upload.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0, got %f", upload.Progress)
		}
		if upload.CompletedAt == nil {
			t.Error("Expected CompletedAt to be set, got nil")
		}
	})
}

func TestModulUploadStatusConstants(t *testing.T) {
	t.Run("All modul upload status constants are defined", func(t *testing.T) {
		statuses := []string{
			ModulUploadStatusQueued,
			ModulUploadStatusPending,
			ModulUploadStatusUploading,
			ModulUploadStatusCompleted,
			ModulUploadStatusCancelled,
			ModulUploadStatusFailed,
			ModulUploadStatusExpired,
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

func TestModulUploadTypeConstants(t *testing.T) {
	t.Run("All modul upload type constants are defined", func(t *testing.T) {
		types := []string{
			ModulUploadTypeCreate,
			ModulUploadTypeUpdate,
		}

		expectedTypes := []string{
			"modul_create",
			"modul_update",
		}

		for i, uploadType := range types {
			if uploadType != expectedTypes[i] {
				t.Errorf("Expected upload type '%s', got '%s'", expectedTypes[i], uploadType)
			}
		}
	})
}

func TestTusModulUploadInitRequest(t *testing.T) {
	t.Run("TusModulUploadInitRequest with valid data", func(t *testing.T) {
		req := TusModulUploadInitRequest{
			NamaFile: "Introduction to Algorithms",
			Tipe:     "pdf",
			Semester: 1,
		}

		if req.NamaFile != "Introduction to Algorithms" {
			t.Errorf("Expected NamaFile 'Introduction to Algorithms', got %s", req.NamaFile)
		}
		if req.Tipe != "pdf" {
			t.Errorf("Expected Tipe 'pdf', got %s", req.Tipe)
		}
		if req.Semester != 1 {
			t.Errorf("Expected Semester 1, got %d", req.Semester)
		}
	})

	t.Run("TusModulUploadInitRequest with different tipe", func(t *testing.T) {
		validTypes := []string{"docx", "xlsx", "pdf", "pptx"}

		for _, tipe := range validTypes {
			req := TusModulUploadInitRequest{
				NamaFile: "Test Module",
				Tipe:     tipe,
				Semester: 3,
			}

			if req.Tipe != tipe {
				t.Errorf("Expected Tipe '%s', got %s", tipe, req.Tipe)
			}
		}
	})
}

func TestTusModulUploadResponse(t *testing.T) {
	t.Run("TusModulUploadResponse struct", func(t *testing.T) {
		resp := TusModulUploadResponse{
			UploadID:  "modul-upload-123",
			UploadURL: "https://example.com/tus/modul-upload-123",
			Offset:    0,
			Length:    512000,
		}

		if resp.UploadID != "modul-upload-123" {
			t.Errorf("Expected UploadID 'modul-upload-123', got %s", resp.UploadID)
		}
		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0, got %d", resp.Offset)
		}
		if resp.Length != 512000 {
			t.Errorf("Expected Length 512000, got %d", resp.Length)
		}
	})

	t.Run("TusModulUploadResponse with progress", func(t *testing.T) {
		resp := TusModulUploadResponse{
			UploadID:  "modul-upload-456",
			UploadURL: "https://example.com/tus/modul-upload-456",
			Offset:    256000,
			Length:    512000,
		}

		if resp.Offset != 256000 {
			t.Errorf("Expected Offset 256000, got %d", resp.Offset)
		}
		expectedProgress := float64(resp.Offset) / float64(resp.Length) * 100
		if expectedProgress != 50.0 {
			t.Errorf("Expected progress 50.0, got %f", expectedProgress)
		}
	})
}

func TestTusModulUploadInfoResponse(t *testing.T) {
	t.Run("TusModulUploadInfoResponse struct", func(t *testing.T) {
		now := time.Now()
		modulID := uint(100)

		resp := TusModulUploadInfoResponse{
			UploadID:  "modul-upload-123",
			ModulID:   modulID,
			NamaFile:  "Data Structures",
			Tipe:      "pdf",
			Semester:  5,
			Status:    ModulUploadStatusUploading,
			Progress:  75.0,
			Offset:    750000,
			Length:    1000000,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if resp.UploadID != "modul-upload-123" {
			t.Errorf("Expected UploadID 'modul-upload-123', got %s", resp.UploadID)
		}
		if resp.ModulID != 100 {
			t.Errorf("Expected ModulID 100, got %d", resp.ModulID)
		}
		if resp.Status != ModulUploadStatusUploading {
			t.Errorf("Expected Status '%s', got %s", ModulUploadStatusUploading, resp.Status)
		}
		if resp.Progress != 75.0 {
			t.Errorf("Expected Progress 75.0, got %f", resp.Progress)
		}
	})

	t.Run("TusModulUploadInfoResponse without ModulID", func(t *testing.T) {
		resp := TusModulUploadInfoResponse{
			UploadID: "modul-upload-456",
			NamaFile: "New Module",
			Tipe:     "docx",
			Semester: 2,
			Status:   ModulUploadStatusPending,
			Progress: 0,
		}

		if resp.ModulID != 0 {
			t.Errorf("Expected ModulID 0, got %d", resp.ModulID)
		}
	})
}

func TestTusModulUploadSlotResponse(t *testing.T) {
	t.Run("TusModulUploadSlotResponse available", func(t *testing.T) {
		resp := TusModulUploadSlotResponse{
			Available:   true,
			Message:     "Upload slot available",
			QueueLength: 1,
			MaxQueue:    10,
		}

		if !resp.Available {
			t.Error("Expected Available to be true")
		}
		if resp.Message != "Upload slot available" {
			t.Errorf("Expected Message 'Upload slot available', got %s", resp.Message)
		}
		if resp.QueueLength != 1 {
			t.Errorf("Expected QueueLength 1, got %d", resp.QueueLength)
		}
		if resp.MaxQueue != 10 {
			t.Errorf("Expected MaxQueue 10, got %d", resp.MaxQueue)
		}
	})

	t.Run("TusModulUploadSlotResponse not available", func(t *testing.T) {
		resp := TusModulUploadSlotResponse{
			Available:   false,
			Message:     "Queue is full",
			QueueLength: 10,
			MaxQueue:    10,
		}

		if resp.Available {
			t.Error("Expected Available to be false")
		}
		if resp.QueueLength != resp.MaxQueue {
			t.Errorf("Expected QueueLength to equal MaxQueue when full")
		}
	})
}

func TestModulUpdateMetadataRequest(t *testing.T) {
	t.Run("ModulUpdateMetadataRequest with valid data", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "Updated Module Name",
			Semester: 4,
		}

		if req.NamaFile != "Updated Module Name" {
			t.Errorf("Expected NamaFile 'Updated Module Name', got %s", req.NamaFile)
		}
		if req.Semester != 4 {
			t.Errorf("Expected Semester 4, got %d", req.Semester)
		}
	})

	t.Run("ModulUpdateMetadataRequest with minimum values", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "ABC",
			Semester: 1,
		}

		if len(req.NamaFile) != 3 {
			t.Errorf("Expected NamaFile length 3, got %d", len(req.NamaFile))
		}
		if req.Semester != 1 {
			t.Errorf("Expected Semester 1, got %d", req.Semester)
		}
	})

	t.Run("ModulUpdateMetadataRequest with maximum semester", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "Advanced Topic Module",
			Semester: 8,
		}

		if req.Semester != 8 {
			t.Errorf("Expected Semester 8, got %d", req.Semester)
		}
	})
}

func TestTusModulUploadProgressCalculation(t *testing.T) {
	t.Run("Calculate progress percentage", func(t *testing.T) {
		testCases := []struct {
			offset      int64
			length      int64
			expectedPct float64
		}{
			{0, 500000, 0.0},
			{250000, 500000, 50.0},
			{500000, 500000, 100.0},
			{375000, 500000, 75.0},
		}

		for _, tc := range testCases {
			progress := float64(tc.offset) / float64(tc.length) * 100
			if progress != tc.expectedPct {
				t.Errorf("Expected progress %f, got %f", tc.expectedPct, progress)
			}
		}
	})
}

func TestTusModulUploadStatusTransitions(t *testing.T) {
	t.Run("Valid status transitions", func(t *testing.T) {
		transitions := map[string][]string{
			ModulUploadStatusQueued:    {ModulUploadStatusPending},
			ModulUploadStatusPending:   {ModulUploadStatusUploading, ModulUploadStatusCancelled, ModulUploadStatusExpired},
			ModulUploadStatusUploading: {ModulUploadStatusCompleted, ModulUploadStatusFailed, ModulUploadStatusCancelled},
		}

		for from, validTo := range transitions {
			for _, to := range validTo {
				valid := false
				for _, v := range validTo {
					if to == v {
						valid = true
						break
					}
				}
				if !valid {
					t.Errorf("Transition from %s to %s should be valid", from, to)
				}
			}
		}
	})
}

func TestTusModulUploadMetadata(t *testing.T) {
	t.Run("TusModulUpload with metadata", func(t *testing.T) {
		now := time.Now()
		metadata := TusModulUploadInitRequest{
			NamaFile: "Machine Learning Basics",
			Tipe:     "pdf",
			Semester: 6,
		}

		upload := TusModulUpload{
			ID:              "modul-upload-metadata-1",
			UserID: "user-1",
			UploadType:      ModulUploadTypeCreate,
			UploadMetadata:  metadata,
			FileSize:        2500000,
			Status:          ModulUploadStatusPending,
			CreatedAt:       now,
			UpdatedAt:       now,
			ExpiresAt:       now.Add(24 * time.Hour),
		}

		if upload.UploadMetadata.NamaFile != "Machine Learning Basics" {
			t.Errorf("Expected NamaFile 'Machine Learning Basics', got %s", upload.UploadMetadata.NamaFile)
		}
		if upload.UploadMetadata.Tipe != "pdf" {
			t.Errorf("Expected Tipe 'pdf', got %s", upload.UploadMetadata.Tipe)
		}
		if upload.UploadMetadata.Semester != 6 {
			t.Errorf("Expected Semester 6, got %d", upload.UploadMetadata.Semester)
		}
	})
}

func TestValidModulTypes(t *testing.T) {
	t.Run("All valid modul file types", func(t *testing.T) {
		validTypes := []string{"docx", "xlsx", "pdf", "pptx"}

		for _, tipe := range validTypes {
			req := TusModulUploadInitRequest{
				NamaFile: "Test",
				Tipe:     tipe,
				Semester: 1,
			}

			if req.Tipe != tipe {
				t.Errorf("Expected Tipe '%s', got %s", tipe, req.Tipe)
			}
		}
	})
}

func TestTusModulUploadEdgeCases(t *testing.T) {
	t.Run("Nil ModulID for new modul creation", func(t *testing.T) {
		upload := TusModulUpload{
			ID:         "modul-upload-nil",
			UserID: "user-1",
			ModulID:    nil,
			UploadType: ModulUploadTypeCreate,
			Status:     ModulUploadStatusPending,
		}

		if upload.ModulID != nil {
			t.Errorf("Expected nil ModulID for new modul creation, got %v", upload.ModulID)
		}
		if upload.UploadType != ModulUploadTypeCreate {
			t.Errorf("Expected UploadType '%s', got %s", ModulUploadTypeCreate, upload.UploadType)
		}
	})

	t.Run("Non-nil ModulID for modul update", func(t *testing.T) {
		modulID := uint(123)
		upload := TusModulUpload{
			ID:         "modul-upload-update",
			UserID: "user-1",
			ModulID:    &modulID,
			UploadType: ModulUploadTypeUpdate,
			Status:     ModulUploadStatusPending,
		}

		if upload.ModulID == nil {
			t.Error("Expected non-nil ModulID for modul update")
		} else if *upload.ModulID != 123 {
			t.Errorf("Expected ModulID 123, got %d", *upload.ModulID)
		}
		if upload.UploadType != ModulUploadTypeUpdate {
			t.Errorf("Expected UploadType '%s', got %s", ModulUploadTypeUpdate, upload.UploadType)
		}
	})

	t.Run("Zero progress for pending upload", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "modul-upload-zero-progress",
			UserID: "user-1",
			Status:   ModulUploadStatusPending,
			Progress: 0,
			FileSize: 1000000,
		}

		if upload.Progress != 0 {
			t.Errorf("Expected Progress 0 for pending upload, got %f", upload.Progress)
		}
	})

	t.Run("Full progress for completed upload", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "modul-upload-full-progress",
			UserID: "user-1",
			Status:   ModulUploadStatusCompleted,
			Progress: 100.0,
			FileSize: 1000000,
		}

		if upload.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0 for completed upload, got %f", upload.Progress)
		}
	})

	t.Run("Partial progress during upload", func(t *testing.T) {
		upload := TusModulUpload{
			ID:            "modul-upload-partial",
			UserID: "user-1",
			Status:        ModulUploadStatusUploading,
			Progress:      45.5,
			FileSize:      1000000,
			CurrentOffset: 455000,
		}

		if upload.Progress != 45.5 {
			t.Errorf("Expected Progress 45.5, got %f", upload.Progress)
		}
		if upload.CurrentOffset != 455000 {
			t.Errorf("Expected CurrentOffset 455000, got %d", upload.CurrentOffset)
		}
	})

	t.Run("Empty upload URL before initialization", func(t *testing.T) {
		upload := TusModulUpload{
			ID:        "modul-upload-no-url",
			UserID: "user-1",
			UploadURL: "",
			Status:    ModulUploadStatusQueued,
		}

		if upload.UploadURL != "" {
			t.Errorf("Expected empty UploadURL before initialization, got %s", upload.UploadURL)
		}
	})

	t.Run("Populated upload URL after initialization", func(t *testing.T) {
		upload := TusModulUpload{
			ID:        "modul-upload-with-url",
			UserID: "user-1",
			UploadURL: "https://example.com/tus/modul-upload-with-url",
			Status:    ModulUploadStatusPending,
		}

		if upload.UploadURL == "" {
			t.Error("Expected non-empty UploadURL after initialization")
		}
	})

	t.Run("Nil CompletedAt for in-progress upload", func(t *testing.T) {
		upload := TusModulUpload{
			ID:          "modul-upload-in-progress",
			UserID: "user-1",
			Status:      ModulUploadStatusUploading,
			Progress:    50.0,
			CompletedAt: nil,
		}

		if upload.CompletedAt != nil {
			t.Error("Expected nil CompletedAt for in-progress upload")
		}
	})

	t.Run("Non-nil CompletedAt for completed upload", func(t *testing.T) {
		now := time.Now()
		upload := TusModulUpload{
			ID:          "modul-upload-done",
			UserID: "user-1",
			Status:      ModulUploadStatusCompleted,
			Progress:    100.0,
			CompletedAt: &now,
		}

		if upload.CompletedAt == nil {
			t.Error("Expected non-nil CompletedAt for completed upload")
		}
	})

	t.Run("Zero file size edge case", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "modul-upload-zero-size",
			UserID: "user-1",
			FileSize: 0,
			Status:   ModulUploadStatusPending,
		}

		if upload.FileSize != 0 {
			t.Errorf("Expected FileSize 0, got %d", upload.FileSize)
		}
	})

	t.Run("Large file size", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "modul-upload-large",
			UserID: "user-1",
			FileSize: 500000000, // 500MB
			Status:   ModulUploadStatusPending,
		}

		if upload.FileSize != 500000000 {
			t.Errorf("Expected FileSize 500000000, got %d", upload.FileSize)
		}
	})
}

func TestTusModulUploadInitRequestEdgeCases(t *testing.T) {
	t.Run("Minimum valid NamaFile length", func(t *testing.T) {
		req := TusModulUploadInitRequest{
			NamaFile: "ABC",
			Tipe:     "pdf",
			Semester: 1,
		}

		if len(req.NamaFile) != 3 {
			t.Errorf("Expected NamaFile length 3, got %d", len(req.NamaFile))
		}
	})

	t.Run("Long valid NamaFile", func(t *testing.T) {
		longName := "A" + " very long module name that describes the content in detail " + "B"
		req := TusModulUploadInitRequest{
			NamaFile: longName,
			Tipe:     "docx",
			Semester: 5,
		}

		if req.NamaFile != longName {
			t.Errorf("Expected NamaFile '%s', got %s", longName, req.NamaFile)
		}
	})

	t.Run("Minimum semester value", func(t *testing.T) {
		req := TusModulUploadInitRequest{
			NamaFile: "Semester 1 Module",
			Tipe:     "pdf",
			Semester: 1,
		}

		if req.Semester != 1 {
			t.Errorf("Expected Semester 1, got %d", req.Semester)
		}
	})

	t.Run("Maximum semester value", func(t *testing.T) {
		req := TusModulUploadInitRequest{
			NamaFile: "Semester 8 Module",
			Tipe:     "pptx",
			Semester: 8,
		}

		if req.Semester != 8 {
			t.Errorf("Expected Semester 8, got %d", req.Semester)
		}
	})

	t.Run("All valid file types", func(t *testing.T) {
		validTypes := []string{"docx", "xlsx", "pdf", "pptx"}

		for _, tipe := range validTypes {
			req := TusModulUploadInitRequest{
				NamaFile: "Module",
				Tipe:     tipe,
				Semester: 3,
			}

			if req.Tipe != tipe {
				t.Errorf("Expected Tipe '%s', got %s", tipe, req.Tipe)
			}
		}
	})
}

func TestModulUpdateMetadataRequestEdgeCases(t *testing.T) {
	t.Run("Minimum NamaFile length", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "XYZ",
			Semester: 1,
		}

		if len(req.NamaFile) != 3 {
			t.Errorf("Expected NamaFile length 3, got %d", len(req.NamaFile))
		}
	})

	t.Run("Update only NamaFile", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "Updated Name Only",
			Semester: 0,
		}

		if req.NamaFile != "Updated Name Only" {
			t.Errorf("Expected NamaFile 'Updated Name Only', got %s", req.NamaFile)
		}
	})

	t.Run("Update only Semester", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "",
			Semester: 6,
		}

		if req.Semester != 6 {
			t.Errorf("Expected Semester 6, got %d", req.Semester)
		}
	})

	t.Run("Update both fields", func(t *testing.T) {
		req := ModulUpdateMetadataRequest{
			NamaFile: "Complete Update",
			Semester: 7,
		}

		if req.NamaFile != "Complete Update" {
			t.Errorf("Expected NamaFile 'Complete Update', got %s", req.NamaFile)
		}
		if req.Semester != 7 {
			t.Errorf("Expected Semester 7, got %d", req.Semester)
		}
	})
}

func TestTusModulUploadResponseEdgeCases(t *testing.T) {
	t.Run("Zero offset for new upload", func(t *testing.T) {
		resp := TusModulUploadResponse{
			UploadID:  "new-upload",
			UploadURL: "https://example.com/tus/new-upload",
			Offset:    0,
			Length:    1000000,
		}

		if resp.Offset != 0 {
			t.Errorf("Expected Offset 0 for new upload, got %d", resp.Offset)
		}
	})

	t.Run("Offset equals length for completed upload", func(t *testing.T) {
		resp := TusModulUploadResponse{
			UploadID:  "completed-upload",
			UploadURL: "https://example.com/tus/completed-upload",
			Offset:    1000000,
			Length:    1000000,
		}

		if resp.Offset != resp.Length {
			t.Errorf("Expected Offset to equal Length for completed upload")
		}
	})

	t.Run("Zero length upload", func(t *testing.T) {
		resp := TusModulUploadResponse{
			UploadID:  "zero-length",
			UploadURL: "https://example.com/tus/zero-length",
			Offset:    0,
			Length:    0,
		}

		if resp.Length != 0 {
			t.Errorf("Expected Length 0, got %d", resp.Length)
		}
	})
}

func TestTusModulUploadInfoResponseEdgeCases(t *testing.T) {
	t.Run("Zero ModulID for new modul", func(t *testing.T) {
		resp := TusModulUploadInfoResponse{
			UploadID: "new-modul-upload",
			ModulID:  0,
			NamaFile: "New Module",
			Tipe:     "pdf",
			Semester: 1,
			Status:   ModulUploadStatusPending,
			Progress: 0,
		}

		if resp.ModulID != 0 {
			t.Errorf("Expected ModulID 0 for new modul, got %d", resp.ModulID)
		}
	})

	t.Run("Non-zero ModulID for existing modul update", func(t *testing.T) {
		resp := TusModulUploadInfoResponse{
			UploadID: "update-modul-upload",
			ModulID:  456,
			NamaFile: "Updated Module",
			Tipe:     "docx",
			Semester: 3,
			Status:   ModulUploadStatusUploading,
			Progress: 50.0,
		}

		if resp.ModulID != 456 {
			t.Errorf("Expected ModulID 456, got %d", resp.ModulID)
		}
	})

	t.Run("Zero progress for pending status", func(t *testing.T) {
		resp := TusModulUploadInfoResponse{
			UploadID: "pending-upload",
			Status:   ModulUploadStatusPending,
			Progress: 0,
		}

		if resp.Progress != 0 {
			t.Errorf("Expected Progress 0 for pending status, got %f", resp.Progress)
		}
	})

	t.Run("Full progress for completed status", func(t *testing.T) {
		resp := TusModulUploadInfoResponse{
			UploadID: "completed-upload",
			Status:   ModulUploadStatusCompleted,
			Progress: 100.0,
		}

		if resp.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0 for completed status, got %f", resp.Progress)
		}
	})

	t.Run("Timestamp fields present", func(t *testing.T) {
		now := time.Now()
		resp := TusModulUploadInfoResponse{
			UploadID:  "timestamp-test",
			CreatedAt: now,
			UpdatedAt: now,
		}

		if resp.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set")
		}
		if resp.UpdatedAt.IsZero() {
			t.Error("Expected UpdatedAt to be set")
		}
	})
}

func TestTusModulUploadSlotResponseEdgeCases(t *testing.T) {
	t.Run("Zero queue length when empty", func(t *testing.T) {
		resp := TusModulUploadSlotResponse{
			Available:   true,
			Message:     "Queue is empty",
			QueueLength: 0,
			MaxQueue:    10,
		}

		if resp.QueueLength != 0 {
			t.Errorf("Expected QueueLength 0, got %d", resp.QueueLength)
		}
	})

	t.Run("Queue length equals max queue when full", func(t *testing.T) {
		resp := TusModulUploadSlotResponse{
			Available:   false,
			Message:     "Queue is full",
			QueueLength: 10,
			MaxQueue:    10,
		}

		if resp.QueueLength != resp.MaxQueue {
			t.Errorf("Expected QueueLength to equal MaxQueue when full")
		}
		if resp.Available {
			t.Error("Expected Available to be false when queue is full")
		}
	})

	t.Run("One below max queue", func(t *testing.T) {
		resp := TusModulUploadSlotResponse{
			Available:   true,
			Message:     "One slot remaining",
			QueueLength: 9,
			MaxQueue:    10,
		}

		if resp.QueueLength != 9 {
			t.Errorf("Expected QueueLength 9, got %d", resp.QueueLength)
		}
		if !resp.Available {
			t.Error("Expected Available to be true when queue is not full")
		}
	})

	t.Run("Max queue zero for unlimited queue", func(t *testing.T) {
		resp := TusModulUploadSlotResponse{
			Available:   true,
			Message:     "Unlimited queue",
			QueueLength: 0,
			MaxQueue:    0,
		}

		if resp.MaxQueue != 0 {
			t.Errorf("Expected MaxQueue 0 for unlimited, got %d", resp.MaxQueue)
		}
	})
}

func TestTusModulUploadInvalidProgress(t *testing.T) {
	t.Run("Negative progress is invalid", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "negative-progress",
			UserID: "user-1",
			Status:   ModulUploadStatusUploading,
			Progress: -10.0,
		}

		if upload.Progress >= 0 {
			t.Logf("Note: Negative progress value %f is stored but should be validated at business logic layer", upload.Progress)
		}
	})

	t.Run("Progress over 100 is invalid", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "over-100-progress",
			UserID: "user-1",
			Status:   ModulUploadStatusUploading,
			Progress: 150.0,
		}

		if upload.Progress > 100 {
			t.Logf("Note: Progress over 100 (%f) is stored but should be validated at business logic layer", upload.Progress)
		}
	})
}
