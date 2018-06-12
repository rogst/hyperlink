package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// Config holds configuration variables
type Config struct {
	Port             int
	Host             string
	HTTPReadTimeout  time.Duration
	HTTPWriteTimeout time.Duration
	HTTPStaticPath   string
	HTTPTemplatePath string
	LogLevel         string
}

// RegisterFlags loads cmdline params into config
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.IntVar(&c.Port, "port", 8080, "Port to listen on")
	f.StringVar(&c.Host, "host", "", "Host IP to listen on")
	f.DurationVar(&c.HTTPReadTimeout, "http.read-timeout", 5*time.Second, "HTTP server read timeout")
	f.DurationVar(&c.HTTPWriteTimeout, "http.write-timeout", 10*time.Second, "HTTP server write timeout")
	f.StringVar(&c.HTTPStaticPath, "http.static-path", "./static", "Path from where to service static files (/static/*)")
	f.StringVar(&c.HTTPTemplatePath, "http.template-path", "./templates", "Path from where to service HTTP templates")
	f.StringVar(&c.LogLevel, "loglevel", "info", "Set log level (debug, info, error)")
}

func main() {
	var config Config
	config.RegisterFlags(flag.CommandLine)
	flag.Parse()

	setLogLevel(config.LogLevel)

	handler := NewHandler(config)
	handler.RegisterRoutes()

	listenAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	server := &http.Server{
		Handler:      handler.router,
		Addr:         listenAddr,
		WriteTimeout: config.HTTPWriteTimeout,
		ReadTimeout:  config.HTTPReadTimeout,
	}

	log.Println("Starting on", listenAddr)
	log.Fatal(server.ListenAndServe())
}

func setLogLevel(level string) {
	switch strings.ToLower(level) {
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	default:
		log.Errorf("Unknown log level %s, defaulting to info", level)
		log.SetLevel(log.InfoLevel)
	}
}
