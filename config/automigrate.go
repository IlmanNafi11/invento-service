package config

import (
	"log"

	"fiber-boiler-plate/internal/domain"
	"gorm.io/gorm"
)

// AutoMigrate runs GORM auto migration for all domain models
func AutoMigrate(db *gorm.DB) {
	log.Println("Menjalankan database migration dengan GORM AutoMigrate...")

	// List semua model yang akan di-migrate
	models := []interface{}{
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
		&domain.User{},
		&domain.RefreshToken{},
		&domain.PasswordResetToken{},
		&domain.Project{},
		&domain.Modul{},
		&domain.TusUpload{},
		&domain.TusModulUpload{},
		&domain.OTP{},
	}

	// Jalankan AutoMigrate
	for _, model := range models {
		if err := db.AutoMigrate(model); err != nil {
			log.Printf("Gagal melakukan migrasi untuk %T: %v", model, err)
		} else {
			log.Printf("✅ Berhasil migrasi: %T", model)
		}
	}

	log.Println("Database migration selesai!")
}
