package backend

import (
	"database/sql"
	"log"
)

type Category struct {
	ID       int
	Name     string
	Selected bool
}

// GetCategories returns all categories from the DB ordered by name.
func GetCategories(DB *sql.DB) []Category {
	cats := []Category{}
	rows, err := DB.Query(`SELECT id, categorie FROM categories ORDER BY categorie`)
	if err != nil {
		log.Println("GetCategories query error:", err)
		return cats
	}
	defer rows.Close()
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name); err != nil {
			log.Println("GetCategories scan error:", err)
			return cats
		}
		cats = append(cats, c)
	}
	return cats
}
