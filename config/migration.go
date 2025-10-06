package config

import (
	"fiber-boiler-plate/internal/domain"
	"log"

	"gorm.io/gorm"
)

func RunMigration(db *gorm.DB) {
	log.Println("Menjalankan database migration...")

	if err := db.AutoMigrate(
		&domain.User{},
		&domain.RefreshToken{},
		&domain.PasswordResetToken{},
		&domain.Role{},
		&domain.Permission{},
		&domain.RolePermission{},
	); err != nil {
		log.Fatal("Gagal melakukan auto migrate:", err)
	}

	log.Println("Database migration selesai")
}
