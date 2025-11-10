package functions

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type LikeRequest struct {
    PostID int `json:"post_id"`
    Value  int `json:"value"` // 1 for like, -1 for dislike
}

func HandleLike(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Vérifie que l'utilisateur est connecté (exemple : via cookie ou session)
        userID, err := GetUserIDFromSession(r, db)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Parse le JSON envoyé depuis le frontend
        var req LikeRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        // Vérifie si un like existe déjà
        var existing int
        err = db.QueryRow("SELECT value FROM likes WHERE user_id = ? AND post_id = ?", userID, req.PostID).Scan(&existing)

        if err == sql.ErrNoRows {
            // Premier like → on insère
            _, err = db.Exec("INSERT INTO likes (user_id, post_id, value) VALUES (?, ?, ?)", userID, req.PostID, req.Value)
        } else if err == nil {
            if existing == req.Value {
                // Si l'utilisateur reclique sur le même bouton → supprime le like
                _, err = db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ?", userID, req.PostID)
            } else {
                // Sinon, il change de choix (like ↔ dislike)
                _, err = db.Exec("UPDATE likes SET value = ? WHERE user_id = ? AND post_id = ?", req.Value, userID, req.PostID)
            }
        }

        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        // Compte les likes et dislikes du post
        var likesCount, dislikesCount int
        db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = 1", req.PostID).Scan(&likesCount)
        db.QueryRow("SELECT COUNT(*) FROM likes WHERE post_id = ? AND value = -1", req.PostID).Scan(&dislikesCount)

        // Réponse JSON
        json.NewEncoder(w).Encode(map[string]int{
            "likes":    likesCount,
            "dislikes": dislikesCount,
        })
    }
}