package backend

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
)

// HandleLike gère les likes et dislikes sans utiliser JSON
func HandleLike(db *sql.DB) http.HandlerFunc {
	
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifie la méthode
		if r.Method != http.MethodPost {
			
			Render(w,http.StatusMethodNotAllowed)
			return
		}

		// Vérifie la session utilisateur
		userID, err := GetUserIDFromSession(r, db)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Récupère les valeurs envoyées depuis un formulaire HTML
		postID := r.FormValue("post_id")
		value := r.FormValue("value") // "1" pour like, "-1" pour dislike
		var dummy int                 // or matching columns
		err = db.QueryRow("SELECT id FROM posts WHERE id = ?", postID).Scan(&dummy)

		if err == sql.ErrNoRows {
			Render(w, http.StatusBadRequest)
			return
		}

		if err != nil {

			log.Println("DB error:", err)
			return
		}
		if postID == "" || value == "" {
			
			Render(w,http.StatusBadRequest)
			return
		}

		// Vérifie si un like existe déjà
		var existing int
		err = db.QueryRow("SELECT kind FROM likes WHERE user_id = ? AND post_id = ?", userID, postID).Scan(&existing)

		if err == sql.ErrNoRows {
			// Premier like
			_, err = db.Exec("INSERT INTO likes (user_id, post_id, kind) VALUES (?, ?, ?)", userID, postID, value)
		} else if err == nil {
			if fmt.Sprint(existing) == value {
				// Même choix -> suppression
				_, err = db.Exec("DELETE FROM likes WHERE user_id = ? AND post_id = ?", userID, postID)
			} else {
				// Changement (like <-> dislike)
				_, err = db.Exec("UPDATE likes SET kind = ? WHERE user_id = ? AND post_id = ?", value, userID, postID)
			}
		}

		if err != nil {
			
			Render(w,http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Referer()+"#"+postID, http.StatusSeeOther)
	}
}

func HandleCommentLike(db *sql.DB) http.HandlerFunc {
	
	return func(w http.ResponseWriter, r *http.Request) {
		// Vérifie la méthode
		if r.Method != http.MethodPost {
			
			Render(w, http.StatusMethodNotAllowed)
			return
		}
		// Vérifie la session utilisateur
		userID, err := GetUserIDFromSession(r, db)
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		// Récupère les valeurs envoyées depuis un formulaire HTML
		commentID := r.FormValue("comment_id")
		value := r.FormValue("value") // "1" pour like, "-1" pour dislike
		var dummy int                 // or matching columns
		err = db.QueryRow("SELECT id FROM comments WHERE id = ?", commentID).Scan(&dummy)

		if err == sql.ErrNoRows {
			Render(w, http.StatusBadRequest)
			return
		}

		if err != nil {

			log.Println("DB error:", err)
			return
		}
		if commentID == "" || value == "" {
			
			Render(w, http.StatusBadRequest)
			return
		}
		// Vérifie si un like existe déjà
		var existing int
		err = db.QueryRow("SELECT kind FROM comments_like WHERE user_id = ? AND comment_id = ?", userID, commentID).Scan(&existing)

		if err == sql.ErrNoRows {
			// Premier like
			_, err = db.Exec("INSERT INTO comments_like (user_id, comment_id, kind) VALUES (?, ?, ?)", userID, commentID, value)
		} else if err == nil {
			if fmt.Sprint(existing) == value {
				// Même choix -> suppression
				_, err = db.Exec("DELETE FROM comments_like WHERE user_id = ? AND comment_id = ?", userID, commentID)
			} else {
				// Changement (like <-> dislike)
				_, err = db.Exec("UPDATE comments_like SET kind = ? WHERE user_id = ? AND comment_id = ?", value, userID, commentID)
			}
		}

		if err != nil {
			
			Render(w, http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, r.Referer()+"#comment"+commentID, http.StatusSeeOther)
	}
}

// InitLikesTable creates the likes table for posts if it doesn't exist.
func InitLikesTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS likes (
		user_id INTEGER NOT NULL,
		post_id INTEGER NOT NULL,
		kind INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		PRIMARY KEY(user_id, post_id)
	)`)
	return err
}

// InitCommentLikesTable creates the likes table for comments if it doesn't exist.
func InitCommentLikesTable(db *sql.DB) error {
	_, err := db.Exec(`
	CREATE TABLE IF NOT EXISTS comments_like (
		user_id INTEGER NOT NULL,
		comment_id INTEGER NOT NULL,
		kind INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		PRIMARY KEY(user_id, comment_id)
	)`)
	return err
}
