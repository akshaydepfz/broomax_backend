package helper

import (
	"database/sql"
	"errors"
	"strings"
)

// ErrNotFound is returned when a row is missing.
var ErrNotFound = errors.New("not found")

// AdminRow is internal DB shape for an admin.
type AdminRow struct {
	ID           int64
	Email        string
	PasswordHash string
	Name         string
}

// GetAdminByEmail loads an admin by email (case-insensitive match on stored email).
func GetAdminByEmail(email string) (AdminRow, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var row AdminRow
	var err error
	if IsPostgres {
		err = DB.QueryRow(
			`SELECT id, email, password_hash, COALESCE(name, '') FROM admins WHERE lower(email) = $1`,
			email,
		).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Name)
	} else {
		err = DB.QueryRow(
			`SELECT id, email, password_hash, COALESCE(name, '') FROM admins WHERE lower(email) = ?`,
			email,
		).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.Name)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return AdminRow{}, ErrNotFound
	}
	return row, err
}

// CreateAdmin inserts a new admin; email is normalized to lower trim.
func CreateAdmin(email, passwordHash, name string) (int64, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	name = strings.TrimSpace(name)
	if IsPostgres {
		var id int64
		err := DB.QueryRow(
			`INSERT INTO admins (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id`,
			email, passwordHash, name,
		).Scan(&id)
		return id, err
	}
	res, err := DB.Exec(
		`INSERT INTO admins (email, password_hash, name) VALUES (?, ?, ?)`,
		email, passwordHash, name,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}
