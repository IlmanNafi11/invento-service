package main

import (
	"context"
	"invento-service/config"
	"invento-service/internal/app"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

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
