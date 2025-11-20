package backend

import (
	"database/sql"
	"net/http"
)

func GetUserIDFromRequest(DB *sql.DB, r *http.Request) int64 {
	c, err := r.Cookie("session_token")
	if err != nil {
		return 0
	}
	token := c.Value

	var userID int64

	err = DB.QueryRow(
		"SELECT user_id FROM sessions WHERE token = ? AND expires_at > datetime('now')",
		token,
	).Scan(&userID)
	if err != nil {
		return 0
	}

	return userID
}

func AuthRequired(DB *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GetUserIDFromRequest(DB, r) == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func NotAuthRequired(DB *sql.DB, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if GetUserIDFromRequest(DB, r) != 0 {
			http.Redirect(w, r, "/post", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}
}