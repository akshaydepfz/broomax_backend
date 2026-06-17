package models

// Admin is a safe admin profile for JSON responses (no secrets).
type Admin struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// AdminLoginRequest is the POST /admin/login body.
type AdminLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// AdminLoginResponseData is the `data` object returned on successful login.
type AdminLoginResponseData struct {
	Token string `json:"token"`
	Admin Admin  `json:"admin"`
}

// CreateAdminRequest is the POST /admin/create-admin body.
type CreateAdminRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}
