package usecase

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/domain"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase/repo"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type tusIntegrationEnv struct {
	cfg            *config.Config
	db             *gorm.DB
	userID         string
	pathResolver   *helper.PathResolver
	projectQueue   *helper.TusQueue
	modulQueue     *helper.TusQueue
	projectManager *helper.TusManager
	modulManager   *helper.TusManager
	uploadUsecase  *tusUploadUsecase
	modulUsecase   *tusModulUsecase
}

func setupTusIntegrationTest(t *testing.T) *tusIntegrationEnv {
	t.Helper()

	baseDir := t.TempDir()
	cfg := &config.Config{
		Upload: config.UploadConfig{
			MaxSize:              5 * 1024 * 1024,
			MaxSizeProject:       5 * 1024 * 1024,
			MaxSizeModul:         2 * 1024 * 1024,
			ChunkSize:            1024,
			MaxConcurrent:        1,
			MaxConcurrentProject: 2,
			MaxConcurrentModul:   2,
			MaxQueueModulPerUser: 3,
			IdleTimeout:          600,
			CleanupInterval:      300,
			TusVersion:           "1.0.0",
			MaxResumeAttempts:    10,
			PathDevelopment:      filepath.Join(baseDir, "uploads"),
			TempPathDevelopment:  filepath.Join(baseDir, "temp"),
		},
		App: config.AppConfig{
			Env: "development",
		},
	}

	dsnSafeName := strings.ReplaceAll(t.Name(), "/", "_")
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", dsnSafeName)
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&domain.User{},
		&domain.Project{},
		&domain.Modul{},
		&domain.TusUpload{},
		&domain.TusModulUpload{},
	)
	require.NoError(t, err)

	userID := "11111111-1111-1111-1111-111111111111"
	require.NoError(t, db.Create(&domain.User{
		ID:       userID,
		Email:    "integration@test.local",
		Name:     "Integration User",
		IsActive: true,
	}).Error)

	pathResolver := helper.NewPathResolver(cfg)
	fileManager := helper.NewFileManager(cfg)

	projectStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSizeProject)
	modulStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSizeModul)

	projectQueue := helper.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	modulQueue := helper.NewTusQueue(cfg.Upload.MaxConcurrentModul)

	projectManager := helper.NewTusManager(projectStore, projectQueue, fileManager, cfg)
	modulManager := helper.NewTusManager(modulStore, modulQueue, fileManager, cfg)

	tusUploadRepo := repo.NewTusUploadRepository(db)
	projectRepo := repo.NewProjectRepository(db)
	tusModulUploadRepo := repo.NewTusModulUploadRepository(db)
	modulRepo := repo.NewModulRepository(db)

	uploadUsecase := NewTusUploadUsecase(
		tusUploadRepo,
		projectRepo,
		nil,
		projectManager,
		fileManager,
		cfg,
	).(*tusUploadUsecase)

	modulUsecase := NewTusModulUsecase(
		tusModulUploadRepo,
		modulRepo,
		modulManager,
		fileManager,
		cfg,
	).(*tusModulUsecase)

	return &tusIntegrationEnv{
		cfg:            cfg,
		db:             db,
		userID:         userID,
		pathResolver:   pathResolver,
		projectQueue:   projectQueue,
		modulQueue:     modulQueue,
		projectManager: projectManager,
		modulManager:   modulManager,
		uploadUsecase:  uploadUsecase,
		modulUsecase:   modulUsecase,
	}
}

func createTestChunk(size int) io.Reader {
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	return bytes.NewReader(data)
}

func integrationModulMetadataHeader(judul, deskripsi string) string {
	enc := func(v string) string {
		return base64.StdEncoding.EncodeToString([]byte(v))
	}
	return fmt.Sprintf("judul %s,deskripsi %s", enc(judul), enc(deskripsi))
}

func TestTusProjectUploadFullFlowIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.uploadUsecase.InitiateUpload(
		env.userID,
		"integration@test.local",
		"mahasiswa",
		3*1024,
		domain.TusUploadInitRequest{NamaProject: "Project Integration", Kategori: "website", Semester: 2},
	)
	require.NoError(t, err)
	require.NotNil(t, resp)

	info, err := env.uploadUsecase.GetUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.UploadStatusPending, info.Status)

	offset, err := env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(1024), offset)

	info, err = env.uploadUsecase.GetUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.UploadStatusUploading, info.Status)
	assert.Equal(t, int64(1024), info.Offset)

	offset, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, offset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(2048), offset)

	offset, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, offset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(3072), offset)

	var upload domain.TusUpload
	require.NoError(t, env.db.Where("id = ?", resp.UploadID).First(&upload).Error)
	assert.Equal(t, domain.UploadStatusCompleted, upload.Status)
	assert.Equal(t, int64(3072), upload.CurrentOffset)

	var project domain.Project
	require.NoError(t, env.db.Where("user_id = ?", env.userID).First(&project).Error)
	assert.FileExists(t, project.PathFile)

	fileInfo, err := os.Stat(project.PathFile)
	require.NoError(t, err)
	assert.Equal(t, int64(3072), fileInfo.Size())
}

func TestTusProjectUploadResumeAfterPauseIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.uploadUsecase.InitiateUpload(
		env.userID,
		"integration@test.local",
		"mahasiswa",
		3*1024,
		domain.TusUploadInitRequest{NamaProject: "Resume Project", Kategori: "mobile", Semester: 4},
	)
	require.NoError(t, err)

	offset, err := env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(1024), offset)

	pausedOffset, length, err := env.uploadUsecase.GetUploadStatus(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1024), pausedOffset)
	assert.Equal(t, int64(3072), length)

	offset, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, pausedOffset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(2048), offset)

	offset, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, offset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(3072), offset)

	var upload domain.TusUpload
	require.NoError(t, env.db.Where("id = ?", resp.UploadID).First(&upload).Error)
	assert.Equal(t, domain.UploadStatusCompleted, upload.Status)
}

func TestTusProjectUploadCancelIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.uploadUsecase.InitiateUpload(
		env.userID,
		"integration@test.local",
		"mahasiswa",
		2*1024,
		domain.TusUploadInitRequest{NamaProject: "Cancel Project", Kategori: "iot", Semester: 5},
	)
	require.NoError(t, err)

	_, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)

	tempUploadPath := env.pathResolver.GetUploadPath(resp.UploadID)
	assert.DirExists(t, tempUploadPath)

	err = env.uploadUsecase.CancelUpload(resp.UploadID, env.userID)
	require.NoError(t, err)

	var upload domain.TusUpload
	require.NoError(t, env.db.Where("id = ?", resp.UploadID).First(&upload).Error)
	assert.Equal(t, domain.UploadStatusCancelled, upload.Status)

	_, statErr := os.Stat(tempUploadPath)
	assert.True(t, os.IsNotExist(statErr))
	assert.Equal(t, 0, env.projectQueue.GetActiveCount())
	assert.True(t, env.projectQueue.CanAcceptUpload())
}

func TestTusModulUploadFullFlowIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.modulUsecase.InitiateModulUpload(
		env.userID,
		2*1024,
		integrationModulMetadataHeader("modul-integration", "deskripsi integration"),
	)
	require.NoError(t, err)
	require.NotNil(t, resp)

	info, err := env.modulUsecase.GetModulUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.ModulUploadStatusPending, info.Status)

	offset, err := env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(1024), offset)

	info, err = env.modulUsecase.GetModulUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.ModulUploadStatusUploading, info.Status)
	assert.Equal(t, int64(1024), info.Offset)

	offset, err = env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, offset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(2048), offset)

	var upload domain.TusModulUpload
	require.NoError(t, env.db.Where("id = ?", resp.UploadID).First(&upload).Error)
	assert.Equal(t, domain.ModulUploadStatusCompleted, upload.Status)
	require.NotNil(t, upload.ModulID)

	var modul domain.Modul
	require.NoError(t, env.db.Where("id = ?", *upload.ModulID).First(&modul).Error)
	assert.FileExists(t, modul.FilePath)

	fileInfo, err := os.Stat(modul.FilePath)
	require.NoError(t, err)
	assert.Equal(t, int64(2048), fileInfo.Size())
}

func TestTusModulUploadResumeAfterPauseIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.modulUsecase.InitiateModulUpload(
		env.userID,
		3*1024,
		integrationModulMetadataHeader("modul-resume", "deskripsi resume"),
	)
	require.NoError(t, err)

	offset, err := env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(1024), offset)

	pausedOffset, length, err := env.modulUsecase.GetModulUploadStatus(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, int64(1024), pausedOffset)
	assert.Equal(t, int64(3072), length)

	offset, err = env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, pausedOffset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(2048), offset)

	offset, err = env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, offset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(3072), offset)

	var upload domain.TusModulUpload
	require.NoError(t, env.db.Where("id = ?", resp.UploadID).First(&upload).Error)
	assert.Equal(t, domain.ModulUploadStatusCompleted, upload.Status)
}

func TestTusModulUploadCancelIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.modulUsecase.InitiateModulUpload(
		env.userID,
		2*1024,
		integrationModulMetadataHeader("modul-cancel", "deskripsi cancel"),
	)
	require.NoError(t, err)

	_, err = env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)

	tempUploadPath := env.pathResolver.GetUploadPath(resp.UploadID)
	assert.DirExists(t, tempUploadPath)

	err = env.modulUsecase.CancelModulUpload(resp.UploadID, env.userID)
	require.NoError(t, err)

	var upload domain.TusModulUpload
	require.NoError(t, env.db.Where("id = ?", resp.UploadID).First(&upload).Error)
	assert.Equal(t, domain.ModulUploadStatusCancelled, upload.Status)

	_, statErr := os.Stat(tempUploadPath)
	assert.True(t, os.IsNotExist(statErr))
}

func TestTusUploadConcurrentSlotsIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	meta := domain.TusUploadInitRequest{NamaProject: "Concurrent", Kategori: "website", Semester: 1}

	first, err := env.uploadUsecase.InitiateUpload(env.userID, "integration@test.local", "mahasiswa", 1024, meta)
	require.NoError(t, err)
	second, err := env.uploadUsecase.InitiateUpload(env.userID, "integration@test.local", "mahasiswa", 1024, meta)
	require.NoError(t, err)

	_, err = env.uploadUsecase.InitiateUpload(env.userID, "integration@test.local", "mahasiswa", 1024, meta)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "slot upload tidak tersedia")

	queuedID := "manually-queued-upload"
	env.projectQueue.Add(queuedID)
	assert.Equal(t, 1, env.projectQueue.GetQueueLength())

	promoted := env.projectQueue.FinishUpload(first.UploadID)
	assert.Equal(t, queuedID, promoted)
	assert.True(t, env.projectQueue.IsActiveUpload(queuedID))
	assert.True(t, env.projectQueue.IsActiveUpload(second.UploadID))
}

func TestTusUploadInvalidFileSizeIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	_, err := env.uploadUsecase.InitiateUpload(
		env.userID,
		"integration@test.local",
		"mahasiswa",
		env.cfg.Upload.MaxSizeProject+1,
		domain.TusUploadInitRequest{NamaProject: "Too Big", Kategori: "website", Semester: 1},
	)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ukuran file melebihi batas maksimal")
}

func TestTusUploadOffsetMismatchIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.modulUsecase.InitiateModulUpload(
		env.userID,
		2*1024,
		integrationModulMetadataHeader("offset-check", "deskripsi offset check"),
	)
	require.NoError(t, err)

	offset, err := env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(1024), offset)

	returnedOffset, err := env.modulUsecase.HandleModulChunk(resp.UploadID, env.userID, 0, createTestChunk(512))
	require.Error(t, err)
	assert.Equal(t, int64(1024), returnedOffset)
	assert.Contains(t, err.Error(), "Offset")
}

func TestTusUploadStatusTransitionsIntegration(t *testing.T) {
	env := setupTusIntegrationTest(t)

	resp, err := env.uploadUsecase.InitiateUpload(
		env.userID,
		"integration@test.local",
		"mahasiswa",
		2*1024,
		domain.TusUploadInitRequest{NamaProject: "Status Flow", Kategori: "deep_learning", Semester: 7},
	)
	require.NoError(t, err)

	info, err := env.uploadUsecase.GetUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.UploadStatusPending, info.Status)

	offset, err := env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, 0, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(1024), offset)

	info, err = env.uploadUsecase.GetUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.UploadStatusUploading, info.Status)

	offset, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, offset, createTestChunk(1024))
	require.NoError(t, err)
	assert.Equal(t, int64(2048), offset)

	info, err = env.uploadUsecase.GetUploadInfo(resp.UploadID, env.userID)
	require.NoError(t, err)
	assert.Equal(t, domain.UploadStatusCompleted, info.Status)

	_, err = env.uploadUsecase.HandleChunk(resp.UploadID, env.userID, offset, createTestChunk(128))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "sudah selesai")

	cancelResp, err := env.uploadUsecase.InitiateUpload(
		env.userID,
		"integration@test.local",
		"mahasiswa",
		1024,
		domain.TusUploadInitRequest{NamaProject: "Cancel Flow", Kategori: "machine_learning", Semester: 8},
	)
	require.NoError(t, err)

	err = env.uploadUsecase.CancelUpload(cancelResp.UploadID, env.userID)
	require.NoError(t, err)

	var cancelledUpload domain.TusUpload
	require.NoError(t, env.db.Where("id = ?", cancelResp.UploadID).First(&cancelledUpload).Error)
	assert.Equal(t, domain.UploadStatusCancelled, cancelledUpload.Status)

	_, err = env.uploadUsecase.HandleChunk(cancelResp.UploadID, env.userID, 0, createTestChunk(128))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "tidak aktif")
}
