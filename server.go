package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
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

func renderTemplate(w http.ResponseWriter, name string, data interface{}) error {
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
	http.HandleFunc("/services", recoverWrapper(servicesHandler))
	http.HandleFunc("/s/deactivate/", recoverWrapper(deactivateHandler))
	http.HandleFunc("/s/activate/", recoverWrapper(activateHandler))
	http.HandleFunc("/s/m/", recoverWrapper(modifyHandler))
	http.HandleFunc("/registry", recoverWrapper(registryHandler))

	fs := http.FileServer(http.Dir("web"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(serverConfig.Host+":"+strconv.Itoa(serverConfig.Port), nil)
}

func recoverWrapper(h func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if re := recover(); re != nil {
				var err error
				fmt.Println("Recovered in registryHandler", re)
				switch t := re.(type) {
				case string:
					err = errors.New(t)
				case error:
					err = t
				default:
					err = errors.New("Unknown error")
				}
				w.WriteHeader(http.StatusOK)
				renderTemplate(w, "error.html", err.Error())
			}
		}()
		h(w, r)
	}
}

func servicesHandler(w http.ResponseWriter, r *http.Request) {
	data := make(map[string]interface{})
	data["services"] = reg.fetchServices()
	renderTemplate(w, r.URL.Path[1:]+".html", data)
}

func deactivateHandler(w http.ResponseWriter, r *http.Request) {
	i := strings.LastIndex(r.URL.Path, "/")
	base64ID := r.URL.Path[i+1:]

	if b, err := base64.StdEncoding.DecodeString(base64ID); err == nil {
		s := string(b)
		j := strings.Index(s, "@")
		name := s[0:j]
		address := s[j+1:]
		reg.deactivateService(name, address)
	}
	http.Redirect(w, r, "/services", http.StatusFound)
}

func activateHandler(w http.ResponseWriter, r *http.Request) {
	i := strings.LastIndex(r.URL.Path, "/")
	base64ID := r.URL.Path[i+1:]

	if b, err := base64.StdEncoding.DecodeString(base64ID); err == nil {
		s := string(b)
		j := strings.Index(s, "@")
		name := s[0:j]
		address := s[j+1:]
		reg.activateService(name, address)
	}

	http.Redirect(w, r, "/services", http.StatusFound)
}

func modifyHandler(w http.ResponseWriter, r *http.Request) {
	metadata := r.URL.Query()

	i := strings.LastIndex(r.URL.Path, "/")
	base64ID := r.URL.Path[i+1:]

	if b, err := base64.StdEncoding.DecodeString(base64ID); err == nil {
		s := string(b)
		j := strings.Index(s, "@")
		name := s[0:j]
		address := s[j+1:]
		reg.updateMetadata(name, address, metadata.Encode())
	}

	http.Redirect(w, r, "/services", http.StatusFound)
}

func registryHandler(w http.ResponseWriter, r *http.Request) {
	oldConfig := serverConfig
	defer func() {
		if re := recover(); re != nil {
			bytes, err := json.MarshalIndent(&oldConfig, "", "\t")
			if err == nil {
				err = ioutil.WriteFile("./config.json", bytes, 0644)
				loadConfig()
			}

			panic(re)
		}
	}()

	if r.Method == "POST" {
		registryType := r.FormValue("registry_type")
		registryURL := r.FormValue("registry_url")
		basePath := r.FormValue("base_path")

		serverConfig.RegistryType = registryType
		serverConfig.RegistryURL = registryURL
		serverConfig.ServiceBaseURL = basePath

		bytes, err := json.MarshalIndent(&serverConfig, "", "\t")
		if err == nil {
			err = ioutil.WriteFile("./config.json", bytes, 0644)
			loadConfig()
		}
	}

	renderTemplate(w, r.URL.Path[1:]+".html", serverConfig)
}

type Registry interface {
	initRegistry()
	fetchServices() []*Service
	deactivateService(name, address string) error
	activateService(name, address string) error
	updateMetadata(name, address string, metadata string) error
}

// Service is a service endpoint
type Service struct {
	ID       string
	Name     string
	Address  string
	Metadata string
	State    string
	Group    string
}
