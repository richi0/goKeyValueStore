// Package: keyvaluestore provides a simple in memory key-value store that supports setting a key-value pair with
// a time-to-live (TTL) in milliseconds, getting a value by key, deleting a key, and getting the length of the store.
package goKeyValueStore

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// A KeyValueStore is a simple key-value store that supports setting a key-value pair with
// a time-to-live (TTL) in milliseconds, getting a value by key,
// deleting a key, and getting the length of the store.
type KeyValueStore struct {
	data         map[string]node
	mu           *sync.RWMutex
	cleanTimeout float32
	cacheFolder  string
}

// NewKeyValueStore creates a new KeyValueStore with a cleanTimeout in seconds.
func NewKeyValueStore(cleanTimeout float32, cacheFolder string) (*KeyValueStore, error) {
	store := &KeyValueStore{
		data:         make(map[string]node),
		mu:           &sync.RWMutex{},
		cleanTimeout: cleanTimeout,
		cacheFolder:  cacheFolder,
	}
	err := store.init()
	if err != nil {
		panic(err)
	}
	go store.clean()
	return store, nil
}

// A node is a key-value pair with a deleteTimestamp.
type node struct {
	Key             string `json:"key"`
	Value           any    `json:"value"`
	DeleteTimestamp int64  `json:"deleteTimestamp"`
}

// newNode creates a new node with a key, value, and TTL.
func newNode(key string, value any, ttl int) node {
	if ttl == 0 {
		ttl = math.MaxInt
	}
	timestamp := time.Now().Add(time.Duration(ttl) * time.Millisecond).UnixMilli()
	return node{Key: key, Value: value, DeleteTimestamp: timestamp}
}

// Set sets a key-value pair with a TTL in milliseconds.
func (d *KeyValueStore) Set(key string, value any, ttl int) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	node := newNode(key, value, ttl)
	d.data[key] = node
	err := d.saveInCache(node)
	if err != nil {
		return err
	}
	return nil
}

// saveInCache saves a node in the cache folder.
func (d *KeyValueStore) saveInCache(node node) error {
	if d.cacheFolder == "" {
		return nil
	}
	data, err := json.Marshal(node)
	if err != nil {
		return err
	}
	fileName, err := d.getFileName(node.Key)
	if err != nil {
		return err
	}
	return os.WriteFile(fileName, data, 0600)
}

// Get gets a value by key. If the key does not exist, the second return value is false.
func (d *KeyValueStore) Get(key string) (any, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	val, ok := d.data[key]
	if !ok || nodeIsExpired(val) {
		return nil, false
	}
	return val.Value, ok
}

// Delete deletes a key. If the key does not exist, this function does nothing.
func (d *KeyValueStore) Delete(key string) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.data, key)
	return d.deleteInCache(key)
}

// deleteInCache deletes a key from the cache folder.
func (d *KeyValueStore) deleteInCache(key string) error {
	if d.cacheFolder == "" {
		return nil
	}
	fileName, err := d.getFileName(key)
	if err != nil {
		return err
	}
	err = os.Remove(fileName)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return nil
}

// getFileName returns the file name for a key in the cache folder.
func (d *KeyValueStore) getFileName(key string) (string, error) {
	sum := sha256.Sum256([]byte(key))
	return filepath.Join(d.cacheFolder, fmt.Sprintf("%s.store.json", hex.EncodeToString(sum[:]))), nil
}

// Length returns the number of key-value pairs in the store.
func (d *KeyValueStore) Length() int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	counter := 0
	for _, node := range d.data {
		if !nodeIsExpired(node) {
			counter++
		}
	}
	return counter
}

// init initializes the KeyValueStore by loading existing key-value pairs from the cache folder.
func (d *KeyValueStore) init() error {
	if d.cacheFolder == "" {
		return nil
	}
	err := os.MkdirAll(d.cacheFolder, 0700)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(d.cacheFolder)
	if err != nil {
		return err
	}
	for _, file := range entries {
		if !strings.HasSuffix(file.Name(), ".store.json") {
			continue
		}
		fileData, err := os.ReadFile(filepath.Join(d.cacheFolder, file.Name()))
		if err != nil {
			return err
		}
		var node node
		err = json.Unmarshal(fileData, &node)
		if err != nil {
			return err
		}
		now := time.Now().UnixMilli()
		timeLeft := node.DeleteTimestamp - now
		if timeLeft < 0 {
			timeLeft = 1
		}
		d.Set(node.Key, node.Value, int(timeLeft))
	}
	return nil
}

// clean deletes expired key-value pairs. The interval of cleaning is determined by cleanTimeout.
func (d *KeyValueStore) clean() error {
	for {
		time.Sleep(time.Duration(d.cleanTimeout) * time.Second)
		d.mu.Lock()
		for key, node := range d.data {
			if nodeIsExpired(node) {
				delete(d.data, key)
				err := d.deleteInCache(key)
				if err != nil {
					panic(err) // this should never happen
				}
			}
		}
		d.mu.Unlock()
	}
}

// nodeIsExpired returns true if a node is expired.
func nodeIsExpired(node node) bool {
	return time.Now().UnixMilli() > node.DeleteTimestamp
}
