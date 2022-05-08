package api

import (
	"flag"
	"time"
)

// Config holds configuration variables
type Config struct {
	Port          int
	Host          string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	IdleTimeout   time.Duration
	StaticPath    string
	TemplatePath  string
	MaxUploadSize string
}

// RegisterFlags loads cmdline params into config
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.IntVar(&c.Port, "http.port", 8080, "Port to listen on")
	f.StringVar(&c.Host, "http.host", "", "Host IP to listen on")
	f.DurationVar(&c.ReadTimeout, "http.read-timeout", 5*time.Second, "HTTP server read timeout")
	f.DurationVar(&c.WriteTimeout, "http.write-timeout", 10*time.Second, "HTTP server write timeout")
	f.DurationVar(&c.IdleTimeout, "http.idle-timeout", 30*time.Second, "HTTP server idle timeout")
	f.StringVar(&c.StaticPath, "http.static-path", "./web/static", "Path from where to service static files (/static/*)")
	f.StringVar(&c.TemplatePath, "http.template-path", "./web/templates", "Path from where to service HTTP templates")
	f.StringVar(&c.MaxUploadSize, "http.max-upload-size", "0", "Max upload size")
}
