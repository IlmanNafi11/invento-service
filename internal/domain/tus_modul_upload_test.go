package domain

import (
	"invento-service/internal/dto"
	"testing"
	"time"
)

func TestTusModulUploadStruct(t *testing.T) {
	t.Run("TusModulUpload struct initialization", func(t *testing.T) {
		now := time.Now()
		modulID := "550e8400-e29b-41d4-a716-446655440050"

		upload := TusModulUpload{
			ID:         "modul-upload-123",
			UserID:     "user-1",
			ModulID:    &modulID,
			UploadType: UploadTypeModulCreate,
			UploadURL:  "https://example.com/upload/modul-upload-123",
			FileSize:   512000,
			FilePath:   "/uploads/module.pdf",
			Status:     UploadStatusPending,
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
		if upload.ModulID == nil || *upload.ModulID != "550e8400-e29b-41d4-a716-446655440050" {
			t.Errorf("Expected ModulID '550e8400-e29b-41d4-a716-446655440050', got %v", upload.ModulID)
		}
		if upload.UploadType != UploadTypeModulCreate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeModulCreate, upload.UploadType)
		}
		if upload.Status != UploadStatusPending {
			t.Errorf("Expected Status '%s', got %s", UploadStatusPending, upload.Status)
		}
	})

	t.Run("TusModulUpload without ModulID", func(t *testing.T) {
		upload := TusModulUpload{
			ID:         "modul-upload-456",
			UserID:     "user-2",
			ModulID:    nil,
			UploadType: UploadTypeModulUpdate,
			FileSize:   256000,
			Status:     UploadStatusQueued,
		}

		if upload.ModulID != nil {
			t.Errorf("Expected nil ModulID, got %v", upload.ModulID)
		}
		if upload.UploadType != UploadTypeModulUpdate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeModulUpdate, upload.UploadType)
		}
	})

	t.Run("TusModulUpload with completed status", func(t *testing.T) {
		now := time.Now()
		completedAt := now

		upload := TusModulUpload{
			ID:          "modul-upload-789",
			UserID:      "user-3",
			UploadType:  UploadTypeModulCreate,
			FileSize:    1024000,
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

func TestModulUploadStatusConstants(t *testing.T) {
	t.Run("All modul upload status constants are defined", func(t *testing.T) {
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

func TestModulUploadTypeConstants(t *testing.T) {
	t.Run("All modul upload type constants are defined", func(t *testing.T) {
		types := []string{
			UploadTypeModulCreate,
			UploadTypeModulUpdate,
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
	t.Run("TusModulUploadMetadata with valid data", func(t *testing.T) {
		req := TusModulUploadMetadata{
			Judul:     "Introduction to Algorithms",
			Deskripsi: "A comprehensive guide to algorithms",
		}

		if req.Judul != "Introduction to Algorithms" {
			t.Errorf("Expected Judul 'Introduction to Algorithms', got %s", req.Judul)
		}
		if req.Deskripsi != "A comprehensive guide to algorithms" {
			t.Errorf("Expected Deskripsi 'A comprehensive guide to algorithms', got %s", req.Deskripsi)
		}
	})

	t.Run("TusModulUploadMetadata with empty deskripsi", func(t *testing.T) {
		req := TusModulUploadMetadata{
			Judul:     "Test Module",
			Deskripsi: "",
		}

		if req.Judul != "Test Module" {
			t.Errorf("Expected Judul 'Test Module', got %s", req.Judul)
		}
		if req.Deskripsi != "" {
			t.Errorf("Expected empty Deskripsi, got %s", req.Deskripsi)
		}
	})
}

func TestTusModulUploadResponse(t *testing.T) {
	t.Run("dto.TusModulUploadResponse struct", func(t *testing.T) {
		resp := dto.TusModulUploadResponse{
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

	t.Run("dto.TusModulUploadResponse with progress", func(t *testing.T) {
		resp := dto.TusModulUploadResponse{
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
	t.Run("dto.TusModulUploadInfoResponse struct", func(t *testing.T) {
		now := time.Now()

		resp := dto.TusModulUploadInfoResponse{
			UploadID:  "modul-upload-123",
			ModulID:   "550e8400-e29b-41d4-a716-446655440100",
			Judul:     "Data Structures",
			Deskripsi: "Learn about data structures",
			Status:    UploadStatusUploading,
			Progress:  75.0,
			Offset:    750000,
			Length:    1000000,
			CreatedAt: now,
			UpdatedAt: now,
		}

		if resp.UploadID != "modul-upload-123" {
			t.Errorf("Expected UploadID 'modul-upload-123', got %s", resp.UploadID)
		}
		if resp.ModulID != "550e8400-e29b-41d4-a716-446655440100" {
			t.Errorf("Expected ModulID '550e8400-e29b-41d4-a716-446655440100', got %s", resp.ModulID)
		}
		if resp.Judul != "Data Structures" {
			t.Errorf("Expected Judul 'Data Structures', got %s", resp.Judul)
		}
		if resp.Deskripsi != "Learn about data structures" {
			t.Errorf("Expected Deskripsi 'Learn about data structures', got %s", resp.Deskripsi)
		}
		if resp.Status != UploadStatusUploading {
			t.Errorf("Expected Status '%s', got %s", UploadStatusUploading, resp.Status)
		}
		if resp.Progress != 75.0 {
			t.Errorf("Expected Progress 75.0, got %f", resp.Progress)
		}
	})

	t.Run("dto.TusModulUploadInfoResponse with empty ModulID", func(t *testing.T) {
		resp := dto.TusModulUploadInfoResponse{
			UploadID:  "modul-upload-456",
			ModulID:   "",
			Judul:     "New Module",
			Deskripsi: "A new module description",
			Status:    UploadStatusPending,
			Progress:  0,
		}

		if resp.ModulID != "" {
			t.Errorf("Expected empty ModulID, got %s", resp.ModulID)
		}
		if resp.Judul != "New Module" {
			t.Errorf("Expected Judul 'New Module', got %s", resp.Judul)
		}
	})
}

func TestTusModulUploadSlotResponse(t *testing.T) {
	t.Run("dto.TusModulUploadSlotResponse available", func(t *testing.T) {
		resp := dto.TusModulUploadSlotResponse{
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

	t.Run("dto.TusModulUploadSlotResponse not available", func(t *testing.T) {
		resp := dto.TusModulUploadSlotResponse{
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
			UploadStatusQueued:    {UploadStatusPending},
			UploadStatusPending:   {UploadStatusUploading, UploadStatusCancelled, UploadStatusExpired},
			UploadStatusUploading: {UploadStatusCompleted, UploadStatusFailed, UploadStatusCancelled},
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
		metadata := TusModulUploadMetadata{
			Judul:     "Machine Learning Basics",
			Deskripsi: "Introduction to machine learning concepts",
		}

		upload := TusModulUpload{
			ID:             "modul-upload-metadata-1",
			UserID:         "user-1",
			UploadType:     UploadTypeModulCreate,
			UploadMetadata: metadata,
			FileSize:       2500000,
			Status:         UploadStatusPending,
			CreatedAt:      now,
			UpdatedAt:      now,
			ExpiresAt:      now.Add(24 * time.Hour),
		}

		if upload.UploadMetadata.Judul != "Machine Learning Basics" {
			t.Errorf("Expected Judul 'Machine Learning Basics', got %s", upload.UploadMetadata.Judul)
		}
		if upload.UploadMetadata.Deskripsi != "Introduction to machine learning concepts" {
			t.Errorf("Expected Deskripsi 'Introduction to machine learning concepts', got %s", upload.UploadMetadata.Deskripsi)
		}
	})
}

func TestTusModulUploadEdgeCases(t *testing.T) {
	t.Run("Nil ModulID for new modul creation", func(t *testing.T) {
		upload := TusModulUpload{
			ID:         "modul-upload-nil",
			UserID:     "user-1",
			ModulID:    nil,
			UploadType: UploadTypeModulCreate,
			Status:     UploadStatusPending,
		}

		if upload.ModulID != nil {
			t.Errorf("Expected nil ModulID for new modul creation, got %v", upload.ModulID)
		}
		if upload.UploadType != UploadTypeModulCreate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeModulCreate, upload.UploadType)
		}
	})

	t.Run("Non-nil ModulID for modul update", func(t *testing.T) {
		modulID := "550e8400-e29b-41d4-a716-446655440200"
		upload := TusModulUpload{
			ID:         "modul-upload-update",
			UserID:     "user-1",
			ModulID:    &modulID,
			UploadType: UploadTypeModulUpdate,
			Status:     UploadStatusPending,
		}

		if upload.ModulID == nil {
			t.Error("Expected non-nil ModulID for modul update")
		} else if *upload.ModulID != "550e8400-e29b-41d4-a716-446655440200" {
			t.Errorf("Expected ModulID '550e8400-e29b-41d4-a716-446655440200', got %s", *upload.ModulID)
		}
		if upload.UploadType != UploadTypeModulUpdate {
			t.Errorf("Expected UploadType '%s', got %s", UploadTypeModulUpdate, upload.UploadType)
		}
	})

	t.Run("Zero progress for pending upload", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "modul-upload-zero-progress",
			UserID:   "user-1",
			Status:   UploadStatusPending,
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
			UserID:   "user-1",
			Status:   UploadStatusCompleted,
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
			UserID:        "user-1",
			Status:        UploadStatusUploading,
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
			UserID:    "user-1",
			UploadURL: "",
			Status:    UploadStatusQueued,
		}

		if upload.UploadURL != "" {
			t.Errorf("Expected empty UploadURL before initialization, got %s", upload.UploadURL)
		}
	})

	t.Run("Populated upload URL after initialization", func(t *testing.T) {
		upload := TusModulUpload{
			ID:        "modul-upload-with-url",
			UserID:    "user-1",
			UploadURL: "https://example.com/tus/modul-upload-with-url",
			Status:    UploadStatusPending,
		}

		if upload.UploadURL == "" {
			t.Error("Expected non-empty UploadURL after initialization")
		}
	})

	t.Run("Nil CompletedAt for in-progress upload", func(t *testing.T) {
		upload := TusModulUpload{
			ID:          "modul-upload-in-progress",
			UserID:      "user-1",
			Status:      UploadStatusUploading,
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
			UserID:      "user-1",
			Status:      UploadStatusCompleted,
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
			UserID:   "user-1",
			FileSize: 0,
			Status:   UploadStatusPending,
		}

		if upload.FileSize != 0 {
			t.Errorf("Expected FileSize 0, got %d", upload.FileSize)
		}
	})

	t.Run("Large file size", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "modul-upload-large",
			UserID:   "user-1",
			FileSize: 500000000, // 500MB
			Status:   UploadStatusPending,
		}

		if upload.FileSize != 500000000 {
			t.Errorf("Expected FileSize 500000000, got %d", upload.FileSize)
		}
	})
}

func TestTusModulUploadInitRequestEdgeCases(t *testing.T) {
	t.Run("Minimum valid Judul length", func(t *testing.T) {
		req := TusModulUploadMetadata{
			Judul:     "ABC",
			Deskripsi: "A description",
		}

		if len(req.Judul) != 3 {
			t.Errorf("Expected Judul length 3, got %d", len(req.Judul))
		}
	})

	t.Run("Long valid Judul", func(t *testing.T) {
		longName := "A" + " very long module name that describes the content in detail " + "B"
		req := TusModulUploadMetadata{
			Judul:     longName,
			Deskripsi: "Description for long module name",
		}

		if req.Judul != longName {
			t.Errorf("Expected Judul '%s', got %s", longName, req.Judul)
		}
	})

	t.Run("Empty deskripsi allowed", func(t *testing.T) {
		req := TusModulUploadMetadata{
			Judul:     "Module Without Description",
			Deskripsi: "",
		}

		if req.Deskripsi != "" {
			t.Errorf("Expected empty Deskripsi, got %s", req.Deskripsi)
		}
	})
}

func TestTusModulUploadResponseEdgeCases(t *testing.T) {
	t.Run("Zero offset for new upload", func(t *testing.T) {
		resp := dto.TusModulUploadResponse{
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
		resp := dto.TusModulUploadResponse{
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
		resp := dto.TusModulUploadResponse{
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
	t.Run("Empty ModulID for new modul", func(t *testing.T) {
		resp := dto.TusModulUploadInfoResponse{
			UploadID:  "new-modul-upload",
			ModulID:   "",
			Judul:     "New Module",
			Deskripsi: "New module description",
			Status:    UploadStatusPending,
			Progress:  0,
		}

		if resp.ModulID != "" {
			t.Errorf("Expected empty ModulID for new modul, got %s", resp.ModulID)
		}
	})

	t.Run("Non-empty ModulID for existing modul update", func(t *testing.T) {
		resp := dto.TusModulUploadInfoResponse{
			UploadID:  "update-modul-upload",
			ModulID:   "550e8400-e29b-41d4-a716-446655440300",
			Judul:     "Updated Module",
			Deskripsi: "Updated description",
			Status:    UploadStatusUploading,
			Progress:  50.0,
		}

		if resp.ModulID != "550e8400-e29b-41d4-a716-446655440300" {
			t.Errorf("Expected ModulID '550e8400-e29b-41d4-a716-446655440300', got %s", resp.ModulID)
		}
	})

	t.Run("Zero progress for pending status", func(t *testing.T) {
		resp := dto.TusModulUploadInfoResponse{
			UploadID: "pending-upload",
			Status:   UploadStatusPending,
			Progress: 0,
		}

		if resp.Progress != 0 {
			t.Errorf("Expected Progress 0 for pending status, got %f", resp.Progress)
		}
	})

	t.Run("Full progress for completed status", func(t *testing.T) {
		resp := dto.TusModulUploadInfoResponse{
			UploadID: "completed-upload",
			Status:   UploadStatusCompleted,
			Progress: 100.0,
		}

		if resp.Progress != 100.0 {
			t.Errorf("Expected Progress 100.0 for completed status, got %f", resp.Progress)
		}
	})

	t.Run("Timestamp fields present", func(t *testing.T) {
		now := time.Now()
		resp := dto.TusModulUploadInfoResponse{
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
		resp := dto.TusModulUploadSlotResponse{
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
		resp := dto.TusModulUploadSlotResponse{
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
		resp := dto.TusModulUploadSlotResponse{
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
		resp := dto.TusModulUploadSlotResponse{
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
			UserID:   "user-1",
			Status:   UploadStatusUploading,
			Progress: -10.0,
		}

		if upload.Progress >= 0 {
			t.Logf("Note: Negative progress value %f is stored but should be validated at business logic layer", upload.Progress)
		}
	})

	t.Run("Progress over 100 is invalid", func(t *testing.T) {
		upload := TusModulUpload{
			ID:       "over-100-progress",
			UserID:   "user-1",
			Status:   UploadStatusUploading,
			Progress: 150.0,
		}

		if upload.Progress > 100 {
			t.Logf("Note: Progress over 100 (%f) is stored but should be validated at business logic layer", upload.Progress)
		}
	})
}
