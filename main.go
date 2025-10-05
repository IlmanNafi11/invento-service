package main

import (
	"fiber-boiler-plate/config"
	"fiber-boiler-plate/internal/app"
	"log"
)

func main() {
	cfg := config.LoadConfig()

	db := config.ConnectDatabase(cfg)

	if cfg.Database.MigrateOnStart && cfg.Database.AutoMigrate {
		log.Printf("🔄 Menjalankan auto migration untuk environment: %s", cfg.App.Env)
		config.RunMigration(db)
	} else {
		log.Println("⏭️  Auto migration dinonaktifkan melalui konfigurasi")
	}

	if cfg.Database.RunSeeder {
		if cfg.App.Env == "production" {
			log.Println("🚨 PERINGATAN: Seeder tidak direkomendasikan untuk production environment!")
			log.Println("🛡️  Melewati eksekusi seeder untuk keamanan production")
			log.Println("ℹ️  Untuk menjalankan seeder di production, ubah APP_ENV ke nilai lain")
		} else {
			log.Printf("🌱 Menjalankan seeder untuk environment: %s", cfg.App.Env)
			config.RunSeeder(db, cfg)
		}
	} else {
		log.Println("⏭️  Seeder dinonaktifkan melalui konfigurasi DB_RUN_SEEDER=false")
	}

	server := app.NewServer(cfg, db)

	log.Printf("Server berjalan di port %s", cfg.App.Port)
	log.Fatal(server.Listen(":" + cfg.App.Port))
}
