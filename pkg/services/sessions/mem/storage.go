package mem

import (
	"git.zam.io/wallet-backend/web-api/pkg/services/sessions"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"sync"
	"time"
)

type valWithExpire struct {
	val       map[string]interface{}
	expireAt  time.Time
	createdAt time.Time
}

// memStorage implements simple in-memory thread-safe storage
type memStorage struct {
	guard  sync.RWMutex
	values map[string]valWithExpire
}

// New returns new in-memory storage
func New() sessions.IStorage {
	return &memStorage{
		values: make(map[string]valWithExpire, 10),
	}
}

// New trivial IStorage implementation
func (s *memStorage) New(data map[string]interface{}, expireAfter time.Duration) (sessions.Token, error) {
	token := sessions.Token(uuid.New().String())

	s.guard.Lock()
	defer s.guard.Unlock()

	s.values[string(token)] = valWithExpire{
		val:       data,
		expireAt:  time.Now().Add(expireAfter),
		createdAt: time.Now(),
	}

	return token, nil
}

// RefreshToken
func (s *memStorage) RefreshToken(oldToken sessions.Token, expireAfter time.Duration) (newToken sessions.Token, err error) {
	if err = validateToken(oldToken); err != nil {
		return
	}

	s.guard.RLock()
	defer s.guard.RUnlock()

	val, ok := s.values[string(oldToken)]
	if !ok {
		err = sessions.ErrNotFound
		return
	}
	if val.expireAt.Before(time.Now()) {
		err = sessions.ErrExpired
		return
	}
	s.values[string(oldToken)] = valWithExpire{
		val:       val.val,
		expireAt:  time.Now().Add(expireAfter),
		createdAt: time.Now(),
	}
	return
}

// Get way more simpler then New
func (s *memStorage) Get(token sessions.Token) (data map[string]interface{}, err error) {
	if err = validateToken(token); err != nil {
		return
	}

	s.guard.RLock()
	defer s.guard.RUnlock()

	val, ok := s.values[string(token)]
	if !ok {
		err = sessions.ErrNotFound
		return
	}
	if !val.expireAt.After(time.Now()) {
		err = sessions.ErrExpired
		return
	}
	data = val.val
	return
}

// Delete token from storage
func (s *memStorage) Delete(token sessions.Token) (err error) {
	if err = validateToken(token); err != nil {
		return
	}

	s.guard.Lock()
	defer s.guard.Unlock()

	_, ok := s.values[string(token)]
	if !ok {
		err = sessions.ErrNotFound
	} else {
		delete(s.values, string(token))
	}
	return
}

// validateToken validates token
func validateToken(token sessions.Token) (err error) {
	_, err = uuid.ParseBytes(token)
	if err != nil {
		err = errors.Wrap(sessions.ErrUnexpectedToken, err.Error())
	}
	return
}
