package config

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func ConnectDatabase(cfg *Config, dbLogger zerolog.Logger) (*gorm.DB, error) {
	// Use Supabase connection URL if available, otherwise fall back to local database
	dsn := cfg.Supabase.DBURL
	if dsn == "" {
		// Fallback to local database for development
		sslMode := "require"
		if cfg.App.Env == EnvDevelopment {
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
		dbLogger.Info().Msg("using local database connection")
	} else {
		dbLogger.Info().Msg("using Supabase database connection")
	}

	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("database config parse: %w", err)
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
		return nil, fmt.Errorf("database connection open: %w", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDB,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, fmt.Errorf("database connection open: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.Performance.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Performance.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(time.Duration(cfg.Performance.DBConnMaxLifetime) * time.Second)
	sqlDB.SetConnMaxIdleTime(time.Duration(cfg.Performance.DBConnMaxIdleTime) * time.Second)

	dbLogger.Info().Msg("connected to PostgreSQL database")
	return db, nil
}
