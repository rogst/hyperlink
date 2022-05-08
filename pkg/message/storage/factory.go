package storage

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/rogst/hyperlink/pkg/message"
	"github.com/rogst/hyperlink/pkg/message/memory"
	"github.com/rogst/hyperlink/pkg/message/redis"
)

// Config holds the storage factory config
type Config struct {
	Client string
	TTL    time.Duration
	Memory memory.Config
	Redis  redis.Config
}

// RegisterFlags registers the config params to cmd flags
func (c *Config) RegisterFlags(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Client, prefix+"client", "memory", "Client used for storage of messages")
	f.DurationVar(&c.TTL, prefix+"ttl", 48*time.Hour, "Time To Live for messages before they expire")
	c.Memory.RegisterFlags(f, prefix+"memory.")
	c.Redis.RegisterFlags(f, prefix+"redis.")
}

// New returns a new StorageClient
func New(cfg Config) (message.StorageClient, error) {
	switch strings.ToLower(cfg.Client) {
	case "memory":
		return memory.New(cfg.Memory, cfg.TTL), nil
	case "redis":
		return redis.New(cfg.Redis, cfg.TTL), nil
	default:
		return nil, fmt.Errorf("unsupported storage client: %s", cfg.Client)
	}
}
