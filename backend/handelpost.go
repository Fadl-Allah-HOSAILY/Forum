package backend

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
)

type PostPageData struct {
	Popup         bool
	Username      string
	Posts         []Datapost
	Error         string
	Cachetitle    string
	Cacheconetent string
	Categories    []string
	Path          string
}

func Handler(DB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			Render(w, 404)
			return

		} else if r.Method != http.MethodGet {
			// return 405 for non-GET methods
			Render(w, http.StatusMethodNotAllowed)
			return
		}
		
		http.Redirect(w, r, "/post", http.StatusSeeOther)
	}
}

func HandlePost(DB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tmp, err := template.ParseFiles("templates/post.html")
		if err != nil {
			log.Printf("template parse error: %v", err)
			Render(w, http.StatusInternalServerError)
			return
		}
		Categories := r.URL.Query().Get("Categories")
		http.SetCookie(w, &http.Cookie{
			Name:  "LastPath",
			Value: r.RequestURI,
			Path:  "/post",
		})
		if r.URL.Path != "/post" {
			Render(w, 404)
			return
		}

		if r.Method == http.MethodGet {
			// r.URL.Query().Has doesn't exist -> use Get()
			if r.URL.Query().Get("title") != "" || r.URL.Query().Get("content") != "" {
				Render(w, http.StatusBadRequest)
				return
			}
			Data := &PostPageData{}
			userid := GetUserIDFromRequest(DB, r)
			username := ""
			if userid != 0 {
				err := DB.QueryRow("SELECT username FROM users WHERE id = ?", userid).Scan(&username)
				if err != nil {
					fmt.Print(err)
					return
				}
			}

			query := r.URL.RawQuery

			if query != "" && !CheckFiltere(w, r, query, username) {

				Render(w, 404)
				return
			}

			post := GetPost(DB, Categories, username, userid)

			Data = &PostPageData{
				Username:   username,
				Posts:      post,
				Categories: []string{"Technology", "Science", "Education", "Engineering", "Entertainment"},
			}

			var buf bytes.Buffer
			if err := tmp.Execute(&buf, Data); err != nil {
				log.Printf("template execute error: %v", err)
				Render(w, http.StatusInternalServerError)
				return
			}

			_, err := buf.WriteTo(w)
			if err != nil {
				log.Printf("write error: %v", err)
			}

			return

		}
		if r.Method == http.MethodPost {
			userId := GetUserIDFromRequest(DB, r)

			if userId == 0 {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			} else {
				if err := r.ParseForm(); err != nil {

					Render(w, http.StatusBadRequest)
					return
				}
				title, titleOK := r.Form["title"]
				content, contentOK := r.Form["content"]

				if !titleOK || !contentOK {

					Render(w, http.StatusBadRequest)

					return
				}
				title[0] = strings.TrimSpace(title[0])
				content[0] = strings.TrimSpace(content[0])

				if len(title[0]) == 0 {
					RenderTemplate(w, "post.html", CheckDataPost(DB, r, "⚠️ Your post needs a title."))
					return
				}

				if len(title[0]) > 30 {
					RenderTemplate(w, "post.html", CheckDataPost(DB, r, "⚠️ Title too long (max 30 characters)."))
					return
				}

				if len(content[0]) == 0 {
					RenderTemplate(w, "post.html", CheckDataPost(DB, r, "⚠️ Your post needs some content."))
					return
				}

				if len(content[0]) < 20 || len(content[0]) > 10000 {
					RenderTemplate(w, "post.html", CheckDataPost(DB, r, "⚠️ Content must be at least 20 characters."))
					return
				}

				// var category []string
				category := r.Form["category_ids"]

				if len(category) == 0 {
					errorMsg := "⚠️ You must choose one category or more"

					RenderTemplate(w, "post.html", CheckDataPost(DB, r, errorMsg))
					return
				}

				insrtpost := `INSERT INTO posts (title,content,user_id) VALUES (?,?,?)`
				stmt, err := DB.Prepare(insrtpost)
				if err != nil {
					fmt.Println(err)
					return
				}

				defer stmt.Close()
				res, err := stmt.Exec(title[0], content[0], userId)
				if err != nil {
					fmt.Println(err)
					return
				}
				IdPost, err := res.LastInsertId()
				if err != nil {
					fmt.Println("Error getting last insert ID:", err)
					Render(w, http.StatusInternalServerError)
					return
				}
				InsertCategoriId(DB, IdPost, category)
				http.Redirect(w, r, "/post", http.StatusSeeOther)
			}
		}
	}
}

func HandlerStatic(DB *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET requests
		if r.Method != http.MethodGet {
			Render(w, 405)
			return
		} else {
			// Check if the requested file exists and is not a directory
			info, err := os.Stat(r.URL.Path[1:])

			if err != nil {
				// file not found
				Render(w, http.StatusNotFound)
				return
			} else if info.IsDir() {
				Render(w, 403)
				return
			} else {
				// Serve the static file
				http.ServeFile(w, r, r.URL.Path[1:])
			}
		}
	}
}
