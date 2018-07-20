package mem

import (
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"gitlab.com/ZamzamTech/wallet-api/services/sessions"
	"sync"
	"time"
)

type valWithExpire struct {
	val       interface{}
	expireAt  time.Time
	createdAt time.Time
}

// memStorage implements simple in-memory thread-safe storage
type memStorage struct {
	guard  sync.RWMutex
	values map[string]valWithExpire
}

// New returns new in-memory storage
func New() nosql.IStorage {
	return &memStorage{
		values: make(map[string]valWithExpire, 10),
	}
}

func (s *memStorage) Get(key string) (data interface{}, err error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	val, ok := s.values[key]
	if !ok {
		err = nosql.ErrNoSuchKeyFound
		return
	}
	if !val.expireAt.IsZero() && !val.expireAt.After(time.Now()) {
		err = nosql.ErrNoSuchKeyFound
		return
	}
	data = val.val
	return
}

func (s *memStorage) Set(key string, data interface{}) error {
	s.guard.Lock()
	defer s.guard.Unlock()

	s.values[key] = valWithExpire{
		val:       data,
		createdAt: time.Now(),
	}

	return nil
}

func (s *memStorage) SetWithExpire(key string, data interface{}, ttl time.Duration) error {
	s.guard.Lock()
	defer s.guard.Unlock()

	s.values[key] = valWithExpire{
		val:       data,
		expireAt:  time.Now().Add(ttl),
		createdAt: time.Now(),
	}

	return nil
}

func (s *memStorage) Delete(key string) (err error) {
	s.guard.RLock()
	defer s.guard.RUnlock()

	val, ok := s.values[key]
	if !ok {
		err = sessions.ErrNotFound
		return
	}
	if val.expireAt.Before(time.Now()) {
		err = sessions.ErrExpired
		return
	}
	s.values[key] = valWithExpire{
		val:       val.val,
		expireAt:  time.Now().Add(val.expireAt.Sub(val.createdAt)),
		createdAt: time.Now(),
	}
	return
}
