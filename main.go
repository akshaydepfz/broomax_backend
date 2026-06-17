package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"oryoo.com/handler"
	"oryoo.com/helper"
)

func writeAuthJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// adminPublicPaths are /admin/* routes that skip Bearer/JWT checks.
// Add new public admin endpoints here (e.g. password reset) so they stay outside the gate.
var adminPublicPaths = map[string]struct{}{
	"/admin/login":        {},
	"/admin/create-admin": {},
}

// adminAuthMiddleware protects /admin/* except paths in adminPublicPaths.
//
// Authenticated requests must send:
//
//	Authorization: Bearer <token>
//
// Where <token> is either:
//   - The value of env ADMIN_BEARER_SECRET (static API secret for scripts/automation), or
//   - A JWT issued by POST /admin/login (signed with JWT_SECRET; validates expiry and issuer).
//
// Prefer the login JWT for interactive admin sessions (it identifies the admin subject).
// Use ADMIN_BEARER_SECRET for trusted backends or ops automation that should not depend on a user login.
func adminAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if strings.HasPrefix(path, "/admin/") {
			if _, ok := adminPublicPaths[path]; ok {
				next.ServeHTTP(w, r)
				return
			}

			auth := r.Header.Get("Authorization")
			const prefix = "Bearer "
			if !strings.HasPrefix(auth, prefix) {
				writeAuthJSON(w, http.StatusUnauthorized, `{"success":false,"error":"missing or invalid authorization"}`)
				return
			}
			raw := strings.TrimSpace(strings.TrimPrefix(auth, prefix))
			if raw == "" {
				writeAuthJSON(w, http.StatusUnauthorized, `{"success":false,"error":"missing or invalid authorization"}`)
				return
			}

			static := os.Getenv("ADMIN_BEARER_SECRET")
			if static != "" && raw == static {
				next.ServeHTTP(w, r)
				return
			}

			if _, err := helper.ParseAndVerifyAdminJWT(raw); err != nil {
				writeAuthJSON(w, http.StatusUnauthorized, `{"success":false,"error":"invalid or expired token"}`)
				return
			}
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	if err := helper.InitDB(); err != nil {
		log.Fatalf("database: %v", err)
	}

	http.HandleFunc("/admin/login", handler.AdminLoginHandler)
	http.HandleFunc("/admin/create-admin", handler.CreateAdminHandler)

	mux := http.DefaultServeMux
	addr := ":8080"
	if p := os.Getenv("PORT"); p != "" {
		addr = ":" + strings.TrimPrefix(strings.TrimSpace(p), ":")
	}

	log.Printf("listening on %s", addr)
	log.Fatal(http.ListenAndServe(addr, enableCors(adminAuthMiddleware(mux))))
}
