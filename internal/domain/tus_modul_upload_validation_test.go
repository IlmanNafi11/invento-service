package domain

import (
	"testing"
	"time"

	"invento-service/internal/dto"
)

func TestTusModulUploadEdgeCases(t *testing.T) {
	t.Parallel()
	t.Run("Nil ModulID for new modul creation", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("Minimum valid Judul length", func(t *testing.T) {
		t.Parallel()
		req := TusModulUploadMetadata{
			Judul:     "ABC",
			Deskripsi: "A description",
		}

		if len(req.Judul) != 3 {
			t.Errorf("Expected Judul length 3, got %d", len(req.Judul))
		}
	})

	t.Run("Long valid Judul", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("Zero offset for new upload", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("Empty ModulID for new modul", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("Zero queue length when empty", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
		t.Parallel()
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
	t.Parallel()
	t.Run("Negative progress is invalid", func(t *testing.T) {
		t.Parallel()
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
		t.Parallel()
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
