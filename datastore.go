package main

import (
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

type HyperLink struct {
	Message  string        `json:"secretMessage,omitempty"`
	File     []byte        `json:"secretFile,omitempty"`
	MaxViews int           `json:"maxViews,omitempty"`
	Views    int           `json:"views,omitempty"`
	ExpireIn time.Duration `json:"expireIn,omitempty"`
	Created  time.Time
}

// Clone creates a new instance of HyperLink with the same values
func (h *HyperLink) Clone() HyperLink {
	return HyperLink{
		Message:  h.Message,
		File:     h.File,
		MaxViews: h.MaxViews,
		ExpireIn: h.ExpireIn,
		Created:  h.Created,
	}
}

type Datastore struct {
	data map[string]*HyperLink
	mtx  sync.RWMutex
}

func NewDatastore(cfg Config) *Datastore {
	return &Datastore{
		data: map[string]*HyperLink{},
	}
}

func (d *Datastore) Add(hyperlink *HyperLink) string {
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

func (d *Datastore) Get(key string) (*HyperLink, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if hyperlink, ok := d.data[key]; ok {
		hyperlink.Views++
		if hyperlink.Views >= hyperlink.MaxViews {
			delete(d.data, key)
		}
		return hyperlink, nil
	}

	return &HyperLink{}, nil
}
