package config

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"

	"gorm.io/gorm"
)

func RunSeeder(db *gorm.DB, cfg *Config) {
	log.Println("Menjalankan seeder: migrations/app/001_create_all_tables.up.sql")

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Gagal mendapatkan koneksi database:", err)
	}

	filePath := "migrations/app/001_create_all_tables.up.sql"

	sqlBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Gagal membaca file seeder %s: %v", filePath, err)
	}

	sqlContent := string(sqlBytes)

	if err := runSQL(sqlDB, sqlContent); err != nil {
		log.Fatalf("Gagal menjalankan seeder %s: %v", filePath, err)
	}

	log.Println("Seeder berhasil dijalankan dari:", filePath)
}

func runSQL(db *sql.DB, query string) error {
	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("gagal menjalankan SQL: %w", err)
	}
	return nil
}
