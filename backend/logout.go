package backend

import (
	"database/sql"
	"log"
	"net/http"
	"path"
	"strings"
	"time"
)

func LogoutHandler(DB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// URL validation
		if len(r.URL.Path) > 2048 || strings.Contains(r.URL.Path, "\x00") || strings.Contains(r.URL.Path, "..") {
			log.Printf("Suspicious logout path: %q", r.URL.Path)
			Render(w, http.StatusBadRequest)
			return
		}
		if r.URL.Path != path.Clean(r.URL.Path) {
			Render(w, http.StatusBadRequest)
			return
		}

		// Method validation
		if r.Method != http.MethodPost {
			Render(w, http.StatusMethodNotAllowed)
			return
		}

		c, err := r.Cookie("session_token")
		if err == nil {
			token := c.Value
			if _, err := DB.Exec("DELETE FROM sessions WHERE token = ?", token); err != nil {
				log.Printf("Error deleting session: %v", err)
			}
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    "",
			Path:     "/",
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
		http.Redirect(w, r, "/", http.StatusSeeOther)
	}
}
