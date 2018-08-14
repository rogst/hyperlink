package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const (
	// KeyLength specifies the length of the random string for the hyperlink key
	KeyLength = 10
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
	Created     time.Time     `json:"created,omitempty"`
	Type        string        `json:"type,omitempty"`
	ContentType string        `json:"contenttype,omitempty"`
	Filename    string        `json:"filename,omitempty"`
}

// Hyperlink is contains the key data and metadata
type Hyperlink struct {
	Data []byte            `json:"data,omitempty"`
	Meta HyperlinkMetadata `json:"meta,omitempty"`
}

// Hyperlog contains log entires and metadata about the a hyperlink
type Hyperlog struct {
	CreatedBy string
	Created   time.Time
	Entries   []HyperlogEntry `json:"entries"`
}

// HyperlogEntry contains a timestamp and fields used to identify Hyperlink visitors
type HyperlogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"useragent"`
}

// Add adds a log entry to the hyperlink log
func (h *Hyperlog) Add(entry HyperlogEntry) {
	h.Entries = append(h.Entries, entry)
}

// Datastore is the object that stores the data and logs for all hyperlinks
type Datastore struct {
	data map[string]*Hyperlink
	logs map[string]*Hyperlog
	mtx  sync.RWMutex
}

// NewDatastore returns a new initialized Datastore
func NewDatastore(cfg Config) *Datastore {
	return &Datastore{
		data: map[string]*Hyperlink{},
		logs: map[string]*Hyperlog{},
	}
}

// Add adds a new hyperlink to the datastore
func (d *Datastore) Add(hyperlink *Hyperlink, creator string) string {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	var key string
	for {
		key = NewDatastoreKey(KeyLength)
		if _, exists := d.data[key]; !exists {
			break
		}
	}

	hyperlink.Meta.Created = time.Now().UTC()
	Hyperlog := &Hyperlog{
		CreatedBy: creator,
		Created:   hyperlink.Meta.Created,
	}

	d.data[key] = hyperlink
	d.logs[key] = Hyperlog

	return key
}

// Get fetches a hyperlink from the Datastore, and adds a log entry to the log
func (d *Datastore) Get(key string, entry HyperlogEntry) (*Hyperlink, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if hyperlink, ok := d.data[key]; ok {
		d.logs[key].Add(entry)
		hyperlink.Meta.Views++
		if hyperlink.Meta.MaxViews > 0 && hyperlink.Meta.Views >= hyperlink.Meta.MaxViews {
			delete(d.data, key)
		}
		if hyperlink.Meta.Created.Add(hyperlink.Meta.ExpireIn).Sub(time.Now().UTC()) < 0 {
			delete(d.data, key)
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
		return hyperlink.Meta, nil
	}

	return HyperlinkMetadata{}, fmt.Errorf("%s was not found", key)
}

// Logs returns the log entries to a hyperlink
func (d *Datastore) Logs(key string) ([]HyperlogEntry, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if hyperlog, ok := d.logs[key]; ok {
		return hyperlog.Entries, nil
	}

	return []HyperlogEntry(nil), fmt.Errorf("%s was not found", key)
}
