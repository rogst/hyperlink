package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/rogst/hyperlink/pkg/api"
	"github.com/rogst/hyperlink/pkg/build"
	"github.com/rogst/hyperlink/pkg/message/storage"
	log "github.com/sirupsen/logrus"
)

var version string = "v0.0.0"

func init() {
	build.Version = version
}

// Config holds configuration variables
type Config struct {
	LogLevel string
	Storage  storage.Config
	API      api.Config
}

// RegisterFlags loads cmdline params into config
func (c *Config) RegisterFlags(f *flag.FlagSet) {
	f.StringVar(&c.LogLevel, "loglevel", "info", fmt.Sprintf("Set log level %v", log.AllLevels))
	c.Storage.RegisterFlags(f, "store.")
	c.API.RegisterFlags(f)
}

func main() {
	var config Config
	config.RegisterFlags(flag.CommandLine)
	flag.Parse()

	setLogLevel(config.LogLevel)
	log.Infoln("Hyperlink", build.Version)

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	webServer := api.New(config.API, config.Storage)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-signals
		cancel()
	}()

	log.Println("Starting on", webServer.GetBindAddress())
	if err := webServer.Run(ctx); err != nil {
		log.Fatalln("failed to start server:", err)
	}
}

func setLogLevel(level string) {
	if l, err := log.ParseLevel(level); err == nil {
		log.SetLevel(l)
		return
	}

	log.Fatalf("Unsupported loglevel \"%s\", valid levels are %v\n", level, log.AllLevels)
}
