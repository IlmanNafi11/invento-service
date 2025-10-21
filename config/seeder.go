package config

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gorm.io/gorm"
)

func RunSeeder(db *gorm.DB, cfg *Config) {
	log.Println("Menjalankan database seeder dari folder migrations/app...")

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Gagal mendapatkan koneksi database:", err)
	}

	seederDir := "migrations/app"

	files, err := filepath.Glob(filepath.Join(seederDir, "*seed*.up.sql"))
	if err != nil {
		log.Fatal("Gagal membaca file seeder:", err)
	}

	if len(files) == 0 {
		log.Println("Tidak ada file seeder ditemukan di folder migrations/app")
		return
	}

	for _, file := range files {
		log.Printf("Menjalankan seeder file: %s", file)

		sqlBytes, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalf("Gagal membaca file %s: %v", file, err)
		}

		sqlContent := string(sqlBytes)
		if err := runSQL(sqlDB, sqlContent); err != nil {
			log.Fatalf("Gagal menjalankan seeder %s: %v", file, err)
		}
	}

	log.Println("Semua seeder SQL berhasil dijalankan!")
}

func runSQL(db *sql.DB, query string) error {
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("gagal menjalankan SQL: %w", err)
	}
	return nil
}
