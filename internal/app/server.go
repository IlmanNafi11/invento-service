package app

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"fiber-boiler-plate/internal/usecase/repo"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
)

func NewServer(cfg *config.Config, db *gorm.DB) *fiber.App {
	appLogger := helper.NewLogger()

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			appLogger.Error("[ERROR_HANDLER] Error occurred: %v, Path: %s, Method: %s", err, c.Path(), c.Method())

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

	app.Use(logger.New())
	app.Use(recover.New())

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
	app.Static("/uploads", pathResolver.GetBasePath())

	userRepo := repo.NewUserRepository(db)
	refreshTokenRepo := repo.NewRefreshTokenRepository(db)
	resetTokenRepo := repo.NewPasswordResetTokenRepository(db)
	roleRepo := repo.NewRoleRepository(db)
	permissionRepo := repo.NewPermissionRepository(db)
	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	projectRepo := repo.NewProjectRepository(db)
	modulRepo := repo.NewModulRepository(db)
	tusUploadRepo := repo.NewTusUploadRepository(db)
	otpRepo := repo.NewOTPRepository(db)

	jwtManager, err := helper.NewJWTManager(cfg)
	if err != nil {
		panic("Gagal inisialisasi JWT Manager: " + err.Error())
	}

	casbinEnforcer, err := helper.NewCasbinEnforcer(db)
	if err != nil {
		panic("Gagal inisialisasi Casbin enforcer: " + err.Error())
	}
	tusStore := helper.NewTusStore(pathResolver, cfg.Upload.MaxSize)
	tusQueue := helper.NewTusQueue(cfg.Upload.MaxConcurrent)
	fileManager := helper.NewFileManager(cfg)
	tusManager := helper.NewTusManager(tusStore, tusQueue, fileManager, cfg)
	tusCleanup := helper.NewTusCleanup(tusUploadRepo, tusStore, cfg.Upload.CleanupInterval, cfg.Upload.IdleTimeout)
	tusCleanup.Start()

	keyRotation := helper.NewKeyRotationScheduler(jwtManager, cfg.JWT.KeyRotationHours)
	keyRotation.Start()

	refreshTokenCleanup := helper.NewRefreshTokenCleanup(refreshTokenRepo, 24)
	refreshTokenCleanup.Start()

	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, resetTokenRepo, roleRepo, cfg)
	authController := http.NewAuthController(authUsecase, cfg)

	authOTPUsecase := usecase.NewAuthOTPUsecase(userRepo, refreshTokenRepo, otpRepo, roleRepo, cfg)
	authOTPController := http.NewAuthOTPController(authOTPUsecase, cfg)

	roleUsecase := usecase.NewRoleUsecase(roleRepo, permissionRepo, rolePermissionRepo, casbinEnforcer)
	roleController := http.NewRoleController(roleUsecase)

	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, projectRepo, modulRepo, casbinEnforcer, pathResolver, cfg, db)
	userController := http.NewUserController(userUsecase)

	projectUsecase := usecase.NewProjectUsecase(projectRepo, fileManager)
	projectController := http.NewProjectController(projectUsecase)

	tusUploadUsecase := usecase.NewTusUploadUsecase(tusUploadRepo, projectRepo, projectUsecase, tusManager, fileManager, cfg)
	tusController := http.NewTusController(tusUploadUsecase, cfg)

	tusModulUploadRepo := repo.NewTusModulUploadRepository(db)
	modulUsecase := usecase.NewModulUsecase(modulRepo)
	tusModulUsecase := usecase.NewTusModulUsecase(tusModulUploadRepo, modulRepo, tusManager, fileManager, cfg)
	modulController := http.NewModulController(modulUsecase, tusModulUsecase, cfg)

	healthUsecase := usecase.NewHealthUsecase(db, cfg)
	healthController := http.NewHealthController(healthUsecase)

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.RefreshToken)
	auth.Post("/register/otp", authOTPController.RegisterWithOTP)
	auth.Post("/register/verify-otp", authOTPController.VerifyRegisterOTP)
	auth.Post("/register/resend-otp", authOTPController.ResendRegisterOTP)
	auth.Post("/reset-password/otp", authOTPController.InitiateResetPassword)
	auth.Post("/reset-password/verify-otp", authOTPController.VerifyResetPasswordOTP)
	auth.Post("/reset-password/confirm-otp", authOTPController.ConfirmResetPasswordWithOTP)
	auth.Post("/reset-password/resend-otp", authOTPController.ResendResetPasswordOTP)

	protected := auth.Group("/", helper.JWTAuthMiddleware(jwtManager))
	protected.Post("logout", authController.Logout)

	role := api.Group("/role", helper.JWTAuthMiddleware(jwtManager))
	role.Get("/permissions", helper.RBACMiddleware(casbinEnforcer, "Permission", "read"), roleController.GetAvailablePermissions)
	role.Get("/", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), roleController.GetRoleList)
	role.Post("/", helper.RBACMiddleware(casbinEnforcer, "Role", "create"), roleController.CreateRole)
	role.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), roleController.GetRoleDetail)
	role.Put("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "update"), roleController.UpdateRole)
	role.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "delete"), roleController.DeleteRole)

	user := api.Group("/user", helper.JWTAuthMiddleware(jwtManager))
	user.Get("/", helper.RBACMiddleware(casbinEnforcer, "User", "read"), userController.GetUserList)
	user.Put("/:id/role", helper.RBACMiddleware(casbinEnforcer, "User", "update"), userController.UpdateUserRole)
	user.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "User", "delete"), userController.DeleteUser)
	user.Get("/:id/files", helper.RBACMiddleware(casbinEnforcer, "User", "read"), userController.GetUserFiles)
	user.Post("/:id/download", helper.RBACMiddleware(casbinEnforcer, "User", "download"), userController.DownloadUserFiles)
	user.Get("/permissions", userController.GetUserPermissions)

	profile := api.Group("/profile", helper.JWTAuthMiddleware(jwtManager))
	profile.Get("/", userController.GetProfile)
	profile.Put("/", userController.UpdateProfile)

	project := api.Group("/project", helper.JWTAuthMiddleware(jwtManager))
	project.Get("/", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.GetList)
	project.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.GetByID)
	project.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "update"), projectController.UpdateMetadata)
	project.Post("/download", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.Download)
	project.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "delete"), projectController.Delete)

	tusUploadCheck := api.Group("/project/upload", helper.JWTAuthMiddleware(jwtManager))
	tusUploadCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.CheckUploadSlot)
	tusUploadCheck.Post("/reset-queue", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), tusController.ResetUploadQueue)

	tusUpload := api.Group("/project/upload", helper.JWTAuthMiddleware(jwtManager), helper.TusProtocolMiddleware(cfg.Upload.TusVersion))
	tusUpload.Post("/", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), tusController.InitiateUpload)
	tusUpload.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), tusController.UploadChunk)
	tusUpload.Head("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetUploadStatus)
	tusUpload.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetUploadInfo)
	tusUpload.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "delete"), tusController.CancelUpload)

	projectUpdate := api.Group("/project/:id", helper.JWTAuthMiddleware(jwtManager))
	projectUpdate.Post("/upload", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Project", "update"), tusController.InitiateProjectUpdateUpload)
	projectUpdate.Patch("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Project", "update"), tusController.UploadProjectUpdateChunk)
	projectUpdate.Head("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetProjectUpdateUploadStatus)
	projectUpdate.Get("/update/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), tusController.GetProjectUpdateUploadInfo)
	projectUpdate.Delete("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Project", "update"), tusController.CancelProjectUpdateUpload)

	modul := api.Group("/modul", helper.JWTAuthMiddleware(jwtManager))
	modul.Get("/", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetList)
	modul.Patch("/:id", helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), modulController.UpdateMetadata)
	modul.Post("/download", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.Download)
	modul.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Modul", "delete"), modulController.Delete)

	tusModulCheck := api.Group("/modul/upload", helper.JWTAuthMiddleware(jwtManager))
	tusModulCheck.Get("/check-slot", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.CheckUploadSlot)

	tusModul := api.Group("/modul/upload", helper.JWTAuthMiddleware(jwtManager), helper.TusProtocolMiddleware(cfg.Upload.TusVersion))
	tusModul.Post("/", helper.RBACMiddleware(casbinEnforcer, "Modul", "create"), modulController.InitiateUpload)
	tusModul.Patch("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "create"), modulController.UploadChunk)
	tusModul.Head("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetUploadStatus)
	tusModul.Get("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetUploadInfo)
	tusModul.Delete("/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "delete"), modulController.CancelUpload)

	modulUpdate := api.Group("/modul/:id", helper.JWTAuthMiddleware(jwtManager))
	modulUpdate.Post("/upload", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), modulController.InitiateModulUpdateUpload)
	modulUpdate.Patch("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), modulController.UploadModulUpdateChunk)
	modulUpdate.Head("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetModulUpdateUploadStatus)
	modulUpdate.Get("/update/:upload_id", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetModulUpdateUploadInfo)
	modulUpdate.Delete("/update/:upload_id", helper.TusProtocolMiddleware(cfg.Upload.TusVersion), helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), modulController.CancelModulUpdateUpload)

	monitoring := api.Group("/monitoring")
	monitoring.Get("/health", healthController.ComprehensiveHealthCheck)
	monitoring.Get("/metrics", healthController.GetSystemMetrics)
	monitoring.Get("/status", healthController.GetApplicationStatus)

	app.Get("/health", healthController.BasicHealthCheck)

	return app
}
