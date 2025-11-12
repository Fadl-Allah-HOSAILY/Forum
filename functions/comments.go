package functions

import (
	"database/sql"
	"fmt"
	"net/http"
)

func HandleAddComment(db *sql.DB) http.HandlerFunc {
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

		// Récupère les champs du formulaire
		postID := r.FormValue("post_id")
		content := r.FormValue("content")

		if postID == "" || content == "" {
			http.Error(w, "Champs manquants", http.StatusBadRequest)
			return
		}

		// Insérer le commentaire
		_, err = db.Exec("INSERT INTO comments (post_id, user_id, content) VALUES (?, ?, ?)", postID, userID, content)
		if err != nil {
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}

		// Récupère la liste mise à jour des commentaires
		rows, err := db.Query(`
			SELECT u.username, c.content, c.created_at
			FROM comments c
			JOIN users u ON u.id = c.user_id
			WHERE c.post_id = ?
			ORDER BY c.created_at DESC`, postID)
		if err != nil {
			http.Error(w, "Erreur base de données", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// On renvoie du HTML directement pour mettre à jour dynamiquement
		var responseHTML string
		for rows.Next() {
			var username, content, createdAt string
			rows.Scan(&username, &content, &createdAt)
			responseHTML += fmt.Sprintf("<p><strong>%s</strong>: %s <em>(%s)</em></p>", username, content, createdAt)
		}

		// Réponse texte/html
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, responseHTML)
	}
}
