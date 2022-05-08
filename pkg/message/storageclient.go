package message

import "context"

// StorageClient provides and interface for reading and writing Messages from different storage types
type StorageClient interface {
	GetMetadata(key string) (Metadata, error)

	GetMessage(key string) (Message, error)
	SetMessage(key string, msg Message) error

	NewMessageKey() string
	Run(ctx context.Context) error
}
