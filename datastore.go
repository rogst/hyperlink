package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const KeyLength = 10
const KeyLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func init() {
	rand.Seed(time.Now().UnixNano())
}

func NewDatastoreKey() string {
	b := make([]byte, KeyLength)
	for i := range b {
		b[i] = KeyLetters[rand.Intn(len(KeyLetters))]
	}

	return string(b)
}

type HyperlinkFileDesc struct {
	Data        []byte
	ContentType string
	Filename    string
}

type Hyperlink struct {
	Message  string        `json:"message,omitempty"`
	MaxViews int           `json:"maxViews,omitempty"`
	Views    int           `json:"views,omitempty"`
	ExpireIn time.Duration `json:"expireIn,omitempty"`
	Created  time.Time
	Type     string
	File     HyperlinkFileDesc
}

// Clone creates a new instance of Hyperlink with the same values
func (h *Hyperlink) Clone() Hyperlink {
	return Hyperlink{
		Message:  h.Message,
		MaxViews: h.MaxViews,
		Views:    h.Views,
		ExpireIn: h.ExpireIn,
		Created:  h.Created,
		Type:     h.Type,
		File: HyperlinkFileDesc{
			Data:        h.File.Data,
			ContentType: h.File.ContentType,
			Filename:    h.File.Filename,
		},
	}
}

type Datastore struct {
	data map[string]*Hyperlink
	mtx  sync.RWMutex
}

func NewDatastore(cfg Config) *Datastore {
	return &Datastore{
		data: map[string]*Hyperlink{},
	}
}

func (d *Datastore) Add(hyperlink *Hyperlink) string {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	var key string
	for {
		key = NewDatastoreKey()
		if _, exists := d.data[key]; !exists {
			break
		}
	}

	hyperlink.Created = time.Now().UTC()
	d.data[key] = hyperlink
	return key
}

func (d *Datastore) Get(key string) (*Hyperlink, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if hyperlink, ok := d.data[key]; ok {
		hyperlink.Views++
		if hyperlink.MaxViews > 0 && hyperlink.Views >= hyperlink.MaxViews {
			delete(d.data, key)
		}
		if hyperlink.Created.Add(hyperlink.ExpireIn).Sub(time.Now().UTC()) < 0 {
			delete(d.data, key)
		}
		return hyperlink, nil
	}

	return &Hyperlink{}, fmt.Errorf("%s was not found", key)
}

func (d *Datastore) Info(key string) (string, string, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if hyperlink, ok := d.data[key]; ok {
		return hyperlink.Type, hyperlink.File.Filename, nil
	}

	return "", "", fmt.Errorf("%s was not found", key)
}
