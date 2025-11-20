package backend

import (
	"bytes"
	"crypto/rand"
	"database/sql"
	"log"
	"net/http"
	"path"
	"regexp"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func SignupHandler(DB *sql.DB) http.HandlerFunc {
	usernameRe := regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

	emailRe := regexp.MustCompile(
		`^[A-Za-z0-9](?:[A-Za-z0-9._%+\-]{0,63}[A-Za-z0-9])?` +
			`@` +
			`[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?` +
			`(?:\.[A-Za-z0-9](?:[A-Za-z0-9\-]{0,61}[A-Za-z0-9])?)+$`)

	return func(w http.ResponseWriter, r *http.Request) {
		if len(r.URL.Path) > 2048 || strings.Contains(r.URL.Path, "\x00") || strings.Contains(r.URL.Path, "..") {
			log.Printf("Suspicious signup path: %q", r.URL.Path)
			Render(w, http.StatusBadRequest)
			return
		}
		if r.URL.Path != path.Clean(r.URL.Path) {
			Render(w, http.StatusBadRequest)
			return
		}

		// --- GET Method ---
		if r.Method == http.MethodGet {
			if r.URL.Query().Get("username") != "" || r.URL.Query().Get("email") != "" || r.URL.Query().Get("password") != "" {
				Render(w, http.StatusBadRequest)
				return
			}
			var buf bytes.Buffer

			err := templates.ExecuteTemplate(&buf, "signup.html", map[string]string{
				"Error": "",
			})
			if err != nil {
				log.Printf("Template render error (GET /signup): %v", err)
				Render(w, http.StatusInternalServerError)
				return
			}

			if _, err := buf.WriteTo(w); err != nil {
				log.Printf("Write error (GET /signup): %v", err)
			}
			return

		}

		// --- POST Method ---
		if err := r.ParseForm(); err != nil {
			Render(w, http.StatusBadRequest)
			return
		}

		username, ok := r.Form[("username")]
		if !ok {
			Render(w, http.StatusBadRequest)
			return
		}
		email := r.Form[("email")]
		if !ok {
			Render(w, http.StatusBadRequest)
			return
		}
		password, ok := r.Form[("password")]
		if !ok {
			Render(w, http.StatusBadRequest)
			return
		}
		username[0] = strings.TrimSpace(username[0])
		email[0] = strings.TrimSpace(email[0])
		password[0] = strings.TrimSpace(password[0])
		if username[0] == "" || email[0] == "" || password[0] == "" {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			templates.ExecuteTemplate(w, "signup.html", map[string]string{"Error": "All fields are required"})
			return
		}
		if !emailRe.MatchString(email[0]) {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			templates.ExecuteTemplate(w, "signup.html", map[string]string{"Error": "Invalid email"})
			return
		}
		if !usernameRe.MatchString(username[0]) {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			templates.ExecuteTemplate(w, "signup.html", map[string]string{"Error": "Invalid username"})
			return
		}
		if len(password[0]) < 8 {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			templates.ExecuteTemplate(w, "signup.html", map[string]string{"Error": "Password must be >= 8 chars"})
			return
		}

		var exists int
		if err := DB.QueryRow("SELECT COUNT(1) FROM users WHERE email = ?", email[0]).Scan(&exists); err != nil {
			log.Printf("DB error (email check): %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}
		if exists > 0 {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			templates.ExecuteTemplate(w, "signup.html", map[string]string{"Error": "Email already taken"})
			return
		}
		exists = 0
		if err := DB.QueryRow("SELECT COUNT(1) FROM users WHERE username = ?", username[0]).Scan(&exists); err != nil {
			log.Printf("DB error (email check): %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}
		if exists > 0 {
			w.WriteHeader(http.StatusNonAuthoritativeInfo)
			templates.ExecuteTemplate(w, "signup.html", map[string]string{"Error": "Username already taken"})
			return
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(password[0]), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("Password hash error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}

		tx, err := DB.Begin()
		if err != nil {
			log.Printf("DB transaction start error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}
		defer tx.Rollback()

		res, err := tx.Exec("INSERT INTO users (username, email, password_hash, created_at) VALUES (?, ?, ?, datetime('now'))", username[0], email[0], string(hash))
		if err != nil {
			log.Printf("Insert user error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}

		userID, err := res.LastInsertId()
		if err != nil {
			log.Printf("Get LastInsertId error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}

		token := generateRandomToken()
		if token == "" {
			log.Printf("Token generation failed")
			Render(w, http.StatusInternalServerError)
			return
		}
		exp := time.Now().Add(24 * time.Hour)

		_, err = tx.Exec("INSERT INTO sessions(token, user_id, expires_at) VALUES (?, ?, ?)", token, userID, exp.Format("2006-01-02 15:04:05"))
		if err != nil {
			log.Printf("Insert session error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}

		if err := tx.Commit(); err != nil {
			log.Printf("DB commit error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    token,
			Expires:  exp,
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})

		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}

func generateRandomToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return ""
	}
	return hexEncode(b)
}

func hexEncode(b []byte) string {
	const hexChars = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = hexChars[int(v>>4)]
		out[i*2+1] = hexChars[int(v&0x0f)]
	}
	return string(out)
}
