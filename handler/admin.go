package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"oryoo.com/helper"
	"oryoo.com/models"
)

type jsonEnvelope struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// AdminLoginHandler handles POST /admin/login.
//
// Request JSON: models.AdminLoginRequest (email, password).
// Success JSON: {"success":true,"data":{"token":"<jwt>","admin":{...}}}
// Errors: {"success":false,"error":"..."} with 4xx/5xx as appropriate.
func AdminLoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, jsonEnvelope{Success: false, Error: "method not allowed"})
		return
	}
	var req models.AdminLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonEnvelope{Success: false, Error: "invalid JSON body"})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, jsonEnvelope{Success: false, Error: "email and password are required"})
		return
	}

	row, err := helper.GetAdminByEmail(req.Email)
	if err != nil {
		if errors.Is(err, helper.ErrNotFound) {
			writeJSON(w, http.StatusUnauthorized, jsonEnvelope{Success: false, Error: "invalid email or password"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, jsonEnvelope{Success: false, Error: "internal error"})
		return
	}
	if !helper.CheckPassword(row.PasswordHash, req.Password) {
		writeJSON(w, http.StatusUnauthorized, jsonEnvelope{Success: false, Error: "invalid email or password"})
		return
	}

	token, err := helper.SignAdminJWT(row.ID, row.Email)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, jsonEnvelope{Success: false, Error: "could not issue token"})
		return
	}

	writeJSON(w, http.StatusOK, jsonEnvelope{
		Success: true,
		Data: models.AdminLoginResponseData{
			Token: token,
			Admin: models.Admin{
				ID:    row.ID,
				Email: row.Email,
				Name:  row.Name,
			},
		},
	})
}

// CreateAdminHandler handles POST /admin/create-admin (bootstrap).
func CreateAdminHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, jsonEnvelope{Success: false, Error: "method not allowed"})
		return
	}
	var req models.CreateAdminRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, jsonEnvelope{Success: false, Error: "invalid JSON body"})
		return
	}
	req.Email = strings.TrimSpace(req.Email)
	req.Password = strings.TrimSpace(req.Password)
	req.Name = strings.TrimSpace(req.Name)
	if req.Email == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, jsonEnvelope{Success: false, Error: "email and password are required"})
		return
	}

	hash, err := helper.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, jsonEnvelope{Success: false, Error: "could not hash password"})
		return
	}

	id, err := helper.CreateAdmin(req.Email, hash, req.Name)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			writeJSON(w, http.StatusConflict, jsonEnvelope{Success: false, Error: "email already registered"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, jsonEnvelope{Success: false, Error: "could not create admin"})
		return
	}

	writeJSON(w, http.StatusCreated, jsonEnvelope{
		Success: true,
		Data: models.Admin{
			ID:    id,
			Email: strings.ToLower(req.Email),
			Name:  req.Name,
		},
	})
}
