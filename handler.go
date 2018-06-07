package main

import (
	"fmt"
	"html/template"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
)

// Handler hols the http handler, and the routes etc
type Handler struct {
	router       *mux.Router
	staticPath   string
	templatePath string
}

// NewHandler returns a new initiated Handler
func NewHandler(cfg Config) *Handler {
	router := mux.NewRouter()
	return &Handler{
		router:       router,
		staticPath:   cfg.HTTPStaticPath,
		templatePath: cfg.HTTPTemplatePath,
	}
}

// RegisterRoutes adds routes to the HTTP router
func (h *Handler) RegisterRoutes() {
	if h.staticPath != "" {
		h.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(h.staticPath))))
	} else {
		h.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", h.handleStatic()))
	}
	h.router.HandleFunc("/", h.handleIndex())
}

func (h *Handler) handleStatic() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if staticDesc, ok := staticFiles[r.URL.Path]; ok {
			w.Header().Set("Content-Type", staticDesc.ContentType)
			w.WriteHeader(http.StatusOK)
			fmt.Fprintln(w, staticDesc.Data)
		} else {
			w.WriteHeader(http.StatusNotFound)
			fmt.Fprintln(w, "These are not the files your looking for")
		}
	}
}

func (h *Handler) handleIndex() http.HandlerFunc {
	var (
		init sync.Once
		tpl  *template.Template
		err  error
	)
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			tpl, err = template.New("html").Parse(indexHTMLTemplate)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		vars := map[string]interface{}{
			"title":    "Hyperlink",
			"headline": "Secure Information Exchange",
		}

		tpl.ExecuteTemplate(w, "html", vars)
	}
}
