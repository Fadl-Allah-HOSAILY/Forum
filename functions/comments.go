package functions

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

type CommentRequest struct {
    PostID  int    `json:"post_id"`
    Content string `json:"content"`
}

func HandleAddComment(db *sql.DB) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Vérifie si l'utilisateur est connecté
        userID, err := GetUserIDFromSession(r, db)
        if err != nil {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }

        // Décoder le JSON envoyé depuis le frontend
        var req CommentRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            http.Error(w, "Invalid request", http.StatusBadRequest)
            return
        }

        // Vérifier que le contenu n’est pas vide
        if req.Content == "" {
            http.Error(w, "Empty comment", http.StatusBadRequest)
            return
        }

        // Insérer le commentaire
        _, err = db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)",
            req.PostID, userID, req.Content)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }

        // Retourner tous les commentaires mis à jour pour ce post
        rows, err := db.Query(`
            SELECT c.id, u.username, c.content, c.created_at
            FROM comments c
            JOIN users u ON u.id = c.user_id
            WHERE c.post_id = ?
            ORDER BY c.created_at DESC`, req.PostID)
        if err != nil {
            http.Error(w, "Database error", http.StatusInternalServerError)
            return
        }
        defer rows.Close()

        type CommentResponse struct {
            ID        int    `json:"id"`
            Username  string `json:"username"`
            Content   string `json:"content"`
            CreatedAt string `json:"created_at"`
        }

        var comments []CommentResponse
        for rows.Next() {
            var c CommentResponse
            rows.Scan(&c.ID, &c.Username, &c.Content, &c.CreatedAt)
            comments = append(comments, c)
        }

        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(comments)
    }
}