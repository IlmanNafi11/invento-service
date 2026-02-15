package testing

import (
	"invento-service/config"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupTestApp creates a Fiber app instance configured for testing
func SetupTestApp(cfg *config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
		EnablePrintRoutes:     false,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"success": false,
				"message": err.Error(),
				"code":    code,
			})
		},
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format:     "[${ip}] ${status} - ${method} ${path}\n",
		TimeFormat: "2006-01-02 15:04:05",
		Output:     nil,
	}))

	// CORS configuration
	if cfg.App.Env == "development" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.App.CorsOriginDev,
			AllowCredentials: true,
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		}))
	} else {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     cfg.App.CorsOriginProd,
			AllowCredentials: true,
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		}))
	}

	return app
}

// TeardownTestApp performs cleanup operations for the test app
func TeardownTestApp(app *fiber.App) error {
	return app.Shutdown()
}
