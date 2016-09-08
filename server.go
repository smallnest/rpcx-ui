package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
)

var templates map[string]*template.Template

// Load templates on program initialisation
func init() {
	//https: //elithrar.github.io/article/approximating-html-template-inheritance/

	if templates == nil {
		templates = make(map[string]*template.Template)
	}

	templatesDir := "./templates/"

	//pages to show indeed
	bases, err := filepath.Glob(templatesDir + "bases/*.html")
	if err != nil {
		log.Fatal(err)
	}

	//widgts, header, footer, sidebar, etc.
	includes, err := filepath.Glob(templatesDir + "includes/*.html")
	if err != nil {
		log.Fatal(err)
	}

	// Generate our templates map from our bases/ and includes/ directories
	for _, base := range bases {
		files := append(includes, base)
		templates[filepath.Base(base)] = template.Must(template.ParseFiles(files...))
	}
}

func renderTemplate(w http.ResponseWriter, name string, data map[string]interface{}) error {
	// Ensure the template exists in the map.
	tmpl, ok := templates[name]
	if !ok {
		return fmt.Errorf("The template %s does not exist.", name)
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	return tmpl.ExecuteTemplate(w, name, data)
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/services", http.StatusFound)
	})
	http.HandleFunc("/services", servicesHandler)
	http.HandleFunc("/hosts", hostsHandler)

	fs := http.FileServer(http.Dir("web"))
	http.Handle("/static", http.StripPrefix("/static/", fs))
	http.ListenAndServe(":8972", nil)
}

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, r.URL.Path[1:]+".html", nil)
}

func hostsHandler(w http.ResponseWriter, r *http.Request) {

	renderTemplate(w, r.URL.Path[1:]+".html", nil)
}
