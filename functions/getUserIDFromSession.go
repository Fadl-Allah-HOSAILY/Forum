package functions

import (
	"database/sql"
	"errors"
	"net/http"
	"time"
)

// GetUserIDFromSession récupère l'ID utilisateur depuis le cookie "session_token"
func GetUserIDFromSession(r *http.Request, db *sql.DB) (int, error) {
    // Lire le cookie envoyé par le navigateur
    cookie, err := r.Cookie("session_token")
    if err != nil {
        return 0, errors.New("no session cookie")
    }

    // Chercher le token dans la base
    var userID int
    var expiresAt time.Time
    err = db.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token = ?", cookie.Value).Scan(&userID, &expiresAt)
    if err != nil {
        return 0, errors.New("invalid session")
    }

    // Vérifier si la session est expirée
    if time.Now().After(expiresAt) {
        // Supprimer la session expirée
        db.Exec("DELETE FROM sessions WHERE token = ?", cookie.Value)
        return 0, errors.New("session expired")
    }

    return userID, nil
}