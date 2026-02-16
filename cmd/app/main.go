package main

import (
	"context"
	"invento-service/config"
	"invento-service/internal/app"
	"os"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

// @title Invento Service API
// @version 1.0
// @description REST API service for managing projects, modules, users, and file uploads with JWT authentication and RBAC.
// @host localhost:3000
// @BasePath /api/v1
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Authorization header using the Bearer scheme. Example: "Bearer {token}"
// @securityDefinitions.apikey CookieAuth
// @in cookie
// @name access_token
// @description Access token stored in HttpOnly cookie (set automatically on login/register/refresh)

func main() {
	// Pre-config logger for fatal startup errors (logger not yet configured)
	bootLogger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	cfg, err := config.LoadConfig()
	if err != nil {
		bootLogger.Fatal().Err(err).Msg("config load failed")
	}
	if err = cfg.Validate(); err != nil {
		bootLogger.Fatal().Err(err).Msg("config validation failed")
	}

	// Apply runtime memory configuration from config
	// This overrides GOMEMLIMIT/GOGC env vars for consistency
	{
		memLimit, memErr := config.ParseMemLimit(cfg.Performance.GoMemLimit)
		if memErr == nil && memLimit > 0 {
			debug.SetMemoryLimit(memLimit)
			bootLogger.Info().Str("gomemlimit", cfg.Performance.GoMemLimit).Int64("bytes", memLimit).Msg("GOMEMLIMIT set")
		} else if memErr != nil {
			bootLogger.Warn().Str("value", cfg.Performance.GoMemLimit).Err(memErr).Msg("invalid GOMEMLIMIT value")
		}
	}
	if cfg.Performance.GoGC >= 0 {
		oldGOGC := debug.SetGCPercent(cfg.Performance.GoGC)
		bootLogger.Info().Int("gogc", cfg.Performance.GoGC).Int("was", oldGOGC).Msg("GOGC set")
	}

	db, err := config.ConnectDatabase(cfg, bootLogger)
	if err != nil {
		bootLogger.Fatal().Err(err).Msg("database connection failed")
	}

	server, err := app.NewServer(cfg, db)
	if err != nil {
		bootLogger.Fatal().Err(err).Msg("server init failed")
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.Listen(":" + cfg.App.Port); err != nil {
			bootLogger.Error().Err(err).Msg("server listen error")
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.ShutdownWithContext(shutdownCtx); err != nil {
		bootLogger.Error().Err(err).Msg("forced shutdown")
	}
	bootLogger.Info().Msg("server stopped gracefully")
}
