package main

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/app"
	"fiber-boiler-plate/internal/helper"
	"log"
	"os"
)

// Standalone seeder command for seeding roles and permissions
// Run with: go run cmd/seed/main.go

func main() {
	log.Println("Memulai seeding roles dan permissions...")

	cfg := config.LoadConfig()
	db := config.ConnectDatabase(cfg)

	casbinEnforcer, err := helper.NewCasbinEnforcer(db)
	if err != nil {
		log.Fatalf("Gagal inisialisasi Casbin enforcer: %v", err)
	}

	seeder := app.NewSeeder(db, casbinEnforcer)
	if err := seeder.Run(); err != nil {
		log.Printf("Seeder gagal: %v", err)
		os.Exit(1)
	}

	log.Println("Seeder berhasil dijalankan!")
	log.Println("Roles yang tersedia: Admin, Dosen, Mahasiswa")
	log.Println("Silakan assign role ke user yang sesuai di Supabase.")
}
