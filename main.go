package main

import (
	"context"
	"invento-service/config"
	"invento-service/internal/app"
	"log"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("config load failed: %v", err)
	}
	if err := cfg.Validate(); err != nil {
		log.Fatalf("config validation failed: %v", err)
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
