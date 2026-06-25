package helper

import (
	"database/sql"
	"errors"
	"strings"
	"time"
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

// AdminCredentialRow is the internal DB shape for admin_credentials.
type AdminCredentialRow struct {
	ID           int64
	Email        string
	PasswordHash string
	LastLoggedAt sql.NullTime
}

// GetAdminCredentialByEmail loads an admin credential by email (case-insensitive).
func GetAdminCredentialByEmail(email string) (AdminCredentialRow, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	var row AdminCredentialRow
	var err error
	if IsPostgres {
		err = DB.QueryRow(
			`SELECT id, email, password_hash, last_logged_at FROM admin_credentials WHERE lower(email) = $1`,
			email,
		).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.LastLoggedAt)
	} else {
		err = DB.QueryRow(
			`SELECT id, email, password_hash, last_logged_at FROM admin_credentials WHERE lower(email) = ?`,
			email,
		).Scan(&row.ID, &row.Email, &row.PasswordHash, &row.LastLoggedAt)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return AdminCredentialRow{}, ErrNotFound
	}
	return row, err
}

// CreateAdminCredential inserts a new admin credential; email is normalized to lower trim.
func CreateAdminCredential(email, passwordHash string) (int64, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if IsPostgres {
		var id int64
		err := DB.QueryRow(
			`INSERT INTO admin_credentials (email, password_hash) VALUES ($1, $2) RETURNING id`,
			email, passwordHash,
		).Scan(&id)
		return id, err
	}
	res, err := DB.Exec(
		`INSERT INTO admin_credentials (email, password_hash) VALUES (?, ?)`,
		email, passwordHash,
	)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateAdminCredentialLastLoggedAt sets last_logged_at to now for the given credential id.
func UpdateAdminCredentialLastLoggedAt(id int64, at time.Time) error {
	if IsPostgres {
		_, err := DB.Exec(
			`UPDATE admin_credentials SET last_logged_at = $1 WHERE id = $2`,
			at, id,
		)
		return err
	}
	_, err := DB.Exec(
		`UPDATE admin_credentials SET last_logged_at = ? WHERE id = ?`,
		at.UTC().Format(time.RFC3339), id,
	)
	return err
}
