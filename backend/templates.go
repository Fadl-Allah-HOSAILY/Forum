package backend

import (
	"bytes"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
)

var templates *template.Template

func LoadTemplates(pattern string) {
	var err error
	templates, err = template.ParseGlob(pattern)
	if err != nil {
		log.Fatalf("Failed to parse templates: %v", err)
	}
}

func IsSafeTemplateName(name string) bool {
	matched, _ := regexp.MatchString(`^[a-zA-Z0-9_\-]+\.html$`, filepath.Base(name))
	return matched
}

func RenderTemplate(w http.ResponseWriter, name string, data any) {
	if templates == nil {
		http.Error(w, "Templates not loaded", http.StatusInternalServerError)
		return
	}

	if !IsSafeTemplateName(name) {
		http.Error(w, "Invalid template name", http.StatusBadRequest)
		return
	}

	var buf bytes.Buffer

	err := templates.ExecuteTemplate(&buf, name, data)
	if err != nil {
		http.Error(w, "Template rendering error: "+err.Error(), http.StatusInternalServerError)
		log.Println("Render error:", err)
		return
	}

	if _, err := buf.WriteTo(w); err != nil {
		log.Println("Write error:", err)
	}
}
