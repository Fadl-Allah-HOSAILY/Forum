package main

import (
	"log"
	"net/http"

	"forum/backend"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	DB := backend.InitDB()

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

	log.Println("Server running at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
