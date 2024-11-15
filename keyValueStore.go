// Package: keyvaluestore provides a simple in memory key-value store that supports setting a key-value pair with
// a time-to-live (TTL) in milliseconds, getting a value by key, deleting a key, and getting the length of the store.
package goKeyValueStore

import (
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
}

// NewKeyValueStore creates a new KeyValueStore with a cleanTimeout in seconds.
func NewKeyValueStore(cleanTimeout float32) *KeyValueStore {
	store := &KeyValueStore{data: make(map[string]node), mu: &sync.RWMutex{}, cleanTimeout: cleanTimeout}
	go store.clean()
	return store
}

// A node is a key-value pair with a deleteTimestamp.
type node struct {
	key             string
	value           interface{}
	deleteTimestamp int64
}

// newNode creates a new node with a key, value, and TTL.
func newNode(key string, value interface{}, ttl int) node {
	timestamp := time.Now().Add(time.Duration(ttl) * time.Millisecond).UnixMilli()
	return node{key: key, value: value, deleteTimestamp: timestamp}
}

// Set sets a key-value pair with a TTL in milliseconds.
func (d *KeyValueStore) Set(key string, value interface{}, ttl int) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.data[key] = newNode(key, value, ttl)
}

// Get gets a value by key. If the key does not exist, the second return value is false.
func (d *KeyValueStore) Get(key string) (interface{}, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	val, ok := d.data[key]
	if !ok || nodeIsExpired(val) {
		return nil, false
	}
	return val.value, ok
}

// Delete deletes a key. If the key does not exist, this function does nothing.
func (d *KeyValueStore) Delete(key string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	delete(d.data, key)
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

// clean deletes expired key-value pairs. The interval of cleaning is determined by cleanTimeout.
func (d *KeyValueStore) clean() {
	for {
		time.Sleep(time.Duration(d.cleanTimeout) * time.Second)
		d.mu.Lock()
		for key, node := range d.data {
			if nodeIsExpired(node) {
				delete(d.data, key)
			}
		}
		d.mu.Unlock()
	}
}

// nodeIsExpired returns true if a node is expired.
func nodeIsExpired(node node) bool {
	return time.Now().UnixMilli() > node.deleteTimestamp
}
