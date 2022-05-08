package message

import "time"

// Message struct
type Message struct {
	Data []byte   `json:"data"`
	Meta Metadata `json:"metadata"`
}

// Metadata contains information about a message
type Metadata struct {
	Created     time.Time `json:"created"`
	Filename    string    `json:"filename,omitempty"`
	ContentType string    `json:"content-type,omitempty"`
}

// New returns a new Message
func New() Message {
	return Message{
		Meta: Metadata{
			Created: time.Now().UTC(),
		},
	}
}
