package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
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
	CleanInterval    time.Duration
	ExpiredTTL       time.Duration
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
	f.DurationVar(&c.CleanInterval, "clean.interval", 5*time.Minute, "How often to purge expired keys")
	f.DurationVar(&c.ExpiredTTL, "expired.ttl", 24*time.Hour, "How long to keep expired keys")
}

func main() {
	var config Config
	config.RegisterFlags(flag.CommandLine)
	flag.Parse()

	setLogLevel(config.LogLevel)

	handler := NewHandler(config)
	handler.RegisterRoutes()
	go handler.RunCleaner(config.CleanInterval)

	listenAddr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	server := &http.Server{
		Handler:      handler.router,
		Addr:         listenAddr,
		WriteTimeout: config.HTTPWriteTimeout,
		ReadTimeout:  config.HTTPReadTimeout,
	}

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signals
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		handler.StopCleaner()
	}()

	log.Println("Starting on", listenAddr)
	log.Fatal(server.ListenAndServe())

	log.Info("Hyperlink stopped")
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

	log.Infoln("LogLevel set to", log.GetLevel().String())
}
