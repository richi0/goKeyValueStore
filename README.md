# goKeyValueStore [![GoDoc](https://pkg.go.dev/badge/goKeyValueStore.svg)](https://pkg.go.dev/github.com/richi0/goKeyValueStore)

goKeyValueStore provides a simple in memory key-value store with support for time-to-live (TTL) for each key.

## What does goKeyValueStore do?

goKeyValueStore allows you to set key-value pairs with an optional TTL, get values by key, delete keys, and automatically clean up expired keys.

## How do I use goKeyValueStore?

### Install

```bash
go get -u github.com/richi0/goKeyValueStore
```

### Example

```go
import (
    "fmt"
    "time"
    "github.com/richi0/goKeyValueStore"
)

func main() {
    store := goKeyValueStore.NewKeyValueStore(1)
    store.Set("key1", "value1", 1000)
    val, ok := store.Get("key1")
    if ok {
        fmt.Println(val) // Output: value1
    }
    time.Sleep(2 * time.Second)
    _, ok = store.Get("key1")
    if !ok {
        fmt.Println("key1 has expired")
    }
}
```

## Documentation

Find the full documentation of the package here: https://pkg.go.dev/github.com/richi0/goKeyValueStore
