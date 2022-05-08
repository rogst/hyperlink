package redis

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"time"

	"github.com/go-redis/redis"
	"github.com/rogst/hyperlink/pkg/message"
)

// Config hols the redis client configuration options
type Config struct {
	Addr     string
	DB       int
	Password string
}

// RegisterFlags binds cmd flags to the configuration options
func (c *Config) RegisterFlags(f *flag.FlagSet, prefix string) {
	f.StringVar(&c.Addr, prefix+"addr", "localhost:6379", "Address (host:port) of the redis server")
	f.IntVar(&c.DB, prefix+"db", 0, "Redis DB index (0-15)")
	f.StringVar(&c.Password, prefix+"password", "", "Redis server password")
}

type redisClient struct {
	db  *redis.Client
	cfg Config
	ttl time.Duration
}

// New returns a new redis client
func New(cfg Config, ttl time.Duration) message.StorageClient {
	return &redisClient{
		db: redis.NewClient(&redis.Options{
			Addr:     cfg.Addr,
			DB:       cfg.DB,
			Password: cfg.Password,
		}),
		cfg: cfg,
		ttl: ttl,
	}
}

func (c *redisClient) GetMetadata(key string) (message.Metadata, error) {
	hash, err := c.db.HGetAll(key).Result()
	if err != nil {
		return message.Metadata{}, err
	} else if len(hash) == 0 {
		return message.Metadata{}, fmt.Errorf("could not find key: %s", key)
	}

	return convertRedisHashToMessage(hash).Meta, nil
}

func (c *redisClient) GetMessage(key string) (message.Message, error) {
	hash, err := c.db.HGetAll(key).Result()
	if err != nil {
		return message.Message{}, err
	} else if len(hash) == 0 {
		return message.Message{}, fmt.Errorf("could not find key: %s", key)
	}

	// Messages are only valid for one view, so we need to remove it
	if err := c.db.Del(key).Err(); err != nil {
		return message.Message{}, err
	}

	return convertRedisHashToMessage(hash), nil
}

func (c *redisClient) SetMessage(key string, msg message.Message) error {
	if err := c.db.HMSet(key, convertMessageToRedisHash(msg)).Err(); err != nil {
		return err
	}

	c.db.Expire(key, c.ttl)
	return nil
}

func (c *redisClient) NewMessageKey() string {
	return message.NewKeyFromRandomLetters(12)
}

// Run not implemented for redis storageclient
func (c *redisClient) Run(ctx context.Context) error {
	return nil
}

func convertRedisHashToMessage(hash map[string]string) message.Message {
	ts, _ := strconv.Atoi(hash["created"])
	return message.Message{
		Data: []byte(hash["data"]),
		Meta: message.Metadata{
			Created:     time.Unix(int64(ts), 0),
			Filename:    hash["filename"],
			ContentType: hash["content-type"],
		},
	}
}

func convertMessageToRedisHash(msg message.Message) map[string]interface{} {
	return map[string]interface{}{
		"data":         string(msg.Data),
		"created":      msg.Meta.Created.Unix(),
		"filename":     msg.Meta.Filename,
		"content-type": msg.Meta.ContentType,
	}
}
