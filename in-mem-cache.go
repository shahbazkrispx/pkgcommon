package pkgcommon

import (
	"fmt"
	"os"
	"strconv"
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

// NewMemCache Create A new In memory Client
func NewMemCache() InMemPostCache {
	expVal, _ := strconv.ParseInt(os.Getenv("IN_MEM_DEFAULT_EXP"), 10, 64)
	cleanUp, _ := strconv.ParseInt(os.Getenv("IN_MEM_CLEANUP_INTERVAL"), 10, 64)

	return &inMemCache{
		DefaultExpiration: time.Duration(int(expVal)) * time.Minute,
		CleanUpInterval:   time.Duration(int(cleanUp)) * time.Minute,
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

func (inMem *inMemCache) Delete(key string) (int64, error) {
	c := inMem.getInMemClient()
	c.Delete(key)
	return 1, nil
}
