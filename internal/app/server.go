package app

import (
	"fmt"
	"invento-service/config"
	_ "invento-service/docs"
	"invento-service/internal/constants"
	"invento-service/internal/controller/base"
	"invento-service/internal/controller/http"
	"invento-service/internal/helper"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/storage"
	supabaseAuth "invento-service/internal/supabase"
	"invento-service/internal/usecase"
	"invento-service/internal/usecase/repo"
	"io"
	"os"
	"runtime"
	"runtime/metrics"
	"strings"
	"time"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/supabase-community/supabase-go"
	"github.com/swaggo/fiber-swagger"
	"gorm.io/gorm"
)

// initLogger creates a zerolog.Logger configured for the given environment and level.
func initLogger(env, level string) zerolog.Logger {
	var output io.Writer
	if env == "development" {
		output = zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "15:04:05"}
	} else {
		output = os.Stderr
	}
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		lvl = zerolog.ErrorLevel // production default
	}
	return zerolog.New(output).Level(lvl).With().Timestamp().Str("service", "invento-service").Logger()
}

func NewServer(cfg *config.Config, db *gorm.DB) (*fiber.App, error) {
	// Initialize structured logger
	appLogger := initLogger(cfg.App.Env, cfg.Logging.Level)

	app := fiber.New(fiber.Config{
		ReadBufferSize:    cfg.Performance.FiberReadBufferSize,
		StreamRequestBody: cfg.Performance.FiberStreamRequestBody,
		Concurrency:       cfg.Performance.FiberConcurrency,
		ReduceMemoryUsage: cfg.Performance.FiberReduceMemory,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			appLogger.Error().Str("path", c.Path()).Str("method", c.Method()).Err(err).Msg("unhandled error")

			if c.Path() != "" && len(c.Path()) >= 8 && c.Path()[:8] == "/uploads" {
				return err
			}

			if err != nil {
				if e, ok := err.(*fiber.Error); ok {
					if e.Code == fiber.StatusNotFound {
						return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
							"status":    "error",
							"message":   "Endpoint tidak ditemukan",
							"code":      404,
							"timestamp": time.Now().Format(time.RFC3339),
						})
					}
				}
			}

			tusVersion := c.Get("Tus-Resumable")
			if tusVersion != "" && (c.Method() == "PATCH" || c.Method() == "HEAD" || c.Method() == "DELETE") {
				c.Set("Tus-Resumable", cfg.Upload.TusVersion)
				return c.SendStatus(fiber.StatusInternalServerError)
			}
			return httputil.SendInternalServerErrorResponse(c)
		},
	})

	// Conditionally enable pprof profiling endpoint
	if cfg.Performance.EnablePprof {
		app.Use(pprof.New())
		appLogger.Info().Msg("pprof profiling enabled at /debug/pprof/*")
	}

	// Apply middleware in order: RequestID -> Logger -> Recover -> CORS
	app.Use(middleware.RequestID())
	app.Use(fiberzerolog.New(fiberzerolog.Config{
		Logger:   &appLogger,
		SkipURIs: []string{"/health", "/uploads"},
	}))
	app.Use(recover.New())

	// Log startup information
	appLogger.Info().Str("app", cfg.App.Name).Str("env", cfg.App.Env).Str("port", cfg.App.Port).Msg("server starting")

	// Start background memory monitor
	if memLimit, err := config.ParseMemLimit(cfg.Performance.GoMemLimit); err == nil {
		startMemoryMonitor(cfg.Performance.MemoryWarningThreshold, memLimit, &appLogger)
		appLogger.Info().Str("gomemlimit", cfg.Performance.GoMemLimit).Float64("threshold_pct", cfg.Performance.MemoryWarningThreshold*100).Msg("memory monitor started")
	}

	corsOrigin := cfg.App.CorsOriginDev
	if cfg.App.Env == "production" {
		corsOrigin = cfg.App.CorsOriginProd
	}

	app.Use(cors.New(cors.Config{
		AllowOrigins:     corsOrigin,
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Tus-Resumable, Upload-Length, Upload-Metadata, Upload-Offset, Content-Length",
		AllowMethods:     "GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS",
		AllowCredentials: true,
		ExposeHeaders:    "Tus-Resumable, Tus-Version, Tus-Extension, Tus-Max-Size, Upload-Offset, Upload-Length, Location",
	}))

	pathResolver := storage.NewPathResolver(cfg)
	cookieHelper := httputil.NewCookieHelper(cfg)
	app.Static("/uploads", pathResolver.GetBasePath())

	userRepo := repo.NewUserRepository(db)
	roleRepo := repo.NewRoleRepository(db)
	permissionRepo := repo.NewPermissionRepository(db)
	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	projectRepo := repo.NewProjectRepository(db)
	modulRepo := repo.NewModulRepository(db, appLogger)
	tusUploadRepo := repo.NewTusUploadRepository(db)
	tusModulUploadRepo := repo.NewTusModulUploadRepository(db)

	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceKey, nil)
	if err != nil {
		return nil, fmt.Errorf("supabase client init: %w", err)
	}

	supabaseServiceKey := cfg.Supabase.ServiceKey
	authURL := cfg.Supabase.URL + "/auth/v1"
	supabaseAuthService, err := supabaseAuth.NewAuthService(authURL, supabaseServiceKey)
	if err != nil {
		return nil, fmt.Errorf("supabase auth service init: %w", err)
	}

	casbinEnforcer, err := helper.NewCasbinEnforcer(db)
	if err != nil {
		return nil, fmt.Errorf("casbin enforcer init: %w", err)
	}
	tusProjectStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSizeProject)
	tusModulStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSizeModul)
	tusQueue := helper.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	tusModulQueue := helper.NewTusQueue(cfg.Upload.MaxQueueModulPerUser)
	fileManager := storage.NewFileManager(cfg)
	tusProjectManager := helper.NewTusManager(tusProjectStore, tusQueue, fileManager, cfg, appLogger)
	tusModulManager := helper.NewTusManager(tusModulStore, tusModulQueue, fileManager, cfg, appLogger)

	if activeIDs, err := tusUploadRepo.GetActiveUploadIDs(); err == nil && len(activeIDs) > 0 {
		tusQueue.LoadFromDB(activeIDs)
	}
	if activeIDs, err := tusModulUploadRepo.GetActiveUploadIDs(); err == nil && len(activeIDs) > 0 {
		tusModulQueue.LoadFromDB(activeIDs)
	}

	authUsecase, err := usecase.NewAuthUsecase(userRepo, roleRepo, supabaseClient, supabaseServiceKey, cfg, appLogger)
	if err != nil {
		return nil, fmt.Errorf("auth usecase init: %w", err)
	}
	authController := http.NewAuthController(authUsecase, cookieHelper, cfg, appLogger)

	roleUsecase := usecase.NewRoleUsecase(roleRepo, permissionRepo, rolePermissionRepo, casbinEnforcer)
	baseCtrl := base.NewBaseController(cfg.Supabase.URL, casbinEnforcer)
	roleController := http.NewRoleController(roleUsecase, baseCtrl)

	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, projectRepo, modulRepo, casbinEnforcer, pathResolver, cfg)
	userController := http.NewUserController(userUsecase)

	projectUsecase := usecase.NewProjectUsecase(projectRepo, fileManager)
	projectController := http.NewProjectController(projectUsecase, cfg.Supabase.URL, casbinEnforcer)

	tusUploadUsecase := usecase.NewTusUploadUsecase(tusUploadRepo, projectRepo, projectUsecase, tusProjectManager, fileManager, cfg)
	tusController := http.NewTusController(tusUploadUsecase, cfg, baseCtrl)

	modulUsecase := usecase.NewModulUsecase(modulRepo)
	tusModulUsecase := usecase.NewTusModulUsecase(tusModulUploadRepo, modulRepo, tusModulManager, fileManager, cfg)
	modulController := http.NewModulController(modulUsecase, cfg, baseCtrl)
	tusModulController := http.NewTusModulController(tusModulUsecase, cfg, baseCtrl)

	tusCleanup := helper.NewTusCleanup(tusUploadRepo, tusModulUploadRepo, tusProjectStore, tusModulStore, cfg.Upload.CleanupInterval, cfg.Upload.IdleTimeout, appLogger)
	tusCleanup.Start()

	statisticUsecase := usecase.NewStatisticUsecase(userRepo, projectRepo, modulRepo, roleRepo, casbinEnforcer, db)
	statisticController := http.NewStatisticController(statisticUsecase)

	healthUsecase := usecase.NewHealthUsecase(db, cfg)
	healthController := http.NewHealthController(healthUsecase)

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/login", authController.Login)
	auth.Post("/register", authController.Register)
	auth.Post("/refresh", authController.RefreshToken)
	auth.Post("/reset-password", authController.RequestPasswordReset)

	protected := auth.Group("/", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	protected.Post("logout", authController.Logout)

	role := api.Group("/role", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	role.Get("/permissions", helper.RBACMiddleware(casbinEnforcer, constants.ResourcePermission, constants.ActionRead), roleController.GetAvailablePermissions)
	role.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionRead), roleController.GetRoleList)
	role.Post("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionCreate), roleController.CreateRole)
	role.Get("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionRead), roleController.GetRoleDetail)
	role.Put("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionUpdate), roleController.UpdateRole)
	role.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionDelete), roleController.DeleteRole)
	role.Get("/:id/users", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionRead), userController.GetUsersForRole)
	role.Post("/:id/users/bulk", helper.RBACMiddleware(casbinEnforcer, constants.ResourceRole, constants.ActionUpdate), userController.BulkAssignRole)

	user := api.Group("/user", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	user.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceUser, constants.ActionRead), userController.GetUserList)
	user.Put("/:id/role", helper.RBACMiddleware(casbinEnforcer, constants.ResourceUser, constants.ActionUpdate), userController.UpdateUserRole)
	user.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceUser, constants.ActionDelete), userController.DeleteUser)
	user.Get("/:id/files", helper.RBACMiddleware(casbinEnforcer, constants.ResourceUser, constants.ActionRead), userController.GetUserFiles)
	user.Post("/:id/download", helper.RBACMiddleware(casbinEnforcer, constants.ResourceUser, constants.ActionDownload), userController.DownloadUserFiles)
	user.Get("/permissions", userController.GetUserPermissions)

	profile := api.Group("/profile", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	profile.Get("/", userController.GetProfile)
	profile.Put("/", userController.UpdateProfile)

	project := api.Group("/project", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	project.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), projectController.GetList)
	project.Get("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), projectController.GetByID)
	project.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionUpdate), projectController.UpdateMetadata)
	project.Post("/download", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), projectController.Download)
	project.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionDelete), projectController.Delete)

	tusUploadCheck := api.Group("/project/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	tusUploadCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), tusController.CheckUploadSlot)
	tusUploadCheck.Post("/reset-queue", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionCreate), tusController.ResetUploadQueue)

	tusUpload := api.Group("/project/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper), helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject))
	tusUpload.Post("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionCreate), tusController.InitiateUpload)
	tusUpload.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionCreate), tusController.UploadChunk)
	tusUpload.Head("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), tusController.GetUploadStatus)
	tusUpload.Get("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), tusController.GetUploadInfo)
	tusUpload.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionDelete), tusController.CancelUpload)

	projectUpdate := api.Group("/project/:id", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	projectUpdate.Post("/upload", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionUpdate), tusController.InitiateProjectUpdateUpload)
	projectUpdate.Patch("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionUpdate), tusController.UploadProjectUpdateChunk)
	projectUpdate.Head("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), tusController.GetProjectUpdateUploadStatus)
	projectUpdate.Get("/update/:upload_id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionRead), tusController.GetProjectUpdateUploadInfo)
	projectUpdate.Delete("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, constants.ResourceProject, constants.ActionUpdate), tusController.CancelProjectUpdateUpload)

	modul := api.Group("/modul", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	modul.Get("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), modulController.GetList)
	modul.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionUpdate), modulController.UpdateMetadata)
	modul.Post("/download", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), modulController.Download)
	modul.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionDelete), modulController.Delete)

	tusModulCheck := api.Group("/modul/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	tusModulCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), tusModulController.CheckUploadSlot)

	tusModul := api.Group("/modul/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper), helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul))
	tusModul.Post("/", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionCreate), tusModulController.InitiateUpload)
	tusModul.Patch("/:upload_id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionCreate), tusModulController.UploadChunk)
	tusModul.Head("/:upload_id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), tusModulController.GetUploadStatus)
	tusModul.Get("/:upload_id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), tusModulController.GetUploadInfo)
	tusModul.Delete("/:upload_id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionDelete), tusModulController.CancelUpload)

	modulUpdate := api.Group("/modul/:id", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	modulUpdate.Post("/upload", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionUpdate), tusModulController.InitiateModulUpdateUpload)
	modulUpdate.Patch("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionUpdate), tusModulController.UploadModulUpdateChunk)
	modulUpdate.Head("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), tusModulController.GetModulUpdateUploadStatus)
	modulUpdate.Get("/update/:upload_id", helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionRead), tusModulController.GetModulUpdateUploadInfo)
	modulUpdate.Delete("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, constants.ResourceModul, constants.ActionUpdate), tusModulController.CancelModulUpdateUpload)

	statistic := api.Group("/statistic", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	statistic.Get("/", statisticController.GetStatistics)

	monitoring := api.Group("/monitoring")
	monitoring.Get("/health", healthController.ComprehensiveHealthCheck)
	monitoring.Get("/metrics", healthController.GetSystemMetrics)
	monitoring.Get("/status", healthController.GetApplicationStatus)

	app.Get("/health", healthController.BasicHealthCheck)

	// Enable Swagger UI
	appLogger.Info().Msg("swagger UI enabled at /swagger/*")
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	return app, nil
}

// startMemoryMonitor starts a background goroutine that periodically checks heap
// memory usage and logs a warning when it exceeds the threshold percentage of GOMEMLIMIT.
// Uses runtime/metrics instead of runtime.ReadMemStats to avoid stop-the-world pauses.
func startMemoryMonitor(thresholdPct float64, limitBytes int64, appLogger *zerolog.Logger) {
	if limitBytes <= 0 || thresholdPct <= 0 {
		return
	}
	thresholdBytes := int64(float64(limitBytes) * thresholdPct)

	go func() {
		samples := []metrics.Sample{
			{Name: "/memory/classes/heap/objects:bytes"},
		}
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			metrics.Read(samples)
			heapBytes := int64(samples[0].Value.Uint64())
			if heapBytes > thresholdBytes {
				appLogger.Warn().
					Int64("heap_bytes", heapBytes).
					Int64("heap_mb", heapBytes/(1024*1024)).
					Int64("limit_mb", limitBytes/(1024*1024)).
					Float64("threshold_pct", thresholdPct*100).
					Int("goroutines", runtime.NumGoroutine()).
					Msg("heap memory approaching GOMEMLIMIT")
			}
		}
	}()
}
