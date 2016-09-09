package main

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
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
	http.HandleFunc("/s/deactivate/", deactivateHandler)
	http.HandleFunc("/s/activate/", activateHandler)
	http.HandleFunc("/hosts", hostsHandler)

	fs := http.FileServer(http.Dir("web"))
	http.Handle("/static", http.StripPrefix("/static/", fs))
	http.ListenAndServe(serverConfig.Host+":"+strconv.Itoa(serverConfig.Port), nil)
}

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["services"] = reg.fetchServices()
	renderTemplate(w, r.URL.Path[1:]+".html", data)
}

func deactivateHandler(w http.ResponseWriter, r *http.Request) {
	//deactivate the service
	i := strings.LastIndex(r.URL.Path, "/")
	base64ID := r.URL.Path[i+1:]

	if b, err := base64.StdEncoding.DecodeString(base64ID); err == nil {
		// TODO
		fmt.Println(string(b))
	}
	http.Redirect(w, r, "/services", http.StatusFound)
}

func activateHandler(w http.ResponseWriter, r *http.Request) {
	//activate the service
	i := strings.LastIndex(r.URL.Path, "/")
	base64ID := r.URL.Path[i+1:]

	if b, err := base64.StdEncoding.DecodeString(base64ID); err == nil {
		// TODO
		fmt.Println(string(b))
	}

	http.Redirect(w, r, "/services", http.StatusFound)
}

func hostsHandler(w http.ResponseWriter, r *http.Request) {

	renderTemplate(w, r.URL.Path[1:]+".html", nil)
}

type Registry interface {
	fetchServices() []*Service
}

// Service is a service endpoint
type Service struct {
	Id       string
	Name     string
	Address  string
	Metadata string
	State    string
}
