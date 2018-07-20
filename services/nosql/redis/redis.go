package redis

import (
	"github.com/go-redis/redis"
	"github.com/segmentio/objconv/json"
	"gitlab.com/ZamzamTech/wallet-api/services/nosql"
	"io"
	"time"
)

// New creates nosql.IStorage wrapper
func New(options *redis.UniversalOptions) (nosql.IStorage, io.Closer) {
	c := universalRedisClientWrapper{
		client: redis.NewUniversalClient(options),
	}
	return c, c
}

// universalRedisClientWrapper wraps redis universal client (e.g. for both single and cluster modes)
type universalRedisClientWrapper struct {
	client redis.UniversalClient
}

// Get gets redis key using GET cmd, trying to unmarshal json into interface{}
func (c universalRedisClientWrapper) Get(key string) (data interface{}, err error) {
	cmd := c.client.Get(key)
	if cmd.Err() != nil {
		if cmd.Err().Error() == "redis: nil" {
			return nil, nosql.ErrNoSuchKeyFound
		}
		err = cmd.Err()
		return
	}

	bytes, _ := cmd.Bytes()
	if bytes == nil {
		return nil, nosql.ErrNoSuchKeyFound
	}
	if len(bytes) == 0 {
		return
	}

	err = json.Unmarshal(bytes, &data)
	return
}

// Set sets redis key value marshaling it's value using json
func (c universalRedisClientWrapper) Set(key string, data interface{}) error {
	return c.SetWithExpire(key, data, 0)
}

// SetWithExpire same as Set but with expiration
func (c universalRedisClientWrapper) SetWithExpire(key string, data interface{}, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	res := c.client.Set(key, bytes, 0)
	return res.Err()
}

// Delete deletes key from redis
func (c universalRedisClientWrapper) Delete(key string) error {
	res := c.client.Del(key)
	if res, err := res.Result(); err != nil || res == 0 {
		if err != nil {
			return err
		}
		return nosql.ErrNoSuchKeyFound
	}
	return nil
}

// Close implements io.Closer interface
func (c universalRedisClientWrapper) Close() error {
	return c.client.Close()
}
