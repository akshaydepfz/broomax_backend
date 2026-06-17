package database

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
)

// ConnectPostgres opens a PostgreSQL connection using a libpq connection string
// (URL form postgres://... or keyword form host=...).
// Caller must not close the pool while the app runs; assign to helper.DB from app init.
func ConnectPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

// PostgresDSNFromEnv builds a DSN from the environment — no credentials in source code.
//
// Preference order:
//  1. DATABASE_URL if it looks like a Postgres URL (postgres:// or postgresql://)
//  2. Keyword DSN built from PGHOST, PGUSER, PGPASSWORD, PGDATABASE (and optional PGPORT, PGSSLMODE)
//
// Returns ("", false) if Postgres is not configured (caller may use SQLite instead).
func PostgresDSNFromEnv() (dsn string, ok bool) {
	if u := strings.TrimSpace(os.Getenv("DATABASE_URL")); u != "" {
		if strings.HasPrefix(u, "postgres://") || strings.HasPrefix(u, "postgresql://") {
			return u, true
		}
	}
	host := strings.TrimSpace(os.Getenv("PGHOST"))
	if host == "" {
		return "", false
	}
	port := strings.TrimSpace(os.Getenv("PGPORT"))
	if port == "" {
		port = "5432"
	}
	user := strings.TrimSpace(os.Getenv("PGUSER"))
	password := os.Getenv("PGPASSWORD")
	dbname := strings.TrimSpace(os.Getenv("PGDATABASE"))
	sslmode := strings.TrimSpace(os.Getenv("PGSSLMODE"))
	if sslmode == "" {
		sslmode = "require"
	}
	dsn = fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
	return dsn, true
}
