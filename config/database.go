package config

import (
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDatabase(cfg *Config) *gorm.DB {
	// Use Supabase connection URL if available, otherwise fall back to local database
	dsn := cfg.Supabase.DBURL
	if dsn == "" {
		// Fallback to local database for development
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable TimeZone=Asia/Jakarta",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
		)
		log.Println("Using local database connection")
	} else {
		log.Println("Using Supabase database connection")
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})

	if err != nil {
		log.Fatal("Gagal menghubungkan ke database:", err)
	}

	log.Println("Berhasil terhubung ke database PostgreSQL")
	return db
}
