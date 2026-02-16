package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	httpcontroller "invento-service/internal/controller/http"
	"invento-service/internal/domain"
	dto "invento-service/internal/dto"
	app_testing "invento-service/internal/testing"
	"invento-service/internal/upload"
)

func TestTusModulController_GetUploadInfo_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusModulControllerUsecase)
	baseCtrl := getTusBaseController()
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusModulController(mockUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Get("/api/v1/tus/modul/:upload_id", controller.GetUploadInfo)

	mockUC.On("GetModulUploadInfo", mock.Anything, "modul-upload-id", "user-123").Return(&dto.TusModulUploadInfoResponse{
		UploadID:  "modul-upload-id",
		Judul:     "Modul Dasar",
		Deskripsi: "Deskripsi modul",
		Status:    domain.UploadStatusUploading,
		Progress:  25,
		Offset:    256,
		Length:    1024,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}, nil)

	resp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/modul/modul-upload-id", nil, "")
	app_testing.AssertSuccess(t, resp)

	mockUC.AssertExpectations(t)
}

func TestTusModulController_CancelUpload_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusModulControllerUsecase)
	baseCtrl := getTusBaseController()
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusModulController(mockUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Delete("/api/v1/tus/modul/:upload_id", controller.CancelUpload)

	mockUC.On("CancelModulUpload", mock.Anything, "modul-upload-id", "user-123").Return(nil)

	req := httptest.NewRequest("DELETE", "/api/v1/tus/modul/modul-upload-id", http.NoBody)
	req.Header.Set("Tus-Resumable", "1.0.0")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 204, resp.StatusCode)

	mockUC.AssertExpectations(t)
}

func TestTusModulController_ModulUpdateEndpoints_Success(t *testing.T) {
	t.Parallel()
	mockUC := new(MockTusModulControllerUsecase)
	baseCtrl := getTusBaseController()
	cfg := getTusTestConfig()
	controller := httpcontroller.NewTusModulController(mockUC, cfg, baseCtrl)

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		setTusAuthenticatedUser(c, "user-123", "test@example.com", "user")
		return c.Next()
	})
	app.Post("/api/v1/tus/modul/:id/update", controller.InitiateModulUpdateUpload)
	app.Patch("/api/v1/tus/modul/:id/update/:upload_id", controller.UploadModulUpdateChunk)
	app.Head("/api/v1/tus/modul/:id/update/:upload_id", controller.GetModulUpdateUploadStatus)
	app.Get("/api/v1/tus/modul/:id/update/:upload_id", controller.GetModulUpdateUploadInfo)
	app.Delete("/api/v1/tus/modul/:id/update/:upload_id", controller.CancelModulUpdateUpload)

	modulID := "550e8400-e29b-41d4-a716-446655440099"
	metadataHeader := encodeTusMetadata(map[string]string{"judul": "Modul Rev", "deskripsi": "rev"})

	mockUC.On("InitiateModulUpdateUpload", mock.Anything, modulID, "user-123", int64(4096), metadataHeader).Return(&dto.TusModulUploadResponse{
		UploadID:  "update-upload-id",
		UploadURL: "/modul/" + modulID + "/update/update-upload-id",
		Offset:    0,
		Length:    4096,
	}, nil).Once()
	mockUC.On("HandleModulUpdateChunk", mock.Anything, modulID, "update-upload-id", "user-123", int64(0), mock.Anything).Return(int64(4), nil).Once()
	mockUC.On("GetModulUpdateUploadStatus", mock.Anything, modulID, "update-upload-id", "user-123").Return(int64(4), int64(4096), nil).Once()
	mockUC.On("GetModulUpdateUploadInfo", mock.Anything, modulID, "update-upload-id", "user-123").Return(&dto.TusModulUploadInfoResponse{UploadID: "update-upload-id"}, nil).Once()
	mockUC.On("CancelModulUpdateUpload", mock.Anything, modulID, "update-upload-id", "user-123").Return(nil).Once()

	initReq := httptest.NewRequest("POST", "/api/v1/tus/modul/"+modulID+"/update", http.NoBody)
	initReq.Header.Set("Tus-Resumable", "1.0.0")
	initReq.Header.Set("Upload-Length", "4096")
	initReq.Header.Set("Upload-Metadata", metadataHeader)
	initResp, initErr := app.Test(initReq)
	assert.NoError(t, initErr)
	assert.Equal(t, 201, initResp.StatusCode)

	chunkReq := httptest.NewRequest("PATCH", "/api/v1/tus/modul/"+modulID+"/update/update-upload-id", bytes.NewReader([]byte("test")))
	chunkReq.Header.Set("Tus-Resumable", "1.0.0")
	chunkReq.Header.Set("Upload-Offset", "0")
	chunkReq.Header.Set("Content-Type", upload.TusContentType)
	chunkReq.Header.Set("Content-Length", "4")
	chunkResp, chunkErr := app.Test(chunkReq)
	assert.NoError(t, chunkErr)
	assert.Equal(t, 204, chunkResp.StatusCode)

	headReq := httptest.NewRequest("HEAD", "/api/v1/tus/modul/"+modulID+"/update/update-upload-id", http.NoBody)
	headReq.Header.Set("Tus-Resumable", "1.0.0")
	headResp, headErr := app.Test(headReq)
	assert.NoError(t, headErr)
	assert.Equal(t, 200, headResp.StatusCode)

	infoResp := app_testing.MakeRequest(app, "GET", "/api/v1/tus/modul/"+modulID+"/update/update-upload-id", nil, "")
	app_testing.AssertSuccess(t, infoResp)

	deleteReq := httptest.NewRequest("DELETE", "/api/v1/tus/modul/"+modulID+"/update/update-upload-id", http.NoBody)
	deleteReq.Header.Set("Tus-Resumable", "1.0.0")
	deleteResp, deleteErr := app.Test(deleteReq)
	assert.NoError(t, deleteErr)
	assert.Equal(t, 204, deleteResp.StatusCode)

	mockUC.AssertExpectations(t)
}
