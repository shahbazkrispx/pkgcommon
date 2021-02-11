package pkgcommon

import (
	"errors"
	"fmt"
	"github.com/go-redis/redis/v7"
	"os"
	"strconv"
	"time"
)

type PostCache interface {
	Set(key string, value string, exp time.Duration) error
	Get(key string) (string,error)
	Exists(key string) int64
	Delete(Key string) (int64,error)

}

type redisCache struct {
	host    string
	db      int
	password string
	expires time.Duration

}

func NewRedisCache() PostCache {
	dbIndex, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	exp, _ := time.ParseDuration(os.Getenv("REDIS_EXPIRY"))
	host := fmt.Sprintf("%s:%s",os.Getenv("REDIS_HOST"),os.Getenv("REDIS_PORT"))

	//return redisCache
	return &redisCache{
		host:    host,
		db:     dbIndex,
		password: os.Getenv("REDIS_PASSWORD"),
		expires:  exp,
	}
}

func (cache *redisCache) getClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     cache.host,
		Password: cache.password,
		DB:       cache.db,
	})
}

func (cache *redisCache) Set(key string, value string, exp time.Duration) error {
	client := cache.getClient()

	_, err := client.Ping().Result()
	if err != nil {
		return err
	}
	// serialize Post object to JSON
	client.Set(key, value, exp)

	return nil
}

func (cache *redisCache) Get(key string) (string,error) {
	client := cache.getClient()

	duration, err := client.TTL(key).Result()

	switch {
	case client.Keys(key) == nil:
		return "", errors.New("Invalid Code")
	case duration.Seconds() == -2:
		return "", errors.New("key does not exist")
	case duration.Seconds() == -1:
		return "", errors.New("The key will not expire.")
	}

	val, err := client.Get(key).Result()

	if err != nil {
		return "", err
	}

	return val, nil

}

func (cache *redisCache) Delete(key string) (int64,error)  {
	client := cache.getClient()
	d, err := client.Del(key).Result()
	if err != nil {
		return 0, err
	}
	return d, err
}

func (cache *redisCache) Exists(key string) int64 {
	client := cache.getClient()

	val, err := client.Exists(key).Result()
	if err != nil {
		panic(err)
		return -1
	}

	return val
}



