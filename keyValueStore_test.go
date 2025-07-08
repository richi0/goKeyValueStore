package goKeyValueStore_test

import (
	"os"
	"testing"
	"time"

	"github.com/richi0/goKeyValueStore"
)

const CACHE_DIR = "test_dir_cache"

func getTestStore() *goKeyValueStore.KeyValueStore {
	os.RemoveAll(CACHE_DIR)
	store, err := goKeyValueStore.NewKeyValueStore(0.5, CACHE_DIR)
	if err != nil {
		panic(err)
	}
	store.Set("key1", "value1", 100)
	store.Set("key2", "value2", 100)
	store.Set("key3", "value3", 100)
	return store
}

type CacheData struct {
	ID   int
	Name string
	List []int
}

func getTestStoreObject() *goKeyValueStore.KeyValueStore {
	os.RemoveAll(CACHE_DIR)
	store, err := goKeyValueStore.NewKeyValueStore(0.5, CACHE_DIR)
	if err != nil {
		panic(err)
	}
	store.Set("key1", CacheData{ID: 1, Name: "value1", List: []int{1, 2, 3}}, 100)
	store.Set("key2", CacheData{ID: 2, Name: "value2", List: []int{4, 5, 6}}, 100)
	store.Set("key3", CacheData{ID: 3, Name: "value3", List: []int{7, 8, 9}}, 100)
	return store
}

func TestKeyValueStoreLength(t *testing.T) {
	store := getTestStore()
	if store.Length() != 3 {
		t.Errorf("Expected length to be 3, got %d", store.Length())
	}
}

func TestKeyValueStoreSet(t *testing.T) {
	store := getTestStore()
	store.Set("key4", "value4", 1)
	if store.Length() != 4 {
		t.Errorf("Expected length to be 4, got %d", store.Length())
	}
}

func TestKeyValueStoreGet(t *testing.T) {
	store := getTestStore()
	val, ok := store.Get("key1")
	if !ok {
		t.Errorf("Expected key1 to be present")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %s", val)
	}
}

func TestGetNonExistentKey(t *testing.T) {
	store := getTestStore()
	_, ok := store.Get("key4")
	if ok {
		t.Errorf("Expected key4 to not be present")
	}
}

func TestKeyValueStoreGetExpiredKey(t *testing.T) {
	store := getTestStore()
	time.Sleep(time.Duration(110) * time.Millisecond) // Wait for key1 to expire
	_, ok := store.Get("key1")
	if ok {
		t.Errorf("Expected key1 to be expired")
	}
}

func TestKeyValueStoreDelete(t *testing.T) {
	store := getTestStore()
	store.Delete("key1")
	if store.Length() != 2 {
		t.Errorf("Expected length to be 2, got %d", store.Length())
	}
}

func TestDeleteNonExistentKey(t *testing.T) {
	store := getTestStore()
	store.Delete("key4")
	if store.Length() != 3 {
		t.Errorf("Expected length to be 3, got %d", store.Length())
	}
	time.Sleep(time.Duration(110) * time.Millisecond) // Wait for keys to expire
	if store.Length() != 0 {
		t.Errorf("Expected length to be 0, got %d", store.Length())
	}
}

func TestKeyValueStoreClean(t *testing.T) {
	store := getTestStore()
	time.Sleep(1 * time.Second)
	if store.Length() != 0 {
		t.Errorf("Expected length to be 0, got %d", store.Length())
	}
	entries, err := os.ReadDir(CACHE_DIR)
	if err != nil {
		t.Error(err)
	}
	if len(entries) != 0 {
		t.Error("All files should be removed")
	}
}

func TestKeyValueStoreCleanDetails(t *testing.T) {
	store := getTestStore()
	store.Set("key4", "value4", 1200)
	time.Sleep(1 * time.Second)
	if store.Length() != 1 {
		t.Errorf("Expected length to be 1, got %d", store.Length())
	}
	entries, err := os.ReadDir(CACHE_DIR)
	if err != nil {
		t.Error(err)
	}
	if len(entries) != 1 {
		t.Error("All files should be removed")
	}
	time.Sleep(1 * time.Second)
}

type testStruct struct {
	Name string
	Age  int
}

func TestSetTypeAsValue(t *testing.T) {
	store := getTestStore()
	store.Set("key4", testStruct{Name: "Alice", Age: 30}, 100)
	val, ok := store.Get("key4")
	if !ok {
		t.Errorf("Expected key4 to be present")
	}
	testData := val.(testStruct)
	if testData.Name != "Alice" {
		t.Errorf("Expected Alice, got %s", testData.Name)
	}
	if testData.Age != 30 {
		t.Errorf("Expected 30, got %d", testData.Age)
	}
}

func TestKeyValueStoreLengthObject(t *testing.T) {
	store := getTestStoreObject()
	if store.Length() != 3 {
		t.Errorf("Expected length to be 3, got %d", store.Length())
	}
}

func TestKeyValueStoreGetObject(t *testing.T) {
	store := getTestStoreObject()
	val, ok := store.Get("key1")
	if !ok {
		t.Errorf("Expected key1 to be present")
	}
	value, ok := val.(CacheData)
	if !ok {
		t.Errorf("Expected type cast to be ok")
	}
	if value.Name != "value1" {
		t.Errorf("Expected value1, got %s", value.Name)
	}
	if value.ID != 1 {
		t.Errorf("Expected 1, got %d", value.ID)
	}
	if len(value.List) != 3 {
		t.Errorf("Expected 3, got %d", len(value.List))
	}
}
