package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed migrations/001_products.sql
var productsMigrationSQL string

// ApplyMigrationSQL runs the full migration script inside a single transaction.
// PostgreSQL parses semicolons and line comments server-side, so the script is
// not split on the client (client-side splitting can drop statements that are
// preceded by -- comments).
func ApplyMigrationSQL(db *sql.DB, raw string) error {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("migration begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.Exec(raw); err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migration commit: %w", err)
	}
	return nil
}

// MigrateProducts applies the embedded product catalog schema (PostgreSQL only).
func MigrateProducts(db *sql.DB) error {
	return ApplyMigrationSQL(db, productsMigrationSQL)
}
