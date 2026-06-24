package database

import (
	"database/sql"
	_ "embed"
	"fmt"
	"strings"
)

//go:embed migrations/001_products.sql
var productsMigrationSQL string

// ApplyMigrationSQL executes SQL statements separated by semicolons.
func ApplyMigrationSQL(db *sql.DB, raw string) error {
	for _, stmt := range splitSQL(raw) {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("migration failed: %w\nstatement: %s", err, truncate(stmt, 120))
		}
	}
	return nil
}

// MigrateProducts applies the embedded product catalog schema (PostgreSQL only).
func MigrateProducts(db *sql.DB) error {
	return ApplyMigrationSQL(db, productsMigrationSQL)
}

func splitSQL(input string) []string {
	var out []string
	var b strings.Builder
	inSingle := false
	inDouble := false
	for i := 0; i < len(input); i++ {
		ch := input[i]
		if ch == '\'' && !inDouble {
			inSingle = !inSingle
		}
		if ch == '"' && !inSingle {
			inDouble = !inDouble
		}
		if ch == ';' && !inSingle && !inDouble {
			stmt := strings.TrimSpace(b.String())
			if stmt != "" && !strings.HasPrefix(stmt, "--") {
				out = append(out, stmt)
			}
			b.Reset()
			continue
		}
		b.WriteByte(ch)
	}
	if tail := strings.TrimSpace(b.String()); tail != "" {
		out = append(out, tail)
	}
	return out
}

func truncate(s string, n int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
