package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	// KeyLetters is the characters used to generate the hyperlink key
	KeyLetters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	// MetaTypeMessage is used to identify message links
	MetaTypeMessage = "message"
	// MetaTypeFile is used to identify file links
	MetaTypeFile = "file"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

// NewDatastoreKey returns a new randomly generated string of specified length
func NewDatastoreKey(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = KeyLetters[rand.Intn(len(KeyLetters))]
	}

	return string(b)
}

// HyperlinkMetadata contains information about the hyperlink key
type HyperlinkMetadata struct {
	MaxViews    int           `json:"maxViews,omitempty"`
	Views       int           `json:"views,omitempty"`
	ExpireIn    time.Duration `json:"expireIn,omitempty"`
	Expired     bool          `json:"expired,omitempty"`
	Created     time.Time     `json:"created,omitempty"`
	Type        string        `json:"type,omitempty"`
	ContentType string        `json:"contenttype,omitempty"`
	Filename    string        `json:"filename,omitempty"`
}

// Hyperlink is contains the key data and metadata
type Hyperlink struct {
	Data []byte            `json:"data,omitempty"`
	Meta HyperlinkMetadata `json:"meta,omitempty"`
	Log  []HyperlogEntry   `json:"log,omitempty"`
}

func (h *Hyperlink) hasExpired() bool {
	if h.Meta.Expired {
		return true
	}

	if h.Meta.MaxViews > 0 && h.Meta.Views > h.Meta.MaxViews {
		h.Meta.Expired = true
	}

	if h.Meta.Created.Add(h.Meta.ExpireIn).Sub(time.Now().UTC()) < 0 {
		h.Meta.Expired = true
	}

	return h.Meta.Expired
}

// HyperlogEntry contains a timestamp and fields used to identify Hyperlink visitors
type HyperlogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"useragent"`
}

// AddLog adds a log entry to the hyperlink log
func (h *Hyperlink) AddLog(entry HyperlogEntry) {
	h.Log = append(h.Log, entry)
}

// Datastore is the object that stores the data and logs for all hyperlinks
type Datastore struct {
	config Config
	data   map[string]*Hyperlink
	mtx    sync.RWMutex
}

// NewDatastore returns a new initialized Datastore
func NewDatastore(cfg Config) *Datastore {
	return &Datastore{
		data:   map[string]*Hyperlink{},
		config: cfg,
	}
}

// Add adds a new hyperlink to the datastore
func (d *Datastore) Add(hyperlink *Hyperlink, entry HyperlogEntry) string {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	var key string
	for {
		key = NewDatastoreKey(d.config.KeyLength)
		if _, exists := d.data[key]; !exists {
			break
		}
	}

	hyperlink.Meta.Created = time.Now().UTC()
	hyperlink.AddLog(entry)

	d.data[key] = hyperlink

	return key
}

// Get fetches a hyperlink from the Datastore, and adds a log entry to the log
func (d *Datastore) Get(key string, entry HyperlogEntry) (*Hyperlink, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if hyperlink, ok := d.data[key]; ok {
		hyperlink.AddLog(entry)
		hyperlink.Meta.Views++
		if hyperlink.hasExpired() {
			d.data[key].Data = nil
			return &Hyperlink{}, fmt.Errorf("%s has expired", key)
		}

		return hyperlink, nil
	}

	return &Hyperlink{}, fmt.Errorf("%s was not found", key)
}

// Info returns the metadata about the hyperlink
func (d *Datastore) Info(key string) (HyperlinkMetadata, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if hyperlink, ok := d.data[key]; ok {
		if hyperlink.hasExpired() {
			return HyperlinkMetadata{}, fmt.Errorf("%s has expired", key)
		}

		return hyperlink.Meta, nil
	}

	return HyperlinkMetadata{}, fmt.Errorf("%s was not found", key)
}

// Logs returns the log entries to a hyperlink
func (d *Datastore) Logs(key string) ([]HyperlogEntry, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if hyperlink, ok := d.data[key]; ok {
		return hyperlink.Log, nil
	}

	return []HyperlogEntry(nil), fmt.Errorf("%s was not found", key)
}

// PurgeExpiredKeys will go through all keys and expire expired keys, and remove keys older than expiredTTL
func (d *Datastore) PurgeExpiredKeys() error {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	for x := range d.data {
		expireIn := d.data[x].Meta.Created.Add(d.data[x].Meta.ExpireIn).Sub(time.Now().UTC())
		if expireIn < -d.config.ExpiredTTL && d.data[x].hasExpired() {
			delete(d.data, x)
		} else if expireIn < 0 && !d.data[x].hasExpired() {
			d.data[x].Data = nil
		}
	}

	return nil
}
