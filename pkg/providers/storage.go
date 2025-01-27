package providers

import (
	"fmt"
	serverconf "git.zam.io/wallet-backend/web-api/config/server"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql"
	"git.zam.io/wallet-backend/web-api/pkg/services/nosql/mem"
	nosqlredis "git.zam.io/wallet-backend/web-api/pkg/services/nosql/redis"
	"github.com/go-redis/redis"
	"io"
	"net/url"
	"strings"
)

// Storage creates nosql storage according to given scheme (only mem, redis and rediss are supported)
func Storage(conf serverconf.Scheme) (nosql.IStorage, io.Closer, error) {
	if conf.Storage.URI == "" {
		conf.Storage.URI = "mem://"
	}
	return storageFromURI(conf.Storage.URI)
}

func storageFromURI(uri string) (nosql.IStorage, io.Closer, error) {
	parsed, err := url.Parse(uri)
	if err != nil {
		return nil, nil, err
	}

	switch parsed.Scheme {
	case "mem":
		return mem.New(), noopCloser{}, nil
	case "redis", "rediss":
		hosts := strings.Split(parsed.Host, ",")
		options := &redis.UniversalOptions{}

		// for multi host uris (means redis cluster)
		if len(hosts) > 1 {
			singleHost := *parsed
			singleHost.Host = hosts[0]

			singleOptions, err := redis.ParseURL(singleHost.String())
			if err != nil {
				return nil, nil, err
			}
			options = singleOptionsToUniversal(singleOptions)
			options.Addrs = hosts
		} else {
			singleOptions, err := redis.ParseURL(uri)
			if err != nil {
				return nil, nil, err
			}
			options = singleOptionsToUniversal(singleOptions)
		}

		client, closer := nosqlredis.New(options)
		return client, closer, nil
	default:
		return nil, nil, fmt.Errorf("unsupported nosql storage scheme %s given by uri %s", parsed.Scheme, uri)
	}
}

// utils
type noopCloser struct{}

func (noopCloser) Close() error {
	return nil
}

func singleOptionsToUniversal(singleOptions *redis.Options) *redis.UniversalOptions {
	return &redis.UniversalOptions{
		Addrs:              []string{singleOptions.Addr},
		DB:                 singleOptions.DB,
		MaxRetries:         singleOptions.MaxRetries,
		OnConnect:          singleOptions.OnConnect,
		Password:           singleOptions.Password,
		DialTimeout:        singleOptions.DialTimeout,
		ReadTimeout:        singleOptions.ReadTimeout,
		WriteTimeout:       singleOptions.WriteTimeout,
		PoolSize:           singleOptions.PoolSize,
		PoolTimeout:        singleOptions.PoolTimeout,
		IdleTimeout:        singleOptions.IdleTimeout,
		IdleCheckFrequency: singleOptions.IdleCheckFrequency,
		TLSConfig:          singleOptions.TLSConfig,
	}
}
