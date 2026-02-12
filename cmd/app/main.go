package main

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/app"
	"fiber-boiler-plate/internal/helper"
	"log"
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

func main() {
	cfg := config.LoadConfig()
	db := config.ConnectDatabase(cfg)

	casbinEnforcer, err := helper.NewCasbinEnforcer(db)
	if err != nil {
		log.Fatalf("Gagal inisialisasi Casbin enforcer: %v", err)
	}

	if cfg.Database.SeedData {
		seeder := app.NewSeeder(db, casbinEnforcer)
		if err := seeder.Run(); err != nil {
			log.Printf("Seeder gagal: %v", err)
		} else {
			log.Println("Seeder berhasil dijalankan")
		}
	}

	server := app.NewServer(cfg, db)

	log.Printf("Server berjalan di port %s", cfg.App.Port)
	log.Fatal(server.Listen(":" + cfg.App.Port))
}
