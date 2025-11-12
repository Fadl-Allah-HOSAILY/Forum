package functions

import (
	"database/sql"
	"fmt"
	"net/http"
)

// HandleLike gère les likes et dislikes sans utiliser JSON
func HandleLike(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifie la méthode
		if r.Method != http.MethodPost {
			http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
			return
		}

		// Vérifie la session utilisateur
		userID, err := GetUserIDFromSession(r, db)
		if err != nil {
			http.Error(w, "Non autorisé", http.StatusUnauthorized)
			return
		}

		// Récupère les valeurs envoyées depuis un formulaire HTML
		postID := r.FormValue("post_id")
		value := r.FormValue("value") // "1" pour like, "-1" pour dislike

		if postID == "" || value == "" {
			http.Error(w, "Paramètres manquants", http.StatusBadRequest)
			return
		}

		// Vérifie si un like existe déjà
		var existing int
		err = db.QueryRow("SELECT value FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existing)

		if err == sql.ErrNoRows {
			// Premier like
			_, err = db.Exec("INSERT INTO likes (user_id, post_id, value) VALUES (?, ?, ?)", userID, postID, value)
		} else if err == nil {
			if fmt.Sprint(existing) == value {
				// Même choix -> suppression
				_, err = db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ?", userID, postID)
			} else {
				// Changement (like <-> dislike)
				_, err = db.Exec("UPDATE likes SET value = ? WHERE user_id = ? AND post_id = ?", value, userID, postID)
			}
		}

		if err != nil {
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}

		// Compte les likes et dislikes
		var likesCount, dislikesCount int
		db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1", postID).Scan(&likesCount)
		db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1", postID).Scan(&dislikesCount)

		// Réponse simple en texte
		fmt.Fprintf(w, "likes=%d;dislikes=%d", likesCount, dislikesCount)
	}
}
