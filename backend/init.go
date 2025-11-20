package backend

import (
	"database/sql"
	"log"
)

func InitDB() *sql.DB {
	var err error
	DB, err := sql.Open("sqlite3", "forum.db")
	if err != nil {
		log.Fatal(err)
	}
	err = DB.Ping()
	if err != nil {
		log.Fatal(err)
	}
	// PRAGMA settings: foreign keys must be enabled per-connection; easiest: exec now (works for the connection used)
	_, err = DB.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		log.Fatal("failed to enable foreign keys:", err)
	}
	_, err = DB.Exec("PRAGMA foreign_keys = ON;")
	if err != nil {
		log.Fatal("failed to enable foreign keys:", err)
	}
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL UNIQUE,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME DEFAULT (datetime('now'))
	);

	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		token TEXT NOT NULL UNIQUE,
		user_id INTEGER NOT NULL,
		expires_at DATETIME NOT NULL,
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS posts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		content TEXT NOT NULL,
		
		created_at DATETIME DEFAULT (datetime('now')),
		updated_at DATETIME DEFAULT (datetime('now')),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
	);

	CREATE TABLE IF NOT EXISTS comments (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		post_id INTEGER NOT NULL,
		user_id INTEGER NOT NULL,
		comment TEXT NOT NULL,
		created_at DATETIME DEFAULT (datetime('now')),
		updated_at DATETIME DEFAULT (datetime('now')),

		FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,

		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE

	);

	CREATE TABLE IF NOT EXISTS likes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		post_id INTEGER,
		comment_id INTEGER,
		kind INTEGER NOT NULL, -- 1 = like, -1 = dislike (toggle)
		created_at DATETIME DEFAULT (datetime('now')),
		FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
		FOREIGN KEY(post_id) REFERENCES posts(id) ON DELETE CASCADE,
		FOREIGN KEY(comment_id) REFERENCES comments(id) ON DELETE CASCADE
	);
	CREATE TABLE IF NOT EXISTS comments_like (
		user_id INTEGER NOT NULL,
		comment_id INTEGER NOT NULL,
		kind INTEGER NOT NULL,
		created_at DATETIME NOT NULL DEFAULT (datetime('now')),
		PRIMARY KEY(user_id, comment_id)
	);

	CREATE TABLE IF NOT EXISTS post_categories (
    post_id INTEGER,
    category_id INTEGER,
    FOREIGN KEY(post_id) REFERENCES posts(id),
    FOREIGN KEY(category_id) REFERENCES categories(id),
    PRIMARY KEY(post_id, category_id)  
);
	`

	_, err = DB.Exec(schema)
	if err != nil {
		log.Fatal(err)
	}
	if !tableExists(DB, "categories") {
		CreateCategoriestable := `CREATE TABLE IF NOT EXISTS categories(
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		categorie TEXT NOT NULL
	);`
		_, err = DB.Exec(CreateCategoriestable)
		if err != nil {
			log.Fatalf("Failed to create table: %v", err)
		}
		WriteCategories(DB)
	}
	return DB
}
