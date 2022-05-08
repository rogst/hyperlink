package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rogst/hyperlink/pkg/message"
	"github.com/rogst/hyperlink/pkg/message/storage"
	log "github.com/sirupsen/logrus"
)

// Server structure
type Server struct {
	config Config
	server http.Server
	store  message.StorageClient
}

// New returns a new server
func New(config Config, storageConfig storage.Config) *Server {
	store, err := storage.New(storageConfig)
	if err != nil {
		log.Fatalln(err)
	}

	handler := NewHandler(config, store)
	handler.RegisterRoutes()

	return &Server{
		config: config,
		server: http.Server{
			Handler:      handler.Router,
			WriteTimeout: config.WriteTimeout,
			ReadTimeout:  config.ReadTimeout,
			IdleTimeout:  config.IdleTimeout,
		},
		store: store,
	}
}

// Run starts the web server
func (s *Server) Run(ctx context.Context) error {
	// Start the backend storage
	go func() {
		if err := s.store.Run(ctx); err != nil {
			log.Errorln(err)
		}
	}()

	go func() {
		<-ctx.Done()
		s.server.Shutdown(ctx)
	}()

	s.server.Addr = s.GetBindAddress()
	if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// GetBindAddress returns the host:port that the server is bound to
func (s *Server) GetBindAddress() string {
	return fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
}
