package pkgcommon

import (
	"fmt"
	"time"

	"github.com/patrickmn/go-cache"
)

type InMemPostCache interface {
	Set(key string, value interface{}, exp time.Duration) error
	Get(key string) (string, error)
	Delete(Key string) (int64, error)
}

type inMemCache struct {
	DefaultExpiration time.Duration
	CleanUpInterval   time.Duration
}

// NewMemCache Create A new In memory Client with default epx and cleanup
func NewMemCache(defaultExp time.Duration, cleanUpExp time.Duration) InMemPostCache {

	return &inMemCache{
		DefaultExpiration: defaultExp,
		CleanUpInterval:   cleanUpExp,
	}
}

func (inMem *inMemCache) getInMemClient() *cache.Cache {
	return cache.New(inMem.DefaultExpiration, inMem.CleanUpInterval)
}

func (inMem *inMemCache) Set(key string, value interface{}, exp time.Duration) error {
	c := inMem.getInMemClient()
	c.Set(key, value, exp)
	return nil
}

func (inMem *inMemCache) Get(key string) (string, error) {
	c := inMem.getInMemClient()
	val, found := c.Get(key)
	if !found {
		return "", fmt.Errorf("%v Not Found", key)
	}

	return val.(string), nil
}

// delete always return 1
func (inMem *inMemCache) Delete(key string) (int64, error) {
	c := inMem.getInMemClient()
	c.Delete(key)
	return 1, nil
}
