package config

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDatabase(cfg *Config) *gorm.DB {
	// Use Supabase connection URL if available, otherwise fall back to local database
	dsn := cfg.Supabase.DBURL
	if dsn == "" {
		// Fallback to local database for development
		sslMode := "require"
		if cfg.App.Env == "development" {
			sslMode = "disable"
		}
		dsn = fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=Asia/Jakarta",
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Name,
			sslMode,
		)
		log.Println("Using local database connection")
	} else {
		log.Println("Using Supabase database connection")
	}

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Gagal parsing konfigurasi database:", err)
	}
	// Disable prepared statement caching for compatibility with Supabase PgBouncer (transaction mode)
	pgxConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	pgxConfig.DialFunc = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, port, splitErr := net.SplitHostPort(addr)
		if splitErr != nil {
			return nil, splitErr
		}
		resolver := &net.Resolver{}
		ips, resolveErr := resolver.LookupHost(ctx, host)
		if resolveErr != nil {
			return nil, resolveErr
		}
		for _, ip := range ips {
			if net.ParseIP(ip).To4() != nil {
				return (&net.Dialer{}).DialContext(ctx, "tcp4", net.JoinHostPort(ip, port))
			}
		}
		return (&net.Dialer{}).DialContext(ctx, "tcp", addr)
	}

	// Use RegisterConnConfig so the custom DialFunc is actually honored.
	// stdlib.OpenDB(config) passes config by value and may ignore DialFunc.
	connStr := stdlib.RegisterConnConfig(pgxConfig)
	sqlDB, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal("Gagal membuka koneksi database:", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		log.Fatal("Gagal menghubungkan ke database:", err)
	}
	sqlDB.SetMaxIdleConns(5)
	sqlDB.SetMaxOpenConns(20)
	sqlDB.SetConnMaxLifetime(30 * time.Minute)
	sqlDB.SetConnMaxIdleTime(5 * time.Minute)

	log.Println("Berhasil terhubung ke database PostgreSQL")
	return db
}
