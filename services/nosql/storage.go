package nosql

import (
	"errors"
	"time"
)

var (
	// ErrNoSuchKeyFound when no data associated with given key found
	ErrNoSuchKeyFound = errors.New("no such key found")
)

// IStorage exposes simple key-value interface which wraps some NoSql DB.
// Also this interface expects that backend DB knows how the data will be serialized.
type IStorage interface {
	// Get returns data associated with given key or return ErrNoSuchKeyFound
	Get(key string) (data interface{}, err error)

	// Set associates given key with given data
	Set(key string, data interface{}) error

	// SetWithExpire associates given key with given data for specified time
	SetWithExpire(key string, data interface{}, ttl time.Duration) error

	// Delete delete value associated with given key from storage, should return ErrNoSuchKey if nothing deleted
	Delete(key string) error
}
