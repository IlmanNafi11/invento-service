package app

import (
	"invento-service/config"
	"invento-service/internal/controller/http"
	"invento-service/internal/domain"
	"invento-service/internal/httputil"
	"invento-service/internal/middleware"
	"invento-service/internal/rbac"
	"invento-service/internal/usecase/repo"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
	fiberSwagger "github.com/swaggo/fiber-swagger"
)

// routeDeps holds all dependencies needed for route registration.
type routeDeps struct {
	authController      *http.AuthController
	roleController      *http.RoleController
	userController      *http.UserController
	projectController   *http.ProjectController
	modulController     *http.ModulController
	tusController       *http.TusController
	tusModulController  *http.TusModulController
	statisticController *http.StatisticController
	healthController    *http.HealthController

	supabaseAuthService domain.AuthService
	userRepo            repo.UserRepository
	cookieHelper        *httputil.CookieHelper
	casbinEnforcer      *rbac.CasbinEnforcer

	cfg       *config.Config
	appLogger zerolog.Logger
}

// registerRoutes sets up all API route groups on the Fiber app.
func registerRoutes(app *fiber.App, deps routeDeps) {
	api := app.Group("/api/v1")

	registerAuthRoutes(api, deps)
	registerRoleRoutes(api, deps)
	registerUserRoutes(api, deps)
	registerProjectRoutes(api, deps)
	registerModulRoutes(api, deps)
	registerStatisticRoutes(api, deps)
	registerMonitoringRoutes(api, deps)

	// Top-level health check (no auth)
	app.Get("/health", deps.healthController.BasicHealthCheck)

	registerSwaggerRoutes(app, deps)
}

// registerAuthRoutes registers /auth routes: login, register, refresh, reset-password, logout.
func registerAuthRoutes(api fiber.Router, deps routeDeps) {
	auth := api.Group("/auth")
	auth.Post("/login", deps.authController.Login)
	auth.Post("/register", deps.authController.Register)
	auth.Post("/refresh", deps.authController.RefreshToken)
	auth.Post("/reset-password", deps.authController.RequestPasswordReset)

	protected := auth.Group("/", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	protected.Post("logout", deps.authController.Logout)
}

// registerRoleRoutes registers /role routes with auth + RBAC middleware.
func registerRoleRoutes(api fiber.Router, deps routeDeps) {
	role := api.Group("/role", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	role.Get("/permissions", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourcePermission, rbac.ActionRead), deps.roleController.GetAvailablePermissions)
	role.Get("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionRead), deps.roleController.GetRoleList)
	role.Post("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionCreate), deps.roleController.CreateRole)
	role.Get("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionRead), deps.roleController.GetRoleDetail)
	role.Put("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionUpdate), deps.roleController.UpdateRole)
	role.Delete("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionDelete), deps.roleController.DeleteRole)
	role.Get("/:id/users", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionRead), deps.userController.GetUsersForRole)
	role.Post("/:id/users/bulk", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceRole, rbac.ActionUpdate), deps.userController.BulkAssignRole)
}

// registerUserRoutes registers /user and /profile routes with auth + RBAC middleware.
func registerUserRoutes(api fiber.Router, deps routeDeps) {
	user := api.Group("/user", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	user.Get("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceUser, rbac.ActionRead), deps.userController.GetUserList)
	user.Put("/:id/role", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceUser, rbac.ActionUpdate), deps.userController.UpdateUserRole)
	user.Delete("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceUser, rbac.ActionDelete), deps.userController.DeleteUser)
	user.Get("/:id/files", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceUser, rbac.ActionRead), deps.userController.GetUserFiles)
	user.Post("/:id/download", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceUser, rbac.ActionDownload), deps.userController.DownloadUserFiles)
	user.Get("/permissions", deps.userController.GetUserPermissions)

	profile := api.Group("/profile", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	profile.Get("/", deps.userController.GetProfile)
	profile.Put("/", deps.userController.UpdateProfile)
}

// registerProjectRoutes registers /project routes including TUS upload and update groups.
func registerProjectRoutes(api fiber.Router, deps routeDeps) {
	project := api.Group("/project", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	project.Get("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.projectController.GetList)
	project.Get("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.projectController.GetByID)
	project.Patch("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionUpdate), deps.projectController.UpdateMetadata)
	project.Post("/download", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.projectController.Download)
	project.Delete("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionDelete), deps.projectController.Delete)

	// TUS upload check (no TUS protocol middleware)
	tusUploadCheck := api.Group("/project/upload", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	tusUploadCheck.Get("/check-slot", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.tusController.CheckUploadSlot)
	tusUploadCheck.Post("/reset-queue", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionCreate), deps.tusController.ResetUploadQueue)

	// TUS upload (with TUS protocol middleware)
	tusUpload := api.Group("/project/upload", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper), middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeProject))
	tusUpload.Post("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionCreate), deps.tusController.InitiateUpload)
	tusUpload.Patch("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionCreate), deps.tusController.UploadChunk)
	tusUpload.Head("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.tusController.GetUploadStatus)
	tusUpload.Get("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.tusController.GetUploadInfo)
	tusUpload.Delete("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionDelete), deps.tusController.CancelUpload)

	// Project update upload
	projectUpdate := api.Group("/project/:id", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	projectUpdate.Post("/upload", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeProject), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionUpdate), deps.tusController.InitiateProjectUpdateUpload)
	projectUpdate.Patch("/update/:upload_id", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeProject), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionUpdate), deps.tusController.UploadProjectUpdateChunk)
	projectUpdate.Head("/update/:upload_id", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeProject), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.tusController.GetProjectUpdateUploadStatus)
	projectUpdate.Get("/update/:upload_id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionRead), deps.tusController.GetProjectUpdateUploadInfo)
	projectUpdate.Delete("/update/:upload_id", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeProject), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceProject, rbac.ActionUpdate), deps.tusController.CancelProjectUpdateUpload)
}

// registerModulRoutes registers /modul routes including TUS upload and update groups.
func registerModulRoutes(api fiber.Router, deps routeDeps) {
	modul := api.Group("/modul", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	modul.Get("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.modulController.GetList)
	modul.Patch("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionUpdate), deps.modulController.UpdateMetadata)
	modul.Post("/download", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.modulController.Download)
	modul.Delete("/:id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionDelete), deps.modulController.Delete)

	// TUS modul upload check (no TUS protocol middleware)
	tusModulCheck := api.Group("/modul/upload", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	tusModulCheck.Get("/check-slot", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.tusModulController.CheckUploadSlot)

	// TUS modul upload (with TUS protocol middleware)
	tusModul := api.Group("/modul/upload", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper), middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeModul))
	tusModul.Post("/", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionCreate), deps.tusModulController.InitiateUpload)
	tusModul.Patch("/:upload_id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionCreate), deps.tusModulController.UploadChunk)
	tusModul.Head("/:upload_id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.tusModulController.GetUploadStatus)
	tusModul.Get("/:upload_id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.tusModulController.GetUploadInfo)
	tusModul.Delete("/:upload_id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionDelete), deps.tusModulController.CancelUpload)

	// Modul update upload
	modulUpdate := api.Group("/modul/:id", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	modulUpdate.Post("/upload", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeModul), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionUpdate), deps.tusModulController.InitiateModulUpdateUpload)
	modulUpdate.Patch("/update/:upload_id", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeModul), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionUpdate), deps.tusModulController.UploadModulUpdateChunk)
	modulUpdate.Head("/update/:upload_id", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeModul), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.tusModulController.GetModulUpdateUploadStatus)
	modulUpdate.Get("/update/:upload_id", middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionRead), deps.tusModulController.GetModulUpdateUploadInfo)
	modulUpdate.Delete("/update/:upload_id", middleware.TusProtocolMiddleware(deps.cfg.Upload.TusVersion, deps.cfg.Upload.MaxSizeModul), middleware.RBACMiddleware(deps.casbinEnforcer, rbac.ResourceModul, rbac.ActionUpdate), deps.tusModulController.CancelModulUpdateUpload)
}

// registerStatisticRoutes registers /statistic routes with auth middleware.
func registerStatisticRoutes(api fiber.Router, deps routeDeps) {
	statistic := api.Group("/statistic", middleware.SupabaseAuthMiddleware(deps.supabaseAuthService, deps.userRepo, deps.cookieHelper))
	statistic.Get("/", deps.statisticController.GetStatistics)
}

// registerMonitoringRoutes registers /monitoring routes (no auth).
func registerMonitoringRoutes(api fiber.Router, deps routeDeps) {
	monitoring := api.Group("/monitoring")
	monitoring.Get("/health", deps.healthController.ComprehensiveHealthCheck)
	monitoring.Get("/metrics", deps.healthController.GetSystemMetrics)
	monitoring.Get("/status", deps.healthController.GetApplicationStatus)
}

// registerSwaggerRoutes enables the Swagger UI endpoint.
func registerSwaggerRoutes(app *fiber.App, deps routeDeps) {
	deps.appLogger.Info().Msg("swagger UI enabled at /swagger/*")
	app.Get("/swagger/*", fiberSwagger.WrapHandler)
}
