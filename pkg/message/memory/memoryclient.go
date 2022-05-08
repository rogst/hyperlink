package memory

import (
	"context"
	"flag"
	"fmt"
	"sync"
	"time"

	"github.com/rogst/hyperlink/pkg/message"
	log "github.com/sirupsen/logrus"
)

// Config hols the memory client configuration options
type Config struct {
	PruneInterval time.Duration
}

// RegisterFlags binds cmd flags to the configuration options
func (c *Config) RegisterFlags(f *flag.FlagSet, prefix string) {
	f.DurationVar(&c.PruneInterval, prefix+"prune-interval", 5*time.Minute, "How often to prune expired messages")
}

type memoryClient struct {
	data map[string]message.Message
	cfg  Config
	mtx  sync.RWMutex
	ttl  time.Duration
}

// New returns a new memory client
func New(cfg Config, ttl time.Duration) message.StorageClient {
	return &memoryClient{
		data: map[string]message.Message{},
		cfg:  cfg,
		ttl:  ttl,
	}
}

func (mc *memoryClient) GetMetadata(key string) (message.Metadata, error) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	if msg, ok := mc.data[key]; ok && time.Since(msg.Meta.Created) < mc.ttl {
		return msg.Meta, nil
	}

	return message.Metadata{}, fmt.Errorf("no message found for key: %s", key)
}

func (mc *memoryClient) GetMessage(key string) (message.Message, error) {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	if msg, ok := mc.data[key]; ok && time.Since(msg.Meta.Created) < mc.ttl {
		// Messages are only valid for one view, so we need to remove it
		delete(mc.data, key)
		return msg, nil
	}

	return message.Message{}, fmt.Errorf("no message found for key: %s", key)
}

func (mc *memoryClient) SetMessage(key string, msg message.Message) error {
	mc.mtx.Lock()
	defer mc.mtx.Unlock()
	mc.data[key] = msg
	return nil
}

func (mc *memoryClient) NewMessageKey() string {
	return message.NewKeyFromRandomLetters(12)
}

// Run is expected to run in the background for pruning old messages
func (mc *memoryClient) Run(ctx context.Context) error {
	cleanup := func() {
		mc.mtx.Lock()
		defer mc.mtx.Unlock()
		for key, msg := range mc.data {
			if time.Since(msg.Meta.Created) > mc.ttl {
				delete(mc.data, key)
				log.Debugln("memoryclient deleted expired message:", key)
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(mc.cfg.PruneInterval):
			log.Debugln("memoryclient running cleanup")
			cleanup()
		}
	}
}
