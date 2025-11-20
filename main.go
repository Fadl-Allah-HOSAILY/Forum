package main

import (
	"database/sql"
	"log"
	"net/http"

	"forum/backend"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	DB, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		log.Fatal(err)
	}
	defer DB.Close()

	// initialize likes tables
	if err := backend.InitLikesTable(DB); err != nil {
		log.Fatalf("failed to init likes table: %v", err)
	}
	if err := backend.InitCommentLikesTable(DB); err != nil {
		log.Fatalf("failed to init comment likes table: %v", err)
	}

	backend.LoadTemplates("templates/*.html")
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.HandleFunc("/post", backend.HandlePost(DB))

	http.HandleFunc("/", backend.Handler(DB))

	http.HandleFunc("/like", backend.HandleLike(DB))
	http.HandleFunc("/commentlike", backend.HandleCommentLike(DB))
	// http.HandleFunc("/static", backend.HandlerStatic)
	http.HandleFunc("/signup", backend.NotAuthRequired(DB, backend.SignupHandler(DB)))
	http.HandleFunc("/login", backend.NotAuthRequired(DB, backend.LoginHandler(DB)))

	http.HandleFunc("/logout", backend.LogoutHandler(DB))
	http.HandleFunc("/comment", backend.HandleAddComment(DB))

	log.Println("Server running at http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}
