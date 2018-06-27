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

type HyperlinkMetadata struct {
	MaxViews    int           `json:"maxViews,omitempty"`
	Views       int           `json:"views,omitempty"`
	ExpireIn    time.Duration `json:"expireIn,omitempty"`
	Created     time.Time     `json:"created,omitempty"`
	Type        string        `json:"type,omitempty"`
	ContentType string        `json:"contenttype,omitempty"`
	Filename    string        `json:"filename,omitempty"`
}

type Hyperlink struct {
	Data []byte            `json:"data,omitempty"`
	Meta HyperlinkMetadata `json:"meta,omitempty"`
}

type HyperLog struct {
	CreatedBy string
	Created   time.Time
	Entries   []HyperLogEntry `json:"entries"`
}

type HyperLogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	IP        string    `json:"ip"`
	UserAgent string    `json:"useragent"`
}

func (h *HyperLog) Add(c ClientInfo) {
	h.Entries = append(h.Entries, HyperLogEntry{
		Timestamp: time.Now().UTC(),
		IP:        c.IP,
		UserAgent: c.UserAgent,
	})
}

type Datastore struct {
	data map[string]*Hyperlink
	logs map[string]*HyperLog
	mtx  sync.RWMutex
}

func NewDatastore(cfg Config) *Datastore {
	return &Datastore{
		data: map[string]*Hyperlink{},
		logs: map[string]*HyperLog{},
	}
}

func (d *Datastore) Add(hyperlink *Hyperlink, creator string) string {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	var key string
	for {
		key = NewDatastoreKey()
		if _, exists := d.data[key]; !exists {
			break
		}
	}

	hyperlink.Meta.Created = time.Now().UTC()
	hyperLog := &HyperLog{
		CreatedBy: creator,
		Created:   hyperlink.Meta.Created,
	}

	d.data[key] = hyperlink
	d.logs[key] = hyperLog

	return key
}

func (d *Datastore) Get(key string, c ClientInfo) (*Hyperlink, error) {
	d.mtx.Lock()
	defer d.mtx.Unlock()

	if hyperlink, ok := d.data[key]; ok {
		d.logs[key].Add(c)
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

func (d *Datastore) Info(key string) (HyperlinkMetadata, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if hyperlink, ok := d.data[key]; ok {
		return hyperlink.Meta, nil
	}

	return HyperlinkMetadata{}, fmt.Errorf("%s was not found", key)
}

func (d *Datastore) Logs(key string) ([]HyperLogEntry, error) {
	d.mtx.RLock()
	defer d.mtx.RUnlock()

	if hyperlog, ok := d.logs[key]; ok {
		return hyperlog.Entries, nil
	}

	return []HyperLogEntry(nil), fmt.Errorf("%s was not found", key)
}
