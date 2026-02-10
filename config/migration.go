package config

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"gorm.io/gorm"
)

func RunMigration(db *gorm.DB) {
	log.Println("Menjalankan database migration...")

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Gagal mendapatkan database instance:", err)
	}

	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{
		MigrationsTable: "schema_migrations",
	})
	if err != nil {
		log.Fatal("Gagal membuat driver migration:", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations/app",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatal("Gagal membuat migration instance:", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		log.Printf("Peringatan saat mendapatkan versi migration: %v", err)
	} else if err == migrate.ErrNilVersion {
		log.Println("Tidak ada migration yang pernah dijalankan sebelumnya")
	} else {
		log.Printf("Versi migration saat ini: %d (dirty: %v)", version, dirty)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("Gagal menjalankan migration:", err)
	}

	if err == migrate.ErrNoChange {
		log.Println("Database sudah up-to-date, tidak ada migration baru")
	} else {
		newVersion, _, _ := m.Version()
		log.Printf("Migration berhasil dijalankan, versi sekarang: %d", newVersion)
	}

	log.Println("Database migration selesai")
}
