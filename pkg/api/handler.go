package api

import (
	"bytes"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"

	"github.com/rogst/hyperlink/pkg/build"
	"github.com/rogst/hyperlink/pkg/message"
)

const dateFormat = "Jan 2 15:04:05 2006 MST"

// Handler hols the http handler, and the routes etc
type Handler struct {
	cfg          Config
	store        message.StorageClient
	Router       *mux.Router
	staticPath   string
	templatePath string
}

// NewHandler returns a new initiated Handler
func NewHandler(cfg Config, store message.StorageClient) *Handler {
	router := mux.NewRouter()
	return &Handler{
		cfg:          cfg,
		store:        store,
		Router:       router,
		staticPath:   cfg.StaticPath,
		templatePath: cfg.TemplatePath,
	}
}

// RegisterRoutes adds routes to the HTTP router
func (h *Handler) RegisterRoutes() {
	h.Router.PathPrefix("/static/").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir(h.staticPath))))

	h.Router.HandleFunc("/api/", h.addMessageHandler()).Methods("POST")
	h.Router.HandleFunc("/api/{key}", h.getMessageHandler()).Methods("GET")

	h.Router.HandleFunc("/{key}/{filename}", h.getMessageHandler()).Methods("GET")
	h.Router.HandleFunc("/{key}", h.messageHandler()).Methods("GET")

	h.Router.HandleFunc("/", h.appHandler())
}

func templateVars() map[string]interface{} {
	return map[string]interface{}{
		"title":   "Hyperlink",
		"today":   time.Now().Format("2006-01-02"),
		"version": build.Version,
	}
}

func (h *Handler) appHandler() http.HandlerFunc {
	masterFile := filepath.Join(h.templatePath, "master.html")
	pageFile := filepath.Join(h.templatePath, "app.html")
	tmpl, err := template.New("master").ParseFiles(masterFile, pageFile)
	if err != nil {
		log.Fatalln("failed to create page template:", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, templateVars())
	}
}

func (h *Handler) messageHandler() http.HandlerFunc {
	masterFile := filepath.Join(h.templatePath, "master.html")
	pageFile := filepath.Join(h.templatePath, "message.html")
	tmpl, err := template.New("master").ParseFiles(masterFile, pageFile)
	if err != nil {
		log.Fatalln("failed to create page template:", err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		vars := templateVars()
		if key, ok := v["key"]; ok {
			if meta, err := h.store.GetMetadata(key); err == nil {
				vars["key"] = key
				vars["type"] = "message"
				vars["link"] = fmt.Sprintf("/api/%s", key)
				if meta.Filename != "" {
					vars["type"] = "file"
					vars["link"] = fmt.Sprintf("/%s/%s", key, meta.Filename)
				}
			} else {
				w.WriteHeader(http.StatusNotFound)
				log.Debugln("messageHandler:", err)
				vars["err"] = err.Error()
			}
		}

		tmpl.Execute(w, vars)
	}
}

func (h *Handler) addMessageHandler() http.HandlerFunc {
	maxSize := parseSizeInBytes(h.cfg.MaxUploadSize)
	return func(w http.ResponseWriter, r *http.Request) {
		if maxSize > 0 && r.ContentLength > maxSize {
			http.Error(w, "Request too large", http.StatusExpectationFailed)
			return
		}

		if maxSize > 0 {
			r.Body = http.MaxBytesReader(w, r.Body, maxSize)
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
		}

		msg := message.New()
		if r.Header.Get("Content-Type") == "application/x-www-form-urlencoded" {
			msg.Data = []byte(r.FormValue("data"))
		} else if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
			file, hdr, err := r.FormFile("data")
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			defer file.Close()

			var buf bytes.Buffer
			_, err = io.Copy(&buf, file)
			if err != nil {
				http.Error(w, "Bad request", http.StatusBadRequest)
				return
			}
			msg.Data = buf.Bytes()
			msg.Meta.ContentType = hdr.Header.Get("Content-Type")
			msg.Meta.Filename = hdr.Filename
		} else {
			http.Error(w, "Unsupported Content-Type", http.StatusBadRequest)
			return
		}

		key := h.store.NewMessageKey()
		err := h.store.SetMessage(key, msg)
		if err != nil {
			log.Errorln("store message failed:", err)
			http.Error(w, "A server-side error occurred", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, key)
	}
}

func (h *Handler) getMessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		var msg message.Message
		var err error
		if key, ok := vars["key"]; ok {
			log.Infoln("Get key:", key, "from:", getClientIP(r), "agent:", r.UserAgent())
			msg, err = h.store.GetMessage(key)
			if err != nil {
				http.Error(w, "Provided key was not found", http.StatusNotFound)
				return
			}
		} else {
			http.Error(w, "No key provided", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", msg.Meta.ContentType)
		_, err = w.Write(msg.Data)
		if err != nil {
			http.Error(w, "IO write error", http.StatusInternalServerError)
			log.Errorln("write message failed:", err)
		}
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

func parseSizeInBytes(size string) int64 {
	s := strings.TrimSpace(size)
	split := make([]string, 2)
	split[0] = s
	for i, r := range s {
		if !unicode.IsDigit(r) {
			split[0], split[1] = s[:i], s[i:]
			break
		}
	}

	v, err := strconv.ParseFloat(split[0], 64)
	if err != nil {
		log.Errorln(err)
		return 0
	}

	switch strings.ToLower(split[1]) {
	case "b", "byte", "bytes":
		return int64(v)
	case "k", "kb", "kilobyte", "kilobytes":
		return int64(v) << 10
	case "m", "mb", "megabyte", "megabytes":
		return int64(v) << 20
	case "g", "gb", "gigabyte", "gigabytes":
		return int64(v) << 30
	case "t", "tb", "terabyte", "terabytes":
		return int64(v) << 40
	default:
		return int64(v)
	}
}
