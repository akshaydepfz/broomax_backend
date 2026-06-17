package helper

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
	"oryoo.com/database"
)

// DB is the application database handle (opened in main).
var DB *sql.DB

// IsPostgres is true when DATABASE_URL selects PostgreSQL.
var IsPostgres bool

// InitDB opens the DB and applies migrations.
//
// If Postgres is configured via database.PostgresDSNFromEnv (DATABASE_URL or PG* vars),
// connects through oryoo.com/database.ConnectPostgres and sets helper.DB.
// Otherwise uses SQLite (default file broomax.db when DATABASE_URL unset, or a sqlite DSN in DATABASE_URL for file:...).
func InitDB() error {
	if pgDSN, ok := database.PostgresDSNFromEnv(); ok {
		db, err := database.ConnectPostgres(pgDSN)
		if err != nil {
			return err
		}
		DB = db
		IsPostgres = true
		return migrate()
	}

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	var openStr string
	if dsn == "" {
		openStr = "file:broomax.db?cache=shared&mode=rwc"
	} else {
		openStr = dsn
	}
	db, err := sql.Open("sqlite", openStr)
	if err != nil {
		return err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return err
	}
	DB = db
	IsPostgres = false
	return migrate()
}

func migrate() error {
	var ddl string
	if IsPostgres {
		ddl = `
CREATE TABLE IF NOT EXISTS admins (
	id BIGSERIAL PRIMARY KEY,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL DEFAULT '',
	created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
`
	} else {
		ddl = `
CREATE TABLE IF NOT EXISTS admins (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	email TEXT NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	name TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
`
	}
	if _, err := DB.Exec(ddl); err != nil {
		return fmt.Errorf("migrate admins: %w", err)
	}
	return nil
}
