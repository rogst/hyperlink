package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// Handler hols the http handler, and the routes etc
type Handler struct {
	db           *Datastore
	router       *mux.Router
	staticPath   string
	templatePath string
	done         chan struct{}
	stopped      chan struct{}
}

// NewHandler returns a new initiated Handler
func NewHandler(cfg Config) *Handler {
	router := mux.NewRouter()
	return &Handler{
		db:           NewDatastore(cfg),
		router:       router,
		staticPath:   cfg.HTTPStaticPath,
		templatePath: cfg.HTTPTemplatePath,
		done:         make(chan struct{}),
		stopped:      make(chan struct{}),
	}
}

// RegisterRoutes adds routes to the HTTP router
func (h *Handler) RegisterRoutes() {
	h.router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(h.staticPath))))

	api := h.router.PathPrefix("/api/").Subrouter()
	api.HandleFunc("/", h.handleAddHyperlink()).Methods("POST")
	api.HandleFunc("/{key}", h.handleGetHyperlink()).Methods("GET")

	h.router.HandleFunc("/{key}/download/{filename}", h.handleDownload())
	h.router.HandleFunc("/{key}/logs", h.handleLogs())
	h.router.HandleFunc("/{key}", h.handleView())

	h.router.HandleFunc("/", h.handleIndex())
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
			if meta, err := h.db.Info(key); err == nil {
				vars["key"] = key
				vars["type"] = meta.Type
				if meta.Type == MetaTypeFile {
					vars["link"] = fmt.Sprintf("/%s/download/%s", key, meta.Filename)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				vars["err"] = err.Error()
			}
		}

		tmpl.Execute(w, vars)
	}
}

func (h *Handler) handleLogs() http.HandlerFunc {
	var (
		init sync.Once
		tmpl *template.Template
		err  error
	)
	return func(w http.ResponseWriter, r *http.Request) {
		init.Do(func() {
			masterFile := filepath.Join(h.templatePath, "master.html")
			viewFile := filepath.Join(h.templatePath, "logs.html")
			tmpl, err = template.New("master").ParseFiles(masterFile, viewFile)
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		v := mux.Vars(r)
		vars := map[string]interface{}{}
		if key, ok := v["key"]; ok {
			if logs, err := h.db.Logs(key); err == nil {
				vars["logs"] = logs
			} else {
				w.WriteHeader(http.StatusNotFound)
				vars["err"] = err.Error()
			}
		}

		tmpl.Execute(w, vars)
	}
}

func (h *Handler) handleDownload() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		if key, ok := v["key"]; ok {
			hyperlink, err := h.db.Get(key, getClientInfo(r))
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Header().Set("Content-Type", hyperlink.Meta.ContentType)
			_, err = w.Write(hyperlink.Data)
			if err != nil {
				log.Error(err)
			}
		} else {
			http.Error(w, "No key provided", http.StatusNotFound)
			return
		}
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

		/* if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			renderError(w, "FILE_TOO_BIG", http.StatusBadRequest)
			return
		} */
		hyperlink := &Hyperlink{
			Meta: HyperlinkMetadata{
				MaxViews: maxViews,
				ExpireIn: expireIn,
			},
		}

		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			hyperlink.Data = []byte(r.FormValue("data"))
			hyperlink.Meta.Type = MetaTypeMessage
		} else if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			file, hdr, err := r.FormFile("data")
			if err != nil {
				http.Error(w, "Failed to read file from request", http.StatusBadRequest)
				return
			}
			defer file.Close()

			var buf bytes.Buffer
			io.Copy(&buf, file)
			hyperlink.Data = buf.Bytes()
			hyperlink.Meta.ContentType = hdr.Header.Get("Content-Type")
			hyperlink.Meta.Filename = hdr.Filename
			hyperlink.Meta.Type = MetaTypeFile
		} else {
			http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
			return
		}

		key := h.db.Add(hyperlink, getClientInfo(r))

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, key)
	}
}

func (h *Handler) handleGetHyperlink() http.HandlerFunc {
	type Response struct {
		Data     string        `json:"data,omitempty"`
		ExpireIn time.Duration `json:"expireIn,omitempty"`
		MaxViews int           `json:"maxViews,omitempty"`
		Views    int           `json:"views,omitempty"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var hyperlink *Hyperlink
		var err error
		if key, ok := vars["key"]; ok {
			hyperlink, err = h.db.Get(key, getClientInfo(r))
			if err != nil {
				http.Error(w, "Provided key was not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "No key provided", http.StatusBadRequest)
			return
		}

		// Create a copy in order to change ExpireIn without changing the stored data
		resp := Response{
			Data:     string(hyperlink.Data),
			ExpireIn: hyperlink.Meta.Created.Add(hyperlink.Meta.ExpireIn).Sub(time.Now().UTC()).Truncate(time.Second),
			MaxViews: hyperlink.Meta.MaxViews,
			Views:    hyperlink.Meta.Views,
		}

		buf := bytes.Buffer{}
		err = json.NewEncoder(&buf).Encode(resp)
		if err != nil {
			http.Error(w, "JSON encoding error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, buf.String())
	}
}

func getClientInfo(r *http.Request) HyperlogEntry {
	return HyperlogEntry{
		Timestamp: time.Now().UTC(),
		IP:        getClientIP(r),
		UserAgent: r.Header.Get("User-Agent"),
	}
}

func getClientIP(r *http.Request) string {
	clientIP := r.Header.Get("x-real-ip")
	if clientIP == "" {
		clientIP = r.Header.Get("x-forwarded-for")
	}
	if clientIP == "" {
		clientIP = r.RemoteAddr
	}

	return clientIP
}

// RunCleaner runs a PurgeExpiredKeys at CleanInterval
func (h *Handler) RunCleaner(interval time.Duration) {
	defer close(h.stopped)

	ticker := time.NewTicker(interval)

	for {
		select {
		case <-ticker.C:
			if err := h.db.PurgeExpiredKeys(); err != nil {
				log.Error(err)
			}
		case <-h.done:
			break
		}
	}
}

// StopCleaner will cause RunCleaner to exit
func (h *Handler) StopCleaner() {
	close(h.done)
	<-h.stopped
}
