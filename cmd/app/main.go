package main

import (
	"context"
	"invento-service/config"
	"invento-service/internal/app"
	"log"
	"os/signal"
	"runtime/debug"
	"syscall"
	"time"
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
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
	}

	// Apply runtime memory configuration from config
	// This overrides GOMEMLIMIT/GOGC env vars for consistency
	if memLimit, err := config.ParseMemLimit(cfg.Performance.GoMemLimit); err == nil && memLimit > 0 {
		debug.SetMemoryLimit(memLimit)
		log.Printf("GOMEMLIMIT set to %s (%d bytes)", cfg.Performance.GoMemLimit, memLimit)
	} else if err != nil {
		log.Printf("WARNING: invalid GOMEMLIMIT value %q: %v", cfg.Performance.GoMemLimit, err)
	}
	if cfg.Performance.GoGC >= 0 {
		oldGOGC := debug.SetGCPercent(cfg.Performance.GoGC)
		log.Printf("GOGC set to %d (was %d)", cfg.Performance.GoGC, oldGOGC)
	}

	db, err := config.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}

	server, err := app.NewServer(cfg, db)
	if err != nil {
		log.Fatalf("server init failed: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := server.Listen(":" + cfg.App.Port); err != nil {
			log.Printf("server listen error: %v", err)
		}
	}()

	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.ShutdownWithContext(shutdownCtx); err != nil {
		log.Printf("forced shutdown: %v", err)
	}
	log.Println("server stopped gracefully")
}
