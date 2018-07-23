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

// memSet implements IStrSet interface
type memSet struct {
	guard sync.RWMutex
	set   strWithExpireSet
}

type strWithExpireSet map[string]time.Time

type notASet struct {}


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

func (s *memStorage) StrSet(key string) nosql.IStrSet {
	var set *memSet
	setRaw, err := s.Get(key)
	if err == nosql.ErrNoSuchKeyFound {
		set = &memSet{set: make(map[string]time.Time)}
		s.Set(key, set)
	} else {
		var ok bool
		set, ok = setRaw.(*memSet)
		if !ok {
			return notASet{}
		}
	}
	return set
}

func (set *memSet) Add(val string) error {
	set.guard.Lock()
	defer set.guard.Unlock()

	set.set[val] = time.Time{}
	return nil
}

func (set *memSet) AddExpire(val string, ttl time.Duration) error {
	set.guard.Lock()
	defer set.guard.Unlock()

	set.set[val] = time.Now().UTC().Add(ttl)
	return nil
}

func (set *memSet) Remove(val string) error {
	set.guard.Lock()
	defer set.guard.Unlock()

	delete(set.set, val)

	return nil
}

func (set *memSet) Check(val string) (bool, error) {
	set.guard.RLock()
	defer set.guard.RUnlock()

	elem, ok := set.set[val]
	if !ok {
		return false, nil
	}

	if elem.Before(time.Now().UTC()) {
		return false, nil
	}

	return true, nil
}

func (set *memSet) List() ([]string, error) {
	set.guard.RLock()
	defer set.guard.RUnlock()

	elements := make([]string, 0, len(set.set))
	for e := range set.set {
		elements = append(elements, e)
	}
	return elements, nil
}

func (notASet) Add(val string) error {
	return nosql.ErrNotStrSet
}

func (notASet) AddExpire(val string, ttl time.Duration) error {
	return nosql.ErrNotStrSet
}

func (notASet) Remove(val string) error {
	return nosql.ErrNotStrSet
}

func (notASet) Check(val string) (bool, error) {
	return false, nosql.ErrNotStrSet
}

func (notASet) List() ([]string, error) {
	return nil, nosql.ErrNotStrSet
}
