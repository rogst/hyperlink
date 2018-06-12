package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

// Handler hols the http handler, and the routes etc
type Handler struct {
	db           *Datastore
	router       *mux.Router
	staticPath   string
	templatePath string
}

// NewHandler returns a new initiated Handler
func NewHandler(cfg Config) *Handler {
	router := mux.NewRouter()
	return &Handler{
		db:           NewDatastore(cfg),
		router:       router,
		staticPath:   cfg.HTTPStaticPath,
		templatePath: cfg.HTTPTemplatePath,
	}
}

// RegisterRoutes adds routes to the HTTP router
func (h *Handler) RegisterRoutes() {
	h.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(h.staticPath))))

	h.router.HandleFunc("/{key}", h.handleView())
	h.router.HandleFunc("/", h.handleIndex())

	api := h.router.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/", h.handleAddHyperlink()).Methods("POST")
	api.HandleFunc("/{key}", h.handleGetHyperlink()).Methods("GET")
}

func (h *Handler) handleIndex() http.HandlerFunc {
	var (
		init sync.Once
		tmpl *template.Template
		err  error
	)
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			masterFile := filepath.Join(h.templatePath, "master.html")
			indexFile := filepath.Join(h.templatePath, "index.html")
			tmpl, err = template.New("master").ParseFiles(masterFile, indexFile)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		vars := map[string]interface{}{
			"title": "Hyperlink",
			"today": time.Now().Format("2006-01-02"),
		}

		tmpl.Execute(w, vars)
	}
}

func (h *Handler) handleView() http.HandlerFunc {
	var (
		init sync.Once
		tmpl *template.Template
		err  error
	)
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			masterFile := filepath.Join(h.templatePath, "master.html")
			viewFile := filepath.Join(h.templatePath, "view.html")
			tmpl, err = template.New("master").ParseFiles(masterFile, viewFile)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		v := mux.Vars(r)
		vars := map[string]interface{}{}
		if key, ok := v["key"]; ok {
			vars["key"] = key
		}

		tmpl.Execute(w, vars)
	}
}

func (h *Handler) handleAddHyperlink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		maxViews, err := strconv.Atoi(r.FormValue("maxViews"))
		if err != nil {
			http.Error(w, "Failed to convert maxViews to a number", http.StatusBadRequest)
			return
		}

		expireIn, err := time.ParseDuration(r.FormValue("expireIn"))
		if err != nil {
			http.Error(w, "Failed to parse expire in time duration", http.StatusBadRequest)
			return
		}

		key := h.db.Add(&HyperLink{
			Message:  r.FormValue("secretMessage"),
			MaxViews: maxViews,
			ExpireIn: expireIn,
		})

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, key)
	}
}

func (h *Handler) handleGetHyperlink() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var hyperlink *HyperLink
		var err error
		if key, ok := vars["key"]; ok {
			hyperlink, err = h.db.Get(key)
			if err != nil {
				http.Error(w, "Provided key was not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "No key provided", http.StatusBadRequest)
			return
		}

		// Create a copy in order to change ExpireIn without changing the stored data
		clone := hyperlink.Clone()
		clone.ExpireIn = time.Now().UTC().Sub(clone.Created.Add(clone.ExpireIn)).Truncate(time.Second)

		buf := bytes.Buffer{}
		err = json.NewEncoder(&buf).Encode(clone)
		if err != nil {
			http.Error(w, "JSON encoding error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, buf.String())
	}
}
