package app

import (
	"context"
	"errors"
	"fmt"
	"invento-service/config"
	"invento-service/internal/controller/base"
	"invento-service/internal/controller/http"
	"invento-service/internal/dto"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/rbac"
	"invento-service/internal/storage"
	"invento-service/internal/upload"
	"invento-service/internal/usecase"
	"invento-service/internal/usecase/repo"
	"io"
	"os"
	"runtime"
	"runtime/metrics"
	"strings"
	"time"

	_ "invento-service/docs"

	apperrors "invento-service/internal/errors"

	supabaseAuth "invento-service/internal/supabase"

	"github.com/gofiber/contrib/fiberzerolog"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/pprof"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/rs/zerolog"
	"github.com/supabase-community/supabase-go"
	"gorm.io/gorm"
)

// initLogger creates a zerolog.Logger configured for the given environment and level.
func initLogger(env, level string) zerolog.Logger {
	var output io.Writer
	if env == config.EnvDevelopment {
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

			// Check for AppError first (application-level structured errors)
			var appErr *apperrors.AppError
			if errors.As(err, &appErr) {
				return c.Status(appErr.HTTPStatus).JSON(dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Status:  "error",
						Message: appErr.Message,
						Code:    appErr.HTTPStatus,
					},
					Timestamp: time.Now(),
				})
			}

			// Check for Fiber framework errors
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				if fiberErr.Code == fiber.StatusNotFound {
					return c.Status(fiber.StatusNotFound).JSON(dto.ErrorResponse{
						BaseResponse: dto.BaseResponse{
							Status:  "error",
							Message: "Endpoint tidak ditemukan",
							Code:    fiber.StatusNotFound,
						},
						Timestamp: time.Now(),
					})
				}
				return c.Status(fiberErr.Code).JSON(dto.ErrorResponse{
					BaseResponse: dto.BaseResponse{
						Status:  "error",
						Message: fiberErr.Message,
						Code:    fiberErr.Code,
					},
					Timestamp: time.Now(),
				})
			}

			// TUS protocol error handling
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
	if cfg.App.Env == config.EnvProduction {
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

	casbinEnforcer, err := rbac.NewCasbinEnforcer(db)
	if err != nil {
		return nil, fmt.Errorf("casbin enforcer init: %w", err)
	}
	tusProjectStore := upload.NewTusStore(pathResolver, cfg.Upload.MaxSizeProject)
	tusModulStore := upload.NewTusStore(pathResolver, cfg.Upload.MaxSizeModul)
	tusQueue := upload.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	tusModulQueue := upload.NewTusQueue(cfg.Upload.MaxQueueModulPerUser)
	fileManager := storage.NewFileManager(cfg)
	tusProjectManager := upload.NewTusManager(tusProjectStore, tusQueue, fileManager, cfg, appLogger)
	tusModulManager := upload.NewTusManager(tusModulStore, tusModulQueue, fileManager, cfg, appLogger)

	if activeIDs, activeErr := tusUploadRepo.GetActiveUploadIDs(context.Background()); activeErr == nil && len(activeIDs) > 0 {
		tusQueue.LoadFromDB(activeIDs)
	}
	if activeIDs, activeErr := tusModulUploadRepo.GetActiveUploadIDs(context.Background()); activeErr == nil && len(activeIDs) > 0 {
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

	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, projectRepo, modulRepo, supabaseAuthService, casbinEnforcer, pathResolver, cfg, appLogger)
	userController := http.NewUserController(userUsecase)

	projectUsecase := usecase.NewProjectUsecase(projectRepo, fileManager)
	projectController := http.NewProjectController(projectUsecase, cfg.Supabase.URL, casbinEnforcer)

	tusUploadUsecase := usecase.NewTusUploadUsecase(tusUploadRepo, projectRepo, projectUsecase, tusProjectManager, fileManager, cfg)
	tusController := http.NewTusController(tusUploadUsecase, cfg, baseCtrl)

	modulUsecase := usecase.NewModulUsecase(modulRepo)
	tusModulUsecase := usecase.NewTusModulUsecase(tusModulUploadRepo, modulRepo, tusModulManager, fileManager, cfg)
	modulController := http.NewModulController(modulUsecase, cfg, baseCtrl)
	tusModulController := http.NewTusModulController(tusModulUsecase, cfg, baseCtrl)

	tusCleanup := upload.NewTusCleanup(tusUploadRepo, tusModulUploadRepo, tusProjectStore, tusModulStore, cfg.Upload.CleanupInterval, cfg.Upload.IdleTimeout, appLogger)
	tusCleanup.Start()

	statisticUsecase := usecase.NewStatisticUsecase(userRepo, projectRepo, modulRepo, roleRepo, casbinEnforcer, db)
	statisticController := http.NewStatisticController(statisticUsecase)

	healthUsecase := usecase.NewHealthUsecase(db, cfg)
	healthController := http.NewHealthController(healthUsecase)

	registerRoutes(app, routeDeps{
		authController:      authController,
		roleController:      roleController,
		userController:      userController,
		projectController:   projectController,
		modulController:     modulController,
		tusController:       tusController,
		tusModulController:  tusModulController,
		statisticController: statisticController,
		healthController:    healthController,
		supabaseAuthService: supabaseAuthService,
		userRepo:            userRepo,
		cookieHelper:        cookieHelper,
		casbinEnforcer:      casbinEnforcer,
		cfg:                 cfg,
		appLogger:           appLogger,
	})

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
