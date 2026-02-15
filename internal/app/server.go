package app

import (
	"invento-service/config"
	_ "invento-service/docs"
	"invento-service/internal/controller/base"
	"invento-service/internal/controller/http"
	"invento-service/internal/helper"
	"invento-service/internal/logger"
	"invento-service/internal/middleware"
	supabaseAuth "invento-service/internal/supabase"
	"invento-service/internal/usecase"
	"invento-service/internal/usecase/repo"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/supabase-community/supabase-go"
	"github.com/swaggo/fiber-swagger"
	"gorm.io/gorm"
)

func NewServer(cfg *config.Config, db *gorm.DB) *fiber.App {
	// Initialize structured logger
	logLevel := logger.ParseLogLevel(cfg.Logging.Level)
	logFormat := logger.ParseLogFormat(cfg.Logging.Format)
	isDevelopment := cfg.App.Env == "development"

	appLogger := logger.NewLogger(logLevel, logFormat)
	if isDevelopment && logFormat == logger.TextFormat {
		// Use default text logger for development
		appLogger = logger.NewDefaultLogger(isDevelopment)
	}

	app := fiber.New(fiber.Config{
		ReadBufferSize: 16384, // 16KB - default 4KB terlalu kecil untuk JWT cookies + Authorization headers
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			requestID := middleware.GetRequestID(c)
			reqLogger := appLogger.WithRequestID(requestID)
			reqLogger.Error("[ERROR_HANDLER] Error occurred", map[string]interface{}{
				"error":  err.Error(),
				"path":   c.Path(),
				"method": c.Method(),
			})

			if c.Path() != "" && len(c.Path()) >= 8 && c.Path()[:8] == "/uploads" {
				return err
			}

			if err != nil {
				if e, ok := err.(*fiber.Error); ok {
					if e.Code == fiber.StatusNotFound {
						return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
							"success":   false,
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
			return helper.SendInternalServerErrorResponse(c)
		},
	})

	// Apply middleware in order: RequestID -> Logger -> Recover -> CORS
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogger(appLogger))
	app.Use(recover.New())

	// Log startup information
	appLogger.Info("Server starting", map[string]interface{}{
		"app":        cfg.App.Name,
		"env":        cfg.App.Env,
		"port":       cfg.App.Port,
		"log_level":  cfg.Logging.Level,
		"log_format": cfg.Logging.Format,
	})

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

	pathResolver := helper.NewPathResolver(cfg)
	cookieHelper := helper.NewCookieHelper(cfg)
	app.Static("/uploads", pathResolver.GetBasePath())

	userRepo := repo.NewUserRepository(db)
	roleRepo := repo.NewRoleRepository(db)
	permissionRepo := repo.NewPermissionRepository(db)
	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	projectRepo := repo.NewProjectRepository(db)
	modulRepo := repo.NewModulRepository(db)
	tusUploadRepo := repo.NewTusUploadRepository(db)
	tusModulUploadRepo := repo.NewTusModulUploadRepository(db)

	// Initialize Supabase client
	supabaseClient, err := supabase.NewClient(cfg.Supabase.URL, cfg.Supabase.ServiceKey, nil)
	if err != nil {
		panic("Gagal inisialisasi Supabase client: " + err.Error())
	}

	supabaseServiceKey := cfg.Supabase.ServiceKey
	authURL := cfg.Supabase.URL + "/auth/v1"
	supabaseAuthService, err := supabaseAuth.NewAuthService(authURL, supabaseServiceKey)
	if err != nil {
		panic("Gagal inisialisasi Supabase Auth service: " + err.Error())
	}

	casbinEnforcer, err := helper.NewCasbinEnforcer(db)
	if err != nil {
		panic("Gagal inisialisasi Casbin enforcer: " + err.Error())
	}
	tusProjectStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSizeProject)
	tusModulStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSizeModul)
	tusQueue := helper.NewTusQueue(cfg.Upload.MaxConcurrentProject)
	tusModulQueue := helper.NewTusQueue(cfg.Upload.MaxQueueModulPerUser)
	fileManager := helper.NewFileManager(cfg)
	tusProjectManager := helper.NewTusManager(tusProjectStore, tusQueue, fileManager, cfg)
	tusModulManager := helper.NewTusManager(tusModulStore, tusModulQueue, fileManager, cfg)

	if activeIDs, err := tusUploadRepo.GetActiveUploadIDs(); err == nil && len(activeIDs) > 0 {
		tusQueue.LoadFromDB(activeIDs)
	}
	if activeIDs, err := tusModulUploadRepo.GetActiveUploadIDs(); err == nil && len(activeIDs) > 0 {
		tusModulQueue.LoadFromDB(activeIDs)
	}

	authUsecase := usecase.NewAuthUsecase(userRepo, roleRepo, supabaseClient, supabaseServiceKey, cfg)
	authController := http.NewAuthController(authUsecase, cookieHelper, cfg)

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

	tusCleanup := helper.NewTusCleanup(tusUploadRepo, tusModulUploadRepo, tusProjectStore, tusModulStore, cfg.Upload.CleanupInterval, cfg.Upload.IdleTimeout)
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
	role.Get("/permissions", helper.RBACMiddleware(casbinEnforcer, "Permission", "read"), roleController.GetAvailablePermissions)
	role.Get("/", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), roleController.GetRoleList)
	role.Post("/", helper.RBACMiddleware(casbinEnforcer, "Role", "create"), roleController.CreateRole)
	role.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), roleController.GetRoleDetail)
	role.Put("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "update"), roleController.UpdateRole)
	role.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "delete"), roleController.DeleteRole)
	role.Get("/:id/users", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), userController.GetUsersForRole)
	role.Post("/:id/users/bulk", helper.RBACMiddleware(casbinEnforcer, "Role", "update"), userController.BulkAssignRole)

	user := api.Group("/user", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	user.Get("/", helper.RBACMiddleware(casbinEnforcer, "User", "read"), userController.GetUserList)
	user.Put("/:id/role", helper.RBACMiddleware(casbinEnforcer, "User", "update"), userController.UpdateUserRole)
	user.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "User", "delete"), userController.DeleteUser)
	user.Get("/:id/files", helper.RBACMiddleware(casbinEnforcer, "User", "read"), userController.GetUserFiles)
	user.Post("/:id/download", helper.RBACMiddleware(casbinEnforcer, "User", "download"), userController.DownloadUserFiles)
	user.Get("/permissions", userController.GetUserPermissions)

	profile := api.Group("/profile", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	profile.Get("/", userController.GetProfile)
	profile.Put("/", userController.UpdateProfile)

	project := api.Group("/project", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	project.Get("/", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.GetList)
	project.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.GetByID)
	project.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "update"), projectController.UpdateMetadata)
	project.Post("/download", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.Download)
	project.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "delete"), projectController.Delete)

	tusUploadCheck := api.Group("/project/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	tusUploadCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.CheckUploadSlot)
	tusUploadCheck.Post("/reset-queue", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), tusController.ResetUploadQueue)

	tusUpload := api.Group("/project/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper), helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject))
	tusUpload.Post("/", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), tusController.InitiateUpload)
	tusUpload.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), tusController.UploadChunk)
	tusUpload.Head("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetUploadStatus)
	tusUpload.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetUploadInfo)
	tusUpload.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "delete"), tusController.CancelUpload)

	projectUpdate := api.Group("/project/:id", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	projectUpdate.Post("/upload", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, "Project", "update"), tusController.InitiateProjectUpdateUpload)
	projectUpdate.Patch("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, "Project", "update"), tusController.UploadProjectUpdateChunk)
	projectUpdate.Head("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetProjectUpdateUploadStatus)
	projectUpdate.Get("/update/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetProjectUpdateUploadInfo)
	projectUpdate.Delete("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeProject), helper.RBACMiddleware(casbinEnforcer, "Project", "update"), tusController.CancelProjectUpdateUpload)

	modul := api.Group("/modul", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	modul.Get("/", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetList)
	modul.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), modulController.UpdateMetadata)
	modul.Post("/download", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.Download)
	modul.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Modul", "delete"), modulController.Delete)

	tusModulCheck := api.Group("/modul/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	tusModulCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), tusModulController.CheckUploadSlot)

	tusModul := api.Group("/modul/upload", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper), helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul))
	tusModul.Post("/", helper.RBACMiddleware(casbinEnforcer, "Modul", "create"), tusModulController.InitiateUpload)
	tusModul.Patch("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "create"), tusModulController.UploadChunk)
	tusModul.Head("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), tusModulController.GetUploadStatus)
	tusModul.Get("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), tusModulController.GetUploadInfo)
	tusModul.Delete("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "delete"), tusModulController.CancelUpload)

	modulUpdate := api.Group("/modul/:id", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	modulUpdate.Post("/upload", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), tusModulController.InitiateModulUpdateUpload)
	modulUpdate.Patch("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), tusModulController.UploadModulUpdateChunk)
	modulUpdate.Head("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), tusModulController.GetModulUpdateUploadStatus)
	modulUpdate.Get("/update/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), tusModulController.GetModulUpdateUploadInfo)
	modulUpdate.Delete("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion, cfg.Upload.MaxSizeModul), helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), tusModulController.CancelModulUpdateUpload)

	statistic := api.Group("/statistic", helper.SupabaseAuthMiddleware(supabaseAuthService, userRepo, cookieHelper))
	statistic.Get("/", statisticController.GetStatistics)

	monitoring := api.Group("/monitoring")
	monitoring.Get("/health", healthController.ComprehensiveHealthCheck)
	monitoring.Get("/metrics", healthController.GetSystemMetrics)
	monitoring.Get("/status", healthController.GetApplicationStatus)

	app.Get("/health", healthController.BasicHealthCheck)

	// Enable Swagger UI
	appLogger.Info("Swagger UI enabled at /swagger/*", nil)
	app.Get("/swagger/*", fiberSwagger.WrapHandler)

	return app
}
