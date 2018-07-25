package redis

import (
	"fmt"
	"git.zam.io/wallet-backend/web-api/services/nosql"
	"github.com/go-redis/redis"
	"github.com/segmentio/objconv/json"
	"io"
	"math"
	"strings"
	"time"
)

// New creates nosql.IStorage wrapper
func New(options *redis.UniversalOptions) (nosql.IStorage, io.Closer) {
	c := clientWrapper{
		client: redis.NewUniversalClient(options),
	}
	return c, c
}

// clientWrapper wraps redis universal client (e.g. for both single and cluster modes)
type clientWrapper struct {
	client redis.UniversalClient
}

// clientSetWrapper implements ISetStr interface
type clientSetWrapper struct {
	clientWrapper
	setKey string
}

// Get gets redis key using GET cmd, trying to unmarshal json into interface{}
func (c clientWrapper) Get(key string) (data interface{}, err error) {
	cmd := c.client.Get(key)
	if cmd.Err() != nil {
		if isNilErr(cmd.Err()) {
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
func (c clientWrapper) Set(key string, data interface{}) error {
	return c.SetWithExpire(key, data, 0)
}

// SetWithExpire same as Set but with expiration
func (c clientWrapper) SetWithExpire(key string, data interface{}, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	res := c.client.Set(key, bytes, ttl)
	return res.Err()
}

// Delete deletes key from redis
func (c clientWrapper) Delete(key string) error {
	res := c.client.Del(key)
	if res, err := res.Result(); err != nil || res == 0 {
		if err != nil {
			return err
		}
		return nosql.ErrNoSuchKeyFound
	}
	return nil
}

// SrtSet
func (c clientWrapper) StrSet(key string) nosql.IStrSet {
	return clientSetWrapper{clientWrapper: c, setKey: key}
}

// Close implements io.Closer interface
func (c clientWrapper) Close() error {
	return c.client.Close()
}

func (c clientSetWrapper) Add(val string) error {
	return c.AddExpire(val, -1)
}

func (c clientSetWrapper) AddExpire(val string, ttl time.Duration) error {
	// cleanup expired tokens

	cmd := c.client.ZRemRangeByScore(c.setKey, "-inf", fmt.Sprintf("%d", time.Now().UTC().Unix()-1))
	if cmd.Err() != nil {
		return coerceRedisErr(cmd.Err())
	}

	var score float64
	if ttl == -1 {
		score = math.Inf(1)
	} else {
		score = float64(time.Now().Add(ttl).UTC().Unix())
	}

	return coerceRedisErr(c.client.ZAdd(c.setKey, redis.Z{
		Score:  score,
		Member: val,
	}).Err())
}

func (c clientSetWrapper) Remove(val string) error {
	cmd := c.client.ZRem(c.setKey, val)
	return coerceRedisErr(cmd.Err())
}

func (c clientSetWrapper) Check(val string) (bool, error) {
	cmd := c.client.ZScore(c.setKey, val)
	if cmd.Err() != nil {
		return false, coerceRedisErr(cmd.Err())
	}

	return cmd.Val() >= float64(time.Now().UTC().Unix()), nil
}

func (c clientSetWrapper) List() ([]string, error) {
	cmd := c.client.ZRangeByScore(c.setKey, redis.ZRangeBy{
		Min:   fmt.Sprintf("%d", time.Now().UTC().Unix()),
		Max:   "+inf",
		Count: 100,
	})
	if cmd.Err() != nil {
		return nil, coerceRedisErr(cmd.Err())
	}

	return cmd.Val(), nil
}

// utils
func coerceRedisErr(err error) error {
	switch {
	case err == nil:
		return nil
	case isNilErr(err):
		return nosql.ErrNoSuchKeyFound
	case isWrongOpErr(err):
		return nosql.ErrNotStrSet
	default:
		return err
	}
}

func isNilErr(err error) bool {
	if err == nil {
		return false
	}
	return err.Error() == "redis: nil"
}

func isWrongOpErr(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "WRONGTYPE")
}
