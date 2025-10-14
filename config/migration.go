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
		&domain.Project{},
		&domain.Modul{},
		&domain.TusUpload{},
		&domain.TusModulUpload{},
	); err != nil {
		log.Fatal("Gagal melakukan auto migrate:", err)
	}

	// Periksa apakah tabel tus_uploads ada dengan struktur lama
	if err := migrateTusUploadsTable(db); err != nil {
		log.Fatal("Gagal migrasi tabel tus_uploads:", err)
	}

	log.Println("Database migration selesai")
}

func migrateTusUploadsTable(db *gorm.DB) error {
	log.Println("Memeriksa struktur tabel tus_uploads...")

	// Cek apakah kolom nama_project masih ada (struktur lama)
	var count int
	err := db.Raw(`
		SELECT COUNT(*) 
		FROM information_schema.columns 
		WHERE table_name = 'tus_uploads' 
		AND column_name = 'nama_project'
		AND table_schema = DATABASE()
	`).Scan(&count).Error

	if err != nil {
		return err
	}

	// Jika masih ada kolom lama, migrasi ke struktur baru
	if count > 0 {
		log.Println("Mendapatkan struktur lama tus_uploads, melakukan migrasi ke struktur baru...")

		// Backup data lama jika ada
		var hasData int
		db.Raw("SELECT COUNT(*) FROM tus_uploads").Scan(&hasData)

		if hasData > 0 {
			log.Printf(" WARNING: Ada %d data di tabel lama, akan di-backup dan migrasi", hasData)
		}

		// Drop dan recreate tabel dengan struktur baru
		if err := db.Exec("DROP TABLE IF EXISTS tus_uploads").Error; err != nil {
			return err
		}

		log.Println("Membuat tabel tus_uploads dengan struktur baru...")
		if err := db.AutoMigrate(&domain.TusUpload{}); err != nil {
			return err
		}

		log.Println("Migrasi tabel tus_uploads berhasil")
	} else {
		log.Println("Tabel tus_uploads sudah menggunakan struktur baru")
	}

	return nil
}
