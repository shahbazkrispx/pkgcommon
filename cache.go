package pkgcommon

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/patrickmn/go-cache"
)

// redis
type RedisCache struct {
	client *redis.Client
}

// in memory
type InMemoryCache struct {
	client *cache.Cache
}

type AppCache interface {
	Set(key string, value interface{}, exp time.Duration) error
	Get(key string) (string, error)
	Exists(key string) int64
	Delete(Key string) (int64, error)
}

var MyCache AppCache

// redis cache
func InitRedisCache() {
	dbIndex, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	host := fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT"))
	password := os.Getenv("REDIS_PASSWORD")

	//return redisCache
	MyCache = &RedisCache{
		client: redis.NewClient(&redis.Options{
			Addr:     host,
			Password: password,
			DB:       dbIndex,
		}),
	}
}

// in momory cache
func InitInMemoryCache() {
	MyCache = &InMemoryCache{
		client: cache.New(5*time.Minute, 10*time.Minute),
	}
}

//Set create a new item in redis with expiry
func (cache *RedisCache) Set(key string, value interface{}, exp time.Duration) error {
	_, err := cache.client.Ping().Result()
	if err != nil {
		return err
	}

	// serialize Post object to JSON
	_, err = cache.client.Set(key, value, exp).Result()
	if err != nil {
		return err
	}

	return nil
}

//Get a item by key from redis
func (cache *RedisCache) Get(key string) (string, error) {

	duration, err := cache.client.TTL(key).Result()

	switch {
	case cache.client.Keys(key) == nil:
		return "", errors.New("Invalid Code")
	case duration.Seconds() == -2:
		return "", errors.New("key does not exist")
	case duration.Seconds() == -1:
		return "", errors.New("The key will not expire  ")
	}

	val, err := cache.client.Get(key).Result()

	if err != nil {
		return "", err
	}

	return val, nil

}

//Delete item from by from redis
func (cache *RedisCache) Delete(key string) (int64, error) {

	d, err := cache.client.Del(key).Result()
	if err != nil {
		return 0, err
	}
	return d, err
}

//Exists check item is exist in redis
func (cache *RedisCache) Exists(key string) int64 {

	val, err := cache.client.Exists(key).Result()
	if err != nil {
		panic(err)
		return -1
	}

	return val
}

//in memory
func (cache *InMemoryCache) Set(key string, value interface{}, exp time.Duration) error {

	cache.client.Set(key, value, exp)
	return nil
}

func (cache *InMemoryCache) Get(key string) (string, error) {

	val, found := cache.client.Get(key)
	if !found {
		return "", fmt.Errorf("%v Not Found", key)
	}

	return val.(string), nil
}

// delete always return 1
func (cache *InMemoryCache) Delete(key string) (int64, error) {

	cache.client.Delete(key)
	return 1, nil
}

//Exists check item is exist in redis
func (cache *InMemoryCache) Exists(key string) int64 {

	_, expTime, found := cache.client.GetWithExpiration(key)
	isExpired := math.Signbit(float64(expTime.Sub(time.Now())))
	if found && isExpired {
		return 1
	}

	return 0
}
