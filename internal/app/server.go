package app

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/controller/http"
	"fiber-boiler-plate/internal/helper"
	"fiber-boiler-plate/internal/usecase"
	"fiber-boiler-plate/internal/usecase/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"gorm.io/gorm"
)

func NewServer(cfg *config.Config, db *gorm.DB) *fiber.App {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return helper.SendInternalServerErrorResponse(c)
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Refresh-Token",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))

	userRepo := repo.NewUserRepository(db)
	refreshTokenRepo := repo.NewRefreshTokenRepository(db)
	resetTokenRepo := repo.NewPasswordResetTokenRepository(db)
	roleRepo := repo.NewRoleRepository(db)
	permissionRepo := repo.NewPermissionRepository(db)
	rolePermissionRepo := repo.NewRolePermissionRepository(db)
	projectRepo := repo.NewProjectRepository(db)
	modulRepo := repo.NewModulRepository(db)

	casbinEnforcer, err := helper.NewCasbinEnforcer(db)
	if err != nil {
		panic("Gagal inisialisasi Casbin enforcer: " + err.Error())
	}

	authUsecase := usecase.NewAuthUsecase(userRepo, refreshTokenRepo, resetTokenRepo, roleRepo, cfg)
	authController := http.NewAuthController(authUsecase)

	roleUsecase := usecase.NewRoleUsecase(roleRepo, permissionRepo, rolePermissionRepo, casbinEnforcer)
	roleController := http.NewRoleController(roleUsecase)

	userUsecase := usecase.NewUserUsecase(userRepo, roleRepo, rolePermissionRepo, casbinEnforcer)
	userController := http.NewUserController(userUsecase)

	projectUsecase := usecase.NewProjectUsecase(projectRepo)
	projectController := http.NewProjectController(projectUsecase)

	modulUsecase := usecase.NewModulUsecase(modulRepo)
	modulController := http.NewModulController(modulUsecase)

	healthUsecase := usecase.NewHealthUsecase(db, cfg)
	healthController := http.NewHealthController(healthUsecase)

	api := app.Group("/api/v1")

	auth := api.Group("/auth")
	auth.Post("/register", authController.Register)
	auth.Post("/login", authController.Login)
	auth.Post("/refresh", authController.RefreshToken)
	auth.Post("/reset-password", authController.ResetPassword)
	auth.Post("/reset-password/confirm", authController.ConfirmResetPassword)

	protected := auth.Group("/", helper.JWTAuthMiddleware(cfg.JWT.Secret))
	protected.Post("logout", authController.Logout)

	role := api.Group("/role", helper.JWTAuthMiddleware(cfg.JWT.Secret))
	role.Get("/permissions", helper.RBACMiddleware(casbinEnforcer, "Permission", "read"), roleController.GetAvailablePermissions)
	role.Get("/", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), roleController.GetRoleList)
	role.Post("/", helper.RBACMiddleware(casbinEnforcer, "Role", "create"), roleController.CreateRole)
	role.Get("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "read"), roleController.GetRoleDetail)
	role.Put("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "update"), roleController.UpdateRole)
	role.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Role", "delete"), roleController.DeleteRole)

	user := api.Group("/user", helper.JWTAuthMiddleware(cfg.JWT.Secret))
	user.Get("/", helper.RBACMiddleware(casbinEnforcer, "User", "read"), userController.GetUserList)
	user.Put("/:id/role", helper.RBACMiddleware(casbinEnforcer, "User", "update"), userController.UpdateUserRole)
	user.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "User", "delete"), userController.DeleteUser)
	user.Get("/:id/files", helper.RBACMiddleware(casbinEnforcer, "User", "read"), userController.GetUserFiles)
	user.Get("/permissions", userController.GetUserPermissions)

	profile := api.Group("/profile", helper.JWTAuthMiddleware(cfg.JWT.Secret))
	profile.Get("/", userController.GetProfile)

	project := api.Group("/project", helper.JWTAuthMiddleware(cfg.JWT.Secret))
	project.Get("/", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.GetList)
	project.Post("/", helper.RBACMiddleware(casbinEnforcer, "Project", "create"), projectController.Create)
	project.Post("/download", helper.RBACMiddleware(casbinEnforcer, "Project", "read"), projectController.Download)
	project.Put("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "update"), projectController.Update)
	project.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Project", "delete"), projectController.Delete)

	modul := api.Group("/modul", helper.JWTAuthMiddleware(cfg.JWT.Secret))
	modul.Get("/", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.GetList)
	modul.Post("/", helper.RBACMiddleware(casbinEnforcer, "Modul", "create"), modulController.Create)
	modul.Post("/download", helper.RBACMiddleware(casbinEnforcer, "Modul", "read"), modulController.Download)
	modul.Put("/:id", helper.RBACMiddleware(casbinEnforcer, "Modul", "update"), modulController.Update)
	modul.Delete("/:id", helper.RBACMiddleware(casbinEnforcer, "Modul", "delete"), modulController.Delete)

	monitoring := api.Group("/monitoring")
	monitoring.Get("/health", healthController.ComprehensiveHealthCheck)
	monitoring.Get("/metrics", healthController.GetSystemMetrics)
	monitoring.Get("/status", healthController.GetApplicationStatus)

	app.Get("/health", healthController.BasicHealthCheck)

	return app
}
